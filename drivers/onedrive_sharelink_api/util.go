package onedrive_sharelink_api

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	stdpath "path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func (d *OnedriveSharelinkAPI) NewNoRedirectCLient() *http.Client {
	return &http.Client{
		Timeout: time.Hour * 48,
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.Conf.TlsInsecureSkipVerify},
		},
		//no redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func (d *OnedriveSharelinkAPI) getCookiesWithPassword(link, password string) (string, error) {
	// Send GET request
	resp, err := http.Get(link)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse the HTML response
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	// Find input fields by their IDs
	var viewstate, eventvalidation, postAction string

	var findInputFields func(*html.Node)
	findInputFields = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			for _, attr := range n.Attr {
				if attr.Key == "id" {
					switch attr.Val {
					case "__VIEWSTATE":
						viewstate = d.getAttrValue(n, "value")
					case "__EVENTVALIDATION":
						eventvalidation = d.getAttrValue(n, "value")
					}
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "form" {
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == "inputForm" {
					postAction = d.getAttrValue(n, "action")
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findInputFields(c)
		}
	}
	findInputFields(doc)

	// Prepare the new URL for the POST request
	linkParts, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	newURL := fmt.Sprintf("%s://%s%s", linkParts.Scheme, linkParts.Host, postAction)

	// Prepare the request body
	data := url.Values{
		"txtPassword":          []string{password},
		"__EVENTVALIDATION":    []string{eventvalidation},
		"__VIEWSTATE":          []string{viewstate},
		"__VIEWSTATEENCRYPTED": []string{""},
	}

	client := d.NewNoRedirectCLient()
	// Send the POST request,no redirect
	resp, err = client.PostForm(newURL, data)
	if err != nil {
		return "", err
	}
	// Extract the desired cookie value
	cookie := resp.Cookies()
	var fedAuthCookie string
	for _, c := range cookie {
		if c.Name == "FedAuth" {
			fedAuthCookie = c.Value
			break
		}
	}
	if fedAuthCookie == "" {
		return "", fmt.Errorf("wrong password")
	}
	return fmt.Sprintf("FedAuth=%s;", fedAuthCookie), nil
}

func (d *OnedriveSharelinkAPI) getAttrValue(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func (d *OnedriveSharelinkAPI) getSharelinkRoot() error {
	if !d.UseSharelinkRoot {
		d.SharelinkRootPath = ""
		return nil
	}
	u, err := url.Parse(d.RedirectUrl)
	if err != nil {
		return err
	}
	id := u.Query().Get("id")
	//url decode
	id, err = url.QueryUnescape(id)
	if err != nil {
		return err
	}
	// we throw ANYTHING before 'Documents'(included, or 'Shared Documents')
	// away, and use the rest as the root id
	//sth like /a/b/c/Documents/d/e/f -> /d/e/f
	id = strings.TrimRight(id, "/")
	parts := strings.Split(id, "/")
	for i, part := range parts {
		if part == "Documents" || part == "Shared Documents" {
			id = strings.Join(parts[i+1:], "/")
			break
		}
	}
	d.SharelinkRootPath = "/" + id
	return nil
}

func (d *OnedriveSharelinkAPI) getHeaders() (http.Header, error) {
	header := http.Header{}
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 Edg/90.0.818.51")
	header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	if d.ShareLinkPassword == "" {
		//no redirect client
		clientNoDirect := d.NewNoRedirectCLient()
		// create a request
		req, err := http.NewRequest("GET", d.ShareLinkURL, nil)
		if err != nil {
			return nil, err
		}
		// set req.Header to d.Header
		req.Header = header
		// request the Sharelink
		answerNoRedirect, err := clientNoDirect.Do(req)
		if err != nil {
			return nil, err
		}
		// get the location
		redirectUrl := answerNoRedirect.Header.Get("Location")
		log.Debugln("redirectUrl:", redirectUrl)
		if redirectUrl == "" {
			return nil, fmt.Errorf("password protected link. Please provide password")
		}
		header.Set("Cookie", answerNoRedirect.Header.Get("Set-Cookie"))
		// set Referer to the redirectUrl
		header.Set("Referer", redirectUrl)
		// set authority to the netloc of the redirectUrl
		//get the netloc
		u, err := url.Parse(redirectUrl)
		if err != nil {
			return nil, err
		}
		header.Set("authority", u.Host)
		// return the header
		return header, nil
	} else {
		cookie, err := d.getCookiesWithPassword(d.ShareLinkURL, d.ShareLinkPassword)
		if err != nil {
			return nil, err
		}
		header.Set("Cookie", cookie)
		// set the referer
		header.Set("Referer", d.ShareLinkURL)
		// set the authority
		header.Set("authority", strings.Split(strings.Split(d.ShareLinkURL, "//")[1], "/")[0])
		// return the header
		return header, nil
	}

}

func (d *OnedriveSharelinkAPI) GetRedirectUrl() (err error) {
	//no redirect client
	clientNoDirect := d.NewNoRedirectCLient()
	// create a request
	req, err := http.NewRequest("GET", d.ShareLinkURL, nil)
	if err != nil {
		return err
	}
	header := req.Header
	d.RedirectUrl = ""
	if d.ShareLinkPassword == "" {
		// set headers
		header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 Edg/90.0.818.51")
		header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
		// set req.Header to Header
		req.Header = header
		// request the Sharelink
		answerNoRedirect, err := clientNoDirect.Do(req)
		if err != nil {
			return err
		}
		// get the location
		d.RedirectUrl = answerNoRedirect.Header.Get("Location")
	} else {
		header = d.Headers
		req.Header = header
		answerNoRedirect, err := clientNoDirect.Do(req)
		if err != nil {
			d.Headers, err = d.getHeaders()
			if err != nil {
				return err
			}
			return d.GetRedirectUrl()
		}
		// get the location
		d.RedirectUrl = answerNoRedirect.Header.Get("Location")
	}
	if d.RedirectUrl == "" {
		return fmt.Errorf("password protected link. Please provide password")
	}
	return nil
}
func (d *OnedriveSharelinkAPI) GetBaseUrl() error {
	isSharepoint := false
	if !strings.Contains(d.RedirectUrl, "-my") {
		isSharepoint = true
	}
	//get the netloc of the ShareLinkURL
	clientNoDirect := d.NewNoRedirectCLient()
	//request the redirectUrl with d.Headers
	req, err := http.NewRequest("GET", d.RedirectUrl, nil)
	if err != nil {
		return err
	}
	req.Header = d.Headers
	answer, err := clientNoDirect.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(answer.Body)
	if err != nil {
		return err
	}
	driveUrl := ""
	if !isSharepoint {
		re := regexp.MustCompile(`".driveUrl":"(.*?)"`)
		// get the first match
		driveUrl = re.FindString(string(body))
		//replace \u002f to /
		driveUrl = strings.Replace(driveUrl, `\u002f`, "/", -1)
		//delete "\".driveUrl\":\" and \""
		driveUrl = strings.Replace(driveUrl, `".driveUrl":"`, ``, -1)
		driveUrl = strings.Replace(driveUrl, `"`, "", -1)
	} else {
		re := regexp.MustCompile(`".driveUrl" : "(.*?)"`)
		// get the first match
		driveUrl = re.FindString(string(body))
		//replace \u002f to /
		driveUrl = strings.Replace(driveUrl, `\u002f`, "/", -1)
		//delete "\".driveUrl\":\" and \""
		driveUrl = strings.Replace(driveUrl, `".driveUrl" : "`, ``, -1)
		driveUrl = strings.Replace(driveUrl, `"`, "", -1)
	}
	d.BaseUrl = driveUrl
	return nil
}
func (d *OnedriveSharelinkAPI) GetMetaUrl(auth bool, path string) string {
	//add d.SharelinkRootPath to path
	path = d.SharelinkRootPath + path
	path = utils.EncodePath(path, true)
	if path == "/" || path == "\\" {
		return fmt.Sprintf("%s/root", d.BaseUrl)
	} else {
		//delete the last "/" or "\"
		path = strings.TrimSuffix(path, "/")
		path = strings.TrimSuffix(path, "\\")
		return fmt.Sprintf("%s/root:%s:", d.BaseUrl, path)
	}
}

func (d *OnedriveSharelinkAPI) refreshToken() error {
	var err error
	for i := 0; i < 3; i++ {
		d.Headers, err = d.getHeaders()
		if err == nil {
			break
		}
	}
	return err
}
func (d *OnedriveSharelinkAPI) Request(url string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()
	req.Header = d.Headers
	//if method is not GET, set Authorization to Bearer
	if method != "GET" {
		req.Header.Set("Authorization", "Bearer")
	}
	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e RespErr
	req.SetError(&e)
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	if e.Error.Code != "" {
		if e.Error.Code == "InvalidAuthenticationToken" || e.Error.Code == "unauthenticated" {
			err = d.refreshToken()
			if err != nil {
				return nil, err
			}
			return d.Request(url, method, callback, resp)
		}
		return nil, errors.New(e.Error.Message)
	}
	return res.Body(), nil
}

func (d *OnedriveSharelinkAPI) getFiles(path string) ([]File, error) {
	var res []File
	nextLink := d.GetMetaUrl(false, path) + "/children?$top=5000&$expand=thumbnails($select=medium)&$select=id,name,size,lastModifiedDateTime,content.downloadUrl,file,parentReference"
	for nextLink != "" {
		var files Files
		_, err := d.Request(nextLink, http.MethodGet, nil, &files)
		if err != nil {
			return nil, err
		}
		res = append(res, files.Value...)
		nextLink = files.NextLink
	}
	return res, nil
}

func (d *OnedriveSharelinkAPI) GetFile(path string) (*File, error) {
	var file File
	u := d.GetMetaUrl(false, path)
	_, err := d.Request(u, http.MethodGet, nil, &file)
	return &file, err
}

func (d *OnedriveSharelinkAPI) upSmall(ctx context.Context, dstDir model.Obj, stream model.FileStreamer) error {
	url := d.GetMetaUrl(false, stdpath.Join(dstDir.GetPath(), stream.GetName())) + "/content"
	data, err := io.ReadAll(stream)
	if err != nil {
		return err
	}
	_, err = d.Request(url, http.MethodPut, func(req *resty.Request) {
		req.SetBody(data).SetContext(ctx)
	}, nil)
	return err
}

func (d *OnedriveSharelinkAPI) upBig(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	url := d.GetMetaUrl(false, stdpath.Join(dstDir.GetPath(), stream.GetName())) + "/createUploadSession"
	res, err := d.Request(url, http.MethodPost, nil, nil)
	if err != nil {
		return err
	}
	uploadUrl := jsoniter.Get(res, "uploadUrl").ToString()
	var finish int64 = 0
	DEFAULT := d.ChunkSize * 1024 * 1024
	for finish < stream.GetSize() {
		if utils.IsCanceled(ctx) {
			return ctx.Err()
		}
		log.Debugf("upload: %d", finish)
		var byteSize int64 = DEFAULT
		left := stream.GetSize() - finish
		if left < DEFAULT {
			byteSize = left
		}
		byteData := make([]byte, byteSize)
		n, err := io.ReadFull(stream, byteData)
		log.Debug(err, n)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("PUT", uploadUrl, bytes.NewBuffer(byteData))
		if err != nil {
			return err
		}
		req = req.WithContext(ctx)
		req.Header.Set("Content-Length", strconv.Itoa(int(byteSize)))
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", finish, finish+byteSize-1, stream.GetSize()))
		finish += byteSize
		res, err := base.HttpClient.Do(req)
		if err != nil {
			return err
		}
		if res.StatusCode != 201 && res.StatusCode != 202 {
			data, _ := io.ReadAll(res.Body)
			res.Body.Close()
			return errors.New(string(data))
		}
		res.Body.Close()
		up(int(finish * 100 / stream.GetSize()))
	}
	return nil
}

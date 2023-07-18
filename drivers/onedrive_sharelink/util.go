package onedrive_sharelink

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/conf"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func NewNoRedirectCLient() *http.Client {
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

func getCookiesWithPassword(link, password string) (string, error) {
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
						viewstate = getAttrValue(n, "value")
					case "__EVENTVALIDATION":
						eventvalidation = getAttrValue(n, "value")
					}
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "form" {
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == "inputForm" {
					postAction = getAttrValue(n, "action")
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

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
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

func getAttrValue(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func (d *OnedriveSharelink) getHeaders() (http.Header, error) {
	header := http.Header{}
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 Edg/90.0.818.51")
	header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	//save timestamp to d.HeaderTime
	d.HeaderTime = time.Now().Unix()
	if d.ShareLinkPassword == "" {
		//no redirect client
		clientNoDirect := NewNoRedirectCLient()
		// create a request
		req, err := http.NewRequest("GET", d.ShareLinkURL, nil)
		if err != nil {
			return nil, err
		}
		// set headers
		header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 Edg/90.0.818.51")
		header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
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
		cookie, err := getCookiesWithPassword(d.ShareLinkURL, d.ShareLinkPassword)
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
func (d *OnedriveSharelink) getFiles(path string) ([]Item, error) {
	//no redirect client
	clientNoDirect := NewNoRedirectCLient()
	// create a request
	req, err := http.NewRequest("GET", d.ShareLinkURL, nil)

	if err != nil {
		return nil, err
	}
	header := req.Header
	redirectUrl := ""
	if d.ShareLinkPassword == "" {
		// set headers
		header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 Edg/90.0.818.51")
		header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
		// set req.Header to Header
		req.Header = header
		// request the Sharelink
		answerNoRedirect, err := clientNoDirect.Do(req)
		if err != nil {
			return nil, err
		}
		// get the location
		redirectUrl = answerNoRedirect.Header.Get("Location")
	} else {
		header = d.Headers
		req.Header = header
		answerNoRedirect, err := clientNoDirect.Do(req)
		if err != nil {
			return nil, err
		}
		// get the location
		redirectUrl = answerNoRedirect.Header.Get("Location")
	}
	redirectSplitURL := strings.Split(redirectUrl, "/")
	// init Headers
	req.Header = d.Headers
	downloadLinkPrefix := ""
	rootFolderPre := ""
	if d.IsSharepoint {
		// Get redirectUrl
		answer, err := clientNoDirect.Get(redirectUrl)
		if err != nil {
			d.Headers, err = d.getHeaders()
			if err != nil {
				return nil, err
			}
			return d.getFiles(path)
		}
		// Use regex 'templateUrl":"(.*?)"' to search the templateUrl
		re := regexp.MustCompile(`templateUrl":"(.*?)"`)
		// get the body of the answer
		body, err := io.ReadAll(answer.Body)
		if err != nil {
			return nil, err
		}
		// search in the answer
		template := re.FindString(string(body))
		// Throw the content before "templateUrl":"(included) away
		template = template[strings.Index(template, "templateUrl\":\"")+len("templateUrl\":\""):]
		// Throw the content after "?id="(included) away
		template = template[:strings.Index(template, "?id=")]
		// Throw the content after last '/'(included) away
		template = template[:strings.LastIndex(template, "/")]
		// add /download.aspx?UniqueId= to the templateUrl
		downloadLinkPrefix = template + "/download.aspx?UniqueId="
		// get params of redirectUrl
		params, err := url.ParseQuery(redirectUrl[strings.Index(redirectUrl, "?")+1:])
		if err != nil {
			return nil, err
		}
		// get id
		rootFolderPre = params.Get("id")
	} else {
		// Throw the content after last '/'(included) away of the redirectUrl
		redirectUrlCut := redirectUrl[:strings.LastIndex(redirectUrl, "/")]
		// add /download.aspx?UniqueId= to the redirectUrl
		downloadLinkPrefix = redirectUrlCut + "/download.aspx?UniqueId="
		// get params of redirectUrl
		params, err := url.ParseQuery(redirectUrl[strings.Index(redirectUrl, "?")+1:])
		if err != nil {
			return nil, err
		}
		// get id
		rootFolderPre = params.Get("id")
	}
	d.downloadLinkPrefix = downloadLinkPrefix
	// url decode the rootFolderPre
	rootFolder, err := url.QueryUnescape(rootFolderPre)
	if err != nil {
		return nil, err
	}
	log.Debugln("rootFolder:", rootFolder)
	// Then we store ANYTHING before 'Documents'(included, or 'Shared Documents') into relativePath. In this case, it's /personal/admin_sakurapy_onmicrosoft_com/Documents'
	// We url encode relativePath and replace _ with %5F and - with %2D store it as relativeUrl. The relativePath/redirectUrl will not be changed.
	// If you need to access subfolder, you need to add the subfolder name after rootFolder, like /personal/admin_sakurapy_onmicrosoft_com/Documents/DMYZ/subfolder
	// We url encode rootFolder and replace _ with %5F and - with %2D store it as rootFolderUrl.
	relativePath := strings.Split(rootFolder, "Documents")[0] + "Documents"
	relativeUrl := url.QueryEscape(relativePath)
	//replace
	relativeUrl = strings.Replace(relativeUrl, "_", "%5F", -1)
	relativeUrl = strings.Replace(relativeUrl, "-", "%2D", -1)
	if path != "/" {
		//add path to rootFolder
		rootFolder = rootFolder + path
	}
	rootFolderUrl := url.QueryEscape(rootFolder)
	//replace
	rootFolderUrl = strings.Replace(rootFolderUrl, "_", "%5F", -1)
	rootFolderUrl = strings.Replace(rootFolderUrl, "-", "%2D", -1)
	log.Debugln("relativePath:", relativePath, "relativeUrl:", relativeUrl, "rootFolder:", rootFolder, "rootFolderUrl:", rootFolderUrl)
	graphqlVar := fmt.Sprintf(`{"query":"query (\n        $listServerRelativeUrl: String!,$renderListDataAsStreamParameters: RenderListDataAsStreamParameters!,$renderListDataAsStreamQueryString: String!\n        )\n      {\n      \n      legacy {\n      \n      renderListDataAsStream(\n      listServerRelativeUrl: $listServerRelativeUrl,\n      parameters: $renderListDataAsStreamParameters,\n      queryString: $renderListDataAsStreamQueryString\n      )\n    }\n      \n      \n  perf {\n    executionTime\n    overheadTime\n    parsingTime\n    queryCount\n    validationTime\n    resolvers {\n      name\n      queryCount\n      resolveTime\n      waitTime\n    }\n  }\n    }","variables":{"listServerRelativeUrl":"%s","renderListDataAsStreamParameters":{"renderOptions":5707527,"allowMultipleValueFilterForTaxonomyFields":true,"addRequiredFields":true,"folderServerRelativeUrl":"%s"},"renderListDataAsStreamQueryString":"@a1=\'%s\'&RootFolder=%s&TryNewExperienceSingle=TRUE"}}`, relativePath, rootFolder, relativeUrl, rootFolderUrl)
	tempHeader := make(http.Header)
	for k, v := range d.Headers {
		tempHeader[k] = v
	}
	tempHeader["Content-Type"] = []string{"application/json;odata=verbose"}

	client := &http.Client{}
	// python: graphqlReq = req.post("/".join(redirectSplitURL[:-3])+"/_api/v2.1/graphql", data=graphqlVar.encode('utf-8'), headers=tempHeader)
	postUrl := strings.Join(redirectSplitURL[:len(redirectSplitURL)-3], "/") + "/_api/v2.1/graphql"
	req, err = http.NewRequest("POST", postUrl, strings.NewReader(graphqlVar))
	if err != nil {
		return nil, err
	}
	req.Header = tempHeader

	resp, err := client.Do(req)
	if err != nil {
		d.Headers, err = d.getHeaders()
		if err != nil {
			return nil, err
		}
		return d.getFiles(path)
	}
	defer resp.Body.Close()
	var graphqlReq GraphQLRequest
	json.NewDecoder(resp.Body).Decode(&graphqlReq)
	log.Debugln("graphqlReq:", graphqlReq)
	filesData := graphqlReq.Data.Legacy.RenderListDataAsStream.ListData.Row
	//if "NextHref" in graphqlReq["data"]["legacy"]["renderListDataAsStream"]["ListData"]:
	if graphqlReq.Data.Legacy.RenderListDataAsStream.ListData.NextHref != "" {
		nextHref := graphqlReq.Data.Legacy.RenderListDataAsStream.ListData.NextHref + "&@a1=REPLACEME&TryNewExperienceSingle=TRUE"
		nextHref = strings.Replace(nextHref, "REPLACEME", "%27"+relativeUrl+"%27", -1)
		log.Debugln("nextHref:", nextHref)
		filesData = append(filesData, graphqlReq.Data.Legacy.RenderListDataAsStream.ListData.Row...)

		listViewXml := graphqlReq.Data.Legacy.RenderListDataAsStream.ViewMetadata.ListViewXml
		log.Debugln("listViewXml:", listViewXml)
		renderListDataAsStreamVar := `{"parameters":{"__metadata":{"type":"SP.RenderListDataParameters"},"RenderOptions":1216519,"ViewXml":"REPLACEME","AllowMultipleValueFilterForTaxonomyFields":true,"AddRequiredFields":true}}`
		listViewXml = strings.Replace(listViewXml, `"`, `\"`, -1)
		renderListDataAsStreamVar = strings.Replace(renderListDataAsStreamVar, "REPLACEME", listViewXml, -1)

		graphqlReqNEW := GraphQLNEWRequest{}
		//python: graphqlReq = req.post("/".join(redirectSplitURL[:-3])+"/_api/web/GetListUsingPath(DecodedUrl=@a1)/RenderListDataAsStream"+nextHref, data=renderListDataAsStreamVar.encode('utf-8'), headers=tempHeader)
		postUrl = strings.Join(redirectSplitURL[:len(redirectSplitURL)-3], "/") + "/_api/web/GetListUsingPath(DecodedUrl=@a1)/RenderListDataAsStream" + nextHref
		req, _ = http.NewRequest("POST", postUrl, strings.NewReader(renderListDataAsStreamVar))
		req.Header = tempHeader

		resp, err := client.Do(req)
		if err != nil {
			d.Headers, err = d.getHeaders()
			if err != nil {
				return nil, err
			}
			return d.getFiles(path)
		}
		defer resp.Body.Close()
		//clear graphqlReq
		json.NewDecoder(resp.Body).Decode(&graphqlReqNEW)
		for graphqlReqNEW.ListData.NextHref != "" {
			//clear graphqlReqNEW
			graphqlReqNEW = GraphQLNEWRequest{}
			//python: graphqlReq = req.post("/".join(redirectSplitURL[:-3])+"/_api/web/GetListUsingPath(DecodedUrl=@a1)/RenderListDataAsStream"+nextHref, data=renderListDataAsStreamVar.encode('utf-8'), headers=tempHeader)
			postUrl = strings.Join(redirectSplitURL[:len(redirectSplitURL)-3], "/") + "/_api/web/GetListUsingPath(DecodedUrl=@a1)/RenderListDataAsStream" + nextHref
			req, _ = http.NewRequest("POST", postUrl, strings.NewReader(renderListDataAsStreamVar))
			req.Header = tempHeader
			resp, err := client.Do(req)
			if err != nil {
				d.Headers, err = d.getHeaders()
				if err != nil {
					return nil, err
				}
				return d.getFiles(path)
			}
			defer resp.Body.Close()
			json.NewDecoder(resp.Body).Decode(&graphqlReqNEW)
			nextHref = graphqlReqNEW.ListData.NextHref + "&@a1=REPLACEME&TryNewExperienceSingle=TRUE"
			nextHref = strings.Replace(nextHref, "REPLACEME", "%27"+relativeUrl+"%27", -1)
			filesData = append(filesData, graphqlReqNEW.ListData.Row...)
		}
		filesData = append(filesData, graphqlReqNEW.ListData.Row...)
	} else {
		filesData = append(filesData, graphqlReq.Data.Legacy.RenderListDataAsStream.ListData.Row...)
	}
	return filesData, nil
}

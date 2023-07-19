package cloudreve_sharelink

import (
	"errors"
	"net/http"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/pkg/cookie"
	json "github.com/json-iterator/go"
)

// do others that not defined in Driver interface

func (d *CloudreveSharelink) request(method string, path string, callback base.ReqCallback, out interface{}) error {
	u := d.Address + "/api/v3" + path
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"Cookie":     "cloudreve-session=" + d.Cookie,
		"Accept":     "application/json, text/plain, */*",
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
	})

	var r Resp

	req.SetResult(&r)

	if callback != nil {
		callback(req)
	}

	resp, err := req.Execute(method, u)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return errors.New(resp.String())
	}

	if r.Code != 0 {
		return errors.New(r.Msg)
	}
	sess := cookie.GetCookie(resp.Cookies(), "cloudreve-session")
	if sess != nil {
		d.Cookie = sess.Value
	}
	if out != nil && r.Data != nil {
		var marshal []byte
		marshal, err = json.Marshal(r.Data)
		if err != nil {
			return err
		}
		err = json.Unmarshal(marshal, out)
		if err != nil {
			return err
		}
	}

	return nil
}
func (d *CloudreveSharelink) checkIfProtected(password string) (bool, error) {
	var Locked ShareLinkInfo
	var err error
	if password == "" {
		err = d.request(http.MethodGet, "/share/info/"+d.SharelinkKey, nil, &Locked)
	} else {
		err = d.request(http.MethodGet, "/share/info/"+d.SharelinkKey+"?password="+password, nil, &Locked)
	}
	if err != nil {
		return false, err
	}
	return Locked.Locked, nil
}

func (d *CloudreveSharelink) login() error {
	var siteConfig Config
	err := d.request(http.MethodGet, "/site/config", nil, &siteConfig)
	if err != nil {
		return err
	}
	return err
}

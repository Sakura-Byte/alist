package cloudreve_sharelink

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/cookie"
	"github.com/go-resty/resty/v2"
	json "github.com/json-iterator/go"
	jsoniter "github.com/json-iterator/go"
)

const loginPath = "/user/session"

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
	//if Username and Userpass is not empty, use it to login
	if d.Username != "" && d.Userpass != "" {
		for i := 0; i < 5; i++ {
			err = d.doLogin(siteConfig.LoginCaptcha)
			if err == nil {
				break
			}
			if err != nil && err.Error() != "CAPTCHA not match." {
				break
			}
		}
	}
	return err
}

func (d *CloudreveSharelink) doLogin(needCaptcha bool) error {
	var captchaCode string
	var err error
	if needCaptcha {
		var captcha string
		err = d.request(http.MethodGet, "/site/captcha", nil, &captcha)
		if err != nil {
			return err
		}
		if len(captcha) == 0 {
			return errors.New("can not get captcha")
		}
		i := strings.Index(captcha, ",")
		dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(captcha[i+1:]))
		vRes, err := base.RestyClient.R().SetMultipartField(
			"image", "validateCode.png", "image/png", dec).
			Post(setting.GetStr(conf.OcrApi))
		if err != nil {
			return err
		}
		if jsoniter.Get(vRes.Body(), "status").ToInt() != 200 {
			return errors.New("ocr error:" + jsoniter.Get(vRes.Body(), "msg").ToString())
		}
		captchaCode = jsoniter.Get(vRes.Body(), "result").ToString()
	}
	var resp Resp
	err = d.request(http.MethodPost, loginPath, func(req *resty.Request) {
		req.SetBody(base.Json{
			"username":    d.Addition.Username,
			"Password":    d.Addition.Userpass,
			"captchaCode": captchaCode,
		})
	}, &resp)
	return err
}

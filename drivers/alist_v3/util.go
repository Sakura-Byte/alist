package alist_v3

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

func (d *AListV3) login() error {
	if d.Username == "" {
		return nil
	}
	var resp common.Resp[LoginResp]
	_, err := d.request("/auth/login", http.MethodPost, func(req *resty.Request) {
		req.SetResult(&resp).SetBody(base.Json{
			"username": d.Username,
			"password": d.Password,
		})
	})
	if err != nil {
		return err
	}
	d.Token = resp.Data.Token
	op.MustSaveDriverStorage(d)
	return nil
}

func (d *AListV3) request(api, method string, callback base.ReqCallback, retry ...bool) ([]byte, error) {
	// Wait for permission from the rate limiter.
	if d.limiter != nil {
		if err := d.limiter.Wait(context.Background()); err != nil {
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}
	}

	url := d.Address + "/api" + api
	req := base.RestyClient.R()
	req.SetHeader("Authorization", d.Token)
	if callback != nil {
		callback(req)
	}
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	log.Debugf("[alist_v3] response body: %s", res.String())
	if res.StatusCode() >= 400 {
		return nil, fmt.Errorf("request failed, status: %s", res.Status())
	}
	code := utils.Json.Get(res.Body(), "code").ToInt()
	if code != 200 {
		if code == 500 {
			for i := 0; i < 10; i++ {
				time.Sleep(200 * time.Millisecond)
				res, err = req.Execute(method, url)
				code = utils.Json.Get(res.Body(), "code").ToInt()
				if code == 200 {
					return res.Body(), nil
				}
			}
			return nil, fmt.Errorf("request failed, code: %d, message: %s", code, utils.Json.Get(res.Body(), "message").ToString())
		} else if (code == 401 || code == 403) && !utils.IsBool(retry...) {
			err = d.login()
			if err != nil {
				return nil, err
			}
			return d.request(api, method, callback, true)
		}
		return nil, fmt.Errorf("request failed, code: %d, message: %s", code, utils.Json.Get(res.Body(), "message").ToString())
	}

	return res.Body(), nil
}

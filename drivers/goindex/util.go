package goindex

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	jsoniter "github.com/json-iterator/go"
)

// Any additional utility functions can be added here.

func (d *GoIndex) getFiles(path string) ([]File, error) {
	path_encoded := url.PathEscape(path)
	URL_list := d.URL + "/0:" + path_encoded
	//replace all %2F to /
	URL_list = strings.ReplaceAll(URL_list, "%2F", "/")
	page_token := "first"
	page_index := 0
	var post_json []byte
	var err error
	res := make([]File, 0)
	for page_token != "" {
		if page_token == "first" {
			page_token = ""
		}
		var resp FolderResp
		post_content := map[string]string{
			"page_index": fmt.Sprintf("%d", page_index),
			"page_token": page_token,
			"password":   "",
			"q":          "",
		}
		post_json, err = jsoniter.Marshal(post_content)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest("POST", URL_list, bytes.NewReader(post_json))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("Connection", "close")
		answer, err := base.HttpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer answer.Body.Close()
		//decode json
		decoder := jsoniter.NewDecoder(answer.Body)
		err = decoder.Decode(&resp)
		if err != nil {
			return nil, err
		}
		page_index++
		page_token = resp.NextPageToken
		res = append(res, resp.Data.Files...)
	}
	return res, nil
}

package onedrive_vercel

import (
	"context"
	"net/url"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
)

type OnedriveVercel struct {
	model.Storage
	Addition
}

func (d *OnedriveVercel) Config() driver.Config {
	return config
}

func (d *OnedriveVercel) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *OnedriveVercel) Init(ctx context.Context) error {
	if len(d.Addition.Address) > 0 && string(d.Addition.Address[len(d.Addition.Address)-1]) == "/" {
		d.Addition.Address = d.Addition.Address[0 : len(d.Addition.Address)-1]
	}
	return nil
}

func (d *OnedriveVercel) Drop(ctx context.Context) error {
	return nil
}

func (d *OnedriveVercel) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	url := d.Address + "/api/?path=" + url.QueryEscape(dir.GetPath())
	next := "first"
	urlnext := ""
	var files []model.Obj
	for next != "" {
		var resp FolderResp
		req := base.RestyClient.R().
			SetResult(&resp)
		if d.Host != "" {
			req.SetHeader("Host", d.Host)
		}
		if next == "first" {
			urlnext = url
		} else {
			urlnext = url + "&next=" + next
		}
		_, err := req.Get(urlnext)
		if err != nil {
			return nil, err
		}
		for _, f := range resp.Folder.Value {
			file := model.ObjThumb{
				Object: model.Object{
					Name:     f.Name,
					Modified: *f.LastModifiedDateTime,
					Size:     f.Size,
					IsFolder: f.Folder != nil,
				},
				Thumbnail: model.Thumbnail{},
			}
			files = append(files, &file)
		}
		next = resp.Next
	}

	return files, nil
}

func (d *OnedriveVercel) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	url := d.Address + "/api/raw/?path=" + url.QueryEscape(file.GetPath())
	req := base.NoRedirectClient.R()
	if d.Host != "" {
		req.SetHeader("Host", d.Host)
	}
	resp, err := req.Get(url)
	if err != nil {
		return nil, err
	}
	return &model.Link{
		URL: resp.Header().Get("Location"),
	}, nil
}

func (d *OnedriveVercel) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return errs.NotImplement
}

func (d *OnedriveVercel) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotImplement
}

func (d *OnedriveVercel) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return errs.NotImplement
}

func (d *OnedriveVercel) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotImplement
}

func (d *OnedriveVercel) Remove(ctx context.Context, obj model.Obj) error {
	return errs.NotImplement
}

func (d *OnedriveVercel) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	return errs.NotImplement
}

var _ driver.Driver = (*OnedriveVercel)(nil)

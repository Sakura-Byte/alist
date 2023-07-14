package goindex

import (
	"context"
	"net/url"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
)

type GoIndex struct {
	model.Storage
	Addition
}

func (d *GoIndex) Config() driver.Config {
	return config
}

func (d *GoIndex) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *GoIndex) Init(ctx context.Context) error {
	if len(d.Addition.URL) > 0 && string(d.Addition.URL[len(d.Addition.URL)-1]) == "/" {
		d.Addition.URL = d.Addition.URL[0 : len(d.Addition.URL)-1]
	}
	return nil
}

func (d *GoIndex) Drop(ctx context.Context) error {
	return nil
}

func (d *GoIndex) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	path := dir.GetPath()
	if path != "/" {
		path = path + "/"
	}
	files, err := d.getFiles(path)
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return fileToObj(src), nil
	})
}

func (d *GoIndex) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	path := url.PathEscape(file.GetPath())
	linkURL := d.URL + "/0:" + path
	return &model.Link{
		URL: linkURL,
	}, nil
}

func (d *GoIndex) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return errs.NotImplement
}

func (d *GoIndex) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotImplement
}

func (d *GoIndex) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return errs.NotImplement
}

func (d *GoIndex) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotImplement
}

func (d *GoIndex) Remove(ctx context.Context, obj model.Obj) error {
	return errs.NotImplement
}

func (d *GoIndex) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	return errs.NotImplement
}

var _ driver.Driver = &GoIndex{}

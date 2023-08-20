package onedrive_sharelink_api

import (
	"context"
	"net/http"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

type OnedriveSharelinkAPI struct {
	model.Storage
	Addition
	AccessToken string
}

func (d *OnedriveSharelinkAPI) Config() driver.Config {
	return config
}

func (d *OnedriveSharelinkAPI) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *OnedriveSharelinkAPI) Init(ctx context.Context) error {
	var err error
	if d.ChunkSize < 1 {
		d.ChunkSize = 5
	}
	d.Headers, err = d.getHeaders()
	if err != nil {
		return err
	}
	err = d.GetRedirectUrl()
	if err != nil {
		return err
	}
	err = d.getSharelinkRoot()
	if err != nil {
		return err
	}
	err = d.GetBaseUrl()
	if err != nil {
		return err
	}
	return d.refreshToken()
}

func (d *OnedriveSharelinkAPI) Drop(ctx context.Context) error {
	return nil
}

func (d *OnedriveSharelinkAPI) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(dir.GetPath())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return fileToObj(src, dir.GetID()), nil
	})
}

func (d *OnedriveSharelinkAPI) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	f, err := d.GetFile(file.GetPath())
	if err != nil {
		return nil, err
	}
	if f.File == nil {
		return nil, errs.NotFile
	}
	return &model.Link{
		URL: f.Url,
	}, nil
}

func (d *OnedriveSharelinkAPI) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	url := d.GetMetaUrl(false, parentDir.GetPath()) + "/children"
	data := base.Json{
		"name":                              dirName,
		"folder":                            base.Json{},
		"@microsoft.graph.conflictBehavior": "rename",
	}
	_, err := d.Request(url, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *OnedriveSharelinkAPI) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotSupport
}

func (d *OnedriveSharelinkAPI) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return errs.NotSupport
}

func (d *OnedriveSharelinkAPI) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	dst, err := d.GetFile(dstDir.GetPath())
	if err != nil {
		return err
	}
	data := base.Json{
		"parentReference": base.Json{
			"driveId": dst.ParentReference.DriveId,
			"id":      dst.Id,
		},
		"name": srcObj.GetName(),
	}
	url := d.GetMetaUrl(false, srcObj.GetPath()) + "/copy"
	_, err = d.Request(url, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *OnedriveSharelinkAPI) Remove(ctx context.Context, obj model.Obj) error {
	return errs.NotSupport
}

func (d *OnedriveSharelinkAPI) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	var err error
	if stream.GetSize() <= 4*1024*1024 {
		err = d.upSmall(ctx, dstDir, stream)
	} else {
		err = d.upBig(ctx, dstDir, stream, up)
	}
	return err
}

var _ driver.Driver = (*OnedriveSharelinkAPI)(nil)

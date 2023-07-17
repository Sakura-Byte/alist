package onedrive_sharelink

import (
	"context"
	"strings"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type OnedriveSharelink struct {
	model.Storage
	Addition
}

func (d *OnedriveSharelink) Config() driver.Config {
	return config
}

func (d *OnedriveSharelink) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *OnedriveSharelink) Init(ctx context.Context) error {
	//init err
	var err error
	// if there is "-my" in the url, it is NOT a sharepoint link
	d.IsSharepoint = !strings.Contains(d.ShareLinkURL, "-my")
	d.Headers, err = d.getHeaders()
	if err != nil {
		return err
	}

	return nil
}

func (d *OnedriveSharelink) Drop(ctx context.Context) error {
	return nil
}

func (d *OnedriveSharelink) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	path := dir.GetPath()
	files, err := d.getFiles(path)
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src Item) (model.Obj, error) {
		return fileToObj(src), nil
	})
}

func (d *OnedriveSharelink) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	//get Id
	uniqueId := file.GetID()
	// cut the first char and the last char
	uniqueId = uniqueId[1 : len(uniqueId)-1]
	url := d.downloadLinkPrefix + uniqueId
	header := d.Headers
	log.Debugln("link url:", url)
	log.Debugf("link header: %+v", header)
	return &model.Link{
		URL:    url,
		Header: header,
	}, nil
}

func (d *OnedriveSharelink) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	// TODO create folder, optional
	return errs.NotImplement
}

func (d *OnedriveSharelink) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO move obj, optional
	return errs.NotImplement
}

func (d *OnedriveSharelink) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	// TODO rename obj, optional
	return errs.NotImplement
}

func (d *OnedriveSharelink) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj, optional
	return errs.NotImplement
}

func (d *OnedriveSharelink) Remove(ctx context.Context, obj model.Obj) error {
	// TODO remove obj, optional
	return errs.NotImplement
}

func (d *OnedriveSharelink) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file, optional
	return errs.NotImplement
}

//func (d *OnedriveSharelink) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*OnedriveSharelink)(nil)

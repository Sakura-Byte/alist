package cloudreve_sharelink

import (
	"context"
	"net/http"

	"net/url"
	"strings"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
)

type CloudreveSharelink struct {
	model.Storage
	Addition
}

func (d *CloudreveSharelink) Config() driver.Config {
	return config
}

func (d *CloudreveSharelink) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *CloudreveSharelink) Init(ctx context.Context) error {
	//clear cookie
	d.Cookie = ""
	d.Sharelink = strings.TrimSuffix(d.Sharelink, "/")
	//if d.Address contains "/s/"
	if strings.Contains(d.Sharelink, "/s/") {
		//ignore former value, overwrite it
		d.SharelinkKey = strings.Split(d.Sharelink, "/s/")[1]
		d.Address = strings.Split(d.Sharelink, "/s/")[0]
	} else {
		return errs.EmptyToken
	}
	// removing trailing slash
	d.Address = strings.TrimSuffix(d.Address, "/")
	var tempIsLocked bool
	tempIsLocked, err := d.checkIfProtected("")
	if err != nil {
		return err
	}
	//detect if link is locked
	if !tempIsLocked {
		if d.Password != "" {
			//clear password
			d.Password = ""
		}
	}
	d.Cookie = ""
	//detect if password is correct
	tempIsLocked, err = d.checkIfProtected(d.Password)
	if err != nil {
		return err
	}
	//still locked
	if tempIsLocked {
		//no password
		if d.Password == "" {
			return errs.EmptyPassword
		} else {
			//wrong password
			return errs.WrongPassword
		}
	}
	return d.login()
}

func (d *CloudreveSharelink) Drop(ctx context.Context) error {
	d.Cookie = ""
	return nil
}

func (d *CloudreveSharelink) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	var r DirectoryResp
	path_encoded := url.PathEscape(dir.GetPath())
	//replace "+" with "%2B"
	path_encoded = strings.Replace(path_encoded, "+", "%2B", -1)
	err := d.request(http.MethodGet, "/share/list/"+d.SharelinkKey+path_encoded, nil, &r)
	if err != nil {
		return nil, err
	}

	return utils.SliceConvert(r.Objects, func(src Object) (model.Obj, error) {
		return objectToObj(src), nil
	})
}

func (d *CloudreveSharelink) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var dUrl string
	//encode every character in path
	path_encoded := url.PathEscape(file.GetPath())
	//replace "+" with "%2B"
	path_encoded = strings.Replace(path_encoded, "+", "%2B", -1)
	err := d.request(http.MethodPut, "/share/download/"+d.SharelinkKey+"?path="+path_encoded, nil, &dUrl)
	//if dUrl is a relative path, add d.Address
	if !strings.HasPrefix(dUrl, "http") {
		dUrl = d.Address + dUrl
	}
	if err != nil {
		return nil, err
	}
	return &model.Link{
		URL: dUrl,
	}, nil
}

func (d *CloudreveSharelink) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return errs.NotImplement
}

func (d *CloudreveSharelink) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotImplement
}

func (d *CloudreveSharelink) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return errs.NotImplement
}

func (d *CloudreveSharelink) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotImplement
}

func (d *CloudreveSharelink) Remove(ctx context.Context, obj model.Obj) error {
	return errs.NotImplement
}

func (d *CloudreveSharelink) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	return errs.NotImplement
}

//func (d *CloudreveSharelink) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*CloudreveSharelink)(nil)

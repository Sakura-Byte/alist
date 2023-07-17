package onedrive_sharelink

import (
	"net/http"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	driver.RootPath
	// define other
	ShareLinkURL       string `json:"url" required:"true"`
	ShareLinkPassword  string `json:"password"`
	IsSharepoint       bool
	downloadLinkPrefix string
	Headers            http.Header
}

var config = driver.Config{
	Name:              "Onedrive Sharelink",
	OnlyProxy:         true,
	NoUpload:          true,
	DefaultRoot:       "/",
	CheckStatus:       false,
	NoOverwriteUpload: false,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &OnedriveSharelink{}
	})
}

package cloudreve_sharelink

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	driver.RootPath
	// define other
	Sharelink    string `json:"sharelink" required:"true"`
	Address      string
	SharelinkKey string
	Password     string `json:"password"`
	Cookie       string
}

var config = driver.Config{
	Name:        "Cloudreve Sharelink",
	DefaultRoot: "/",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &CloudreveSharelink{}
	})
}

package onedrive_vercel

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootPath
	Address string `json:"url" required:"true"`
	Host    string `json:"host"`
}

var config = driver.Config{
	Name:        "Onedrive_Vercel",
	LocalSort:   true,
	NoUpload:    true,
	DefaultRoot: "/",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &OnedriveVercel{}
	})
}

package goindex

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootPath
	URL string `json:"url" required:"true"`
}

var config = driver.Config{
	Name:        "GoIndex",
	LocalSort:   true,
	NoUpload:    true,
	DefaultRoot: "/",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &GoIndex{}
	})
}

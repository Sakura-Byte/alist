package alist_v3

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
	"golang.org/x/time/rate"
)

type Addition struct {
	driver.RootPath
	Address            string `json:"url" required:"true"`
	MetaPassword       string `json:"meta_password"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	Token              string `json:"token"`
	PassUAToUpsteam    bool   `json:"pass_ua_to_upsteam" default:"true"`
	CustomDownloadHost string `json:"custom_download_host"`
	RPSLimit           int    `json:"rps_limit" default:"3" description:"Requests per second limit"`
}

var config = driver.Config{
	Name:             "AList V3",
	LocalSort:        true,
	DefaultRoot:      "/",
	ProxyRangeOption: true,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		d := &AListV3{}
		// Set default if not provided.
		if d.RPSLimit <= 0 {
			d.RPSLimit = 10
		}
		// Initialize the rate limiter: limit and burst are set to RPSLimit.
		d.limiter = rate.NewLimiter(rate.Limit(d.RPSLimit), d.RPSLimit)
		return d
	})
}

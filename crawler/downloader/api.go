package downloader

import (
	"net/http"
	"psearch/util/errors"
)

type DownloaderApi struct {
	Addr string
}

func (self *DownloaderApi) Download(url string) (*http.Response, error) {
	resp, err := http.Get(self.Addr + (&Downloader{}).ApiUrl() + "?url=" + url)
	return resp, errors.NewErr(err)
}

package downloader

import (
	"net/http"
	"psearch/util/errors"
)

type DownloaderApi struct {
	Addr string
}

func (self *DownloaderApi) Download(url string) (*http.Response, error) {
	resp, err := http.Get("http://" + self.Addr + (&Downloader{}).ApiUrl() + "?url=" + url)
	return resp, errors.NewErr(err)
}

func (self *DownloaderApi) DownloadAll(urls []string) (*http.Response, error) {
	q := ""
	for i, u := range urls {
		if i == 0 {
			q += "?"
		} else {
			q += "&"
		}
		q += "url=" + u
	}
	resp, err := http.Get("http://" + self.Addr + (&Downloader{}).ApiUrl() + q)
	return resp, errors.NewErr(err)
}

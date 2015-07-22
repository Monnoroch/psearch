package downloader

import (
	"io/ioutil"
	"net/http"
	"psearch/util/errors"
)

type Downloader struct{}

func (self *Downloader) Download(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.NewErr(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.NewErr(err)
	}

	return string(body), nil
}

func (self *Downloader) DownloadAll(urls []string) ([]string, error) {
	res := make([]string, len(urls))
	for i, url := range urls {
		b, err := self.Download(url)
		if err != nil {
			return nil, err
		}

		res[i] = b
	}
	return res, nil
}

type DownloaderServer struct {
	Downloader *Downloader
}

func (self *DownloaderServer) Download(args *Args, result *string) error {
	r, err := self.Downloader.Download(args.Url)
	if err != nil {
		return err
	}

	*result = r
	return nil
}

func (self *DownloaderServer) DownloadAll(args *ArgsAll, result *[]string) error {
	r, err := self.Downloader.DownloadAll(args.Urls)
	if err != nil {
		return err
	}

	*result = r
	return nil
}

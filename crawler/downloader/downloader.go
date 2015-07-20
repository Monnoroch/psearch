package downloader

import (
	"io/ioutil"
	"net/http"
	"psearch/util"
	"psearch/util/errors"
)

type Downloader struct{}

func (self *Downloader) ApiUrl() string {
	return "/dl"
}

func (self *Downloader) Download(w http.ResponseWriter, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.NewErr(err)
	}
	defer resp.Body.Close()

	// TODO: do not read all, io.Copy instead
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.NewErr(err)
	}

	return util.SendJson(w, map[string]string{url: string(body)})
}

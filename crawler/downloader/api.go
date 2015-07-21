package downloader

import (
	"encoding/json"
	"net/http"
	"psearch/util/errors"
	"psearch/util/iter"
)

type DownloadResult struct {
	Status string "json:`status`"
	Res    string "json:`res`"
}

type DownloaderApi struct {
	Addr string
}

func (self *DownloaderApi) ApiUrl() string {
	return "/dl"
}

func (self *DownloaderApi) DownloadAll(urls iter.Iterator) (map[string]DownloadResult, error) {
	req, err := http.NewRequest("GET", "http://"+self.Addr+self.ApiUrl(), iter.ReadDelim(urls, []byte("\t")))
	if err != nil {
		return nil, errors.NewErr(err)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, errors.NewErr(err)
	}
	defer resp.Body.Close()

	var res map[string]DownloadResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.NewErr(err)
	}

	return res, nil
}

package downloader

import (
	"encoding/json"
	"psearch/util"
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
	resp, err := util.DoIterTsvRequest("http://"+self.Addr+self.ApiUrl(), urls)
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

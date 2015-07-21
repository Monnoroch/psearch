package downloader

import (
	"encoding/json"
	"net/http"
	"psearch/util/errors"
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

func (self *DownloaderApi) DownloadAll(urls []string) (map[string]DownloadResult, error) {
	q := ""
	for i, u := range urls {
		if i == 0 {
			q += "?"
		} else {
			q += "&"
		}
		q += "url=" + u
	}
	resp, err := http.Get("http://" + self.Addr + self.ApiUrl() + q)
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

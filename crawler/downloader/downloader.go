package downloader

import (
	"io/ioutil"
	"net/http"
	"psearch/util"
	"psearch/util/errors"
	"sync"
)

type Downloader struct{}

func (self *Downloader) DownloadAll(w http.ResponseWriter, urls []string) error {
	type resultData struct {
		Err error
		Res string
	}

	res := make([]resultData, len(urls))
	wg := sync.WaitGroup{}
	wg.Add(len(urls))
	for i, u := range urls {
		go func(idx int, url string) {
			defer wg.Done()

			resp, err := http.Get(url)
			if err != nil {
				res[idx].Err = errors.NewErr(err)
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				res[idx].Err = errors.NewErr(err)
				return
			}

			res[idx].Res = string(body)
		}(i, u)
	}
	wg.Wait()

	result := map[string]DownloadResult{}
	for i, v := range res {
		r := DownloadResult{}
		if v.Err != nil {
			r.Status = "error"
			r.Res = v.Err.Error()
		} else {
			r.Status = "ok"
			r.Res = v.Res
		}
		result[urls[i]] = r
	}
	return util.SendJson(w, result)
}

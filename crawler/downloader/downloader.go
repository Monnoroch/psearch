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

	result := map[string]map[string]string{}
	for i, v := range res {
		m := map[string]string{}
		if v.Err != nil {
			m["status"] = "error"
			m["res"] = v.Err.Error()
		} else {
			m["status"] = "ok"
			m["res"] = v.Res
		}
		result[urls[i]] = m
	}
	return util.SendJson(w, result)
}

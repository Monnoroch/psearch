package downloader

import (
	"io/ioutil"
	"net/http"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/iter"
	"sync"
)

type Downloader struct{}

func (self *Downloader) DownloadAll(w http.ResponseWriter, urls iter.Iterator) error {
	type resultData struct {
		Err error
		Res string
		Url string
	}

	results := make([]*resultData, 0)
	wg := sync.WaitGroup{}
	err := iter.Do(urls, func(u string) error {
		res := &resultData{
			Url: u,
		}
		results = append(results, res)
		wg.Add(1)
		go func(res *resultData) {
			defer wg.Done()

			resp, err := http.Get(res.Url)
			if err != nil {
				res.Err = errors.NewErr(err)
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				res.Err = errors.NewErr(err)
				return
			}

			res.Res = string(body)
		}(res)
		return nil
	})
	if err != nil {
		return err
	}

	wg.Wait()

	result := map[string]DownloadResult{}
	for _, v := range results {
		r := DownloadResult{}
		if v.Err != nil {
			r.Status = "error"
			r.Res = v.Err.Error()
		} else {
			r.Status = "ok"
			r.Res = v.Res
		}
		result[v.Url] = r
	}
	return util.SendJson(w, result)
}

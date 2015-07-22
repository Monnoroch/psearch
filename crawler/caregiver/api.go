package caregiver

import (
	"net/http"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/iter"
)

type CaregiverApi struct {
	Addr    string
	stopped bool
}

func (self *CaregiverApi) PushUrl() string {
	return "/push"
}

func (self *CaregiverApi) PullUrl() string {
	return "/pull"
}

func (self *CaregiverApi) PushUrls(urls iter.Iterator) (*http.Response, error) {
	resp, err := util.DoIterTsvRequest("http://"+self.Addr+self.PushUrl(), urls)
	if err != nil {
		return nil, errors.NewErr(err)
	}

	return resp, nil
}

type Response struct {
	Resp *http.Response
	Err  error
}

func (self *CaregiverApi) PullUrls(res chan Response) {
	self.stopped = false
	client := &http.Client{}
	url := "http://" + self.Addr + self.PullUrl()
	for {
		// this as almost correct, but we don't mind making one more request, very rarely
		if self.stopped {
			break
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			res <- Response{
				Resp: nil,
				Err:  errors.NewErr(err),
			}
			continue
		}

		resp, err := client.Do(req)
		res <- Response{
			Resp: resp,
			Err:  errors.NewErr(err),
		}
	}
}

func (self *CaregiverApi) Stop() {
	self.stopped = true
}

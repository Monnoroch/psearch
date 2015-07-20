package gatekeeper

import (
	"io"
	"net/http"
	"psearch/util/errors"
)

type GatekeeperApi struct {
	Addr string
}

func (self *GatekeeperApi) FindUrl() string {
	return "/find"
}

func (self *GatekeeperApi) ReadUrl() string {
	return "/read"
}

func (self *GatekeeperApi) WriteUrl() string {
	return "/write"
}

func (self *GatekeeperApi) Find(url string) (*http.Response, error) {
	resp, err := http.Get("http://" + self.Addr + self.FindUrl() + "?url=" + url)
	return resp, errors.NewErr(err)
}

func (self *GatekeeperApi) Read(url string) (*http.Response, error) {
	resp, err := http.Get("http://" + self.Addr + self.ReadUrl() + "?url=" + url)
	return resp, errors.NewErr(err)
}

func (self *GatekeeperApi) Write(url string, r io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("GET", "http://"+self.Addr+self.WriteUrl()+"?url="+url, r)
	if err != nil {
		return nil, errors.NewErr(err)
	}

	resp, err := (&http.Client{}).Do(req)
	return resp, errors.NewErr(err)
}

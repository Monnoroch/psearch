package dns

import (
	"net/http"
	"psearch/util/errors"
)

type ResolverApi struct {
	Addr string
}

func (self *ResolverApi) Resolve(url string) (*http.Response, error) {
	resp, err := http.Get("http://" + self.Addr + (&Resolver{}).ApiUrl() + "?host=" + url)
	return resp, errors.NewErr(err)
}

func (self *ResolverApi) ResolveAll(urls []string) (*http.Response, error) {
	q := ""
	for i, u := range urls {
		if i == 0 {
			q += "?"
		} else {
			q += "&"
		}
		q += "host=" + u
	}
	resp, err := http.Get("http://" + self.Addr + (&Resolver{}).ApiUrl() + q)
	return resp, errors.NewErr(err)
}

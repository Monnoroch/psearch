package dns

import (
	"net/http"
	"psearch/util/errors"
	"strings"
)

type ResolverApi struct {
	prefix string
}

func (self *ResolverApi) ApiUrl() string {
	return "/res"
}

func NewResolverApi(addr string) ResolverApi {
	return ResolverApi{
		prefix: "http://" + addr + (&ResolverApi{}).ApiUrl() + "?host=",
	}
}

func (self *ResolverApi) Resolve(url string) (*http.Response, error) {
	resp, err := http.Get(self.prefix + url)
	return resp, errors.NewErr(err)
}

func (self *ResolverApi) ResolveAll(hosts []string) (*http.Response, error) {
	resp, err := http.Get(self.prefix + strings.Join(hosts, "&host="))
	return resp, errors.NewErr(err)
}

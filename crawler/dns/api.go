package dns

import (
	"encoding/json"
	"net/http"
	"psearch/util/errors"
	"strings"
)

type HostResult struct {
	Status string   "json:`status`"
	Err    string   "json:`err`"
	Ips    []string "json:`ips`"
}

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

func (self *ResolverApi) ResolveAll(hosts []string) (map[string]HostResult, error) {
	resp, err := http.Get(self.prefix + strings.Join(hosts, "&host="))
	if err != nil {
		return nil, errors.NewErr(err)
	}
	defer resp.Body.Close()

	var res map[string]HostResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.NewErr(err)
	}

	return res, nil
}

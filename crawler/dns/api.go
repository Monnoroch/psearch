package dns

import (
	"encoding/json"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/iter"
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

func NewResolverApi(addr string) *ResolverApi {
	return &ResolverApi{
		prefix: "http://" + addr + (&ResolverApi{}).ApiUrl(),
	}
}

func (self *ResolverApi) ResolveAll(hosts iter.Iterator) (map[string]HostResult, error) {
	resp, err := util.DoIterTsvRequest(self.prefix, hosts)
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

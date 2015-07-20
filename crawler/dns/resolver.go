package dns

import (
	"net"
	"net/http"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/log"
	"sync"
	"time"
)

type dataT struct {
	ips []net.IP
	end time.Time
}

type Resolver struct {
	cacheTime time.Duration
	cache     map[string]dataT
}

func NewResolver(cacheTime time.Duration) *Resolver {
	return &Resolver{
		cacheTime: cacheTime,
		cache:     map[string]dataT{},
	}
}

func (self *Resolver) ApiUrl() string {
	return "/res"
}

func (self *Resolver) Resolve(host string) ([]net.IP, error) {
	now := time.Now()

	r, ok := self.cache[host]
	if ok && now.Before(r.end) {
		return r.ips, nil
	}

	res, err := net.LookupIP(host)
	if err != nil {
		if ok {
			return r.ips, errors.NewErr(err)
		} else {
			return nil, errors.NewErr(err)
		}
	}

	self.cache[host] = dataT{
		ips: res,
		end: now.Add(self.cacheTime),
	}
	return res, nil
}

func (self *Resolver) ResolveAll(w http.ResponseWriter, hosts []string) error {
	now := time.Now()

	type resultData struct {
		Err error
		Res []net.IP
	}

	results := make([]resultData, len(hosts))

	wg := sync.WaitGroup{}
	wg.Add(len(hosts))
	for i, h := range hosts {
		r, ok := self.cache[h]
		if ok && now.Before(r.end) {
			results[i].Res = r.ips
			continue
		}

		go func(idx int, host string) {
			defer wg.Done()

			res, err := net.LookupIP(host)
			if err != nil {
				if ok {
					results[idx].Res = r.ips
					results[idx].Err = errors.NewErr(err)
				} else {
					results[idx].Err = errors.NewErr(err)
				}
			}

			self.cache[host] = dataT{
				ips: res,
				end: now.Add(self.cacheTime),
			}

			results[idx].Res = res
		}(i, h)
	}
	wg.Wait()

	result := map[string]map[string]interface{}{}
	for i, v := range results {
		m := map[string]interface{}{}
		if v.Res != nil {
			m["status"] = "ok"
			m["res"] = v.Res
			if v.Err != nil {
				log.Errorln(v.Err)
			}
		} else {
			m["status"] = "error"
			m["res"] = v.Err.Error()
		}
		result[hosts[i]] = m
	}
	return util.SendJson(w, result)
}

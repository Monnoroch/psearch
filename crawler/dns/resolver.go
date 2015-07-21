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

	result := map[string]HostResult{}
	for i, v := range results {
		r := HostResult{}
		if v.Err != nil {
			r.Err = v.Err.Error()
			if v.Res != nil {
				log.Errorln(v.Err)
			}
		}

		if v.Res != nil {
			r.Status = "ok"
			r.Ips = make([]string, 0, len(v.Res))
			for _, v := range v.Res {
				r.Ips = append(r.Ips, v.String())
			}
		} else {
			r.Status = "error"
		}
		result[hosts[i]] = r
	}
	return util.SendJson(w, result)
}

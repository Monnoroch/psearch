package dns

import (
	"net"
	"net/http"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/iter"
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
	mutex     sync.RWMutex
}

func NewResolver(cacheTime time.Duration) *Resolver {
	return &Resolver{
		cacheTime: cacheTime,
		cache:     map[string]dataT{},
	}
}

func (self *Resolver) ResolveAll(w http.ResponseWriter, hosts iter.Iterator) error {
	type resultData struct {
		Err  error
		Res  []net.IP
		host string
	}

	results := make([]*resultData, 0)
	wg := sync.WaitGroup{}
	now := time.Now()
	err := iter.Do(hosts, func(h string) error {
		res := &resultData{
			host: h,
		}
		results = append(results, res)
		self.mutex.RLock()
		r, ok := self.cache[h]
		self.mutex.RUnlock()
		if ok && now.Before(r.end) {
			res.Res = r.ips
			return nil
		}

		wg.Add(1)
		go func(res *resultData, r *dataT, ok bool) {
			defer wg.Done()

			val, err := net.LookupIP(res.host)
			if err != nil {
				if ok {
					res.Res = r.ips
					res.Err = errors.NewErr(err)
				} else {
					res.Err = errors.NewErr(err)
				}
			}

			d := dataT{
				ips: val,
				end: now.Add(self.cacheTime),
			}
			self.mutex.Lock()
			self.cache[res.host] = d
			self.mutex.Unlock()

			res.Res = val
		}(res, &r, ok)
		return nil
	})
	if err != nil {
		return err
	}

	wg.Wait()

	result := map[string]HostResult{}
	for _, v := range results {
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
		result[v.host] = r
	}
	return util.SendJson(w, result)
}

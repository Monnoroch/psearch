package dns

import (
	"net"
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
	mutex     sync.RWMutex
}

func NewResolver(cacheTime time.Duration) *Resolver {
	return &Resolver{
		cacheTime: cacheTime,
		cache:     map[string]dataT{},
	}
}

func (self *Resolver) load(data *dataT, res *[]string) {
	*res = make([]string, len(data.ips))
	for i, v := range data.ips {
		(*res)[i] = v.String()
	}
}

func (self *Resolver) Resolve(host string) ([]string, error) {
	log.Printf("Resolver.Resolve(%s)\n", host)
	self.mutex.RLock()
	r, hacheHit := self.cache[host]
	self.mutex.RUnlock()

	var res []string
	if hacheHit && time.Now().Before(r.end) {
		self.load(&r, &res)
		log.Printf("Resolver.Resolve(%s) OK (cache)!\n", host)
		return res, nil
	}

	val, err := net.LookupIP(host)
	if err != nil {
		err = errors.NewErr(err)
		if hacheHit {
			self.load(&r, &res)
			log.Errorln(err)
			log.Printf("Resolver.Resolve(%s) OK (cache, but error)!\n", host)
			return res, nil
		}
		return nil, err
	}

	d := dataT{
		ips: make([]net.IP, 0, len(val)),
		end: time.Now().Add(self.cacheTime),
	}
	for _, v := range val {
		// if ip4 := v.To4(); len(ip4) == net.IPv4len {
		d.ips = append(d.ips, v)
		// }
	}

	self.mutex.Lock()
	self.cache[host] = d
	self.mutex.Unlock()

	self.load(&d, &res)
	log.Printf("Resolver.Resolve(%s) OK!\n", host)
	return res, nil
}

func (self *Resolver) ResolveAll(hosts []string) ([][]string, error) {
	log.Printf("Resolver.ResolveAll(%#v)\n", hosts)
	res := make([][]string, len(hosts))
	for i, host := range hosts {
		b, err := self.Resolve(host)
		if err != nil {
			return nil, err
		}

		res[i] = b
	}
	log.Printf("Resolver.ResolveAll(%#v) OK\n", hosts)
	return res, nil
}

type ResolverServer struct {
	Resolver *Resolver
}

func (self *ResolverServer) Resolve(args *Args, result *[]string) error {
	r, err := self.Resolver.Resolve(args.Host)
	if err != nil {
		log.Errorln(err, args)
		return err
	}

	*result = r
	return nil
}

func (self *ResolverServer) ResolveAll(args *ArgsAll, result *[][]string) error {
	r, err := self.Resolver.ResolveAll(args.Hosts)
	if err != nil {
		log.Errorln(err, args)
		return err
	}

	*result = r
	return nil
}

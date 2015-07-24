package caregiver

import (
	"net/url"
	"psearch/crawler/dns"
	"psearch/crawler/downloader"
	"psearch/util/errors"
	"psearch/util/log"
	"strings"
	"sync"
	"time"
)

type hostData struct {
	timeout  time.Duration
	maxCount uint
	urls     LockedQueue
	mutex    sync.Mutex
	end      time.Time
}

type Caregiver struct {
	downloader      downloader.DownloaderClient
	dns             *dns.ResolverClient
	hosts           map[string]*hostData
	mutex           sync.Mutex
	defaultTimeout  time.Duration
	defaultMaxCount uint
	data            map[string]string
	dataMutex       sync.RWMutex
	pullTimeout     time.Duration
	WorkTimeout     time.Duration
}

func NewCaregiver(dlAddr, dnsAddr string, defaultMaxCount uint, defaultTimeout, pullTimeout, workTimeout time.Duration) (*Caregiver, error) {
	dlc, err := downloader.NewDownloaderClient(dlAddr)
	if err != nil {
		return nil, err
	}

	res := &Caregiver{
		downloader:      dlc,
		hosts:           map[string]*hostData{},
		defaultTimeout:  defaultTimeout,
		defaultMaxCount: defaultMaxCount,
		data:            map[string]string{},
		pullTimeout:     pullTimeout,
		WorkTimeout:     workTimeout,
	}
	if dnsAddr != "" {
		dnc, err := dns.NewResolverClient(dnsAddr)
		if err != nil {
			return nil, err
		}

		res.dns = &dnc
	}
	return res, nil
}

func (self *Caregiver) PushUrls(urls []string) error {
	log.Printf("Caregiver.PushUrls(%#v)\n", urls)
	data := map[string][]string{}
	for _, u := range urls {
		url, err := url.Parse(u)
		if err != nil {
			return errors.NewErr(err)
		}

		path := url.Path
		if len(path) != 0 && path[0] == '/' {
			path = path[1:]
		}
		if url.RawQuery != "" {
			path += "?"
			path += url.RawQuery
		}
		data[url.Host] = append(data[url.Host], path)
	}

	for host, urls := range data {
		self.mutex.Lock()
		hosts, ok := self.hosts[host]
		if !ok {
			hosts = self.getData(host)
			self.hosts[host] = hosts
		}
		self.mutex.Unlock()
		hosts.urls.EnqueueAll(urls...)
	}
	log.Printf("Caregiver.PushUrls(%#v) OK\n", urls)
	return nil
}

func (self *Caregiver) PullUrls() map[string]string {
	log.Printf("Caregiver.PullUrls()\n")
	for {
		self.dataMutex.RLock()
		sleep := len(self.data) == 0
		self.dataMutex.RUnlock()
		if !sleep {
			break
		}
		time.Sleep(self.pullTimeout)
	}

	self.dataMutex.Lock()
	defer self.dataMutex.Unlock()
	res := self.data
	self.data = map[string]string{}
	log.Printf("Caregiver.PullUrls() OK (%v)\n", len(res))
	return res
}

func (self *Caregiver) Start() error {
	log.Printf("Caregiver.Start()\n")
	for {
		// выделим хосты, которын можно по таймауту качать
		data := map[string]*hostData{}
		self.mutex.Lock()
		now := time.Now()
		for k, v := range self.hosts {
			if v.urls.Len() != 0 && v.end.Before(now) {
				data[k] = v
			}
		}
		self.mutex.Unlock()
		if len(data) == 0 {
			time.Sleep(self.WorkTimeout)
			continue
		}

		log.Printf("Caregiver.Start(): start downloading\n")

		// резолвим dns
		hosts := make([]string, 0, len(data))
		for k, _ := range data {
			hosts = append(hosts, k)
		}

		var err error
		ips := map[string][]string{}
		// if self.dns != nil {
		// 	ips, err = self.dns.ResolveAll(hosts)
		// 	if err != nil {
		// 		return err
		// 	}
		// }

		now = time.Now()
		// можно качать!
		urls := []string{}
		for k, v := range data {
			/*
				TODO: запросы по ip почему-то не работаютю.
				Подозреваю, что сервер читает RequestURL и нужно будет реализовать свой стек http, чтобы все заработало.
			*/

			ip := k
			if v, ok := ips[k]; ok && len(v) != 0 {
				ip = v[0]
				if strings.Contains(ip, ":") {
					ip = "[" + ip + "]:80"
				}
			}
			host := "http://" + ip
			us := v.urls.DequeueN(v.maxCount)
			for i := 0; i < len(us); i += 1 {
				if len(us[i]) != 0 {
					us[i] = host + "/" + us[i]
				} else {
					us[i] = host
				}
			}
			urls = append(urls, us...)
		}

		log.Printf("Caregiver.Start(): collected urls %#v\n", urls)
		docs, err := self.downloader.DownloadAll(urls)
		if err != nil {
			log.Errorln(err)
		}

		now = time.Now()
		for _, v := range data {
			v.end = now.Add(v.timeout)
		}

		for i, v := range docs {
			if v != "" {
				self.dataMutex.Lock()
				self.data[urls[i]] = v
				self.dataMutex.Unlock()
			} else {
				log.Errorln("Couldn't download url "+urls[i]+",", v)
			}
		}

		log.Printf("Caregiver.Start(): downloaded urls %#v\n", urls)
	}
}

func (self *Caregiver) getData(host string) *hostData {
	return &hostData{
		timeout:  self.defaultTimeout,
		maxCount: self.defaultMaxCount,
	}
}

type CaregiverServer struct {
	Caregiver *Caregiver
}

func (self *CaregiverServer) PushUrls(args *Args, result *struct{}) error {
	err := self.Caregiver.PushUrls(args.Urls)
	if err != nil {
		log.Errorln(err, args)
	}
	return err
}

func (self *CaregiverServer) PullUrls(args *struct{}, result *map[string]string) error {
	*result = self.Caregiver.PullUrls()
	return nil
}

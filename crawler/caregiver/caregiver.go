package caregiver

import (
	"net/http"
	"net/url"
	"psearch/crawler/dns"
	"psearch/crawler/downloader"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/iter"
	"psearch/util/log"
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
	downloader      downloader.DownloaderApi
	dns             dns.ResolverApi
	hosts           map[string]*hostData
	mutex           sync.Mutex
	defaultTimeout  time.Duration
	defaultMaxCount uint
	data            map[string]string
	dataMutex       sync.RWMutex
	pullTimeout     time.Duration
	WorkTimeout     time.Duration
}

func NewCaregiver(dlAddr, dnsAddr string, defaultMaxCount uint, defaultTimeout, pullTimeout, workTimeout time.Duration) *Caregiver {
	return &Caregiver{
		downloader: downloader.DownloaderApi{
			Addr: dlAddr,
		},
		dns:             dns.NewResolverApi(dnsAddr),
		hosts:           map[string]*hostData{},
		defaultTimeout:  defaultTimeout,
		defaultMaxCount: defaultMaxCount,
		data:            map[string]string{},
		pullTimeout:     pullTimeout,
		WorkTimeout:     workTimeout,
	}
}

func (self *Caregiver) PushUrls(urls iter.Iterator) error {
	data := map[string][]string{}
	err := iter.Do(urls, func(u string) error {
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
		return nil
	})
	if err != nil {
		return err
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
	return nil
}

func (self *Caregiver) PullUrls(w http.ResponseWriter) error {
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
	res := util.SendJson(w, self.data)
	self.data = map[string]string{}
	return res
}

func (self *Caregiver) Start() error {
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

		// резолвим dns
		hosts := make([]string, 0, len(data))
		for k, _ := range data {
			hosts = append(hosts, k)
		}

		ips, err := self.dns.ResolveAll(iter.Array(hosts))
		if err != nil {
			return err
		}

		// если какие-то хосты не зарезолвились, удалим
		now = time.Now()
		for k, v := range ips {
			if v.Ips == nil {
				delete(data, k)
			}
		}

		// можно качать!
		// TODO: iter.Generator!
		urls := []string{}
		for k, v := range data {
			host := "http://" + k + "/"
			us := v.urls.DequeueN(v.maxCount)
			for i := 0; i < len(us); i += 1 {
				us[i] = host + us[i]
			}
			urls = append(urls, us...)
		}

		docs, err := self.downloader.DownloadAll(iter.Array(urls))
		if err != nil {
			return err
		}

		now = time.Now()
		for _, v := range data {
			v.end = now.Add(v.timeout)
		}

		for k, v := range docs {
			if v.Status == "ok" {
				self.dataMutex.Lock()
				self.data[k] = v.Res
				self.dataMutex.Unlock()
			} else {
				log.Errorln("Couldn't download url "+k+",", v)
			}
		}
	}
	return nil
}

func (self *Caregiver) getData(host string) *hostData {
	return &hostData{
		timeout:  self.defaultTimeout,
		maxCount: self.defaultMaxCount,
	}
}

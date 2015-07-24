package spider

import (
	"net/url"
	"psearch/crawler/caregiver"
	"psearch/gatekeeper"
	"psearch/util/log"
	"regexp"
	"sync"
	"time"
)

var urlRegex *regexp.Regexp = regexp.MustCompile(`<a\s.*?\s?href\s*?=\s*?['"]\s*?(?P<url>.+?)\s*?['"]`)

type Spider struct {
	gk        gatekeeper.GatekeeperClient
	cg        caregiver.CaregiverClient
	urls      caregiver.LockedQueue
	waitUrls  map[string]struct{}
	waitMutex sync.Mutex
	doneUrls  map[string]struct{}
	interval  time.Duration
	pushCnt   uint
}

func NewSpider(gk, cg string, interval time.Duration, pushCnt uint) (*Spider, error) {
	gkc, err := gatekeeper.NewGatekeeperClient(gk)
	if err != nil {
		return nil, err
	}

	cgc, err := caregiver.NewCaregiverClient(cg)
	if err != nil {
		return nil, err
	}

	return &Spider{
		gk:       gkc,
		cg:       cgc,
		waitUrls: map[string]struct{}{},
		doneUrls: map[string]struct{}{},
		interval: interval,
		pushCnt:  pushCnt,
	}, nil
}

func (self *Spider) RunPusher() error {
	log.Printf("Spider.RunPusher()\n")
	for {
		// попробуем достать урлы, которые надо обойти
		urls := self.urls.DequeueN(self.pushCnt)
		if urls == nil {
			time.Sleep(self.interval)
			continue
		}

		log.Printf("Spider.RunPusher(): push urls %#v\n", urls)

		// скажем менеджеру загрузки их обойти
		if err := self.cg.PushUrls(urls); err != nil {
			return err
		}

		// положим их в список ожидающих
		self.waitMutex.Lock()
		for _, url := range urls {
			self.waitUrls[url] = struct{}{}
		}
		self.waitMutex.Unlock()
		log.Printf("Spider.RunPusher(): pushed urls %#v\n", urls)
	}
}

func (self *Spider) RunPuller() error {
	log.Printf("Spider.RunPuller()\n")
	for {
		now := time.Now()
		// получим документы
		urls, err := self.cg.PullUrls()
		if err != nil {
			return err
		}

		log.Printf("Spider.RunPuller(): pulled urls\n")

		// запишем их в хранилище
		toDel := []string{}
		for url, v := range urls {
			_, err := self.gk.Write(url, v)
			if err != nil {
				toDel = append(toDel, url)
			}
		}

		if len(toDel) != 0 {
			// те, что не получилось, попросим перекачать потом
			log.Printf("Spider.RunPuller(): reenqueue urls %#v\n", toDel)
			self.urls.EnqueueAll(toDel...)
			for _, url := range toDel {
				delete(urls, url)
			}
		}

		// те, что получилось удалим из ожидания
		self.waitMutex.Lock()
		for url, _ := range urls {
			delete(self.waitUrls, url)
		}
		self.waitMutex.Unlock()

		// и отметим, что выкачали
		for url, _ := range urls {
			self.doneUrls[url] = struct{}{}
		}

		log.Printf("Spider.RunPuller(): find urls\n")

		// теперь распарсим документы на предмет урлов и добавим их в список желаемых
		newUrls := make([]string, 0, 100 /*TODO: ajust multiplier at runtime?*/ *len(urls))
		for k, v := range urls {
			curr, err := url.Parse(k)
			if err != nil {
				return err
			}

			matches := urlRegex.FindAllStringSubmatch(v, -1)
			for _, v := range matches {
				u := v[1]
				if parsed, err := url.Parse(u); err == nil {
					parsed.Fragment = ""
					if parsed.Scheme == "" {
						parsed.Scheme = curr.Scheme
					}
					if parsed.Host == "" {
						parsed.Host = curr.Host
					}
					newUrls = append(newUrls, parsed.String())
				}
			}
		}

		// проверим, есть ли такие урлы в ожидании, если есть, отменим их
		self.waitMutex.Lock()
		for i := 0; i < len(newUrls); i += 1 {
			if _, ok := self.waitUrls[newUrls[i]]; ok {
				newUrls[i] = ""
			}
		}
		self.waitMutex.Unlock()
		// проверим, есть ли такие урлы в скачанных, если есть, отменим их
		for i := 0; i < len(newUrls); i += 1 {
			if newUrls[i] == "" {
				continue
			}
			if _, ok := self.doneUrls[newUrls[i]]; ok {
				newUrls[i] = ""
			}
		}

		// TODO: тут еще по robots.txt для хоста вычистить урлы

		// проверим, есть ли такие урлы в хранилище, если есть, отменим их
		for i := 0; i < len(newUrls); i += 1 {
			if newUrls[i] == "" {
				continue
			}

			_, ok, err := self.gk.Find(newUrls[i])
			if err != nil {
				return err
			}

			if ok {
				newUrls[i] = ""
			}
		}

		tmp := map[string]struct{}{}
		for _, url := range newUrls {
			if url != "" {
				tmp[url] = struct{}{}
			}
		}

		// положим оставшиеся в очередь
		toQ := make([]string, 0, len(tmp))
		for url, _ := range tmp {
			toQ = append(toQ, url)
		}

		if len(toQ) != 0 {
			self.AddUrls(toQ)
		}

		passed := time.Now().Sub(now)
		if self.interval > passed {
			time.Sleep(self.interval - passed)
		}
	}
}

// TODO: тут эта штука должна распределять урлы по шардам и по сети раздавать
func (self *Spider) AddUrls(urls []string) {
	log.Printf("Spider.AddUrls(%#v)\n", urls)
	self.urls.EnqueueAll(urls...)
	// положим их в список ожидающих
	self.waitMutex.Lock()
	for _, url := range urls {
		self.waitUrls[url] = struct{}{}
	}
	self.waitMutex.Unlock()
	log.Printf("Spider.AddUrls(%#v) OK\n", urls)
}

type SpiderServer struct {
	Spider *Spider
}

func (self *SpiderServer) AddUrl(args *Args, result *struct{}) error {
	self.Spider.AddUrls([]string{args.Url})
	return nil
}

func (self *SpiderServer) AddUrls(args *ArgsAll, result *struct{}) error {
	self.Spider.AddUrls(args.Urls)
	return nil
}

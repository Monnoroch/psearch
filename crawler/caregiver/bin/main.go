package main

import (
	"flag"
	"net/http"
	"psearch/crawler/caregiver"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/graceful"
	"psearch/util/iter"
	"psearch/util/log"
	"strconv"
	"strings"
	"time"
)

func main() {
	var help = flag.Bool("help", false, "print help")
	var port = flag.Int("port", -1, "port to listen")
	var dlerArrd = flag.String("dl", "", "downloader address")
	var dnsAddr = flag.String("dns", "", "dns resolver address")
	var defaultMaxCount = flag.Int("maxcount", 1, "default maximum count urls for specific host to download at once")
	var defaultTimeout = flag.Int("timeout", 0, "default timeout between downloads for specific host (im ms)")
	var pullTimeout = flag.Int("pull-timeout", 100, "pull check timeout (im ms)")
	var workTimeout = flag.Int("work-timeout", 1000, "pull check timeout (im ms)")
	var gracefulRestart = graceful.SetFlag()
	flag.Parse()

	if *help || *port == -1 || *dlerArrd == "" || *dnsAddr == "" {
		flag.PrintDefaults()
		return
	}

	ct := caregiver.NewCaregiver(
		*dlerArrd,
		*dnsAddr,
		uint(*defaultMaxCount),
		time.Duration(*defaultTimeout)*time.Millisecond,
		time.Duration(*pullTimeout)*time.Millisecond,
		time.Duration(*workTimeout)*time.Millisecond,
	)
	server := graceful.NewServer(http.Server{})

	http.HandleFunc("/favicon.ico", graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})))

	http.HandleFunc((&caregiver.CaregiverApi{}).PushUrl(), graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		r.ParseForm()
		urls := util.GetParams(r, "url")
		log.Println("Push urls", urls)
		if err := ct.PushUrls(iter.Chain(iter.Array(urls), iter.Delim(r.Body, '\t'))); err != nil {
			return err
		}

		log.Println("Pushed URLs: " + strings.Join(urls, ", "))
		return nil
	})))

	http.HandleFunc((&caregiver.CaregiverApi{}).PullUrl(), graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		if err := ct.PullUrls(w); err != nil {
			return err
		}

		log.Println("Pulled URLs!")
		return nil
	})))

	server.SetSighup()

	go func() {
		for {
			if err := ct.Start(); err != nil {
				log.Errorln(err)
				time.Sleep(ct.WorkTimeout)
				continue
			}
			break
		}
	}()

	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	if err := server.Restart(); err != nil {
		log.Fatal(err)
	}
}

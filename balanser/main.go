package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"psearch/balanser/chooser"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/graceful"
	"psearch/util/log"
	"strconv"
)

type Urls []string

func (self *Urls) String() string {
	return fmt.Sprintf("%d", *self)
}

func (self *Urls) Set(value string) error {
	*self = append(*self, value)
	return nil
}

func main() {
	var help = flag.Bool("help", false, "print help")
	var port = flag.Int("port", -1, "port to listen")
	var rout = flag.String("rout", "random", "routing policy: random, roundrobin")
	var urls Urls
	flag.Var(&urls, "url", "backend url")
	var gracefulRestart = graceful.SetFlag()
	flag.Parse()

	if *help || *port == -1 || urls == nil {
		flag.PrintDefaults()
		return
	}

	router, err := chooser.NewChooser(*rout, urls)
	if err != nil {
		flag.PrintDefaults()
		return
	}

	server := graceful.NewServer(http.Server{})

	// NOTE: for testing only, so browser wouldn't spam.
	http.HandleFunc("/favicon.ico", graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})))

	http.HandleFunc("/", graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != "GET" {
			return util.ClientError(errors.New("Can only process GET requests!"))
		}

		failed := map[string]struct{}{}
		for {
			if len(failed) == len(urls) {
				return errors.New("All backends broken!")
			}

			backend := router.Choose()
			if _, ok := failed[backend]; ok {
				continue
			}

			r.URL.Scheme = "http"
			r.URL.Host = backend
			nreq, err := http.NewRequest("GET", r.URL.String(), nil)
			if err != nil {
				return errors.NewErr(err)
			}

			nreq.Header = r.Header
			resp, err := (&http.Client{}).Do(nreq)
			if err != nil {
				failed[backend] = struct{}{}
				log.Errorln(err)
				continue
			}
			defer resp.Body.Close()

			for k, hs := range resp.Header {
				for _, val := range hs {
					w.Header().Add(k, val)
				}
			}
			w.WriteHeader(resp.StatusCode)
			_, err = io.Copy(w, resp.Body)
			if err == nil {
				log.Println("Chosen backend: " + backend)
			}
			return errors.NewErr(err)
		}
	})))

	server.SetSighup()

	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	if err := server.Restart(); err != nil {
		log.Fatal(err)
	}
}

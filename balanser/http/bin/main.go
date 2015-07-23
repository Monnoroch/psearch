package main

import (
	"flag"
	"fmt"
	"net/http"
	"psearch/balanser"
	bhttp "psearch/balanser/http"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/graceful"
	ghttp "psearch/util/graceful/http"
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

	router, err := balanser.NewChooser(*rout, urls)
	if err != nil {
		flag.PrintDefaults()
		return
	}

	balanser := bhttp.NewBalanser(router, urls)
	server := ghttp.NewServer(&http.Server{})

	// NOTE: for testing only, so browser wouldn't spam.
	http.HandleFunc("/favicon.ico", ghttp.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})))

	http.HandleFunc("/", ghttp.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		errs, err := balanser.Request(w, r)
		for _, e := range errs {
			log.Println(e)
		}
		return err
	})))

	graceful.SetSighup(server)

	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	if err := graceful.Restart(server); err != nil {
		log.Fatal(err)
	}
}

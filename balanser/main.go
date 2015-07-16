package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"psearch/balanser/chooser"
	"psearch/util"
	"psearch/util/errors"
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

	http.HandleFunc("/favicon.ico", util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}))

	http.HandleFunc("/", util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		backend := router.Choose()
		r.URL.Scheme = "http"
		r.URL.Host = backend
		resp, err := http.Get(r.URL.String())
		if err != nil {
			return errors.NewErr(err)
		}
		defer resp.Body.Close()

		_, err = io.Copy(w, resp.Body)
		if err == nil {
			log.Println("Chosen backend: " + backend + "   " + r.URL.String())
		}
		return errors.NewErr(err)
	}))

	if err := http.ListenAndServe(":"+strconv.Itoa(*port), nil); err != nil {
		log.Fatal(errors.NewErr(err))
	}
}

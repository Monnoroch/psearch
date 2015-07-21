package main

import (
	"flag"
	"net/http"
	"psearch/crawler/dns"
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
	var cacheTime = flag.Int("cachetime", -1, "time to keep resolved records in cache (in seconds)")
	var gracefulRestart = graceful.SetFlag()
	flag.Parse()

	if *help || *port == -1 || *cacheTime == -1 {
		flag.PrintDefaults()
		return
	}

	resolver := dns.NewResolver(time.Duration(*cacheTime) * time.Second)

	server := graceful.NewServer(http.Server{})

	http.HandleFunc("/favicon.ico", graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})))

	http.HandleFunc((&dns.ResolverApi{}).ApiUrl(), graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		r.ParseForm()
		hosts := util.GetParams(r, "host")
		if err := resolver.ResolveAll(w, iter.Chain(iter.Array(hosts), iter.Delim(r.Body, '\t'))); err != nil {
			return err
		}

		log.Println("Resolved hosts: " + strings.Join(hosts, ", "))
		return nil
	})))

	server.SetSighup()

	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	if err := server.Restart(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"flag"
	"net/http"
	"psearch/crawler/downloader"
	"psearch/gatekeeper"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/graceful"
	"psearch/util/log"
	"strconv"
	"strings"
)

func main() {
	var help = flag.Bool("help", false, "print help")
	var port = flag.Int("port", -1, "port to listen")
	var gk = flag.String("gk", "", "gatekeeper address")
	var gracefulRestart = graceful.SetFlag()
	flag.Parse()

	if *help || *port == -1 {
		flag.PrintDefaults()
		return
	}

	dl := &downloader.Downloader{}
	if *gk != "" {
		dl.Gk = &gatekeeper.GatekeeperApi{Addr: *gk}
	}
	server := graceful.NewServer(http.Server{})

	http.HandleFunc("/favicon.ico", graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})))

	http.HandleFunc((&downloader.DownloaderApi{}).ApiUrl(), graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		r.ParseForm()
		urls, err := util.GetParams(r, "url")
		if err != nil {
			return util.ClientError(err)
		}

		if err := dl.DownloadAll(w, urls); err != nil {
			return err
		}

		log.Println("Served URLs: " + strings.Join(urls, ", "))
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

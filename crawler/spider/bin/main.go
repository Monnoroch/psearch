package main

import (
	"flag"
	"net/rpc"
	"psearch/crawler/spider"
	"psearch/util/errors"
	"psearch/util/graceful"
	gjsonrpc "psearch/util/graceful/jsonrpc"
	"psearch/util/log"
	"strconv"
	"time"
)

func main() {
	var help = flag.Bool("help", false, "print help")
	var port = flag.Int("port", -1, "port to listen")
	var gkArrd = flag.String("gatekeeper", "", "gatekeeper address")
	var cgAddr = flag.String("caregiver", "", "caregiver address")
	var vint = flag.Int("interval", 1, "sleep interval")
	var pushCnt = flag.Int("push-cnt", 10, "urls to push to the caregiver at a time")
	var gracefulRestart = graceful.SetFlag()
	flag.Parse()

	if *help || *port == -1 || *gkArrd == "" || *cgAddr == "" {
		flag.PrintDefaults()
		return
	}

	sp, err := spider.NewSpider(*gkArrd, *cgAddr, time.Duration(*vint)*time.Second, uint(*pushCnt))
	if err != nil {
		log.Fatal(err)
	}

	srv := rpc.NewServer()
	srv.Register(&spider.SpiderServer{sp})

	server := gjsonrpc.NewServer(srv)
	graceful.SetSighup(server)

	go func() {
		for {
			if err := sp.RunPusher(); err != nil {
				log.Errorln(err)
			}
		}
	}()

	go func() {
		for {
			if err := sp.RunPuller(); err != nil {
				log.Errorln(err)
			}
		}
	}()

	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	if err := graceful.Restart(server); err != nil {
		log.Fatal(err)
	}
}

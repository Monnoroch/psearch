package main

import (
	"flag"
	"net/rpc"
	"psearch/crawler/dns"
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
	var cacheTime = flag.Int("cachetime", -1, "time to keep resolved records in cache (in seconds)")
	var gracefulRestart = graceful.SetFlag()
	flag.Parse()

	if *help || *port == -1 || *cacheTime == -1 {
		flag.PrintDefaults()
		return
	}

	srv := rpc.NewServer()
	srv.Register(&dns.ResolverServer{dns.NewResolver(time.Duration(*cacheTime) * time.Second)})
	log.Println(srv)

	server := gjsonrpc.NewServer(srv)
	graceful.SetSighup(server)

	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	if err := graceful.Restart(server); err != nil {
		log.Fatal(err)
	}
}

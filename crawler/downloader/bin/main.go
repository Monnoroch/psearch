package main

import (
	"flag"
	"net/rpc"
	"psearch/crawler/downloader"
	"psearch/util/errors"
	"psearch/util/graceful"
	gjsonrpc "psearch/util/graceful/jsonrpc"
	"psearch/util/log"
	"strconv"
)

func main() {
	var help = flag.Bool("help", false, "print help")
	var port = flag.Int("port", -1, "port to listen")
	var gracefulRestart = graceful.SetFlag()
	flag.Parse()

	if *help || *port == -1 {
		flag.PrintDefaults()
		return
	}

	srv := rpc.NewServer()
	srv.Register(&downloader.DownloaderServer{&downloader.Downloader{}})

	server := gjsonrpc.NewServer(srv)
	graceful.SetSighup(server)

	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	if err := graceful.Restart(server); err != nil {
		log.Fatal(err)
	}
}

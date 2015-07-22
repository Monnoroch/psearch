package main

import (
	"flag"
	"net/rpc"
	"psearch/gatekeeper"
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
	var dir = flag.String("dir", "", "data directory")
	var maxFileSize = flag.Int("max-size", 10*1024*1024, "maximum file size")
	var maxTime = flag.Int("max-time", 1*60, "maximum time between sync calls (in seconds)")
	var gracefulRestart = graceful.SetFlag()
	flag.Parse()

	if *help || *port == -1 {
		flag.PrintDefaults()
		return
	}

	gk, err := gatekeeper.NewGatekeeper(*dir, uint64(*maxFileSize), time.Duration(*maxTime)*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	srv := rpc.NewServer()
	srv.Register(&gatekeeper.GatekeeperServer{gk})

	server := gjsonrpc.NewServer(srv)
	graceful.SetSighup(server)

	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	if err := graceful.Restart(server); err != nil {
		log.Fatal(err)
	}
}

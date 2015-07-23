package main

import (
	"flag"
	"fmt"
	"net/rpc"
	"psearch/balanser"
	btcp "psearch/balanser/tcp"
	// "psearch/util"
	"io"
	"net"
	"psearch/util/errors"
	"psearch/util/graceful"
	grpc "psearch/util/graceful/rpc"
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

	balanser := btcp.NewBalanser(router, urls)
	server := grpc.NewServer(rpc.NewServer(), func(srv *rpc.Server, conn io.ReadWriteCloser) {
		if err := balanser.Request(conn.(*net.TCPConn)); err != nil {
			log.Errorln(err)
		}
	})

	graceful.SetSighup(server)

	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	if err := graceful.Restart(server); err != nil {
		log.Fatal(err)
	}
}

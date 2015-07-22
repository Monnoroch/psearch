package main

import (
	"flag"
	"net/rpc"
	"psearch/crawler/caregiver"
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
	var dlerArrd = flag.String("dl", "", "downloader address")
	var dnsAddr = flag.String("dns", "", "dns resolver address")
	var defaultMaxCount = flag.Int("maxcount", 1, "default maximum count urls for specific host to download at once")
	var defaultTimeout = flag.Int("timeout", 0, "default timeout between downloads for specific host (im ms)")
	var pullTimeout = flag.Int("pull-timeout", 100, "pull check timeout (im ms)")
	var workTimeout = flag.Int("work-timeout", 1000, "pull check timeout (im ms)")
	var gracefulRestart = graceful.SetFlag()
	flag.Parse()

	if *help || *port == -1 || *dlerArrd == "" {
		flag.PrintDefaults()
		return
	}

	ct, err := caregiver.NewCaregiver(
		*dlerArrd,
		*dnsAddr,
		uint(*defaultMaxCount),
		time.Duration(*defaultTimeout)*time.Millisecond,
		time.Duration(*pullTimeout)*time.Millisecond,
		time.Duration(*workTimeout)*time.Millisecond,
	)
	if err != nil {
		log.Fatal(err)
	}

	srv := rpc.NewServer()
	srv.Register(&caregiver.CaregiverServer{ct})

	server := gjsonrpc.NewServer(srv)
	graceful.SetSighup(server)

	go func() {
		for {
			if err := ct.Start(); err != nil {
				log.Errorln(err)
				time.Sleep(ct.WorkTimeout)
				continue
			}
			break
		}
	}()

	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	if err := graceful.Restart(server); err != nil {
		log.Fatal(err)
	}
}

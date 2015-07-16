package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/graceful"
	"psearch/util/log"
	"strconv"
	"syscall"
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

	server := graceful.NewServer(http.Server{})

	http.HandleFunc("/favicon.ico", graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})))

	http.HandleFunc("/dl", graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		r.ParseForm()
		url, err := util.GetParam(r, "url")
		if err != nil {
			return err
		}

		resp, err := http.Get(url)
		if err != nil {
			return errors.NewErr(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.NewErr(err)
		}

		err = util.SendJson(w, map[string]string{url: string(body)})
		if err == nil {
			log.Println("Served URL: " + url)
		}
		return err
	})))

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)
	go func() {
		for {
			s := <-signals
			switch s {
			case syscall.SIGHUP:
				server.Stop()
				log.Println("stopped!")
			default:
				log.Println("Unknown signal.")
			}
		}
	}()

	log.Println("serving!")
	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Println("served err!")
		log.Fatal(errors.NewErr(err))
	}

	log.Println("served ok!")
	if err := server.Restart(); err != nil {
		log.Fatal(err)
	}
}

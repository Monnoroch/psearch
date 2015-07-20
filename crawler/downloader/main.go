package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/graceful"
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

	server.SetSighup()

	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	if err := server.Restart(); err != nil {
		log.Fatal(err)
	}
}

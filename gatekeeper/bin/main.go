package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"psearch/gatekeeper"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/graceful"
	"psearch/util/log"
	"strconv"
	"time"
)

func getUrl(r *http.Request) (string, string, error) {
	r.ParseForm()
	u, err := util.GetParam(r, "url")
	if err != nil {
		return "", "", util.ClientError(err)
	}

	url, err := gatekeeper.UrlTransform(u)
	if err != nil {
		return "", "", util.ClientError(err)
	}

	return u, url, nil
}

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

	server := graceful.NewServer(http.Server{})

	http.HandleFunc("/favicon.ico", graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})))

	http.HandleFunc("/find", graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		u, url, err := getUrl(r)
		if err != nil {
			return err
		}

		val, ok := gk.Find(url)
		if ok {
			log.Printf("Found URL: %s: %+v\n", u, val)
		} else {
			log.Printf("Not found URL: %s\n", u)
		}
		return util.SendJson(w, val)
	})))

	http.HandleFunc("/write", graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		u, url, err := getUrl(r)
		if err != nil {
			return err
		}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return errors.NewErr(err)
		}

		val, err := gk.Write(u, url, data)
		if err != nil {
			return err
		}

		log.Printf("Wrote URL: %s: %+v\n", u, val)
		return util.SendJson(w, val)
	})))

	http.HandleFunc("/read", graceful.CreateHandler(server, util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		u, url, err := getUrl(r)
		if err != nil {
			return err
		}

		val, ok := gk.Find(url)
		if !ok {
			return util.ClientError(errors.New("Url " + u + " not found!"))
		}

		if err := gk.Read(url, val, w); err != nil {
			return err
		}

		log.Printf("Read URL: %s: %+v\n", u, val)
		return util.SendJson(w, val)
	})))

	server.SetSighup()

	if err := server.ListenAndServe(":"+strconv.Itoa(*port), *gracefulRestart); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	if err := server.Restart(); err != nil {
		log.Fatal(err)
	}
}

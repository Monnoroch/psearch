package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/log"
	"strconv"
)

func main() {
	var help = flag.Bool("help", false, "print help")
	var port = flag.Int("port", -1, "port to listen")
	flag.Parse()

	if *help || *port == -1 {
		flag.PrintDefaults()
		return
	}

	http.HandleFunc("/dl", util.CreateErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
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
	}))

	if err := http.ListenAndServe(":"+strconv.Itoa(*port), nil); err != nil {
		log.Fatal(errors.NewErr(err))
	}
}

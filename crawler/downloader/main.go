package main

import (
	"flag"
	"net/http"
	"io/ioutil"
	"log"
	"strconv"
	"psearch/util/errors"
	"psearch/util"
)

func main() {
	var port = flag.Int("port", -1, "port to listen")
	flag.Parse()

	if *port == -1 {
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

		return util.SendJson(w, map[string]string{url: string(body)})
	}))

	if err := http.ListenAndServe(":"+strconv.Itoa(*port), nil); err != nil {
		log.Fatal(err)
	}
}

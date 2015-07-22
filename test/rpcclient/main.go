package main

import (
	"encoding/json"
	"flag"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/log"
)

func main() {
	var help = flag.Bool("help", false, "print help")
	var addr = flag.String("addr", "", "address to dial")
	var method = flag.String("method", "", "rpc method")
	var parg = flag.String("arg", "", "method argument")
	flag.Parse()

	if *help || *addr == "" || *method == "" || *parg == "" {
		flag.PrintDefaults()
		return
	}

	client, err := util.JsonRpcDial(*addr)
	if err != nil {
		log.Fatal(err)
	}

	var arg interface{}
	if err := json.Unmarshal([]byte(*parg), &arg); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	var res interface{}
	if err := client.Call(*method, arg, &res); err != nil {
		log.Fatal(errors.NewErr(err))
	}

	log.Printf("%+v\n", res)
}

package util

import (
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"psearch/util/errors"
	"psearch/util/iter"
	"psearch/util/log"
	"strconv"
	"time"
)

func RndSeed() {
	rand.Seed(time.Now().Unix())
}

func SendJson(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return errors.NewErr(err)
	}
	return nil
}

func GetParam(r *http.Request, name string) (string, error) {
	sval, ok := r.Form[name]
	if !ok || len(sval) != 1 {
		return "", errors.New("No " + name + " in request params!")
	}

	return sval[0], nil
}

func GetParamOr(r *http.Request, name string, def string) (string, error) {
	sval, ok := r.Form[name]
	if !ok {
		return def, nil
	}

	if len(sval) != 1 {
		return "", errors.New("Too many " + name + " in request params!")
	}

	return sval[0], nil
}

func GetParamInt(r *http.Request, name string) (int, error) {
	sval, err := GetParam(r, name)
	if err != nil {
		return -1, err
	}

	val, err := strconv.Atoi(sval)
	if err != nil {
		return -1, errors.NewErr(err)
	}

	return val, nil
}

func GetParams(r *http.Request, name string) []string {
	sval, ok := r.Form[name]
	if !ok {
		return []string{}
	}

	return sval
}

///////////////

type ClientErrorT struct {
	Err error
}

func (self ClientErrorT) Error() string {
	return self.Err.Error()
}

func ClientError(err error) error {
	return ClientErrorT{err}
}

type ErrorHandlerT func(w http.ResponseWriter, r *http.Request) error

func CreateErrorHandler(handler ErrorHandlerT) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			log.Errorln(err)
			if _, ok := err.(ClientErrorT); ok {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

///////////////

func DoIterTsvRequest(url string, it iter.Iterator) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, iter.ReadDelim(it, []byte("\t")))
	if err != nil {
		return nil, errors.NewErr(err)
	}
	req.ContentLength = -1

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.NewErr(err)
	}

	return resp, nil
}

func JsonRpcDial(addr string) (*rpc.Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.NewErr(err)
	}

	return jsonrpc.NewClient(conn), nil
}

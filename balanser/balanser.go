package balanser

import (
	"io"
	"net/http"
	"net/url"
	"psearch/balanser/chooser"
	"psearch/util"
	"psearch/util/errors"
)

type Balanser struct {
	count  int
	router chooser.BackendChooser
}

func NewBalanser(rout chooser.BackendChooser, urls []string) *Balanser {
	return &Balanser{
		count:  len(urls),
		router: rout,
	}
}

func (self *Balanser) Request(w http.ResponseWriter, r *http.Request) ([]error, error) {
	if r.Method != "GET" {
		return nil, util.ClientError(errors.New("Can only process GET requests!"))
	}

	url := url.URL{
		Scheme:   r.URL.Scheme,
		Opaque:   r.URL.Opaque,
		User:     r.URL.User,
		Host:     r.URL.Host,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
		Fragment: r.URL.Fragment,
	}
	url.Scheme = "http"

	var resErrors []error = nil
	failed := map[string]struct{}{}
	for {
		if len(failed) == self.count {
			return resErrors, errors.New("All backends broken!")
		}

		backend := self.router.Choose(r)
		if _, ok := failed[backend]; ok {
			continue
		}

		url.Host = backend
		nreq, err := http.NewRequest("GET", url.String(), nil)
		if err != nil {
			return resErrors, errors.NewErr(err)
		}

		nreq.Header = r.Header
		resp, err := (&http.Client{}).Do(nreq)
		if err != nil {
			failed[backend] = struct{}{}
			resErrors = append(resErrors, errors.NewErr(err))
			continue
		}
		defer resp.Body.Close()

		for k, hs := range resp.Header {
			for _, val := range hs {
				w.Header().Add(k, val)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		return resErrors, errors.NewErr(err)
	}
}

func NewChooser(name string, backends []string) (chooser.BackendChooser, error) {
	if name == "random" {
		return chooser.NewRandomChooser(backends), nil
	}
	if name == "roundrobin" {
		return chooser.NewRoundRobinChooser(backends), nil
	}
	return nil, errors.New("Unknown name: \"" + name + "\".")
}

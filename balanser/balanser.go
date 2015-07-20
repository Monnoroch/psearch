package balanser

import (
	"io"
	"net/http"
	"psearch/balanser/chooser"
	"psearch/util"
	"psearch/util/errors"
)

type Balanser struct {
	Count  int
	Router chooser.BackendChooser
}

func NewBalanser(rout string, urls []string) (*Balanser, error) {
	c, err := chooser.NewChooser(rout, urls)
	if err != nil {
		return nil, err
	}

	return &Balanser{
		Count:  len(urls),
		Router: c,
	}, nil
}

func (self *Balanser) Request(w http.ResponseWriter, r *http.Request) ([]error, error) {
	if r.Method != "GET" {
		return nil, util.ClientError(errors.New("Can only process GET requests!"))
	}

	var resErrors []error = nil
	failed := map[string]struct{}{}
	for {
		if len(failed) == self.Count {
			return resErrors, errors.New("All backends broken!")
		}

		backend := self.Router.Choose()
		if _, ok := failed[backend]; ok {
			continue
		}

		r.URL.Scheme = "http"
		r.URL.Host = backend
		nreq, err := http.NewRequest("GET", r.URL.String(), nil)
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

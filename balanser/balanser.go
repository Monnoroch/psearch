package balanser

import (
	"io"
	"net/http"
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

	var resErrors []error = nil
	failed := map[string]struct{}{}
	for {
		if len(failed) == self.count {
			return resErrors, errors.New("All backends broken!")
		}

		backend := self.router.Choose()
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

package tcp

import (
	"io"
	"net"
	"psearch/balanser/chooser"
	"psearch/util/errors"
	"psearch/util/log"
)

type Balanser struct {
	count    int
	router   chooser.BackendChooser
	backends map[string]net.Conn
}

func NewBalanser(rout chooser.BackendChooser, urls []string) (*Balanser, error) {
	backends := map[string]net.Conn{}
	any := false
	for _, url := range urls {
		conn, err := net.Dial("tcp", url)
		if err != nil {
			log.Errorln(err)
		}
		any = true
		backends[url] = conn
	}
	if !any {
		return nil, errors.New("All backends are broken!")
	}

	return &Balanser{
		count:    len(urls),
		router:   rout,
		backends: backends,
	}, nil
}

func (self *Balanser) Close() error {
	var err error
	for _, v := range self.backends {
		err = v.Close()
	}
	return err
}

func (self *Balanser) Request(conn io.ReadWriteCloser) error {
	failed := map[string]struct{}{}
	for {
		if len(failed) == self.count {
			return errors.New("All backends are broken!")
		}

		backend, err := self.router.Choose()
		if err != nil {
			return err
		}

		if _, ok := failed[backend]; ok {
			continue
		}

		client, ok := self.backends[backend]
		if !ok || client == nil {
			failed[backend] = struct{}{}
			continue
		}

		if _, err := io.Copy(client, conn); err != nil {
			return errors.NewErr(err)
		}

		return nil
	}
}

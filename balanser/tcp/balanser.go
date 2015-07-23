package tcp

import (
	"io"
	"net"
	"psearch/balanser/chooser"
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

func (self *Balanser) Request(client *net.TCPConn) error {
	defer client.Close()

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

		conn, err := net.Dial("tcp", backend)
		if err != nil {
			failed[backend] = struct{}{}
			continue
		}

		server := conn.(*net.TCPConn)

		clientClosed := make(chan error, 1)
		go func() {
			_, err := io.Copy(server, client)
			clientClosed <- errors.NewErr(err)
		}()

		serverClosed := make(chan error, 1)
		go func() {
			_, err := io.Copy(client, server)
			serverClosed <- errors.NewErr(err)
		}()

		select {
		case err = <-clientClosed:
			server.SetLinger(0)
			server.CloseRead()
		case err = <-serverClosed:
		}

		return err
	}
}

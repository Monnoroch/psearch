package graceful

import (
	"flag"
	"net"
	"net/http"
	"os"
	"psearch/util/errors"
	"sync"
	"sync/atomic"
	"syscall"
)

type gracefulRestartData struct {
	stopped int32
	wg      sync.WaitGroup
}

func (self *gracefulRestartData) Stop() {
	atomic.StoreInt32(&self.stopped, 1)
}

func (self *gracefulRestartData) Stopped() bool {
	return atomic.LoadInt32(&self.stopped) == 1
}

func (self *gracefulRestartData) Inc() {
	self.wg.Add(1)
}

func (self *gracefulRestartData) Dec() {
	self.wg.Done()
}

func (self *gracefulRestartData) Wait() {
	self.wg.Wait()
}

type gracefulRestartError struct{}

func (self gracefulRestartError) Error() string {
	return "gracefulRestartError{}"
}

var gracefulRestartErrorVal gracefulRestartError

type gracefulListener struct {
	base   net.Listener
	server *Server
}

func (self *gracefulListener) Accept() (c net.Conn, err error) {
	if self.server.restart.Stopped() {
		return nil, gracefulRestartErrorVal
	}

	return self.base.Accept()
}

func (self *gracefulListener) Close() error {
	if self.server.restart.Stopped() {
		return nil
	}

	return self.base.Close()
}

func (self *gracefulListener) Addr() net.Addr {
	return self.base.Addr()
}

type Server struct {
	base     http.Server
	restart  gracefulRestartData
	listener gracefulListener
}

func NewServer(server http.Server) *Server {
	return &Server{
		base:    server,
		restart: gracefulRestartData{},
	}
}

func (self *Server) Stop() {
	self.restart.Stop()
}

func (self *Server) Restart() error {
	f, err := self.listener.base.(*net.TCPListener).File()
	if err != nil {
		return errors.NewErr(err)
	}

	newargs := make([]string, len(os.Args)+1)
	for i, v := range os.Args {
		newargs[i] = v
	}
	newargs[len(os.Args)] = "-graceful"

	_, err = syscall.ForkExec(os.Args[0], newargs, &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), f.Fd()},
	})
	if err != nil {
		return errors.NewErr(err)
	}
	return nil
}

func (self *Server) Serve(l net.Listener) error {
	self.listener = gracefulListener{
		base:   l,
		server: self,
	}

	err := self.base.Serve(&self.listener)
	self.restart.Wait()
	if err == gracefulRestartErrorVal {
		return nil
	}
	return errors.NewErr(err)
}

func getListener(addr string, graceful bool) (net.Listener, error) {
	var listener net.Listener = nil
	var err error = nil
	if graceful {
		file := os.NewFile(3, "")
		listener, err = net.FileListener(file)
		if err != nil {
			file.Close()
			return nil, errors.NewErr(err)
		}
	} else {
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return nil, errors.NewErr(err)
		}
	}
	return listener, nil
}

func (self *Server) ListenAndServe(addr string, graceful bool) error {
	l, err := getListener(addr, graceful)
	if err != nil {
		return err
	}

	return self.Serve(l)
}

func CreateHandler(server *Server, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server.restart.Inc()
		defer server.restart.Dec()
		handler(w, r)
	}
}

func SetFlag() *bool {
	return flag.Bool("graceful", false, "file descriptor for graceful restart (internal use only!)")
}

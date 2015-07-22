package graceful

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"psearch/util/errors"
	"sync"
	"sync/atomic"
	"syscall"
)

func SetFlag() *bool {
	return flag.Bool("graceful", false, "file descriptor for graceful restart (internal use only!)")
}

type GracefulRestartData struct {
	stopped int32
	wg      sync.WaitGroup
}

func (self *GracefulRestartData) Stop() {
	atomic.StoreInt32(&self.stopped, 1)
}

func (self *GracefulRestartData) Stopped() bool {
	return atomic.LoadInt32(&self.stopped) == 1
}

func (self *GracefulRestartData) Inc() {
	self.wg.Add(1)
}

func (self *GracefulRestartData) Dec() {
	self.wg.Done()
}

func (self *GracefulRestartData) Wait() {
	self.wg.Wait()
}

type gracefulRestartError struct{}

func (self gracefulRestartError) Error() string {
	return "gracefulRestartError{}"
}

var NeedRestart gracefulRestartError

type GracefulListener struct {
	base   *net.TCPListener
	server GServer
}

func NewGracefulListener(base *net.TCPListener, srv GServer) *GracefulListener {
	return &GracefulListener{
		base:   base,
		server: srv,
	}
}

func (self *GracefulListener) Accept() (c net.Conn, err error) {
	if self.server.RestartData().Stopped() {
		return nil, NeedRestart
	}

	return self.base.Accept()
}

func (self *GracefulListener) Close() error {
	if self.server.RestartData().Stopped() {
		return nil
	}

	return self.base.Close()
}

func (self *GracefulListener) Addr() net.Addr {
	return self.base.Addr()
}

func (self *GracefulListener) File() (*os.File, error) {
	f, err := self.base.File()
	if err != nil {
		return nil, errors.NewErr(err)
	}

	return f, nil
}

type GServer interface {
	RestartData() *GracefulRestartData
	Listener() *GracefulListener
}

func SetSighup(srv GServer) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)
	go func() {
		for {
			s := <-signals
			switch s {
			case syscall.SIGHUP:
				Stop(srv)
			default:
			}
		}
	}()
}

func Stop(srv GServer) {
	srv.RestartData().Stop()
}

func Restart(srv GServer) error {
	f, err := srv.Listener().File()
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

func GetListener(addr string, graceful bool) (*net.TCPListener, error) {
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
	return listener.(*net.TCPListener), nil
}

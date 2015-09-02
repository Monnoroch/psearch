package http

import (
	"net"
	"net/http"
	"psearch/util/errors"
	"psearch/util/graceful"
)

type Server struct {
	base     *http.Server
	restart  graceful.GracefulRestartData
	listener *graceful.GracefulListener
}

func NewServer(server *http.Server) *Server {
	return &Server{
		base: server,
	}
}

func (self *Server) Listener() *graceful.GracefulListener {
	return self.listener
}

func (self *Server) RestartData() *graceful.GracefulRestartData {
	return &self.restart
}

func (self *Server) Serve(l *net.TCPListener) error {
	self.listener = graceful.NewGracefulListener(l, self)
	err := self.base.Serve(self.listener)
	self.restart.Wait()
	if err == graceful.NeedRestart {
		return nil
	}
	return errors.NewErr(err)
}

func (self *Server) ListenAndServe(addr string, restart bool) error {
	l, err := graceful.GetListener(addr, restart)
	if err != nil {
		return err
	}

	return self.Serve(l)
}

func CreateHandler(server *Server, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server.restart.Inc()
		defer server.restart.Dec()
		handler.ServeHTTP(w, r)
	}
}

package rpc

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"psearch/util/errors"
	"psearch/util/graceful"
)

type Server struct {
	base     *rpc.Server
	restart  graceful.GracefulRestartData
	listener *graceful.GracefulListener
}

func NewServer(server *rpc.Server) *Server {
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

	var err error
	var conn net.Conn
	for {
		conn, err = self.listener.Accept()
		if err != nil {
			break
		}

		go self.base.ServeCodec(jsonrpc.NewServerCodec(conn))
	}

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

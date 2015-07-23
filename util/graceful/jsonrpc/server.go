package jsonrpc

import (
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
	grpc "psearch/util/graceful/rpc"
)

func NewServer(server *rpc.Server) *grpc.Server {
	return grpc.NewServer(server, func(srv *rpc.Server, conn io.ReadWriteCloser) {
		srv.ServeCodec(jsonrpc.NewServerCodec(conn))
	})
}

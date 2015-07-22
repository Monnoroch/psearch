package dns

import (
	"net/rpc"
	"psearch/util"
	"psearch/util/errors"
)

type Args struct {
	Host string `json:"host"`
}

type ArgsAll struct {
	Hosts []string `json:"hosts"`
}

type ResolverClient struct {
	*rpc.Client
}

func NewResolverClient(addr string) (ResolverClient, error) {
	c, err := util.JsonRpcDial(addr)
	if err != nil {
		return ResolverClient{}, errors.NewErr(err)
	}

	return ResolverClient{c}, nil
}

func (self *ResolverClient) Resolve(host string) ([]string, error) {
	var res []string
	if err := self.Call("ResolverServer.Resolve", Args{Host: host}, &res); err != nil {
		return nil, errors.NewErr(err)
	}

	return res, nil
}

func (self *ResolverClient) ResolveAll(hosts []string) ([][]string, error) {
	var res [][]string
	if err := self.Call("ResolverServer.ResolveAll", ArgsAll{Hosts: hosts}, &res); err != nil {
		return nil, errors.NewErr(err)
	}

	return res, nil
}

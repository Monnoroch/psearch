package gatekeeper

import (
	"net/rpc"
	"psearch/util"
	"psearch/util/errors"
)

type Value struct {
	FNum   uint   `json:"fnum"`
	Offset uint64 `json:"offset"`
	Len    uint64 `json:"len"`
}

type FindArgs struct {
	Url string `json:"url"`
}

type FindResult struct {
	Val *Value `json:"val,omitempty"`
}

type ReadResult struct {
	FindResult
	Body *string `json:"body,omitempty"`
}

type WriteArgs struct {
	FindArgs
	Body string `json:"body"`
}

type WriteResult struct {
	Val Value `json:"val"`
}

type GatekeeperClient struct {
	*rpc.Client
}

func NewGatekeeperClient(addr string) (GatekeeperClient, error) {
	c, err := util.JsonRpcDial(addr)
	if err != nil {
		return GatekeeperClient{}, errors.NewErr(err)
	}

	return GatekeeperClient{c}, nil
}

func (self *GatekeeperClient) Find(url string) (Value, bool, error) {
	var res FindResult
	if err := self.Call("GatekeeperServer.Find", FindArgs{Url: url}, &res); err != nil {
		return Value{}, false, errors.NewErr(err)
	}

	if res.Val == nil {
		return Value{}, false, nil
	}
	return *res.Val, true, nil
}

func (self *GatekeeperClient) Read(url string) (Value, bool, string, error) {
	var res ReadResult
	if err := self.Call("GatekeeperServer.Read", FindArgs{Url: url}, &res); err != nil {
		return Value{}, false, "", errors.NewErr(err)
	}

	if res.Val == nil {
		return Value{}, false, "", nil
	}
	return *res.Val, true, *res.Body, nil
}

func (self *GatekeeperClient) Write(url string, body string) (Value, error) {
	var res WriteResult
	if err := self.Call("GatekeeperServer.Write", WriteArgs{FindArgs{Url: url}, body}, &res); err != nil {
		return Value{}, errors.NewErr(err)
	}

	return res.Val, nil
}

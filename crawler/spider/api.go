package spider

import (
	"net/rpc"
	"psearch/util"
	"psearch/util/errors"
)

type Args struct {
	Url string `json:"url"`
}

type ArgsAll struct {
	Urls []string `json:"urls"`
}

type SpiderClient struct {
	*rpc.Client
}

func NewSpiderClient(addr string) (SpiderClient, error) {
	c, err := util.JsonRpcDial(addr)
	if err != nil {
		return SpiderClient{}, errors.NewErr(err)
	}

	return SpiderClient{c}, nil
}

func (self *SpiderClient) AddUrl(url string) error {
	return errors.NewErr(self.Call("SpiderServer.AddUrl", Args{Url: url}, &struct{}{}))
}

func (self *SpiderClient) AddUrls(urls []string) error {
	return errors.NewErr(self.Call("SpiderServer.AddUrls", ArgsAll{Urls: urls}, &struct{}{}))
}

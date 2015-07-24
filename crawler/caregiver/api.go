package caregiver

import (
	"net/rpc"
	"psearch/util"
	"psearch/util/errors"
)

type Args struct {
	Urls []string `json:"urls"`
}

type CaregiverClient struct {
	*rpc.Client
}

func NewCaregiverClient(addr string) (CaregiverClient, error) {
	c, err := util.JsonRpcDial(addr)
	if err != nil {
		return CaregiverClient{}, errors.NewErr(err)
	}

	return CaregiverClient{c}, nil
}

func (self *CaregiverClient) PushUrls(urls []string) error {
	var res struct{}
	return errors.NewErr(self.Call("CaregiverServer.PushUrls", Args{Urls: urls}, &res))
}

func (self *CaregiverClient) PullUrls() (map[string]string, error) {
	var res map[string]string
	if err := self.Call("CaregiverServer.PullUrls", struct{}{}, &res); err != nil {
		return nil, errors.NewErr(err)
	}

	return res, nil
}

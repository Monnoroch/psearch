package downloader

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

type DownloaderClient struct {
	*rpc.Client
}

func NewDownloaderClient(addr string) (DownloaderClient, error) {
	c, err := util.JsonRpcDial(addr)
	if err != nil {
		return DownloaderClient{}, errors.NewErr(err)
	}

	return DownloaderClient{c}, nil
}

func (self *DownloaderClient) Download(url string) (string, error) {
	var res string
	if err := self.Call("DownloaderServer.Download", Args{Url: url}, &res); err != nil {
		return "", errors.NewErr(err)
	}

	return string(res), nil
}

func (self *DownloaderClient) DownloadAll(urls []string) ([]string, error) {
	var res []string
	if err := self.Call("DownloaderServer.DownloadAll", ArgsAll{Urls: urls}, &res); err != nil {
		return nil, errors.NewErr(err)
	}

	return res, nil
}

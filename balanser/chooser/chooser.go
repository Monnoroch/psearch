package chooser

import (
	"math/rand"
	"net/http"
	"hash"
	"psearch/util"
)

type BackendChooser interface {
	Choose(req *http.Request) (string, error)
}

type randomChooser struct {
	backends []string
}

func (self *randomChooser) Choose(req *http.Request) (string, error) {
	return self.backends[rand.Int31n(int32(len(self.backends)))], nil
}

func NewRandomChooser(backends []string) BackendChooser {
	return &randomChooser{
		backends: backends,
	}
}

type roundRobinChooser struct {
	backends []string
	num      uint
}

func (self *roundRobinChooser) Choose(req *http.Request) (string, error) {
	res := self.backends[self.num]
	self.num += 1
	if self.num == uint(len(self.backends)) {
		self.num = 0
	}
	return res, nil
}

func NewRoundRobinChooser(backends []string) BackendChooser {
	return &roundRobinChooser{
		backends: backends,
		num:      0,
	}
}

type paramHashChooser struct {
	backends []string
	param      string
	hash hash.Hash32
}

func (self *paramHashChooser) Choose(req *http.Request) (string, error) {
	p, err := util.GetParamOr(req, self.param, "")
	if err != nil {
		return "", err
	}

	self.hash.Reset()
	self.hash.Write([]byte(p))
	return self.backends[self.hash.Sum32() % uint32(len(self.backends))], nil
}

func NewParamHashChooser(backends []string, hash hash.Hash32, param string) BackendChooser {
	return &paramHashChooser{
		backends: backends,
		hash: hash,
		param:      param,
	}
}

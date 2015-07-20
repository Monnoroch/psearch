package chooser

import (
	"math/rand"
	"net/http"
	"psearch/util/errors"
)

type BackendChooser interface {
	Choose(req *http.Request) string
}

type randomChooser struct {
	backends []string
}

func (self *randomChooser) Choose(req *http.Request) string {
	return self.backends[rand.Int31n(int32(len(self.backends)))]
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

func (self *roundRobinChooser) Choose(req *http.Request) string {
	res := self.backends[self.num]
	self.num += 1
	if self.num == uint(len(self.backends)) {
		self.num = 0
	}
	return res
}

func NewRoundRobinChooser(backends []string) BackendChooser {
	return &roundRobinChooser{
		backends: backends,
		num:      0,
	}
}

func NewChooser(name string, backends []string) (BackendChooser, error) {
	if name == "random" {
		return NewRandomChooser(backends), nil
	}
	if name == "roundrobin" {
		return NewRoundRobinChooser(backends), nil
	}
	return nil, errors.New("Unknown name: \"" + name + "\".")
}

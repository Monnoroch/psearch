package chooser

import (
	"math/rand"
)

type BackendChooser interface {
	Choose() (string, error)
}

type randomChooser struct {
	backends []string
}

func (self *randomChooser) Choose() (string, error) {
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

func (self *roundRobinChooser) Choose() (string, error) {
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

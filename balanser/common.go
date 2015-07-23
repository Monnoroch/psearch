package balanser

import (
	"psearch/balanser/chooser"
	"psearch/util/errors"
)

func NewChooser(name string, backends []string) (chooser.BackendChooser, error) {
	if name == "random" {
		return chooser.NewRandomChooser(backends), nil
	}
	if name == "roundrobin" {
		return chooser.NewRoundRobinChooser(backends), nil
	}
	return nil, errors.New("Unknown name: \"" + name + "\".")
}

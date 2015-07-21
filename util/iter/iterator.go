package iter

import (
	"bufio"
	"io"
)

type EndIteration struct{}

func (self EndIteration) Error() string {
	return "EndIteration{}"
}

var EOI = EndIteration{}

type Iterator interface {
	Next() (string, error)
}

func Collect(it Iterator) ([]string, error) {
	res := []string{}
	for v, err := it.Next(); err != EOI; v, err = it.Next() {
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

type MapIterator struct {
	it Iterator
	f func(string) (string, error)
}

func (self MapIterator) Next() (string, error) {
	res, err := self.it.Next()
	if err != nil {
		return "", err
	}

	return self.f(res)
}

func Map(it Iterator, f func(string) (string, error)) Iterator {
	return MapIterator{
		it: it,
		f: f,
	}
}

type ChainIterator struct {
	its  []Iterator
	curr int
}

func (self ChainIterator) Next() (string, error) {
	if len(self.its) == self.curr {
		return "", EOI
	}

	res, err := self.its[self.curr].Next()
	if err == EOI {
		self.curr += 1
		return self.Next()
	}
	return res, err
}

func Chain(its ...Iterator) Iterator {
	return ChainIterator{
		its:  its,
		curr: 0,
	}
}


type ArrayIterator struct {
	arr  []string
	curr int
}

func (self ArrayIterator) Next() (string, error) {
	if len(self.arr) == self.curr {
		return "", EOI
	}

	self.curr += 1
	return self.arr[self.curr-1], nil
}

func Array(arr []string) Iterator {
	return ArrayIterator{
		arr:  arr,
		curr: 0,
	}
}

type DelimIterator struct {
	delim  byte
	reader *bufio.Reader
}

func (self DelimIterator) Next() (string, error) {
	next, err := self.reader.ReadString(self.delim)
	if err == io.EOF {
		return "", EOI
	}
	return next, err
}

func Delim(r io.Reader, delim byte) Iterator {
	return DelimIterator{
		delim:  delim,
		reader: bufio.NewReader(r),
	}
}

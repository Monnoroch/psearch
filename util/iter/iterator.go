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

func JoinDelim(it Iterator, delim string) (string, error) {
	res := ""
	first := true
	for v, err := it.Next(); err != EOI; v, err = it.Next() {
		if err != nil {
			return "", err
		}

		if first {
			first = false
		} else {
			res += delim
		}
		res += v
	}
	return res, nil
}

func Count(it Iterator) (uint, error) {
	res := uint(0)
	for _, err := it.Next(); err != EOI; _, err = it.Next() {
		if err != nil {
			return 0, err
		}
		res += 1
	}
	return res, nil
}

type MapIterator struct {
	it Iterator
	f  func(string) (string, error)
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
		f:  f,
	}
}

func Do(it Iterator, f func(string) error) error {
	for v, err := it.Next(); err != EOI; v, err = it.Next() {
		if err != nil {
			return err
		}
		if err := f(v); err != nil {
			return err
		}
	}
	return nil
}

func DoEnum(it Iterator, f func(uint, string) error) error {
	idx := uint(0)
	for v, err := it.Next(); err != EOI; v, err = it.Next() {
		if err != nil {
			return err
		}
		if err := f(idx, v); err != nil {
			return err
		}
		idx += 1
	}
	return nil
}

func WriteDelim(it Iterator, w io.Writer, delim []byte) error {
	return DoEnum(it, func(i uint, v string) error {
		if i != 0 {
			_, err := w.Write(delim)
			if err != nil {
				return err
			}
		}
		_, err := w.Write([]byte(v))
		if err != nil {
			return err
		}
		return nil
	})
}

type IteratorReader struct {
	it           Iterator
	val          string
	written      uint
	delim        []byte
	writtenDelim uint
}

func (self *IteratorReader) Read(p []byte) (int, error) {
	if self.written == uint(len(self.val)) {
		if self.writtenDelim == uint(len(self.delim)) {
			v, err := self.it.Next()
			if err == EOI {
				return 0, io.EOF
			}
			if err != nil {
				return 0, err
			}

			self.val = v
			self.written = 0
			self.writtenDelim = 0
			return self.Read(p)
		}

		needWrite := uint(len(p))
		cnt := uint(len(self.delim))
		if needWrite <= cnt {
			copy(p, self.delim[:needWrite])
			self.writtenDelim += needWrite
			return int(needWrite), nil
		} else {
			copy(p[:cnt], self.delim)
			self.writtenDelim += cnt
			n, err := self.Read(p[cnt:])
			if err != nil {
				return 0, err
			}
			return int(cnt) + n, nil
		}
	}

	needWrite := uint(len(p))
	cnt := uint(len(self.val))
	if needWrite <= cnt {
		copy(p, self.val[:needWrite])
		self.writtenDelim += needWrite
		return int(needWrite), nil
	} else {
		copy(p[:cnt], self.val)
		self.writtenDelim += cnt
		n, err := self.Read(p[cnt:])
		if err != nil {
			return 0, err
		}
		return int(cnt) + n, nil
	}
}

func ReadDelim(it Iterator, delim []byte) io.Reader {
	return &IteratorReader{
		it:           it,
		delim:        delim,
		written:      0,
		writtenDelim: 0,
	}
}

type ChainIterator struct {
	its  []Iterator
	curr int
}

func (self *ChainIterator) Next() (string, error) {
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
	return &ChainIterator{
		its:  its,
		curr: 0,
	}
}

type ArrayIterator struct {
	arr  []string
	curr int
}

func (self *ArrayIterator) Next() (string, error) {
	if len(self.arr) == self.curr {
		return "", EOI
	}

	self.curr += 1
	return self.arr[self.curr-1], nil
}

func Array(arr []string) Iterator {
	return &ArrayIterator{
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
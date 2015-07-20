package util

import (
	"encoding/binary"
	"io"
	"os"
	"psearch/util/errors"
	"strconv"
)

type FileReader struct {
	file   *os.File
	offset uint64
	buf    []byte
}

func Open(file string) (*FileReader, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.NewErr(err)
	}

	return &FileReader{
		file:   f,
		offset: 0,
		buf:    []byte{0},
	}, nil
}

func (self *FileReader) Close() error {
	return errors.NewErr(self.file.Close())
}

func (self *FileReader) Seek(offset int64, whence int) (ret int64, err error) {
	ret, err = self.file.Seek(offset, whence)
	if err == io.EOF {
		return ret, err
	}

	return ret, errors.NewErr(err)
}

func (self *FileReader) Read(b []byte) (n int, err error) {
	n, err = self.file.Read(b)
	if err == io.EOF {
		return n, err
	}

	return n, errors.NewErr(err)
}

func (self *FileReader) ReadByte() (byte, error) {
	n, err := self.file.Read(self.buf)
	if err == io.EOF {
		return 0, err
	}
	if err != nil {
		return 0, errors.NewErr(err)
	}

	if n != 1 {
		return 0, errors.New("Read " + strconv.Itoa(n) + " bytes, instead of one!")
	}

	self.offset += 1
	return self.buf[0], nil
}

func (self *FileReader) UnreadByte() error {
	if self.offset == 0 {
		return nil
	}

	if _, err := self.Seek(int64(self.offset-1), 0); err != nil {
		return err
	}
	self.offset -= 1
	return nil
}

func (self *FileReader) Write(b []byte) (n int, err error) {
	n, err = self.file.Write(b)
	if err == io.EOF {
		return n, err
	}
	return n, errors.NewErr(err)
}

func (self *FileReader) ReadLenval() (uint64, []byte, error) {
	start := self.offset
	l, err := binary.ReadUvarint(self)
	diff := self.offset - start
	if err == io.EOF {
		return diff, nil, err
	}
	if err != nil {
		return diff, nil, errors.NewErr(err)
	}

	res := make([]byte, l)
	n, err := self.Read(res)
	diff += uint64(n)
	if err != nil {
		return diff, nil, err
	}
	if n != int(l) {
		return diff, nil, errors.New("Read " + strconv.Itoa(n) + " bytes, instead of " + strconv.Itoa(int(l)) + "!")
	}

	return diff, res, nil
}

func (self *FileReader) SkipLenval() (uint64, error) {
	start := self.offset
	l, err := binary.ReadUvarint(self)
	diff := self.offset - start
	if err == io.EOF {
		return diff, err
	}
	if err != nil {
		return diff, errors.NewErr(err)
	}

	if _, err := self.Seek(int64(l), 1); err != nil {
		return diff, err
	}

	return l + diff, nil
}

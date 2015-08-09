package gatekeeper

import (
	"encoding/binary"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"psearch/util"
	"psearch/util/errors"
	"psearch/util/log"
	"psearch/util/trie"
	"sort"
	"strconv"
	"strings"
	"time"
)

func UrlTransform(key string) (string, error) {
	u, err := url.Parse(string(key))
	if err != nil {
		return "", errors.NewErr(err)
	}

	arr := strings.Split(u.Host, ".")
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}

	u.Host = strings.Join(arr, ".")
	return u.String(), nil
}

type gkFile struct {
	file   *os.File
	offset uint64
	end    time.Time
}

func (self *gkFile) WriteLenval(b []byte) (n int, err error) {
	buf := make([]byte, 8)
	num := binary.PutUvarint(buf, uint64(len(b)))
	n1, err := self.file.Write(buf[:num])
	if err != nil {
		return n1, errors.NewErr(err)
	}

	n2, err := self.file.Write(b)
	return n1 + n2, errors.NewErr(err)
}

func (self *gkFile) Close() error {
	return errors.NewErr(self.file.Close())
}

type Gatekeeper struct {
	dir         string
	maxTime     time.Duration
	maxFileSize uint64
	fNum        uint
	file        gkFile
	trie        trie.Trie
}

type valT struct {
	num  uint
	file os.FileInfo
}

type valTArr []valT

func (self valTArr) Len() int {
	return len(self)
}

func (self valTArr) Less(i, j int) bool {
	return self[i].num < self[j].num
}

func (self valTArr) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func NewGatekeeper(dir string, maxFileSize uint64, maxTime time.Duration) (*Gatekeeper, error) {
	self := &Gatekeeper{
		dir:         dir,
		maxTime:     maxTime,
		maxFileSize: maxFileSize,
		fNum:        0,
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.NewErr(err)
	}

	arr := make(valTArr, len(files))
	for i, f := range files {
		num, err := strconv.Atoi(f.Name())
		if err != nil {
			return nil, errors.NewErr(err)
		}

		arr[i] = valT{
			num:  uint(num),
			file: f,
		}
	}

	sort.Sort(arr)
	counts := map[uint]uint{}
	for _, f := range arr {
		counts[f.num] = 0
		self.fNum = f.num + 1
		err := self.load(self.dir+"/"+f.file.Name(), f.num, counts)
		if err != nil {
			return nil, err
		}
	}

	for k, v := range counts {
		if v == 0 {
			if err := os.Remove(self.dir + "/" + strconv.Itoa(int(k))); err != nil {
				return nil, errors.NewErr(err)
			}
		}
	}

	return self, nil
}

func (self *Gatekeeper) load(name string, num uint, counts map[uint]uint) error {
	log.Printf("Gatekeeper.load(%v, %v)\n", name, num)
	file, err := util.Open(name)
	if err != nil {
		return err
	}

	cnt := uint64(0)
	for {
		n1, key, err := file.ReadLenval()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		u, err := UrlTransform(string(key))
		if err != nil {
			return err
		}

		n2, err := file.SkipLenval()
		if err != nil {
			return err
		}

		nv := Value{
			FNum:   num,
			Offset: cnt,
			Len:    n1 + n2,
		}
		old := self.trie.Add([]byte(u), nv)
		if old != nil {
			old := old.(Value)
			if old.Len != 0 {
				counts[old.FNum] -= 1
			}
		}

		cnt += n1 + n2
		counts[num] += 1
	}
	return nil
}

func (self *Gatekeeper) Close() error {
	return self.file.Close()
}

func (self *Gatekeeper) nextFile() error {
	if err := self.file.Close(); err != nil {
		return err
	}

	self.fNum += 1
	f, err := os.Create(self.dir + "/" + strconv.Itoa(int(self.fNum)))
	if err != nil {
		return errors.NewErr(err)
	}

	self.file = gkFile{
		file:   f,
		offset: 0,
		end:    time.Now().Add(self.maxTime),
	}
	return nil
}

func (self *Gatekeeper) Write(url, key string, data []byte) (Value, error) {
	log.Printf("Gatekeeper.Write(%v, %v)\n", url, key)

	if self.file.file == nil {
		f, err := os.Create(self.dir + "/" + strconv.Itoa(int(self.fNum)))
		if err != nil {
			return Value{}, errors.NewErr(err)
		}

		self.file = gkFile{
			file:   f,
			offset: 0,
			end:    time.Now().Add(self.maxTime),
		}
	} else if self.file.offset >= self.maxFileSize {
		if err := self.nextFile(); err != nil {
			return Value{}, err
		}
	} else if self.file.end.Before(time.Now()) {
		if err := self.file.file.Sync(); err != nil {
			return Value{}, errors.NewErr(err)
		}
		self.file.end = time.Now().Add(self.maxTime)
	}

	offset := self.file.offset

	cnt := 0
	n, err := self.file.WriteLenval([]byte(url))
	if err != nil {
		return Value{}, err
	}
	cnt += n

	n, err = self.file.WriteLenval(data)
	if err != nil {
		return Value{}, err
	}
	cnt += n

	self.file.offset += uint64(cnt)

	res := Value{
		FNum:   self.fNum,
		Offset: offset,
		Len:    uint64(cnt),
	}

	self.trie.Add([]byte(key), res)

	// TODO: remove
	self.file.file.Sync()

	log.Printf("Gatekeeper.Write(%v, %v) OK (%+v)\n", url, key, res)
	return res, nil
}

func (self *Gatekeeper) Read(val Value) (string, error) {
	log.Printf("Gatekeeper.Read(%+v)\n", val)
	f, err := util.Open(self.dir + "/" + strconv.Itoa(int(val.FNum)))
	if err != nil {
		return "", err
	}

	if _, err := f.Seek(int64(val.Offset), 0); err != nil {
		return "", err
	}

	if _, err := f.SkipLenval(); err != nil {
		return "", err
	}

	_, res, err := f.ReadLenval()
	if err != nil {
		return "", err
	}

	log.Printf("Gatekeeper.Read(%+v) OK\n", val)
	return string(res), nil
}

func (self *Gatekeeper) Find(key string) (Value, bool) {
	log.Printf("Gatekeeper.Find(%+v)\n", key)
	res, ok := self.trie.Find([]byte(key))
	log.Printf("Gatekeeper.Find(%+v) OK (%v, %v)\n", key, res, ok)
	if !ok {
		return Value{}, false
	}

	return res.(Value), true
}

func (self *Gatekeeper) TrieSize() uint {
	return self.trie.Count
}

type GatekeeperServer struct {
	Gatekeeper *Gatekeeper
}

func (self *GatekeeperServer) Find(args *FindArgs, result *FindResult) error {
	key, err := UrlTransform(args.Url)
	if err != nil {
		log.Errorln(err, args)
		return err
	}

	if r, ok := self.Gatekeeper.Find(key); ok {
		*result = FindResult{
			Val: &r,
		}
	}
	return nil
}

func (self *GatekeeperServer) Read(args *FindArgs, result *ReadResult) error {
	key, err := UrlTransform(args.Url)
	if err != nil {
		log.Errorln(err, args)
		return err
	}

	r, ok := self.Gatekeeper.Find(key)
	if !ok {
		return nil
	}

	data, err := self.Gatekeeper.Read(r)
	if err != nil {
		log.Errorln(err, args)
		return err
	}

	*result = ReadResult{FindResult{Val: &r}, &data}
	return nil
}

func (self *GatekeeperServer) Write(args *WriteArgs, result *FindResult) error {
	key, err := UrlTransform(args.Url)
	if err != nil {
		log.Errorln(err, args)
		return err
	}

	r, err := self.Gatekeeper.Write(args.Url, key, []byte(args.Body))
	if err != nil {
		log.Errorln(err, args)
		return err
	}

	*result = FindResult{
		Val: &r,
	}
	return nil
}

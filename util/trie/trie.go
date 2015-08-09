package trie

type nodeT struct {
	val  interface{}
	next map[byte]*nodeT
}

func (self *nodeT) Add(key []byte, val interface{}, n int) (interface{}, uint) {
	if n == len(key) {
		res := self.val
		self.val = val
		return res, 0
	}

	b := key[n]
	if self.next == nil {
		t := nodeT{}
		res, cnt := t.Add(key, val, n+1)
		self.next = map[byte]*nodeT{
			b: &t,
		}
		return res, cnt + 1
	}

	t, ok := self.next[b]
	if ok {
		return t.Add(key, val, n+1)
	} else {
		t := nodeT{}
		res, cnt := t.Add(key, val, n+1)
		self.next[b] = &t
		return res, cnt + 1
	}

}

func (self *nodeT) Find(key []byte, n int) (interface{}, bool) {
	if n == len(key) {
		if self.val != nil {
			return self.val, true
		} else {
			return nil, false
		}
	}

	if self.next == nil {
		return nil, false
	}

	b := key[n]
	t, ok := self.next[b]
	if !ok {
		return nil, false
	}

	return t.Find(key, n+1)
}

type Trie struct {
	root  nodeT
	Count uint
}

func (self *Trie) Add(key []byte, val interface{}) interface{} {
	res, cnt := self.root.Add(key, val, 0)
	self.Count += cnt
	return res
}

func (self *Trie) Find(key []byte) (interface{}, bool) {
	return self.root.Find(key, 0)
}

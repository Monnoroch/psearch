package trie

type Value struct {
	FNum   uint   "json:`fnum`"
	Offset uint64 "json:`offset`"
	Len    uint64 "json:`len`"
}

type nodeT struct {
	val  Value
	next map[byte]*nodeT
}

func (self *nodeT) Add(key string, val Value, n int) (Value, int) {
	if n == len(key) {
		res := self.val
		self.val = val
		return res, 0
	}

	b := byte(key[n])
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

func (self *nodeT) Find(key string, n int) (Value, bool) {
	if n == len(key) {
		if self.val.Len != 0 {
			return self.val, true
		} else {
			return Value{}, false
		}
	}

	if self.next == nil {
		return Value{}, false
	}

	b := byte(key[n])
	t, ok := self.next[b]
	if !ok {
		return Value{}, false
	}

	return t.Find(key, n+1)
}

type Trie struct {
	root  nodeT
	Count int
}

func (self *Trie) Add(key string, val Value) Value {
	res, cnt := self.root.Add(key, val, 0)
	self.Count += cnt
	return res
}

func (self *Trie) Find(key string) (Value, bool) {
	return self.root.Find(key, 0)
}

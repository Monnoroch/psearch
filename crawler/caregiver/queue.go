package caregiver

import "sync"

type Queue struct {
	array []string
	start uint
}

func NewQueue(cap uint) Queue {
	return Queue{
		array: make([]string, 0, cap),
		start: 0,
	}
}

func (self *Queue) Len() uint {
	return uint(len(self.array)) - self.start
}

func (self *Queue) Enqueue(val string) {
	self.array = append(self.array, val)
	// total := len(self.array) + 1
	// if cap(self.array) < total {
	// 	self.realloc(1)
	// }
	// self.array = self.array[:total]
	// self.array[total-1] = val
}

func (self *Queue) EnqueueAll(vals ...string) {
	for _, v := range vals {
		self.Enqueue(v)
	}
	// n := len(vals)
	// lastLen := self.Len()
	// total := len(self.array) + n
	// if cap(self.array) < total {
	// 	self.realloc(uint(n))
	// }
	// self.array = self.array[:self.Len()+uint(n)]
	// copy(self.array[lastLen:], vals)
}

func (self *Queue) TryDequeue() (string, bool) {
	if self.Len() == 0 {
		return "", false
	}
	self.start += 1
	return self.array[self.start-1], true
}

func (self *Queue) DequeueN(n uint) []string {
	minLen := self.Len()
	if n < minLen {
		minLen = n
	}
	if minLen == 0 {
		return nil
	}

	res := make([]string, minLen)
	copy(res, self.array[self.start:self.start+minLen])
	self.start += minLen
	return res
}

func (self *Queue) realloc(need uint) {
	if self.start >= need {
		cnt := copy(self.array, self.array[self.start:])
		self.array = self.array[:cnt]
		self.start = 0
		return
	}

	l := self.Len()
	arr := make([]string, l, ((l+need)*3)/2)
	copy(arr, self.array[self.start:])
	self.array = arr
	self.start = 0
}

type LockedQueue struct {
	queue Queue
	mutex sync.Mutex
}

func (self *LockedQueue) Len() uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.queue.Len()
}

func (self *LockedQueue) Enqueue(val string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.queue.Enqueue(val)
}

func (self *LockedQueue) EnqueueAll(vals ...string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.queue.EnqueueAll(vals...)
}

func (self *LockedQueue) TryDequeue() (string, bool) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.queue.TryDequeue()
}

func (self *LockedQueue) DequeueN(n uint) []string {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.queue.DequeueN(n)
}

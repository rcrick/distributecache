package lru

import "container/list"

type Lru struct {
	maxBytes  int64
	nBytes    int64
	l         *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

func NewLru(maxBytes int64, onEvicted func(key string, value Value)) *Lru {
	return &Lru{
		maxBytes:  maxBytes,
		l:         new(list.List),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (lru *Lru) Get(key string) (value Value, ok bool) {
	if ele, ok := lru.cache[key]; ok {
		lru.l.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (lru *Lru) RemoveOldest() {
	ele := lru.l.Back()
	if ele != nil {
		kv := ele.Value.(*entry)
		lru.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		lru.l.Remove(ele)
		delete(lru.cache, kv.key)
		if lru.OnEvicted != nil {
			lru.OnEvicted(kv.key, kv.value)
		}
	}
}

func (lru *Lru) Add(key string, value Value) {
	if ele, ok := lru.cache[key]; ok {
		kv := ele.Value.(*entry)
		lru.nBytes += int64(value.Len() - kv.value.Len())
		kv.value = value
	} else {
		ele := lru.l.PushFront(&entry{key: key, value: value})
		lru.cache[key] = ele
		lru.nBytes += int64(len(key)) + int64(value.Len())
	}
	for lru.maxBytes != 0 && lru.maxBytes < lru.nBytes {
		lru.RemoveOldest()
	}
}

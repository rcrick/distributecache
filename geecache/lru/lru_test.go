// Package lru provides ...
package lru

import (
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}
func Test_Get(t *testing.T) {
	lru := NewLru(int64(0), nil)
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}

}

func Test_RemoveOldeest(t *testing.T) {
	lru := NewLru(int64(10), nil)
	lru.Add("key1", String("1"))
	lru.Add("key2", String("1"))
	lru.Add("k", String("1"))

	if _, ok := lru.Get("key1"); ok {
		t.Fatalf("RemoveOldest key1 failed")
	}

}

func Test_OnEvicted(t *testing.T) {
	keys := make([]string, 0)
	lru := NewLru(10, func(key string, value Value) {
		keys = append(keys, key)
	})
	lru.Add("key1", String("11"))
	lru.Add("key2", String("11"))
	lru.Add("key3", String("11"))
	lru.Add("key3", String("11"))

	expect := []string{"key1", "key2"}

	if !reflect.DeepEqual(keys, expect) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}

}

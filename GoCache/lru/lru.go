package lru

import "container/list"

type Cache struct {
	maxBytes  int64
	nbytes    int64
	ll        *list.List
	cache     map[interface{}]*list.Element
	onEvicted func(key string, value Value)
}
type entry struct { // list 中存储的数据
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[interface{}]*list.Element),
		onEvicted: onEvicted,
	}
}
func (c *Cache) Get(key string) (value Value, ok bool) {
	if e, ok := c.cache[key]; ok {
		c.ll.MoveToFront(e)
		kv := e.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}
func (c *Cache) Remove() {
	element := c.ll.Back()
	if element == nil {
		return
	}
	c.removeElement(element)
}

func (c *Cache) removeElement(element *list.Element) {
	c.ll.Remove(element)
	kv := element.Value.(*entry)
	delete(c.cache, kv.key)
	c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
	if c.onEvicted != nil {
		c.onEvicted(kv.key, kv.value)
	}
}
func (c *Cache) Add(key string, value Value) {
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element)
		kv := element.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		c.cache[key] = c.ll.PushFront(&entry{key, value})
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.Remove()
	}
}
func (c *Cache) RemoveKey(key string) {
	if element, ok := c.cache[key]; ok {
		c.removeElement(element)
	}
}
func (c *Cache) Len() int {
	return c.ll.Len()
}

func (c *Cache) Bytes() int64 {
	return c.nbytes
}

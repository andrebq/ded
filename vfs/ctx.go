package vfs

import (
	"fmt"
)

func (c *Context) Put(k, v interface{}) {
	c.data[k] = v
}

func (c *Context) Get(k interface{}) (interface{}, bool) {
	val, ok := c.data[k]
	return val, ok
}

func (c *Context) MustGet(k interface{}) interface{} {
	val, ok := c.Get(k)
	if !ok {
		panic(fmt.Errorf("Key %v not found", k))
	}
	return val
}

func (c *Context) Delete(k interface{}) (interface{}, bool) {
	val, ok := c.Get(k)
	if ok {
		delete(c.data, k)
	}
	return val, ok
}

func (c *Context) Clear() {
	c.data = make(map[interface{}]interface{})
}

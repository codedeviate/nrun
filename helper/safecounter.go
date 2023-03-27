package helper

import "sync"

type SafeCounter struct {
	v   int
	mux sync.Mutex
}

func (c *SafeCounter) Inc() {
	c.mux.Lock()
	c.v++
	c.mux.Unlock()
}

func (c *SafeCounter) Dec() {
	c.mux.Lock()
	c.v--
	c.mux.Unlock()
}

func (c *SafeCounter) Value() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.v
}

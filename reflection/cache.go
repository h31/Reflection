package main

import (
	"sync"
	"time"
)

type Cache struct {
	Values *JsonMap
	FilledAt time.Time
	Valid bool
	lock sync.Mutex
}

func (c *Cache) invalidate() {
	c.Valid = false
}

func (c *Cache) isStillValid(timeout time.Duration) (bool, *JsonMap) {
	if c.Valid && time.Since(c.FilledAt) < timeout {
		return true, c.Values
	} else {
		c.Valid = false
		return false, nil
	}
}

func (c *Cache) fill(data *JsonMap) {
	c.Values = data
	c.FilledAt = time.Now()
	c.Valid = true
}

func (c *Cache) GetOrFill(dest JsonMap, fillFunc func(dest JsonMap)) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ok, values := propsCache.isStillValid(15 * time.Second); ok {
		dest.addAll(*values)
	} else {
		newValues := make(JsonMap)
		fillFunc(newValues)
		dest.addAll(newValues)
		propsCache.fill(&newValues)
	}
}
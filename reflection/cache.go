package main

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Cache struct {
	Timeout  time.Duration
	Values   map[string]JsonMap
	FilledAt map[string]time.Time
	lock     sync.Mutex
}

func (c *Cache) isStillValid(hash string, timeout time.Duration) (bool, JsonMap) {
	value, hasValue := c.Values[hash]
	filledAt, hasFillTime := c.FilledAt[hash]
	if hasValue && hasFillTime && time.Since(filledAt) < timeout {
		return true, value
	} else {
		return false, nil
	}
}

func (c *Cache) fill(hash string, data JsonMap) {
	if c.Values == nil {
		c.Values = make(map[string]JsonMap)
	}
	c.Values[hash] = data

	if c.FilledAt == nil {
		c.FilledAt = make(map[string]time.Time)
	}
	c.FilledAt[hash] = time.Now()
}

func (c *Cache) GetOrFill(hash string, dest JsonMap, fillFunc func(dest JsonMap)) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ok, values := c.isStillValid(hash, c.Timeout); ok {
		log.WithField("hash", hash).Debug("Got info from cache")
		dest.addAll(values)
	} else {
		log.WithField("hash", hash).Debug("Executing callback to fill the cache")
		newValues := make(JsonMap)
		fillFunc(newValues)
		dest.addAll(newValues)
		c.fill(hash, newValues)
	}
}

func (m JsonMap) addAll(source JsonMap) {
	for key, value := range source {
		m[key] = value
	}
}

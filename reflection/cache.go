package main

import (
	"github.com/h31/Reflection/qBT"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Cache struct {
	Timeout  time.Duration
	Values   map[qBT.Hash]JsonMap
	FilledAt map[qBT.Hash]time.Time
	lock     sync.Mutex
}

func (c *Cache) isStillValid(hash qBT.Hash, timeout time.Duration) (bool, JsonMap) {
	value, hasValue := c.Values[hash]
	filledAt, hasFillTime := c.FilledAt[hash]
	if hasValue && hasFillTime && time.Since(filledAt) < timeout {
		return true, value
	} else {
		return false, nil
	}
}

func (c *Cache) fill(hash qBT.Hash, data JsonMap) {
	if c.Values == nil {
		c.Values = make(map[qBT.Hash]JsonMap)
	}
	c.Values[hash] = data

	if c.FilledAt == nil {
		c.FilledAt = make(map[qBT.Hash]time.Time)
	}
	c.FilledAt[hash] = time.Now()
}

func (c *Cache) GetOrFill(hash qBT.Hash, dest JsonMap, cacheAllowed bool, fillFunc func(dest JsonMap)) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if isCached, values := c.isStillValid(hash, c.Timeout); cacheAllowed && isCached {
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

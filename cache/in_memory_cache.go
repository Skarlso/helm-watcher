package cache

import "sync"

type HelmCache struct {
	cache map[string]string
	mut   *sync.RWMutex
}

func NewCache() *HelmCache {
	return &HelmCache{
		cache: make(map[string]string),
		mut:   &sync.RWMutex{},
	}
}

func (c *HelmCache) Add(key, value string) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.cache[key] = value
}

func (c *HelmCache) Remove(key string) {
	c.mut.Lock()
	defer c.mut.Unlock()
	delete(c.cache, key)
}

func (c *HelmCache) Get(key string) string {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.cache[key]
}

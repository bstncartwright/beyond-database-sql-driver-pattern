package cache

type InMemoryCache struct {
	cache map[string]interface{}
}

func NewInMemory() Cache {
	return &InMemoryCache{}
}

func (c *InMemoryCache) Get(key string) (any, bool) {
	v, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	return v, true
}

func (c *InMemoryCache) Put(key string, value any) error {
	c.cache[key] = value
	return nil
}

package cache

type InMemoryCache struct {
	cache map[string]any
}

func NewInMemory() Cache {
	return &InMemoryCache{make(map[string]any)}
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

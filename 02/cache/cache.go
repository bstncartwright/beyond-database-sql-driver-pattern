package cache

type Cache interface {
	Get(key string) (any, bool)
	Put(key string, value any) error
}

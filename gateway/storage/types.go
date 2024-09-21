package storage

type CacheInterface interface {
	Get(key string) bool
	Put(key string) string
}

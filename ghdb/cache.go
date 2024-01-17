package ghdb

type Cache interface {
	Clear()
	Get(key string) string
	Set(key string, value string)
	Remove(key string)
	GetOrSet(key string, setter func() string) string
}

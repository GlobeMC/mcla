package ghdb

import (
	"sync"
)

// Cache must be thread safe
type Cache interface {
	Clear()
	Get(key string) string
	Set(key string, value string)
	Remove(key string)
	GetOrSet(key string, setter func() string) string
}

type memoryCache struct {
	l sync.RWMutex
	m map[string]string
}

func NewInMemoryCache() Cache {
	return &memoryCache{
		m: make(map[string]string),
	}
}

func (m *memoryCache) Clear() {
	m.l.Lock()
	defer m.l.Unlock()
	clear(m.m)
}

func (m *memoryCache) Get(key string) string {
	m.l.RLock()
	defer m.l.RUnlock()
	return m.m[key]
}

func (m *memoryCache) Set(key string, value string) {
	m.l.Lock()
	defer m.l.Unlock()
	m.m[key] = value
}

func (m *memoryCache) Remove(key string) {
	m.l.Lock()
	defer m.l.Unlock()
	delete(m.m, key)
}

func (m *memoryCache) GetOrSet(key string, setter func() string) string {
	m.l.RLock()
	v, ok := m.m[key]
	m.l.RUnlock()
	if !ok {
		m.l.Lock()
		defer m.l.Unlock()
		v, ok = m.m[key]
		if !ok {
			v = setter()
			m.m[key] = v
		}
	}
	return v
}

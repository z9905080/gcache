package gcache

import (
	"context"
	"sync"
	"time"
)

type MemoryCacheManager struct {
	*sync.RWMutex
	ctx      context.Context
	CacheMap map[string]*MemoryCache
}

func (m *MemoryCacheManager) AddCache(mCacheName string) {
	m.Lock()
	defer m.Unlock()
	m.CacheMap[mCacheName] = &MemoryCache{
		Cache:       make(map[string]*cacheST, 0),
		Lock:        new(sync.RWMutex),
		InitHashKey: "000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000",
	}
}

func (m *MemoryCacheManager) RemoveCache(mCacheName string) {
	m.Lock()
	defer m.Unlock()
	delete(m.CacheMap, mCacheName)
}

func (m *MemoryCacheManager) FlushAll() {
	m.Lock()
	defer m.Unlock()
	m.CacheMap = make(map[string]*MemoryCache, 0)
}

func (m *MemoryCacheManager) GetCache(mCacheName string) CacheInterface {
	m.RLock()
	defer m.RUnlock()
	return m.CacheMap[mCacheName]
}

func (m *MemoryCacheManager) Check() {
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case <-m.ctx.Done():
			{
				break
			}
		case <-ticker:
			{
				m.RLock()
				for _, item := range m.CacheMap {
					item.Check()
				}
				m.RUnlock()
			}
		}
	}
}

// NewMemoryCacheManager 新的Cache控管中心
func NewMemoryCacheManager() CacheManager {
	manager := &MemoryCacheManager{
		RWMutex:  new(sync.RWMutex),
		ctx:      context.Background(),
		CacheMap: make(map[string]*MemoryCache, 0),
	}
	return manager
}

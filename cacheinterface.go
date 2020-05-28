package gcache

type CacheManager interface {
	AddCache(mCacheName string)
	RemoveCache(mCacheName string)
	GetCache(mCacheName string) CacheInterface
	FlushAll()
	Check()
}

type ConstructFunc func() CacheManager

func Start(mFunc ConstructFunc) CacheManager {
	manager := mFunc()
	go manager.Check()
	return manager
}

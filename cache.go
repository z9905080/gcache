package gcache

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/minio/highwayhash"
	"log"
	"sync"
	"time"
)

type CacheInterface interface {
	Remember(key string, expireTime int, argsMaps map[int]interface{}, getDataFunc GetDataFunc) (interface{}, error)
	Forget(key string)
	Check()
}

// cacheST 要快取的資料及時間
type cacheST struct {
	Data       interface{}
	ExpireTime time.Time
}

// MemoryCache 搜集結構map
type MemoryCache struct {
	Cache map[string]*cacheST
	Lock  *sync.RWMutex
}

func (c *MemoryCache) Check() {
	for key, cache := range c.Cache {
		if time.Now().After(cache.ExpireTime) {
			c.Lock.Lock()
			delete(c.Cache, key)
			c.Lock.Unlock()
			log.Println("刪除過期Key:", key)
		}
	}
}

type GetDataFunc func(argsMaps map[int]interface{}) (interface{}, error)

// Remember 將資料記錄到cache
func (c *MemoryCache) Remember(key string, expireTime int, argsMaps map[int]interface{}, getDataFunc GetDataFunc) (interface{}, error) {

	// 這套件必須給定初始Hash值,才可以去Write (可以考慮替換 crypto/sha256)
	cHashKey, dErr := hex.DecodeString("000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000") // use your own key here
	if dErr != nil {
		log.Printf("Cannot decode hex key: %v", dErr) // add error handling
		return "", dErr
	}

	hash, getHashErr := highwayhash.New(cHashKey)
	if getHashErr != nil {
		return "", getHashErr
	}
	hash.Write([]byte(key + fmt.Sprintf("%v", argsMaps)))

	checksum := hash.Sum(nil)
	newHashKey := hex.EncodeToString(checksum)

	// 取得快取資料
	if cacheData, errOfGetCacheData := c.getCacheData(newHashKey); errOfGetCacheData == nil {
		return cacheData, nil
	}
	c.Lock.Lock()
	defer c.Lock.Unlock()

	data, err := getDataFunc(argsMaps)
	if err != nil {
		return "", err
	}

	c.Cache[newHashKey] = &cacheST{
		Data:       data,
		ExpireTime: time.Now().Add(time.Duration(expireTime) * time.Minute),
	}
	return c.Cache[newHashKey].Data, nil
}

// RememberNew 將資料記錄到cache(增加是否強迫更新參數)
func (c *MemoryCache) RememberNew(key string, expireTime int, argsMaps map[int]interface{}, isForce bool, getDataFunc GetDataFunc) (interface{}, error) {

	// 如果非必更新的話
	if !isForce {
		// 取得快取資料
		if cacheData, errOfGetCacheData := c.getCacheData(key); errOfGetCacheData == nil {
			return cacheData, nil
		}
	}
	c.Lock.Lock()
	defer c.Lock.Unlock()
	data, err := getDataFunc(argsMaps)
	if err != nil {
		return "", err
	}
	c.Cache[key] = &cacheST{
		Data:       data,
		ExpireTime: time.Now().Add(time.Duration(expireTime) * time.Minute),
	}
	return c.Cache[key].Data, nil
}

func (c *MemoryCache) getCacheData(key string) (interface{}, error) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	if item, isExist := c.Cache[key]; isExist {
		if time.Now().Before(item.ExpireTime) {
			return item.Data, nil
		}
	}
	return nil, errors.New("取無資料")
}

// Forget 清除某一筆Cache資料with Key
func (c *MemoryCache) Forget(key string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	delete(c.Cache, key)
}

// Flush 將整個Cache清空
func (c *MemoryCache) Flush() {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.Cache = make(map[string]*cacheST, 0)
}

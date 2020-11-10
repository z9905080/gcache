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
	Remember(key string, expireTime int, argsMaps map[int]interface{}, isForce bool, getDataFunc GetDataFunc) (interface{}, error)
	Forget(key string, argsMap map[int]interface{}) error
	ForgetByHashKey(hashKey string)
	GetHashKey(originKey string, argsMap map[int]interface{}) (string, error)
	Check()
}

// cacheST 要快取的資料及時間
type cacheST struct {
	Data       interface{}
	ExpireTime time.Time
}

// MemoryCache 搜集結構map
type MemoryCache struct {
	Cache       map[string]*cacheST
	Lock        *sync.RWMutex
	InitHashKey string
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

func (c *MemoryCache) GetHashKey(originKey string, argsMap map[int]interface{}) (string, error) {
	// 這套件必須給定初始Hash值,才可以去Write (可以考慮替換 crypto/sha256)
	cHashKey, dErr := hex.DecodeString(c.InitHashKey) // use your own key here
	if dErr != nil {
		log.Printf("Cannot decode hex key: %v", dErr) // add error handling
		return "", dErr
	}

	hash, getHashErr := highwayhash.New(cHashKey)
	if getHashErr != nil {
		return "", getHashErr
	}
	hash.Write([]byte(originKey + fmt.Sprintf("%v", argsMap)))

	checksum := hash.Sum(nil)
	newHashKey := hex.EncodeToString(checksum)

	return newHashKey, nil
}

// Remember 將資料記錄到cache
func (c *MemoryCache) Remember(key string, expireTime int, argsMaps map[int]interface{}, isForce bool, getDataFunc GetDataFunc) (interface{}, error) {

	newHashKey, hashErr := c.GetHashKey(key, argsMaps)
	if hashErr != nil {
		return nil, hashErr
	}

	// 如果非必更新的話
	if !isForce {
		// 取得快取資料
		if cacheData, errOfGetCacheData := c.getCacheData(key); errOfGetCacheData == nil {
			return cacheData, nil
		}
	}

	c.Lock.Lock()
	defer c.Lock.Unlock()

	// 第二次取得快取資料,防止同時間卡在Lock處
	// 這裡不上Read Lock, 因已經上Write Lock
	if item, isExist := c.Cache[key]; isExist {
		if time.Now().Before(item.ExpireTime) {
			return item.Data, nil
		}
	}

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
func (c *MemoryCache) Forget(key string, argsMap map[int]interface{}) error {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if hashKey, hashErr := c.GetHashKey(key, argsMap); hashErr != nil {
		return hashErr
	} else {
		delete(c.Cache, hashKey)
		return nil
	}
}

// ForgetByHashKey 清除某一筆Cache資料with HashKey
func (c *MemoryCache) ForgetByHashKey(hashKey string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	delete(c.Cache, hashKey)
}

// Flush 將整個Cache清空
func (c *MemoryCache) Flush() {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.Cache = make(map[string]*cacheST, 0)
}

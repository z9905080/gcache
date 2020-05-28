package gcache

import (
	"log"
	"testing"
	"time"
)

func TestA(t *testing.T) {
	manager := Start(NewMemoryCacheManager)
	manager.AddCache("he")

	go func() {
		argsMap := map[int]interface{}{
			1: "BB",
		}

		heCache := manager.GetCache("he")
		for i := 0; i < 100000000; i++ {
			data, err := heCache.Remember("a", 1, argsMap, false, func(argsMaps map[int]interface{}) (interface{}, error) {
				return "A", nil
			})
			log.Println("A:", data, err)
			//time.Sleep(1 * time.Second)
		}
		log.Println("finish A")
	}()

	argsMap := map[int]interface{}{
		1: "AA",
	}
	time.Sleep(1 * time.Second)

	heCache := manager.GetCache("he")
	for i := 0; i < 10000000; i++ {
		data, err := heCache.Remember("a", 1, argsMap, false, func(argsMaps map[int]interface{}) (interface{}, error) {
			return "B", nil
		})
		//time.Sleep(1 * time.Second)
		log.Println("B:", data, err)
	}
}

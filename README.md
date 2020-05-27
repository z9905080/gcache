# gcache
cache of golang

## Install

`go get -u github.com/z9905080/gcache`

## Usage

 ```	
    manager := NewMemoryCacheManager()
 	manager.AddCache("hello")
	
    argsMap := map[int]interface{}{
		1: "paramData",
	}
	data, err := manager.GetCache("hello").Remember("a", 1, argsMap, func(argsMaps map[int]interface{}) (interface{}, error) {
		return "B", nil
	})
	log.Println("B:", data, err)
```
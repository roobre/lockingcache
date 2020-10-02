// mapcache is an in-memory implementation of tcache using maps
package mapcache

import (
	"roob.re/tcache"
	"sync"
)

type MapCache struct {
	mutex       sync.RWMutex
	collections map[string]*mapTable
}

func New() *MapCache {
	return &MapCache{
		collections: map[string]*mapTable{},
	}
}

func (mc *MapCache) From(key string) tcache.Table {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mt, found := mc.collections[key]
	if !found {
		mt = &mapTable{
			rows: map[string]*mapEntry{},
		}
		mc.collections[key] = mt
	}

	return mt
}

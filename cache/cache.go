package cache

import (
	"github.com/patrickmn/go-cache"
)

var c = cache.New(cache.NoExpiration, cache.NoExpiration)

func Put(prefix string, k string, x interface{}) {
	c.Set(prefix+k, x, cache.NoExpiration)
}

func Get(prefix string, k string) (interface{}, bool) {
	return c.Get(prefix + k)
}

func Remove(prefix string, k string) {
	c.Delete(prefix + k)
}

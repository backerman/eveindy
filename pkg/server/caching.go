/*
Copyright © 2014–6 Brad Ackerman.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package server

import (
	"time"

	"github.com/backerman/evego"
	"github.com/pmylund/go-cache"
)

type gocache struct {
	c *cache.Cache
}

func (g *gocache) Close() error {
	return nil
}

func (g *gocache) Get(key string) ([]byte, bool) {
	cached, found := g.c.Get(key)
	if found {
		cachedBlob := cached.([]byte)
		return cachedBlob, true
	}
	return nil, false
}

func (g *gocache) Put(key string, val []byte, expires time.Time) error {
	cacheDuration := expires.Sub(time.Now())
	if cacheDuration > 0 {
		g.c.Set(key, val, cacheDuration)
	}
	return nil
}

// InMemCache returns a new cache object that the Eve-Central interface
// will use to cache returned results.
func InMemCache() evego.Cache {
	c := cache.New(5*time.Minute, 30*time.Second)
	return &gocache{c}
}

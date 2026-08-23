package query

import (
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"sync"
)

type RouteSnapshot struct {
	Carrier string
	Stops   []string
}
type RouteCache struct {
	mu     sync.RWMutex
	routes map[string]RouteSnapshot
}

func NewRouteCache() *RouteCache { return &RouteCache{routes: map[string]RouteSnapshot{}} }
func (c *RouteCache) Put(id string, r RouteSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.routes[id] = RouteSnapshot{Carrier: r.Carrier, Stops: append([]string(nil), r.Stops...)}
}
func (c *RouteCache) Get(id string) (RouteSnapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.routes[id]
	if !ok {
		return RouteSnapshot{}, false
	}
	return RouteSnapshot{Carrier: r.Carrier, Stops: append([]string(nil), r.Stops...)}, true
}
func (c *RouteCache) ForUser(_ domain.User, id string) (RouteSnapshot, bool) {
	return c.forUserSnapshot(id)
}

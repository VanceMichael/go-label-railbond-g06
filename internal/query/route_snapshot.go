package query

func (c *RouteCache) forUserSnapshot(id string) (RouteSnapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	snapshot, ok := c.routes[id]
	if !ok {
		return RouteSnapshot{}, false
	}
	return snapshot, true
}

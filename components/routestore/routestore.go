package routestore

import (
	"sync"

	rm "github.com/aarthikrao/timeMachine/models/routemodels"
)

type RouteStore struct {
	m  map[string]*rm.Route
	mu sync.RWMutex
}

func InitRouteStore() *RouteStore {
	return &RouteStore{
		m: make(map[string]*rm.Route),
	}
}

func (rs *RouteStore) AddRoute(id string, route *rm.Route) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.m[id] = route
}

func (rs *RouteStore) RemoveRoute(id string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	delete(rs.m, id)
}

func (rs *RouteStore) GetRoute(id string) *rm.Route {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	return rs.m[id]
}

// Snapshot returns the current snapshot of the route store
func (rs *RouteStore) Snapshot() map[string]*rm.Route {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	return rs.m // TODO: Safe to return map ?
}

// Loads the map to the route store
func (rs *RouteStore) Load(m map[string]*rm.Route) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.m = m
}

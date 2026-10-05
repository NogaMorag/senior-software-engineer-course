// Package projects is a tiny multi-tenant resource: every project belongs to one tenant.
package projects

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"sync"
)

// Project is one tenant-owned object.
type Project struct {
	ID       string `json:"id"`
	TenantID string `json:"-"`
	Name     string `json:"name"`
}

// ErrNotFound means "no such project for this tenant". It is the same error
// whether the id does not exist or belongs to someone else.
var ErrNotFound = errors.New("project not found")

// MemoryStore keeps projects in memory. A PostgreSQL store would put the
// tenant into every WHERE clause in the same way.
type MemoryStore struct {
	mu   sync.RWMutex
	byID map[string]Project
}

// NewMemoryStore returns a store seeded with ps.
func NewMemoryStore(ps ...Project) *MemoryStore {
	s := &MemoryStore{byID: make(map[string]Project, len(ps))}
	for _, p := range ps {
		s.byID[p.ID] = p
	}
	return s
}

// Get returns the project only if it belongs to tenantID.
// The tenant is part of the lookup key, not a check the caller may forget.
func (s *MemoryStore) Get(_ context.Context, tenantID, id string) (Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.byID[id]
	if !ok || p.TenantID != tenantID {
		return Project{}, ErrNotFound
	}
	return p, nil
}

// List returns the tenant's projects ordered by id.
func (s *MemoryStore) List(_ context.Context, tenantID string) ([]Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Project{}
	for _, p := range s.byID {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	slices.SortFunc(out, func(a, b Project) int { return cmp.Compare(a.ID, b.ID) })
	return out, nil
}

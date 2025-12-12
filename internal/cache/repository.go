package cache

import (
	"context"
)

// EntityCacheHelper provides methods for caching single entity lookups (GetByID).
// Wraps Get/Set operations with appropriate key building and TTL handling.
type EntityCacheHelper struct {
	client   *Client
	resource string // e.g., "player", "team", "game"
}

// NewEntityCacheHelper creates a helper for caching entity reads.
func NewEntityCacheHelper(client *Client, resource string) *EntityCacheHelper {
	return &EntityCacheHelper{
		client:   client,
		resource: resource,
	}
}

// Get attempts to retrieve a cached entity by ID.
// Returns true if cache hit, false if miss or cache disabled.
func (h *EntityCacheHelper) Get(ctx context.Context, id string, dest any) bool {
	if h.client == nil {
		return false
	}

	key := h.client.EntityKey(h.resource, id)
	return h.client.Get(ctx, key, dest)
}

// Set stores an entity in cache with entity TTL.
func (h *EntityCacheHelper) Set(ctx context.Context, id string, entity any) error {
	if h.client == nil {
		return nil
	}

	key := h.client.EntityKey(h.resource, id)
	return h.client.Set(ctx, key, entity, h.client.config.TTLs.Entity)
}

// GetOrCompute implements cache-aside pattern for entity lookups.
// Checks cache first, calls compute function on miss, stores result.
func (h *EntityCacheHelper) GetOrCompute(ctx context.Context, id string, compute func() (any, error)) (any, error) {
	if h.client == nil {
		return compute()
	}

	key := h.client.EntityKey(h.resource, id)
	return h.client.GetOrCompute(ctx, key, h.client.config.TTLs.Entity, compute)
}

// Delete removes an entity from cache (for explicit invalidation after writes).
func (h *EntityCacheHelper) Delete(ctx context.Context, id string) error {
	if h.client == nil {
		return nil
	}

	key := h.client.EntityKey(h.resource, id)
	return h.client.Delete(ctx, key)
}

// ListCacheHelper provides methods for caching collection queries with filters.
// Handles parameter normalization and hashing for stable cache keys.
type ListCacheHelper struct {
	client   *Client
	resource string
}

// NewListCacheHelper creates a helper for caching list/collection queries.
func NewListCacheHelper(client *Client, resource string) *ListCacheHelper {
	return &ListCacheHelper{
		client:   client,
		resource: resource,
	}
}

// Get attempts to retrieve cached list results.
func (h *ListCacheHelper) Get(ctx context.Context, params map[string]string, dest any) bool {
	if h.client == nil {
		return false
	}

	key := h.client.ListKey(h.resource, params)
	return h.client.Get(ctx, key, dest)
}

// Set stores list results in cache with list TTL.
func (h *ListCacheHelper) Set(ctx context.Context, params map[string]string, results any) error {
	if h.client == nil {
		return nil
	}

	key := h.client.ListKey(h.resource, params)
	return h.client.Set(ctx, key, results, h.client.config.TTLs.List)
}

// GetOrCompute implements cache-aside pattern for list queries.
func (h *ListCacheHelper) GetOrCompute(ctx context.Context, params map[string]string, compute func() (any, error)) (any, error) {
	if h.client == nil {
		return compute()
	}

	key := h.client.ListKey(h.resource, params)
	return h.client.GetOrCompute(ctx, key, h.client.config.TTLs.List, compute)
}

// InvalidateAll removes all cached list results for this resource.
// Use when data changes that could affect any list query.
func (h *ListCacheHelper) InvalidateAll(ctx context.Context) (int, error) {
	if h.client == nil {
		return 0, nil
	}

	prefix := h.client.KeyPrefix(KeyTypeList, h.resource)
	return h.client.InvalidateByPrefix(ctx, prefix)
}

// FilterToParamMap converts a filter struct to normalized parameter map for caching.
// TODO: reflection or type-specific implementations; repositories should implement custom converters
func FilterToParamMap(filter any) map[string]string {
	return make(map[string]string)
}

// CachedRepository provides a complete caching layer for repositories.
// Combines entity and list caching helpers.
type CachedRepository struct {
	Entity *EntityCacheHelper
	List   *ListCacheHelper
}

// NewCachedRepository creates a cached repository helper for a given resource type.
func NewCachedRepository(client *Client, resource string) *CachedRepository {
	if client == nil {
		return &CachedRepository{Entity: &EntityCacheHelper{}, List: &ListCacheHelper{}}
	}

	return &CachedRepository{
		Entity: NewEntityCacheHelper(client, resource),
		List:   NewListCacheHelper(client, resource),
	}
}

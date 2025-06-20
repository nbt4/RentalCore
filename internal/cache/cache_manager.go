package cache

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IsExpired checks if the cache item has expired
func (c *CacheItem) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// MemoryCache provides in-memory caching with TTL support
type MemoryCache struct {
	items map[string]*CacheItem
	mutex sync.RWMutex
	ttl   time.Duration
}

// NewMemoryCache creates a new memory cache instance
func NewMemoryCache(defaultTTL time.Duration) *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*CacheItem),
		ttl:   defaultTTL,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Set stores an item in cache with default TTL
func (m *MemoryCache) Set(key string, value interface{}) {
	m.SetWithTTL(key, value, m.ttl)
}

// SetWithTTL stores an item in cache with custom TTL
func (m *MemoryCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.items[key] = &CacheItem{
		Data:      value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Get retrieves an item from cache
func (m *MemoryCache) Get(key string) (interface{}, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.items[key]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// Delete removes an item from cache
func (m *MemoryCache) Delete(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.items, key)
}

// Clear removes all items from cache
func (m *MemoryCache) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.items = make(map[string]*CacheItem)
}

// GetStats returns cache statistics
func (m *MemoryCache) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	totalItems := len(m.items)
	expiredItems := 0

	for _, item := range m.items {
		if item.IsExpired() {
			expiredItems++
		}
	}

	return map[string]interface{}{
		"total_items":   totalItems,
		"expired_items": expiredItems,
		"active_items":  totalItems - expiredItems,
	}
}

// cleanup removes expired items periodically
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()
		for key, item := range m.items {
			if item.IsExpired() {
				delete(m.items, key)
			}
		}
		m.mutex.Unlock()
	}
}

// CacheManager provides application-level caching
type CacheManager struct {
	lookupCache  *MemoryCache
	queryCache   *MemoryCache
	sessionCache *MemoryCache
}

// NewCacheManager creates a new cache manager
func NewCacheManager() *CacheManager {
	return &CacheManager{
		lookupCache:  NewMemoryCache(30 * time.Minute), // Lookup data (products, categories)
		queryCache:   NewMemoryCache(5 * time.Minute),  // Query results
		sessionCache: NewMemoryCache(60 * time.Minute), // Session data
	}
}

// CacheKey generates cache keys for different data types
func (cm *CacheManager) CacheKey(dataType, identifier string) string {
	return fmt.Sprintf("%s:%s", dataType, identifier)
}

// GetLookupData retrieves cached lookup data (products, categories, etc.)
func (cm *CacheManager) GetLookupData(key string) (interface{}, bool) {
	return cm.lookupCache.Get(key)
}

// SetLookupData caches lookup data with longer TTL
func (cm *CacheManager) SetLookupData(key string, data interface{}) {
	cm.lookupCache.Set(key, data)
}

// GetQueryResult retrieves cached query results
func (cm *CacheManager) GetQueryResult(key string) (interface{}, bool) {
	return cm.queryCache.Get(key)
}

// SetQueryResult caches query results with shorter TTL
func (cm *CacheManager) SetQueryResult(key string, data interface{}) {
	cm.queryCache.Set(key, data)
}

// GetSessionData retrieves cached session data
func (cm *CacheManager) GetSessionData(key string) (interface{}, bool) {
	return cm.sessionCache.Get(key)
}

// SetSessionData caches session data
func (cm *CacheManager) SetSessionData(key string, data interface{}) {
	cm.sessionCache.Set(key, data)
}

// InvalidateLookupData removes lookup data from cache
func (cm *CacheManager) InvalidateLookupData(pattern string) {
	// Simple pattern matching for cache invalidation
	cm.lookupCache.Clear() // For simplicity, clear all lookup cache
}

// InvalidateQueryCache removes query results from cache
func (cm *CacheManager) InvalidateQueryCache(pattern string) {
	cm.queryCache.Clear() // For simplicity, clear all query cache
}

// GetAllStats returns statistics for all caches
func (cm *CacheManager) GetAllStats() map[string]interface{} {
	return map[string]interface{}{
		"lookup_cache":  cm.lookupCache.GetStats(),
		"query_cache":   cm.queryCache.GetStats(),
		"session_cache": cm.sessionCache.GetStats(),
	}
}

// CachedLookupService provides cached lookup operations
type CachedLookupService struct {
	cache *CacheManager
}

// NewCachedLookupService creates a new cached lookup service
func NewCachedLookupService(cache *CacheManager) *CachedLookupService {
	return &CachedLookupService{cache: cache}
}

// GetProducts returns cached product list
func (cls *CachedLookupService) GetProducts() ([]interface{}, error) {
	key := "products:all"
	
	if data, found := cls.cache.GetLookupData(key); found {
		if products, ok := data.([]interface{}); ok {
			return products, nil
		}
	}

	// If not in cache, would fetch from database here
	// For now, return empty slice
	return []interface{}{}, nil
}

// GetCategories returns cached category list
func (cls *CachedLookupService) GetCategories() ([]interface{}, error) {
	key := "categories:all"
	
	if data, found := cls.cache.GetLookupData(key); found {
		if categories, ok := data.([]interface{}); ok {
			return categories, nil
		}
	}

	// If not in cache, would fetch from database here
	return []interface{}{}, nil
}

// GetCustomersForDropdown returns cached customer list for dropdowns
func (cls *CachedLookupService) GetCustomersForDropdown() ([]interface{}, error) {
	key := "customers:dropdown"
	
	if data, found := cls.cache.GetLookupData(key); found {
		if customers, ok := data.([]interface{}); ok {
			return customers, nil
		}
	}

	// If not in cache, would fetch from database here
	return []interface{}{}, nil
}

// InvalidateProductCache removes product-related cache entries
func (cls *CachedLookupService) InvalidateProductCache() {
	cls.cache.InvalidateLookupData("products")
}

// QueryCache provides caching for complex queries
type QueryCache struct {
	cache *CacheManager
}

// NewQueryCache creates a new query cache
func NewQueryCache(cache *CacheManager) *QueryCache {
	return &QueryCache{cache: cache}
}

// GetDeviceList returns cached device list
func (qc *QueryCache) GetDeviceList(filterHash string) (interface{}, bool) {
	key := fmt.Sprintf("devices:list:%s", filterHash)
	return qc.cache.GetQueryResult(key)
}

// SetDeviceList caches device list result
func (qc *QueryCache) SetDeviceList(filterHash string, data interface{}) {
	key := fmt.Sprintf("devices:list:%s", filterHash)
	qc.cache.SetQueryResult(key, data)
}

// GetJobList returns cached job list
func (qc *QueryCache) GetJobList(filterHash string) (interface{}, bool) {
	key := fmt.Sprintf("jobs:list:%s", filterHash)
	return qc.cache.GetQueryResult(key)
}

// SetJobList caches job list result
func (qc *QueryCache) SetJobList(filterHash string, data interface{}) {
	key := fmt.Sprintf("jobs:list:%s", filterHash)
	qc.cache.SetQueryResult(key, data)
}

// HashFilters creates a hash for filter parameters to use as cache key
func (qc *QueryCache) HashFilters(filters interface{}) string {
	// Simple JSON-based hash for cache keys
	jsonData, _ := json.Marshal(filters)
	return fmt.Sprintf("%x", jsonData)
}

// Global cache manager instance
var GlobalCache *CacheManager

// InitializeCache initializes the global cache manager
func InitializeCache() {
	GlobalCache = NewCacheManager()
}
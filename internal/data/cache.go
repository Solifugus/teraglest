package data

import (
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a cached asset with metadata
type CacheEntry struct {
	Data      interface{}   // The actual asset data
	LoadedAt  time.Time     // When the asset was loaded
	Size      int64         // Size in bytes (for memory tracking)
	RefCount  int           // Reference count for usage tracking
	AssetType string        // Type of asset ("xml", "g3d", "texture", "audio")
}

// AssetCache provides thread-safe caching for game assets
type AssetCache struct {
	entries map[string]*CacheEntry // Key: asset path, Value: cached entry
	mutex   sync.RWMutex           // Thread-safe access

	// Statistics
	hits   int64 // Cache hits
	misses int64 // Cache misses

	// Configuration
	maxMemoryMB int64 // Maximum memory usage in MB (0 = unlimited)
	maxEntries  int   // Maximum number of entries (0 = unlimited)
}

// NewAssetCache creates a new asset cache with optional limits
func NewAssetCache(maxMemoryMB int64, maxEntries int) *AssetCache {
	return &AssetCache{
		entries:     make(map[string]*CacheEntry),
		maxMemoryMB: maxMemoryMB,
		maxEntries:  maxEntries,
	}
}

// Get retrieves an asset from the cache
func (cache *AssetCache) Get(path string) (interface{}, bool) {
	cache.mutex.Lock() // Use full lock for stats updates
	defer cache.mutex.Unlock()

	entry, exists := cache.entries[path]
	if !exists {
		cache.misses++
		return nil, false
	}

	// Update reference count and hit statistics
	entry.RefCount++
	cache.hits++

	return entry.Data, true
}

// Put stores an asset in the cache
func (cache *AssetCache) Put(path string, data interface{}, assetType string, size int64) error {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// Check if we need to enforce memory limits
	if cache.maxMemoryMB > 0 {
		currentMemoryMB := cache.getMemoryUsageMB()
		if currentMemoryMB+size/1024/1024 > cache.maxMemoryMB {
			// Try to evict some entries
			evicted := cache.evictLeastRecentlyUsed(size)
			if !evicted {
				return fmt.Errorf("cache memory limit exceeded and could not evict entries for asset: %s", path)
			}
		}
	}

	// Check entry count limits
	if cache.maxEntries > 0 && len(cache.entries) >= cache.maxEntries {
		// Evict oldest entry
		cache.evictOldest()
	}

	// Store the entry
	cache.entries[path] = &CacheEntry{
		Data:      data,
		LoadedAt:  time.Now(),
		Size:      size,
		RefCount:  1,
		AssetType: assetType,
	}

	return nil
}

// Remove removes an asset from the cache
func (cache *AssetCache) Remove(path string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	delete(cache.entries, path)
}

// Clear removes all assets from the cache
func (cache *AssetCache) Clear() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.entries = make(map[string]*CacheEntry)
	cache.hits = 0
	cache.misses = 0
}

// GetStats returns cache statistics
func (cache *AssetCache) GetStats() CacheStats {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	stats := CacheStats{
		TotalEntries: len(cache.entries),
		Hits:         cache.hits,
		Misses:       cache.misses,
		MemoryUsageMB: cache.getMemoryUsageMB(),
	}

	if cache.hits+cache.misses > 0 {
		stats.HitRatio = float64(cache.hits) / float64(cache.hits+cache.misses)
	}

	// Count assets by type
	stats.AssetCounts = make(map[string]int)
	for _, entry := range cache.entries {
		stats.AssetCounts[entry.AssetType]++
	}

	return stats
}

// CacheStats provides cache performance and usage statistics
type CacheStats struct {
	TotalEntries  int            // Total number of cached assets
	Hits          int64          // Number of cache hits
	Misses        int64          // Number of cache misses
	HitRatio      float64        // Hit ratio (hits / (hits + misses))
	MemoryUsageMB int64          // Current memory usage in MB
	AssetCounts   map[string]int // Count of assets by type
}

// getMemoryUsageMB calculates current memory usage in MB (internal method)
func (cache *AssetCache) getMemoryUsageMB() int64 {
	var totalBytes int64
	for _, entry := range cache.entries {
		totalBytes += entry.Size
	}
	return totalBytes / 1024 / 1024
}

// evictLeastRecentlyUsed removes entries with lowest reference count to free memory
func (cache *AssetCache) evictLeastRecentlyUsed(neededBytes int64) bool {
	// Find entries with lowest reference count
	type entryInfo struct {
		path     string
		refCount int
		size     int64
		loadedAt time.Time
	}

	var candidates []entryInfo
	for path, entry := range cache.entries {
		candidates = append(candidates, entryInfo{
			path:     path,
			refCount: entry.RefCount,
			size:     entry.Size,
			loadedAt: entry.LoadedAt,
		})
	}

	// Sort by reference count (ascending), then by age (oldest first)
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[i].refCount > candidates[j].refCount ||
				(candidates[i].refCount == candidates[j].refCount && candidates[i].loadedAt.After(candidates[j].loadedAt)) {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Evict entries until we have enough space
	var freedBytes int64
	for _, candidate := range candidates {
		delete(cache.entries, candidate.path)
		freedBytes += candidate.size
		if freedBytes >= neededBytes {
			return true
		}
	}

	return freedBytes >= neededBytes
}

// evictOldest removes the oldest entry from the cache
func (cache *AssetCache) evictOldest() {
	var oldestPath string
	var oldestTime time.Time

	// Find the oldest entry
	for path, entry := range cache.entries {
		if oldestPath == "" || entry.LoadedAt.Before(oldestTime) {
			oldestPath = path
			oldestTime = entry.LoadedAt
		}
	}

	if oldestPath != "" {
		delete(cache.entries, oldestPath)
	}
}

// PrintStats prints cache statistics for debugging
func (cache *AssetCache) PrintStats() {
	stats := cache.GetStats()

	fmt.Println("Asset Cache Statistics:")
	fmt.Printf("  Total Entries: %d\n", stats.TotalEntries)
	fmt.Printf("  Cache Hits: %d\n", stats.Hits)
	fmt.Printf("  Cache Misses: %d\n", stats.Misses)
	fmt.Printf("  Hit Ratio: %.2f%%\n", stats.HitRatio*100)
	fmt.Printf("  Memory Usage: %d MB\n", stats.MemoryUsageMB)
	fmt.Println("  Assets by Type:")
	for assetType, count := range stats.AssetCounts {
		fmt.Printf("    %s: %d\n", assetType, count)
	}
}
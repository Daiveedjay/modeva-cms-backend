package category_cache

import (
	"sync"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
)

const TTL = 5 * time.Minute

// ── Full category tree cache ─────────────────────────────────────────────────
// Stores parents with children preloaded + product counts.
// Both GetCategories and GetAllParentCategories read from this.

type treeEntry struct {
	parents       []models.Category
	productCounts map[string]int
	fetchedAt     time.Time
}

var (
	treeMu    sync.RWMutex
	treeCache *treeEntry
)

func GetTree() (parents []models.Category, productCounts map[string]int, ok bool) {
	treeMu.RLock()
	defer treeMu.RUnlock()
	if treeCache != nil && time.Since(treeCache.fetchedAt) < TTL {
		return treeCache.parents, treeCache.productCounts, true
	}
	return nil, nil, false
}

func SetTree(parents []models.Category, productCounts map[string]int) {
	treeMu.Lock()
	defer treeMu.Unlock()
	treeCache = &treeEntry{
		parents:       parents,
		productCounts: productCounts,
		fetchedAt:     time.Now(),
	}
}

// ── Sub-categories cache ─────────────────────────────────────────────────────

type subEntry struct {
	data      []models.Category
	fetchedAt time.Time
}

var (
	subMu    sync.RWMutex
	subCache *subEntry
)

func GetSubs() ([]models.Category, bool) {
	subMu.RLock()
	defer subMu.RUnlock()
	if subCache != nil && time.Since(subCache.fetchedAt) < TTL {
		return subCache.data, true
	}
	return nil, false
}

func SetSubs(data []models.Category) {
	subMu.Lock()
	defer subMu.Unlock()
	subCache = &subEntry{data: data, fetchedAt: time.Now()}
}

// ── Invalidate everything (call on any category create/update/delete) ────────

func Invalidate() {
	treeMu.Lock()
	treeCache = nil
	treeMu.Unlock()

	subMu.Lock()
	subCache = nil
	subMu.Unlock()
}

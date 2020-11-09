package blockstore

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	CacheName, _ = tag.NewKey("cache_name")
)

// CacheMeasures groups all metrics emitted by the blockstore caches.
var CacheMeasures = struct {
	HitRatio     *stats.Float64Measure
	Hits         *stats.Int64Measure
	Misses       *stats.Int64Measure
	Adds         *stats.Int64Measure
	Updates      *stats.Int64Measure
	Evictions    *stats.Int64Measure
	CostAdded    *stats.Int64Measure
	CostEvicted  *stats.Int64Measure
	SetsDropped  *stats.Int64Measure
	SetsRejected *stats.Int64Measure
	GetsDropped  *stats.Int64Measure
	GetsKept     *stats.Int64Measure
}{
	HitRatio:     stats.Float64("blockstore/cache/hit_ratio", "Hit ratio of blockstore cache", stats.UnitDimensionless),
	Hits:         stats.Int64("blockstore/cache/hits", "Total number of hits at blockstore cache", stats.UnitDimensionless),
	Misses:       stats.Int64("blockstore/cache/misses", "Total number of misses at blockstore cache", stats.UnitDimensionless),
	Adds:         stats.Int64("blockstore/cache/adds", "Total number of adds to blockstore cache", stats.UnitDimensionless),
	Updates:      stats.Int64("blockstore/cache/updates", "Total number of updates in blockstore cache", stats.UnitDimensionless),
	Evictions:    stats.Int64("blockstore/cache/evictions", "Total number of evictions from blockstore cache", stats.UnitDimensionless),
	CostAdded:    stats.Int64("blockstore/cache/cost_added", "Total cost (byte size) of entries added into blockstore cache", stats.UnitBytes),
	CostEvicted:  stats.Int64("blockstore/cache/cost_evicted", "Total cost (byte size) of entries evicted by blockstore cache", stats.UnitBytes),
	SetsDropped:  stats.Int64("blockstore/cache/sets_dropped", "Total number of sets dropped by blockstore cache", stats.UnitDimensionless),
	SetsRejected: stats.Int64("blockstore/cache/sets_rejected", "Total number of sets rejected by blockstore cache", stats.UnitDimensionless),
	GetsDropped:  stats.Int64("blockstore/cache/gets_dropped", "Total number of gets dropped by blockstore cache", stats.UnitDimensionless),
	GetsKept:     stats.Int64("blockstore/cache/gets_kept", "Total number of gets serviced by blockstore cache", stats.UnitDimensionless),
}

// CacheViews groups all cache-related default views.
var CacheViews = struct {
	HitRatio     *view.View
	Hits         *view.View
	Misses       *view.View
	Adds         *view.View
	Updates      *view.View
	Evictions    *view.View
	CostAdded    *view.View
	CostEvicted  *view.View
	SetsDropped  *view.View
	SetsRejected *view.View
	GetsDropped  *view.View
	GetsKept     *view.View
}{
	HitRatio: &view.View{
		Measure:     CacheMeasures.HitRatio,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{CacheName},
	},
	Hits: &view.View{
		Measure:     CacheMeasures.Hits,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{CacheName},
	},
	Misses: &view.View{
		Measure:     CacheMeasures.Misses,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{CacheName},
	},
	Adds: &view.View{
		Measure:     CacheMeasures.Adds,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{CacheName},
	},
	Updates: &view.View{
		Measure:     CacheMeasures.Updates,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{CacheName},
	},
	Evictions: &view.View{
		Measure:     CacheMeasures.Evictions,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{CacheName},
	},
	CostAdded: &view.View{
		Measure:     CacheMeasures.CostAdded,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{CacheName},
	},
	CostEvicted: &view.View{
		Measure:     CacheMeasures.CostEvicted,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{CacheName},
	},
	SetsDropped: &view.View{
		Measure:     CacheMeasures.SetsDropped,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{CacheName},
	},
	SetsRejected: &view.View{
		Measure:     CacheMeasures.SetsRejected,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{CacheName},
	},
	GetsDropped: &view.View{
		Measure:     CacheMeasures.GetsDropped,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{CacheName},
	},
	GetsKept: &view.View{
		Measure:     CacheMeasures.GetsKept,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{CacheName},
	},
}

// DefaultViews exports all default views for this package.
var DefaultViews = []*view.View{
	CacheViews.HitRatio,
	CacheViews.Hits,
	CacheViews.Misses,
	CacheViews.Adds,
	CacheViews.Updates,
	CacheViews.Evictions,
	CacheViews.CostAdded,
	CacheViews.CostEvicted,
	CacheViews.SetsDropped,
	CacheViews.SetsRejected,
	CacheViews.GetsDropped,
	CacheViews.GetsKept,
}

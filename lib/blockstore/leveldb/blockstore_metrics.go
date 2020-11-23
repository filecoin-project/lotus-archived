package leveldbbs

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	// TagName records the blockstore name.
	TagName, _ = tag.NewKey("name")

	// TagLevel records the level in level-related metrics.
	TagLevel, _ = tag.NewKey("level")
)

var Metrics = struct {
	LevelSize           *stats.Int64Measure
	LevelTableCount     *stats.Int64Measure
	LevelRead           *stats.Int64Measure
	LevelWrite          *stats.Int64Measure
	LevelCompactionTime *stats.Int64Measure

	BlockCacheSize *stats.Int64Measure
	OpenedTables   *stats.Int64Measure

	IORead  *stats.Int64Measure
	IOWrite *stats.Int64Measure

	AliveSnapshots *stats.Int64Measure
	AliveIterators *stats.Int64Measure

	WriteDelay         *stats.Int64Measure
	WriteDelayDuration *stats.Int64Measure
	WritePaused        *stats.Int64Measure
}{
	LevelSize:           stats.Int64("leveldb/level_size", "LSM level size", stats.UnitBytes),
	LevelTableCount:     stats.Int64("leveldb/level_tables", "LSM level tables", stats.UnitDimensionless),
	LevelRead:           stats.Int64("leveldb/level_read", "LSM level bytes read", stats.UnitBytes),
	LevelWrite:          stats.Int64("leveldb/level_written", "LSM level bytes written", stats.UnitBytes),
	LevelCompactionTime: stats.Int64("leveldb/level_compact_time", "LSM level compaction time", stats.UnitMilliseconds),

	BlockCacheSize: stats.Int64("leveldb/block_cache/size", "block cache size", stats.UnitBytes),
	OpenedTables:   stats.Int64("leveldb/open_tables", "open tables", stats.UnitDimensionless),

	IORead:  stats.Int64("leveldb/io_read", "total bytes read", stats.UnitBytes),
	IOWrite: stats.Int64("leveldb/io_write", "total bytes written", stats.UnitBytes),

	AliveSnapshots: stats.Int64("leveldb/alive_snapshots", "snapshots alive", stats.UnitDimensionless),
	AliveIterators: stats.Int64("leveldb/alive_iterators", "iterators alive", stats.UnitDimensionless),

	WriteDelay:         stats.Int64("leveldb/write_delay_count", "write delay events", stats.UnitDimensionless),
	WriteDelayDuration: stats.Int64("leveldb/write_delay_duration", "write delay duration", stats.UnitMilliseconds),

	WritePaused: stats.Int64("leveldb/write_paused", "whether or not write is currently paused", stats.UnitDimensionless),
}

var Views = struct {
	LevelSize           *view.View
	LevelTableCount     *view.View
	LevelRead           *view.View
	LevelWrite          *view.View
	LevelCompactionTime *view.View
	BlockCacheSize      *view.View
	OpenedTables        *view.View
	IORead              *view.View
	IOWrite             *view.View
	AliveSnapshots      *view.View
	AliveIterators      *view.View
	WriteDelay          *view.View
	WriteDelayDuration  *view.View
	WritePaused         *view.View
}{
	LevelSize: &view.View{
		Measure:     Metrics.LevelSize,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName, TagLevel},
	},
	LevelTableCount: &view.View{
		Measure:     Metrics.LevelTableCount,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName, TagLevel},
	},
	LevelRead: &view.View{
		Measure:     Metrics.LevelRead,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName, TagLevel},
	},
	LevelWrite: &view.View{
		Measure:     Metrics.LevelWrite,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName, TagLevel},
	},
	LevelCompactionTime: &view.View{
		Measure:     Metrics.LevelCompactionTime,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName, TagLevel},
	},
	BlockCacheSize: &view.View{
		Measure:     Metrics.BlockCacheSize,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName},
	},
	OpenedTables: &view.View{
		Measure:     Metrics.OpenedTables,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName},
	},
	IORead: &view.View{
		Measure:     Metrics.IORead,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName},
	},
	IOWrite: &view.View{
		Measure:     Metrics.IOWrite,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName},
	},
	AliveSnapshots: &view.View{
		Measure:     Metrics.AliveSnapshots,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName},
	},
	AliveIterators: &view.View{
		Measure:     Metrics.AliveIterators,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName},
	},
	WriteDelay: &view.View{
		Measure:     Metrics.WriteDelay,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName},
	},
	WriteDelayDuration: &view.View{
		Measure:     Metrics.WriteDelayDuration,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName},
	},
	WritePaused: &view.View{
		Measure:     Metrics.WritePaused,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{TagName},
	},
}

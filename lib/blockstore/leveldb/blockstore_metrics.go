package leveldbbs

import (
	"go.opencensus.io/stats"
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
	LevelSize:           stats.Int64("leveldb/level/size", "LSM level size", stats.UnitBytes),
	LevelTableCount:     stats.Int64("leveldb/level/tables", "LSM level tables", stats.UnitDimensionless),
	LevelRead:           stats.Int64("leveldb/level/read", "LSM level bytes read", stats.UnitBytes),
	LevelWrite:          stats.Int64("leveldb/level/written", "LSM level bytes written", stats.UnitBytes),
	LevelCompactionTime: stats.Int64("leveldb/level/compact_time", "LSM level compaction time", stats.UnitMilliseconds),

	BlockCacheSize: stats.Int64("leveldb/block_cache/size", "block cache size", stats.UnitBytes),
	OpenedTables:   stats.Int64("leveldb/open_tables", "open tables", stats.UnitDimensionless),

	IORead:  stats.Int64("leveldb/io/read", "total bytes read", stats.UnitBytes),
	IOWrite: stats.Int64("leveldb/io/written", "total bytes written", stats.UnitBytes),

	AliveSnapshots: stats.Int64("leveldb/alive_snapshots", "snapshots alive", stats.UnitDimensionless),
	AliveIterators: stats.Int64("leveldb/alive_iterators", "iterators alive", stats.UnitDimensionless),

	WriteDelay:         stats.Int64("leveldb/write/delay_count", "write delay events", stats.UnitDimensionless),
	WriteDelayDuration: stats.Int64("leveldb/write/delay_duration", "write delay duration", stats.UnitMilliseconds),

	WritePaused: stats.Int64("leveldb/write/paused", "whether or not write is currently paused", stats.UnitDimensionless),
}

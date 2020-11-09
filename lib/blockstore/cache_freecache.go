package blockstore

import (
	"context"
	"io"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"github.com/coocood/freecache"
)

var (
	HasFalse = byte(0)
	HasTrue  = byte(1)

	// Sentinel values for false and true.
	HasFalseVal = []byte{HasFalse}
	HasTrueVal  = []byte{HasTrue}
)

type FreecacheCachingBlockstore struct {
	blockCache  *freecache.Cache
	existsCache *freecache.Cache

	inner blockstore.Blockstore
}

var _ blockstore.Blockstore = (*FreecacheCachingBlockstore)(nil)

func WrapFreecacheCache(ctx context.Context, inner blockstore.Blockstore) (*FreecacheCachingBlockstore, error) {
	c := &FreecacheCachingBlockstore{
		blockCache:  freecache.NewCache(1 << 28), // 512MiB.
		existsCache: freecache.NewCache(1 << 26), // 64MiB.
		inner:       inner,
	}

	go func() {
		blockCacheTag, err := tag.New(ctx, tag.Insert(CacheName, "block_cache"))
		if err != nil {
			log.Warnf("blockstore metrics: failed to instantiate block cache tag: %s", err)
			return
		}
		existsCacheTag, err := tag.New(ctx, tag.Insert(CacheName, "exists_cache"))
		if err != nil {
			log.Warnf("blockstore metrics: failed to instantiate exists cache tag: %s", err)
			return
		}

		recordMetrics := func(ctx context.Context, c *freecache.Cache) {
			stats.Record(ctx,
				CacheMeasures.HitRatio.M(c.HitRate()),
				CacheMeasures.Hits.M(c.HitCount()),
				CacheMeasures.Misses.M(c.MissCount()),
				CacheMeasures.Entries.M(c.EntryCount()),
				CacheMeasures.Updates.M(c.OverwriteCount()),
				CacheMeasures.QueriesServed.M(c.LookupCount()),
				CacheMeasures.Evictions.M(c.EvacuateCount()),
			)
		}
		for {
			select {
			case <-time.After(CacheMetricsEmitInterval):
				recordMetrics(blockCacheTag, c.blockCache)
				recordMetrics(existsCacheTag, c.existsCache)
			case <-ctx.Done():
				return // yield
			}
		}
	}()

	return c, nil
}

// Close clears and closes all caches. It also closes the underlying blockstore,
// if it implements io.Closer.
func (c *FreecacheCachingBlockstore) Close() error {
	c.blockCache.Clear()
	c.existsCache.Clear()
	if closer, ok := c.inner.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (c *FreecacheCachingBlockstore) Get(cid cid.Cid) (blocks.Block, error) {
	k := []byte(cid.Hash())
	// check the has cache.
	if has, err := c.existsCache.Get(k); err == nil && has[0] == HasFalse {
		// we know we don't have the item; short-circuit.
		return nil, ErrNotFound
	}
	// check the block cache.
	if data, err := c.blockCache.Get(k); err == nil {
		return blocks.NewBlockWithCid(data, cid)
	}
	// fall back to the inner store.
	res, err := c.inner.Get(cid)
	if err != nil {
		if err == ErrNotFound {
			// inform the has cache that the item does not exist.
			_ = c.existsCache.Set(k, HasFalseVal, 0)
		}
		return res, err
	}
	_ = c.existsCache.Set(k, HasTrueVal, 0)
	_ = c.blockCache.Set(k, res.RawData(), 0)
	return res, err
}

func (c *FreecacheCachingBlockstore) GetSize(cid cid.Cid) (int, error) {
	k := []byte(cid.Hash())
	// check the has cache.
	if has, err := c.existsCache.Get(k); err == nil && has[0] == HasFalse {
		// we know we don't have the item; short-circuit.
		return -1, ErrNotFound
	}
	res, err := c.inner.GetSize(cid)
	if err != nil {
		if err == ErrNotFound {
			// inform the exists cache that the item does not exist.
			_ = c.existsCache.Set(k, HasFalseVal, 0)
		}
		return res, err
	}
	_ = c.existsCache.Set(k, HasTrueVal, 0)
	return res, err
}

func (c *FreecacheCachingBlockstore) Has(cid cid.Cid) (bool, error) {
	k := []byte(cid.Hash())
	if has, err := c.existsCache.Get(k); err == nil {
		return has[0] == HasTrue, nil
	}
	has, err := c.inner.Has(cid)
	if err != nil {
		return has, err
	}
	if has {
		_ = c.existsCache.Set(k, HasTrueVal, 0)
	} else {
		_ = c.existsCache.Set(k, HasFalseVal, 0)
	}
	return has, err
}

func (c *FreecacheCachingBlockstore) Put(block blocks.Block) error {
	k := []byte(block.Cid().Hash())
	if exists := c.probabilisticExists(k); exists {
		return nil
	}
	err := c.inner.Put(block)
	if err != nil {
		return err
	}
	_ = c.blockCache.Set(k, block.RawData(), 0)
	_ = c.existsCache.Set(k, HasTrueVal, 0)
	return err
}

func (c *FreecacheCachingBlockstore) PutMany(blks []blocks.Block) error {
	miss := make([]blocks.Block, 0, len(blks))
	for _, b := range blks {
		k := []byte(b.Cid().Hash())
		if c.probabilisticExists(k) {
			continue
		}
		miss = append(miss, b)
	}
	if len(miss) == 0 {
		// nothing to add.
		return nil
	}

	err := c.inner.PutMany(miss)
	if err != nil {
		return err
	}
	for _, b := range miss {
		k := []byte(b.Cid().Hash())
		_ = c.blockCache.Set(k, b.RawData(), 0)
		_ = c.existsCache.Set(k, HasTrueVal, 0)
	}
	return err
}

func (c *FreecacheCachingBlockstore) AllKeysChan(ctx context.Context) (<-chan cid.Cid, error) {
	return c.inner.AllKeysChan(ctx)
}

func (c *FreecacheCachingBlockstore) DeleteBlock(cid cid.Cid) error {
	k := []byte(cid.Hash())
	err := c.inner.DeleteBlock(cid)
	if err != nil {
		return err
	}
	c.blockCache.Del(k)
	_ = c.existsCache.Set(k, HasFalseVal, 0)
	return err
}

func (c *FreecacheCachingBlockstore) probabilisticExists(k []byte) bool {
	if has, err := c.existsCache.Get(k); err == nil {
		return has[0] == HasTrue
	}
	// may have paged out of the exists cache, but still present in the block cache.
	if _, err := c.blockCache.Get(k); err == nil {
		_ = c.existsCache.Set(k, HasTrueVal, 0) // update the exists cache.
		return true
	}
	// NOTE: we _could_ check the size cache, but if two caches have already
	// missed, it's likely that the size cache would miss too.
	return false
}

func (c *FreecacheCachingBlockstore) HashOnRead(enabled bool) {
	c.inner.HashOnRead(enabled)
}

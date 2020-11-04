package blockstore

import (
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"golang.org/x/xerrors"
)

type RistrettoCachingBlockstore struct {
	cache *ristretto.Cache

	inner blockstore.Blockstore
}

var _ blockstore.Blockstore = (*RistrettoCachingBlockstore)(nil)

func WrapRistrettoCache(inner blockstore.Blockstore) (*RistrettoCachingBlockstore, error) {
	opts := &ristretto.Config{
		NumCounters: 10_000_000, // assumes we're going to be storing 1MM objects (docs say to x10 that)
		MaxCost:     1 << 27,    // 256MiB.
		BufferItems: 64,
		Metrics:     true,
	}

	cache, err := ristretto.NewCache(opts)
	if err != nil {
		return nil, xerrors.Errorf("failed to create ristretto cache: %w", err)
	}

	c := &RistrettoCachingBlockstore{
		cache: cache,
		inner: inner,
	}

	go func() {
		for range time.Tick(2 * time.Second) {
			fmt.Println(cache.Metrics.String())
		}
	}()

	return c, nil
}

func (c *RistrettoCachingBlockstore) Get(cid cid.Cid) (blocks.Block, error) {
	if obj, ok := c.cache.Get([]byte(cid.Hash())); ok {
		return obj.(blocks.Block), nil
	}
	res, err := c.inner.Get(cid)
	if err != nil {
		return res, err
	}
	_ = c.cache.Set([]byte(cid.Hash()), res, int64(len(res.RawData())))
	return res, err
}

func (c *RistrettoCachingBlockstore) GetSize(cid cid.Cid) (int, error) {
	return c.inner.GetSize(cid)
}

func (c *RistrettoCachingBlockstore) Has(cid cid.Cid) (bool, error) {
	return c.inner.Has(cid)
}

func (c *RistrettoCachingBlockstore) Put(block blocks.Block) error {
	if _, ok := c.cache.Get([]byte(block.Cid().Hash())); ok {
		return nil
	}
	err := c.inner.Put(block)
	if err != nil {
		return err
	}
	_ = c.cache.Set([]byte(block.Cid().Hash()), block, int64(len(block.RawData())))
	return err
}

func (c *RistrettoCachingBlockstore) PutMany(blks []blocks.Block) error {
	miss := make([]blocks.Block, 0, len(blks))
	for _, b := range blks {
		if _, ok := c.cache.Get([]byte(b.Cid().Hash())); ok {
			continue
		}
		miss = append(miss, b)
	}
	if len(miss) == 0 {
		return nil
	}

	err := c.inner.PutMany(miss)
	if err != nil {
		return err
	}
	for _, b := range miss {
		_ = c.cache.Set([]byte(b.Cid().Hash()), b, int64(len(b.RawData())))
	}
	return err
}

func (c *RistrettoCachingBlockstore) AllKeysChan(ctx context.Context) (<-chan cid.Cid, error) {
	return c.inner.AllKeysChan(ctx)
}

func (c *RistrettoCachingBlockstore) DeleteBlock(cid cid.Cid) error {
	c.cache.Del(cid)
	return c.inner.DeleteBlock(cid)
}

func (c *RistrettoCachingBlockstore) HashOnRead(enabled bool) {
	c.inner.HashOnRead(enabled)
}

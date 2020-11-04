package blockstore

import (
	"context"

	"github.com/dgraph-io/ristretto"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"golang.org/x/xerrors"
)

type CachingBlockstore struct {
	cache *ristretto.Cache

	inner blockstore.Blockstore
}

var _ blockstore.Blockstore = (*CachingBlockstore)(nil)

func WrapCaching(inner blockstore.Blockstore) (*CachingBlockstore, error) {
	opts := &ristretto.Config{
		NumCounters: 10_000_000,
		MaxCost:     1 << 28,
		BufferItems: 64,
		Metrics:     true,
	}

	cache, err := ristretto.NewCache(opts)
	if err != nil {
		return nil, xerrors.Errorf("failed to create ristretto cache: %w", err)
	}

	c := &CachingBlockstore{
		cache: cache,
		inner: inner,
	}

	return c, nil
}

func (c *CachingBlockstore) Get(cid cid.Cid) (blocks.Block, error) {
	if obj, ok := c.cache.Get(cid); ok {
		return obj.(blocks.Block), nil
	}
	res, err := c.inner.Get(cid)
	if err != nil {
		return res, err
	}
	_ = c.cache.Set(cid, res, int64(len(res.RawData())))
	return res, err
}

func (c *CachingBlockstore) GetSize(cid cid.Cid) (int, error) {
	return c.inner.GetSize(cid)
}

func (c *CachingBlockstore) Has(cid cid.Cid) (bool, error) {
	return c.inner.Has(cid)
}

func (c *CachingBlockstore) Put(block blocks.Block) error {
	if _, ok := c.cache.Get(block.Cid()); ok {
		return nil
	}
	err := c.inner.Put(block)
	if err != nil {
		return err
	}
	_ = c.cache.Set(block.Cid(), block, int64(len(block.RawData())))
	return err
}

func (c *CachingBlockstore) PutMany(blks []blocks.Block) error {
	miss := make([]blocks.Block, 0, len(blks))
	for _, b := range blks {
		if _, ok := c.cache.Get(b.Cid()); ok {
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
		_ = c.cache.Set(b.Cid(), b, int64(len(b.RawData())))
	}
	return err
}

func (c *CachingBlockstore) AllKeysChan(ctx context.Context) (<-chan cid.Cid, error) {
	return c.inner.AllKeysChan(ctx)
}

func (c *CachingBlockstore) DeleteBlock(cid cid.Cid) error {
	c.cache.Del(cid)
	return c.inner.DeleteBlock(cid)
}

func (c *CachingBlockstore) HashOnRead(enabled bool) {
	c.inner.HashOnRead(enabled)
}

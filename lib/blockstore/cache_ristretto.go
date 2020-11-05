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
	blockCache *ristretto.Cache
	hasCache   *ristretto.Cache
	sizeCache  *ristretto.Cache

	inner blockstore.Blockstore
}

var _ blockstore.Blockstore = (*RistrettoCachingBlockstore)(nil)

func WrapRistrettoCache(inner blockstore.Blockstore) (*RistrettoCachingBlockstore, error) {
	blockCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10_000_000, // assumes we're going to be storing 1MM objects (docs say to x10 that)
		MaxCost:     1 << 27,    // 256MiB.
		BufferItems: 64,
		Metrics:     true,
	})
	if err != nil {
		return nil, xerrors.Errorf("failed to create ristretto block cache: %w", err)
	}

	sizeCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10_000_000, // assumes we're going to be storing 1MM objects (docs say to x10 that)
		MaxCost:     1 << 23,    // 8MiB.
		BufferItems: 64,
		Metrics:     true,
	})
	if err != nil {
		return nil, xerrors.Errorf("failed to create ristretto size cache: %w", err)
	}

	hasCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10_000_000, // assumes we're going to be storing 1MM objects (docs say to x10 that)
		MaxCost:     1 << 20,    // 1MiB.
		BufferItems: 64,
		Metrics:     true,
	})
	if err != nil {
		return nil, xerrors.Errorf("failed to create ristretto has cache: %w", err)
	}

	c := &RistrettoCachingBlockstore{
		blockCache: blockCache,
		sizeCache:  sizeCache,
		hasCache:   hasCache,
		inner:      inner,
	}

	go func() {
		for range time.Tick(2 * time.Second) {
			fmt.Println("block cache:", blockCache.Metrics.String())
			fmt.Println("size cache:", sizeCache.Metrics.String())
			fmt.Println("has cache:", hasCache.Metrics.String())
		}
	}()

	return c, nil
}

func (c *RistrettoCachingBlockstore) Get(cid cid.Cid) (blocks.Block, error) {
	b := []byte(cid.Hash())
	if obj, ok := c.blockCache.Get(b); ok {
		return obj.(blocks.Block), nil
	}
	res, err := c.inner.Get(cid)
	if err != nil {
		return res, err
	}
	c.hasCache.Set(b, true, 1)
	_ = c.blockCache.Set(b, res, int64(len(res.RawData())))
	return res, err
}

func (c *RistrettoCachingBlockstore) GetSize(cid cid.Cid) (int, error) {
	return c.inner.GetSize(cid)
}

func (c *RistrettoCachingBlockstore) Has(cid cid.Cid) (bool, error) {
	return c.inner.Has(cid)
}

func (c *RistrettoCachingBlockstore) Put(block blocks.Block) error {
	if _, ok := c.blockCache.Get([]byte(block.Cid().Hash())); ok {
		return nil
	}
	err := c.inner.Put(block)
	if err != nil {
		return err
	}
	_ = c.blockCache.Set([]byte(block.Cid().Hash()), block, int64(len(block.RawData())))
	return err
}

func (c *RistrettoCachingBlockstore) PutMany(blks []blocks.Block) error {
	miss := make([]blocks.Block, 0, len(blks))
	for _, b := range blks {
		if _, ok := c.blockCache.Get([]byte(b.Cid().Hash())); ok {
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
		_ = c.blockCache.Set([]byte(b.Cid().Hash()), b, int64(len(b.RawData())))
	}
	return err
}

func (c *RistrettoCachingBlockstore) AllKeysChan(ctx context.Context) (<-chan cid.Cid, error) {
	return c.inner.AllKeysChan(ctx)
}

func (c *RistrettoCachingBlockstore) DeleteBlock(cid cid.Cid) error {
	c.blockCache.Del(cid)
	return c.inner.DeleteBlock(cid)
}

func (c *RistrettoCachingBlockstore) HashOnRead(enabled bool) {
	c.inner.HashOnRead(enabled)
}

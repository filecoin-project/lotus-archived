package leveldbbs

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	logger "github.com/ipfs/go-log/v2"

	"github.com/filecoin-project/lotus/lib/blockstore"
)

var (
	// MetricsFrequency defines the frequency at which we'll record leveldb metrics.
	MetricsFrequency = 5 * time.Second

	log = logger.Logger("leveldbbs")
)

// Options is an alias of syndtr/goleveldb/opt.Options which might be extended
// in the future.
type Options struct {
	opt.Options

	// Name is the name by which to identify this blockstore in the system.
	// It is used to namespace metrics.
	Name string
}

// Blockstore is a leveldb-backed IPLD blockstore.
type Blockstore struct {
	DB *leveldb.DB

	// status guarded by atomic; 0 for active, 1 for closed.
	status  int32
	closing chan struct{}
	wg      sync.WaitGroup
}

var _ blockstore.Blockstore = (*Blockstore)(nil)
var _ blockstore.Viewer = (*Blockstore)(nil)
var _ io.Closer = (*Blockstore)(nil)

func DefaultOptions(name string) *Options {
	var opts Options
	opts.Filter = filter.NewBloomFilter(10)
	opts.BlockCacheCapacity = 512 << 20
	opts.WriteBuffer = 16 << 20
	opts.DisableSeeksCompaction = true
	opts.Name = name
	return &opts
}

// Open creates a new leveldb-backed blockstore, with the supplied options.
func Open(path string, opts *Options) (*Blockstore, error) {
	var err error
	var db *leveldb.DB

	if path == "" {
		db, err = leveldb.Open(storage.NewMemStorage(), &opts.Options)
	} else {
		db, err = leveldb.OpenFile(path, &opts.Options)
		if errors.IsCorrupted(err) && !opts.GetReadOnly() {
			log.Warnf("leveldb blockstore appears corrupted; recovering")
			db, err = leveldb.RecoverFile(path, &opts.Options)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb blockstore: %w", err)
	}

	bs := &Blockstore{
		DB:      db,
		closing: make(chan struct{}),
	}

	ctx := context.Background()
	if opts.Name != "" {
		var err error
		ctx, err = tag.New(context.Background(), tag.Insert(TagName, opts.Name))
		if err != nil {
			log.Warnf("failed to instantiate metrics tag; metrics will be untagged; err: %s", err)
		}
	}

	go bs.recordMetrics(ctx, MetricsFrequency)

	return bs, nil
}

func (b *Blockstore) recordMetrics(ctx context.Context, freq time.Duration) {
	defer b.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Warnf("metrics crashed; reporting interrupted: %s", r)
		}
	}()

	var st leveldb.DBStats
	for {
		select {
		case <-time.After(freq):
			if err := b.DB.Stats(&st); err != nil {
				log.Warnf("failed to acquire leveldb stats: %s", err)
				continue
			}

			var writePaused int64
			if st.WritePaused {
				writePaused = 1
			}

			// record simple metrics first.
			stats.Record(ctx,
				Metrics.BlockCacheSize.M(int64(st.BlockCacheSize)),
				Metrics.IORead.M(int64(st.IORead)),
				Metrics.IOWrite.M(int64(st.IOWrite)),
				Metrics.AliveIterators.M(int64(st.AliveIterators)),
				Metrics.AliveSnapshots.M(int64(st.AliveSnapshots)),
				Metrics.OpenedTables.M(int64(st.OpenedTablesCount)),
				Metrics.WriteDelay.M(int64(st.WriteDelayCount)),
				Metrics.WriteDelayDuration.M(st.WriteDelayDuration.Milliseconds()),
				Metrics.WritePaused.M(writePaused),
			)

			// looking at source, level metrics are guaranteed to be equal
			// length; just in case, we recover from panics.
			for i := 0; i < len(st.LevelRead); i++ {
				if err := stats.RecordWithTags(ctx,
					[]tag.Mutator{tag.Insert(TagLevel, strconv.Itoa(i))},
					Metrics.LevelRead.M(st.LevelRead[i]),
					Metrics.LevelWrite.M(st.LevelWrite[i]),
					Metrics.LevelTableCount.M(int64(st.LevelTablesCounts[i])),
					Metrics.LevelSize.M(st.LevelSizes[i]),
					Metrics.LevelCompactionTime.M(st.LevelDurations[i].Milliseconds()),
				); err != nil {
					log.Warnf("failed to record metrics for level %d: %s", i, err)
				}
			}

		case <-b.closing:
			return
		}
	}
}

// Close closes the store. If the store has already been closed, this noops and
// returns an error, even if the first closure resulted in error.
func (b *Blockstore) Close() error {
	if !atomic.CompareAndSwapInt32(&b.status, 0, 1) {
		// already closed, or closing.
		b.wg.Wait()
		return nil
	}

	close(b.closing)
	b.wg.Wait()

	return b.DB.Close()
}

// View implements blockstore.Viewer, which leverages zero-copy read-only
// access to values.
func (b *Blockstore) View(cid cid.Cid, fn func([]byte) error) error {
	switch v, err := b.DB.Get(cid.Hash(), nil); err {
	case nil:
		return fn(v)
	case leveldb.ErrNotFound:
		return blockstore.ErrNotFound
	default:
		return fmt.Errorf("failed to view block from leveldb blockstore: %w", err)
	}
}

// Has implements Blockstore.Has.
func (b *Blockstore) Has(cid cid.Cid) (bool, error) {
	switch found, err := b.DB.Has(cid.Hash(), nil); err {
	case nil:
		return found, nil
	default:
		return false, fmt.Errorf("failed to check if block exists in leveldb blockstore: %w", err)
	}
}

// Get implements Blockstore.Get.
func (b *Blockstore) Get(cid cid.Cid) (blocks.Block, error) {
	if !cid.Defined() {
		return nil, blockstore.ErrNotFound
	}

	switch v, err := b.DB.Get(cid.Hash(), nil); err {
	case nil:
		return blocks.NewBlockWithCid(v, cid)
	case leveldb.ErrNotFound:
		return nil, blockstore.ErrNotFound
	default:
		return nil, fmt.Errorf("failed to get block from leveldb blockstore: %w", err)
	}
}

// GetSize implements Blockstore.GetSize.
func (b *Blockstore) GetSize(cid cid.Cid) (int, error) {
	switch v, err := b.DB.Get(cid.Hash(), nil); err {
	case nil:
		return len(v), nil
	case leveldb.ErrNotFound:
		return -1, blockstore.ErrNotFound
	default:
		return -1, fmt.Errorf("failed to get size of block from leveldb blockstore: %w", err)
	}
}

// Put implements Blockstore.Put.
func (b *Blockstore) Put(block blocks.Block) error {
	err := b.DB.Put(block.Cid().Hash(), block.RawData(), nil)
	if err != nil {
		err = fmt.Errorf("failed to put block in leveldb blockstore: %w", err)
	}

	return err
}

// PutMany implements Blockstore.PutMany.
func (b *Blockstore) PutMany(blocks []blocks.Block) error {
	batch := new(leveldb.Batch)
	for _, blk := range blocks {
		batch.Put(blk.Cid().Hash(), blk.RawData())
	}

	err := b.DB.Write(batch, nil)
	if err != nil {
		err = fmt.Errorf("failed to put many blocks in leveldb blockstore: %w", err)
	}
	return err
}

// DeleteBlock implements Blockstore.DeleteBlock.
func (b *Blockstore) DeleteBlock(cid cid.Cid) error {
	err := b.DB.Delete(cid.Hash(), nil)
	if err != nil {
		err = fmt.Errorf("failed to delete block from leveldb blockstore: %w", err)
	}
	return err
}

// AllKeysChan implements Blockstore.AllKeysChan.
func (b *Blockstore) AllKeysChan(ctx context.Context) (<-chan cid.Cid, error) {
	it := b.DB.NewIterator(nil, nil)
	ch := make(chan cid.Cid)
	go func() {
		defer close(ch)
		defer it.Release()

		for it.Next() {
			if ctx.Err() != nil {
				return // context has fired.
			}

			select {
			case ch <- cid.NewCidV1(cid.Raw, it.Key()):
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// HashOnRead implements Blockstore.HashOnRead. It is not supported by this
// blockstore.
func (b *Blockstore) HashOnRead(_ bool) {
	log.Warnf("called HashOnRead on leveldb blockstore; function not supported; ignoring")
}

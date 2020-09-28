package store_test

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/filecoin-project/lotus/chain/types"

	delay "github.com/ipfs/go-ipfs-delay"

	"github.com/ipfs/go-datastore/delayed"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/gen"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types/mock"
	"github.com/filecoin-project/lotus/lib/blockstore"
	datastore "github.com/ipfs/go-datastore"
	syncds "github.com/ipfs/go-datastore/sync"
	"github.com/stretchr/testify/assert"
)

func TestIndexSeeks(t *testing.T) {
	cg, err := gen.NewGenerator()
	if err != nil {
		t.Fatal(err)
	}

	gencar, err := cg.GenesisCar()
	if err != nil {
		t.Fatal(err)
	}

	gen := cg.Genesis()

	ctx := context.TODO()

	nbs := blockstore.NewTemporarySync()
	cs := store.NewChainStore(nbs, syncds.MutexWrap(datastore.NewMapDatastore()), nil)

	_, err = cs.Import(bytes.NewReader(gencar))
	if err != nil {
		t.Fatal(err)
	}

	cur := mock.TipSet(gen)
	if err := cs.PutTipSet(ctx, mock.TipSet(gen)); err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, cs.SetGenesis(gen))

	// Put 113 blocks from genesis
	for i := 0; i < 113; i++ {
		nextts := mock.TipSet(mock.MkBlock(cur, 1, 1))

		if err := cs.PutTipSet(ctx, nextts); err != nil {
			t.Fatal(err)
		}
		cur = nextts
	}

	// Put 50 null epochs + 1 block
	skip := mock.MkBlock(cur, 1, 1)
	skip.Height += 50

	skipts := mock.TipSet(skip)

	if err := cs.PutTipSet(ctx, skipts); err != nil {
		t.Fatal(err)
	}

	ts, err := cs.GetTipsetByHeight(ctx, skip.Height-10, skipts, false)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, abi.ChainEpoch(164), ts.Height())

	for i := 0; i <= 113; i++ {
		ts3, err := cs.GetTipsetByHeight(ctx, abi.ChainEpoch(i), skipts, false)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, abi.ChainEpoch(i), ts3.Height())
	}
}

func TestGetTipsetByHeight(b *testing.T) {
	runCount := 20
	datastoreLatency := 100 * time.Microsecond
	tipsetCount := 50000
	rnd := rand.New(rand.NewSource(2))

	cg, err := gen.NewGenerator()
	if err != nil {
		b.Fatal(err)
	}

	gencar, err := cg.GenesisCar()
	if err != nil {
		b.Fatal(err)
	}

	gen := cg.Genesis()

	ctx := context.TODO()

	dl := delay.Fixed(0)
	bsds := syncds.MutexWrap(delayed.New(datastore.NewMapDatastore(), dl))
	nbs := blockstore.NewBlockstore(bsds)
	cs := store.NewChainStore(nbs, syncds.MutexWrap(datastore.NewMapDatastore()), nil)

	_, err = cs.Import(bytes.NewReader(gencar))
	if err != nil {
		b.Fatal(err)
	}

	cur := mock.TipSet(gen)
	if err := cs.PutTipSet(ctx, mock.TipSet(gen)); err != nil {
		b.Fatal(err)
	}
	assert.NoError(b, cs.SetGenesis(gen))

	fmt.Printf("Creating %d tipsets\n", tipsetCount)
	for i := 0; i < tipsetCount; i++ {
		// Add an average of 5 blocks to the tipset
		var blks []*types.BlockHeader
		tipsetWidth := 1 + rnd.Intn(9)
		for tsi := 0; tsi < tipsetWidth; tsi++ {
			blks = append(blks, mock.MkBlock(cur, 1, uint64(tsi)))
		}

		nextts := mock.TipSet(blks...)
		if err := cs.PutTipSet(ctx, nextts); err != nil {
			b.Fatal(err)
		}
		cur = nextts

		if (i+1)%10000 == 0 {
			fmt.Printf("  %d / %d\n", i+1, tipsetCount)
		}
	}

	cs = store.NewChainStore(nbs, syncds.MutexWrap(datastore.NewMapDatastore()), nil)
	chainStoreIndex := store.NewChainIndex(cs.LoadTipSet, cs.LoadFirstBlock)

	fmt.Printf("Using datastore with %s latency\n", datastoreLatency)
	dl.Set(datastoreLatency)

	getTSByHeight := func(to abi.ChainEpoch) {
		t := time.Now()

		_, err := chainStoreIndex.GetTipsetByHeight(ctx, cur, to)
		if err != nil {
			b.Fatal(err)
		}

		elapsed := time.Since(t)
		fmt.Printf("  height %d: %s\n", to, elapsed)
	}

	fmt.Println("\nRunning GetTipsetByHeight at regular intervals:")
	for i := 0; i < runCount; i++ {
		to := abi.ChainEpoch(tipsetCount - i*(tipsetCount/runCount))
		getTSByHeight(to)
	}

	fmt.Println("\nSample GetTipsetByHeight:")
	sampleEpochs := []abi.ChainEpoch{
		abi.ChainEpoch(tipsetCount - 10),
		abi.ChainEpoch(tipsetCount/2 - 10),
		abi.ChainEpoch(tipsetCount/4 - 10),
		abi.ChainEpoch(tipsetCount/10 - 10),
	}
	for _, to := range sampleEpochs {
		getTSByHeight(to)
	}
}

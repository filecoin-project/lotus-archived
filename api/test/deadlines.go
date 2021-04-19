package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/network"
	cbor "github.com/ipfs/go-ipld-cbor"

	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/extern/sector-storage/mock"
	"github.com/filecoin-project/lotus/node/impl"
)

type dltest struct {
	t   *testing.T
	ctx context.Context

	client                 *impl.FullNodeAPI
	minerA, minerB, minerC TestStorageNode
	maddrA, maddrB, maddrC address.Address
}

// TestDeadlineToggling:
// * spins up a v3 network (miner A)
// * creates an inactive miner (miner B)
// * creates another miner, pledges a sector, waits for power (miner C)
//
// * goes through v4 upgrade
// * goes through PP
// * creates minerD
// * makes sure that miner B/D are inactive, A/C still are
// * pledges sectors on miner B/D
// * disables post on miner C
// * goes through PP
// * asserts that miner C loses power
// * asserts that miner B/D is active and has power
// * disables post on miner B
// * goes through another PP
// * asserts that miner B loses power
func TestDeadlineToggling(t *testing.T, b APIBuilder, blocktime time.Duration) {
	var upgradeH abi.ChainEpoch = 4000
	var provingPeriod abi.ChainEpoch = 2880

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n, sn := b(t, []FullNodeOpts{FullNodeWithActorsV4At(upgradeH)}, OneMiner)

	client := n[0].FullNode.(*impl.FullNodeAPI)
	minerA := sn[0]

	{
		addrinfo, err := client.NetAddrsListen(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if err := minerA.NetConnect(ctx, addrinfo); err != nil {
			t.Fatal(err)
		}
	}

	defaultFrom, err := client.WalletDefaultAddress(ctx)
	require.NoError(t, err)

	maddrA, err := minerA.ActorAddress(ctx)
	require.NoError(t, err)

	build.Clock.Sleep(time.Second)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for ctx.Err() == nil {
			build.Clock.Sleep(blocktime)
			if err := minerA.MineOne(ctx, MineNext); err != nil {
				if ctx.Err() != nil {
					// context was canceled, ignore the error.
					return
				}
				t.Error(err)
			}
		}
	}()
	defer func() {
		cancel()
		<-done
	}()

	minerB := n[0].Stb(ctx, t, TestSpt, defaultFrom)
	minerC := n[0].Stb(ctx, t, TestSpt, defaultFrom)

	maddrB, err := minerB.ActorAddress(ctx)
	require.NoError(t, err)
	maddrC, err := minerC.ActorAddress(ctx)
	require.NoError(t, err)

	ssz, err := minerC.ActorSectorSize(ctx, maddrC)
	require.NoError(t, err)

	// pledge sectors on C, go through a PP, check for power
	{
		pledgeSectors(t, ctx, minerC, 10, 0, nil)

		di, err := client.StateMinerProvingDeadline(ctx, maddrC, types.EmptyTSK)
		require.NoError(t, err)

		fmt.Printf("Running one proving period (miner C)\n")
		fmt.Printf("End for head.Height > %d\n", di.PeriodStart+di.WPoStProvingPeriod*2)

		for {
			head, err := client.ChainHead(ctx)
			require.NoError(t, err)

			if head.Height() > di.PeriodStart+di.WPoStProvingPeriod*2 {
				fmt.Printf("Now head.Height = %d\n", head.Height())
				break
			}
			build.Clock.Sleep(blocktime)
		}

		expectedPower := types.NewInt(uint64(ssz) * 10)

		p, err := client.StateMinerPower(ctx, maddrC, types.EmptyTSK)
		require.NoError(t, err)

		// make sure it has gained power.
		require.Equal(t, p.MinerPower.RawBytePower, expectedPower)
	}

	// go through upgrade + PP
	for {
		head, err := client.ChainHead(ctx)
		require.NoError(t, err)

		if head.Height() > upgradeH+provingPeriod {
			fmt.Printf("Now head.Height = %d\n", head.Height())
			break
		}
		build.Clock.Sleep(blocktime)
	}

	nv, err := client.StateNetworkVersion(ctx, types.EmptyTSK)
	require.NoError(t, err)
	require.GreaterOrEqual(t, nv, network.Version12)

	minerD := n[0].Stb(ctx, t, TestSpt, defaultFrom)

	maddrD, err := minerD.ActorAddress(ctx)
	require.NoError(t, err)

	checkMiner := func(m TestStorageNode, ma address.Address, power abi.StoragePower, active bool) {
		p, err := client.StateMinerPower(ctx, ma, types.EmptyTSK)
		require.NoError(t, err)

		// make sure it has gained power.
		require.Equal(t, p.MinerPower.RawBytePower, power)

		mact, err := client.StateGetActor(ctx, ma, types.EmptyTSK)
		require.NoError(t, err)

		mst, err := miner.Load(adt.WrapStore(ctx, cbor.NewCborStore(blockstore.NewAPIBlockstore(client))), mact)
		require.NoError(t, err)

		act, err := mst.DeadlineCronActive()
		require.NoError(t, err)
		require.Equal(t, active, act)
	}

	// first round of miner checks
	checkMiner(minerA, maddrA, types.NewInt(uint64(ssz)*2), true)
	checkMiner(minerC, maddrC, types.NewInt(uint64(ssz)*10), true)

	checkMiner(minerB, maddrB, types.NewInt(0), false)
	checkMiner(minerD, maddrD, types.NewInt(0), false)

	// pledge sectors on minerB/minerD, stop post on minerC
	pledgeSectors(t, ctx, minerB, 8, 0, nil)
	checkMiner(minerB, maddrB, types.NewInt(0), true)

	pledgeSectors(t, ctx, minerD, 9, 0, nil)
	checkMiner(minerD, maddrD, types.NewInt(0), true)

	minerC.StorageMiner.(*impl.StorageMinerAPI).IStorageMgr.(*mock.SectorMgr).Fail()

	// go through another PP
	for {
		head, err := client.ChainHead(ctx)
		require.NoError(t, err)

		if head.Height() > upgradeH+(provingPeriod*3) {
			fmt.Printf("Now head.Height = %d\n", head.Height())
			break
		}
		build.Clock.Sleep(blocktime)
	}

	// second round of miner checks
	checkMiner(minerA, maddrA, types.NewInt(uint64(ssz)*2), true)
	checkMiner(minerC, maddrC, types.NewInt(0), true)
	checkMiner(minerB, maddrB, types.NewInt(uint64(ssz)*8), true)
	checkMiner(minerB, maddrD, types.NewInt(uint64(ssz)*9), true)

	// disable post on minerB
	minerB.StorageMiner.(*impl.StorageMinerAPI).IStorageMgr.(*mock.SectorMgr).Fail()

	// go through another PP
	for {
		head, err := client.ChainHead(ctx)
		require.NoError(t, err)

		if head.Height() > upgradeH+(provingPeriod*5) {
			fmt.Printf("Now head.Height = %d\n", head.Height())
			break
		}
		build.Clock.Sleep(blocktime)
	}

	// third round of miner checks
	checkMiner(minerA, maddrA, types.NewInt(uint64(ssz)*2), true)
	checkMiner(minerC, maddrC, types.NewInt(0), true)
	checkMiner(minerB, maddrB, types.NewInt(0), true)
	checkMiner(minerB, maddrD, types.NewInt(uint64(ssz)*9), true)
}

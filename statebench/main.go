package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	logging "github.com/ipfs/go-log"
	cbg "github.com/whyrusleeping/cbor-gen"

	"github.com/filecoin-project/lotus/chain/gen"
	"github.com/filecoin-project/lotus/chain/stmgr"
	"github.com/filecoin-project/lotus/chain/types"

	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/builtin/market"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
	"github.com/filecoin-project/specs-actors/actors/crypto"
)

func mustEnc(v cbg.CBORMarshaler) []byte {

	buf := new(bytes.Buffer)
	if err := v.MarshalCBOR(buf); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func main() {
	logging.SetDebugLogging()
	ctx := context.TODO()
	cg, err := gen.NewGenerator()
	if err != nil {
		panic(err)
	}

	signMessage := func(m *types.Message) *types.SignedMessage {
		sig, err := cg.Wallet().Sign(context.TODO(), m.From, m.Cid().Bytes())
		if err != nil {
			panic(err)
		}

		return &types.SignedMessage{
			Message:   *m,
			Signature: *sig,
		}
	}

	var dealMakers []address.Address
	for i := 0; i < 30; i++ {
		a, err := cg.Wallet().GenerateKey(crypto.SigTypeSecp256k1)
		if err != nil {
			panic(err)
		}
		dealMakers = append(dealMakers, a)
		fmt.Println("deal maker: ", a)
	}

	var messages []*types.SignedMessage
	cg.GetMessages = func(cg *gen.ChainGen) ([]*types.SignedMessage, error) {
		m := messages
		messages = nil
		return m, nil
	}

	sm := stmgr.NewStateManager(cg.ChainStore())
	banker, err := sm.GetActor(cg.Banker(), cg.CurTipset.TipSet())
	if err != nil {
		panic(err)
	}
	bnonce := banker.Nonce
	fmt.Println("banker balance: ", banker.Balance)

	for _, dm := range dealMakers {
		sm1 := signMessage(&types.Message{
			From:     cg.Banker(),
			To:       dm,
			Nonce:    bnonce,
			Value:    types.NewInt(11000000),
			GasLimit: types.NewInt(10000),
			GasPrice: types.NewInt(0),
		})
		bnonce++

		sm2 := signMessage(&types.Message{
			From:     dm,
			To:       builtin.StorageMarketActorAddr,
			Method:   builtin.MethodsMarket.AddBalance,
			Value:    types.NewInt(10000000),
			Params:   mustEnc(&dm),
			GasLimit: types.NewInt(10000),
			GasPrice: types.NewInt(0),
		})

		messages = append(messages, sm1, sm2)
	}

	minerAddr := cg.Miners[0]
	messages = append(messages, signMessage(&types.Message{
		From:   cg.Banker(),
		To:     builtin.StorageMarketActorAddr,
		Method: builtin.MethodsMarket.AddBalance,
		Nonce:  bnonce,

		Value:    types.NewInt(1000000000),
		Params:   mustEnc(&minerAddr),
		GasLimit: types.NewInt(10000),
		GasPrice: types.NewInt(0),
	}))

	if _, err := cg.NextTipSet(); err != nil {
		panic(err)
	}

	worker, err := stmgr.GetMinerWorker(ctx, sm, cg.CurTipset.TipSet(), minerAddr)
	if err != nil {
		panic(err)
	}

	var nextSectorID abi.SectorNumber
	{
		sectors, err := stmgr.GetMinerSectorSet(ctx, sm, cg.CurTipset.TipSet(), minerAddr)
		if err != nil {
			panic(err)
		}
		last := sectors[len(sectors)-1]
		nextSectorID = last.ID + 1
	}
	proveCommitAt := make(map[abi.ChainEpoch][]abi.SectorNumber)
	nextDealID := abi.DealID(2) // THIS IS UNGLY but it is correct

	for i := 0; i < 100; i++ {
		curEpoch := cg.CurTipset.TipSet().Height()
		fmt.Println("Block: ", curEpoch)

		wact, err := sm.GetActor(worker, cg.CurTipset.TipSet())
		if err != nil {
			panic(err)
		}

		wnonce := wact.Nonce
		for _, dm := range dealMakers {
			{
				deal := &market.ClientDealProposal{
					Proposal: market.DealProposal{
						PieceCID:  builtin.AccountActorCodeID, // whatever CID, its not like we're checking
						PieceSize: 128,
						Client:    dm,
						Provider:  minerAddr,

						StartEpoch:           curEpoch + 14, // this number is weird
						EndEpoch:             curEpoch + 200,
						StoragePricePerEpoch: big.NewInt(1),

						ProviderCollateral: big.NewInt(1),
						ClientCollateral:   big.NewInt(1),
					}}

				propBytes := mustEnc(&deal.Proposal)
				sig, err := cg.Wallet().Sign(ctx, dm, propBytes)
				if err != nil {
					panic(err)
				}
				deal.ClientSignature.Type = sig.Type
				deal.ClientSignature.Data = sig.Data

				params := mustEnc(&market.PublishStorageDealsParams{Deals: []market.ClientDealProposal{*deal}})

				sm := signMessage(&types.Message{
					From:     worker,
					To:       builtin.StorageMarketActorAddr,
					Method:   builtin.MethodsMarket.PublishStorageDeals,
					Value:    types.NewInt(0),
					GasLimit: types.NewInt(10000000),
					GasPrice: types.NewInt(0),
					Nonce:    wnonce,
					Params:   params,
				})
				messages = append(messages, sm)
				wnonce++
			}

			{
				params := mustEnc(&miner.SectorPreCommitInfo{
					RegisteredProof: abi.RegisteredProof_StackedDRG2KiBSeal,
					SectorNumber:    abi.SectorNumber(nextSectorID),
					SealedCID:       builtin.AccountActorCodeID,
					Expiration:      abi.ChainEpoch(curEpoch + 300),
					DealIDs:         []abi.DealID{nextDealID}, // THIS IS UGLY
				})
				nextDealID++

				sm := signMessage(&types.Message{
					From:     worker,
					To:       minerAddr,
					Method:   builtin.MethodsMiner.PreCommitSector,
					Value:    types.NewInt(0),
					GasLimit: types.NewInt(10000000),
					GasPrice: types.NewInt(0),
					Nonce:    wnonce,
					Params:   params,
				})

				proveCommitEpoch := curEpoch + miner.PreCommitChallengeDelay + 2
				bucket, _ := proveCommitAt[proveCommitEpoch]
				proveCommitAt[proveCommitEpoch] = append(bucket, nextSectorID)

				messages = append(messages, sm)
				wnonce++
				nextSectorID++
			}

		}

		toCommit, _ := proveCommitAt[curEpoch]
		for _, sid := range toCommit {

			params := mustEnc(&miner.ProveCommitSectorParams{
				SectorNumber: abi.SectorNumber(sid),
			})

			sm := signMessage(&types.Message{
				From:     worker,
				To:       minerAddr,
				Method:   builtin.MethodsMiner.ProveCommitSector,
				Value:    types.NewInt(0),
				GasLimit: types.NewInt(10000000),
				GasPrice: types.NewInt(0),
				Nonce:    wnonce,
				Params:   params,
			})
			messages = append(messages, sm)
			wnonce++
		}

		mts, err := cg.NextTipSet()
		if err != nil {
			panic(err)
		}

		fmt.Println(mts.TipSet.Blocks)

	}
}

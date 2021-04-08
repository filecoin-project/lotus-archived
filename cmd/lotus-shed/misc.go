package main

import (
	"fmt"
	"strconv"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin"
	"github.com/filecoin-project/lotus/chain/actors/builtin/multisig"
	"github.com/filecoin-project/lotus/chain/types"
	lcli "github.com/filecoin-project/lotus/cli"
	cbor "github.com/ipfs/go-ipld-cbor"
	"github.com/urfave/cli/v2"
)

var miscCmd = &cli.Command{
	Name:  "misc",
	Usage: "Assorted unsorted commands for various purposes",
	Flags: []cli.Flag{},
	Subcommands: []*cli.Command{
		dealStateMappingCmd,
		msigOwnershipCheckCmd,
	},
}

var dealStateMappingCmd = &cli.Command{
	Name: "deal-state",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return cli.ShowCommandHelp(cctx, cctx.Command.Name)
		}

		num, err := strconv.Atoi(cctx.Args().First())
		if err != nil {
			return err
		}

		ststr, ok := storagemarket.DealStates[uint64(num)]
		if !ok {
			return fmt.Errorf("no such deal state %d", num)
		}
		fmt.Println(ststr)
		return nil
	},
}

var msigOwnershipCheckCmd = &cli.Command{
	Name: "ownership-check",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return cli.ShowCommandHelp(cctx, cctx.Command.Name)
		}

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		defer closer()
		ctx := lcli.ReqContext(cctx)

		addr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		ida, err := api.StateLookupID(ctx, addr, types.EmptyTSK)
		if err != nil {
			return err
		}

		addrs, err := api.StateListActors(ctx, types.EmptyTSK)
		if err != nil {
			return err
		}

		bs := blockstore.NewAPIBlockstore(api)
		cst := cbor.NewCborStore(bs)
		store := adt.WrapStore(ctx, cst)

		total := abi.NewTokenAmount(0)
		for _, a := range addrs {

			act, err := api.StateGetActor(ctx, a, types.EmptyTSK)
			if err != nil {
				return err
			}

			if builtin.IsMultisigActor(act.Code) {
				mst, err := multisig.Load(store, act)
				if err != nil {
					return err
				}

				sigs, err := mst.Signers()
				if err != nil {
					return err
				}

				for _, signer := range sigs {
					if signer == ida {
						fmt.Printf("%s\t%s\n", a, types.FIL(act.Balance))
						total = types.BigAdd(total, act.Balance)
						break
					}
				}

			}
		}
		fmt.Printf("total: %s\n", types.FIL(total))
		return nil
	},
}

package main

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/ipfs/go-cid"
	"github.com/urfave/cli/v2"
)

type message struct {
	from  address.Address
	to    address.Address
	nonce uint64
}

type tipsetMessages struct {
	// from -> message set
	msgs map[message]map[cid.Cid]*types.Message
}

type messageObj struct {
	Epoch abi.ChainEpoch
	To    address.Address
	From  address.Address

	Nonce uint64

	Value abi.TokenAmount

	GasLimit   int64
	GasFeeCap  abi.TokenAmount
	GasPremium abi.TokenAmount

	Cid cid.Cid
}

var findDupeNonceCmd = &cli.Command{
	Name: "find-dupe-nonce",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "tipset",
			Value: "",
		},
		&cli.Uint64Flag{
			Name:  "toheight",
			Usage: "don't look before given block height",
		},
	},
	Action: func(cctx *cli.Context) error {
		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		defer closer()
		ctx := lcli.ReqContext(cctx)

		toh := abi.ChainEpoch(cctx.Uint64("toheight"))

		ts, err := lcli.LoadTipSet(ctx, cctx, api)
		if err != nil {
			return err
		}

		if ts == nil {
			head, err := api.ChainHead(ctx)
			if err != nil {
				return err
			}
			ts = head
		}

		cur := ts
		//fmt.Printf("epoch\tfrom\tto\tnonce\tmethod\tvalue\tgas_limit\tgas_fee_cap\tgas_premium\tcid\n")
		for cur.Height() > toh {
			messageTrack := &tipsetMessages{}
			messageTrack.msgs = make(map[message]map[cid.Cid]*types.Message)

			if ctx.Err() != nil {
				return ctx.Err()
			}

			tipsetMessages := []*types.Message{}

			for _, blockheader := range cur.Blocks() {
				blockMessages, err := api.ChainGetBlockMessages(ctx, blockheader.Cid())
				if err != nil {
					return err
				}

				tipsetMessages = append(tipsetMessages, blockMessages.BlsMessages...)
				for _, secpmsg := range blockMessages.SecpkMessages {
					tipsetMessages = append(tipsetMessages, &secpmsg.Message)
				}
			}

			msgs := tipsetMessages
			for _, m := range msgs {

				if m.Method != 0 {
					continue
				}

				mm := message{
					from:  m.From,
					to:    m.To,
					nonce: m.Nonce,
				}

				if _, ok := messageTrack.msgs[mm]; !ok {
					messageTrack.msgs[mm] = make(map[cid.Cid]*types.Message)
				}

				messageTrack.msgs[mm][m.Cid()] = m
			}

			for _, mset := range messageTrack.msgs {
				if len(mset) > 1 {
					for _, msg := range mset {
						msgd := messageObj{
							Epoch: cur.Height(),
							To:    msg.To,
							From:  msg.From,

							Nonce: msg.Nonce,

							Value: msg.Value,

							GasLimit:   msg.GasLimit,
							GasFeeCap:  msg.GasFeeCap,
							GasPremium: msg.GasPremium,

							Cid: msg.Cid(),
						}
						jdata, err := json.Marshal(msgd)
						if err != nil {
							return err
						}

						fmt.Println(string(jdata))

					}
				}

			}

			next, err := api.ChainGetTipSet(ctx, cur.Parents())
			if err != nil {
				return err
			}

			cur = next
		}

		return nil
	},
}

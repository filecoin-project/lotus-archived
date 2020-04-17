package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/lotus/chain/wallet"
	"github.com/filecoin-project/lotus/lib/sigs"
	"math/rand"
	"os"
	"time"

	"github.com/filecoin-project/specs-actors/actors/crypto"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	lcli "github.com/filecoin-project/lotus/cli"

	"gopkg.in/urfave/cli.v2"
)

func main() {
	app := &cli.App{
		Name:  "chain-noise",
		Usage: "Generate some spam transactions in the network",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "repo",
				EnvVars: []string{"LOTUS_PATH"},
				Hidden:  true,
				Value:   "~/.lotus", // TODO: Consider XDG_DATA_HOME
			},
		},
		Commands: []*cli.Command{runCmd,
			badMsgNoSenderCmd},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

var runCmd = &cli.Command{
	Name: "run",
	Action: func(cctx *cli.Context) error {
		addr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		sendSmallFundsTxs(ctx, api, addr, 5)
		return nil
	},
}

var badMsgNoSenderCmd = &cli.Command{
	Name: "no-sender",
	Action: func(cctx *cli.Context) error {
		ctx := lcli.ReqContext(cctx)
		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		k, err := wallet.GenerateKey(crypto.SigTypeSecp256k1)
		if err != nil {
			return err
		}

		to, err := api.WalletNew(ctx, crypto.SigTypeSecp256k1)
		if err != nil {
			return err
		}

		defer closer()

		sendTx(ctx, api, k, to, 5)
		return nil
	},
}

func sendTx(ctx context.Context, api api.FullNode, k *wallet.Key, to address.Address, rate int) error {
	msg := &types.Message{
		From:     k.Address,
		To:       to,
		Value:    types.NewInt(1),
		GasLimit: 100000,
		GasPrice: types.NewInt(0),
	}

	sig, err := sigs.Sign(wallet.ActSigType(k.Type), k.PrivateKey, msg.Cid().Bytes())
	if err != nil {
		return err
	}

	smsg := &types.SignedMessage{
		Message:   *msg,
		Signature: *sig,
	}

	cid, err := api.MpoolPush(ctx, smsg)

	if err != nil {
		return err
	}
	fmt.Println("Message sent: ", cid)

	return nil
}

func sendSmallFundsTxs(ctx context.Context, api api.FullNode, from address.Address, rate int) error {
	var sendSet []address.Address
	for i := 0; i < 20; i++ {
		naddr, err := api.WalletNew(ctx, crypto.SigTypeSecp256k1)
		if err != nil {
			return err
		}

		sendSet = append(sendSet, naddr)
	}

	tick := time.NewTicker(time.Second / time.Duration(rate))
	for {
		select {
		case <-tick.C:
			msg := &types.Message{
				From:     from,
				To:       sendSet[rand.Intn(20)],
				Value:    types.NewInt(1),
				GasLimit: 100000,
				GasPrice: types.NewInt(0),
			}

			smsg, err := api.MpoolPushMessage(ctx, msg)
			if err != nil {
				return err
			}
			fmt.Println("Message sent: ", smsg.Cid())
		case <-ctx.Done():
			return nil
		}
	}

	return nil
}

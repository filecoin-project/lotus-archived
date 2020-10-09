package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/urfave/cli/v2"
)

var miscCmd = &cli.Command{
	Name:  "misc",
	Usage: "Assorted unsorted commands for various purposes",
	Flags: []cli.Flag{},
	Subcommands: []*cli.Command{
		dealStateMappingCmd,
		epochToTimeCmd,
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

var epochToTimeCmd = &cli.Command{
	Name: "epoch-to-time",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return cli.ShowCommandHelp(cctx, cctx.Command.Name)
		}

		num, err := strconv.Atoi(cctx.Args().First())
		if err != nil {
			return err
		}

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		gen, err := api.ChainGetGenesis(ctx)
		if err != nil {
			return err
		}

		if gen.Height() != 0 {
			return fmt.Errorf("genesis height not zero")
		}

		adjustedTimestamp := gen.MinTimestamp() + uint64(num*30)

		t := time.Unix(int64(adjustedTimestamp), 0)
		fmt.Println(t)
		return nil
	},
}

package main

import (
	"fmt"
	"sort"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/urfave/cli/v2"

	lapi "github.com/filecoin-project/lotus/api"
	lcli "github.com/filecoin-project/lotus/cli"
)

var statCmd = &cli.Command{
	Name: "stat",
	Subcommands: []*cli.Command{
		statWinnersCmd,
	},
}

var statWinnersCmd = &cli.Command{
	Name: "winners",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "length",
			Value: 100,
		},
	},
	Action: func(cctx *cli.Context) error {
		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		defer closer()
		ctx := lcli.ReqContext(cctx)

		cur, err := api.ChainHead(ctx)
		if err != nil {
			return err
		}

		wins := make(map[address.Address]int)

		var total int

		length := cctx.Int("length")
		for i := 0; i < length; {
			fmt.Println(i)
			for _, b := range cur.Blocks() {
				wins[b.Miner] += int(b.ElectionProof.WinCount)
				total += int(b.ElectionProof.WinCount)
			}

			next, err := api.ChainGetTipSet(ctx, cur.Parents())
			if err != nil {
				return err
			}
			i += int(cur.Height() - next.Height())
			cur = next
		}

		fmt.Printf("Effective E: %0.2f\n", float64(total)/float64(length))

		power := make(map[address.Address]*lapi.MinerPower)
		var addrs []address.Address

		for a := range wins {
			pow, err := api.StateMinerPower(ctx, a, cur.Key())
			if err != nil {
				return err
			}

			power[a] = pow
			addrs = append(addrs, a)
		}

		sort.Slice(addrs, func(i, j int) bool {
			powi := power[addrs[i]]
			powj := power[addrs[j]]

			return big.Cmp(powi.MinerPower.QualityAdjPower, powj.MinerPower.QualityAdjPower) > 0
		})

		fmt.Printf("Miner\tPowerRatio\tWinRatio\tDelta\tDeltaPerc\n")
		for _, a := range addrs[:50] {
			pow := power[a]

			actualFrac := 100 * (float64(wins[a]) / float64(5*length))
			expFrac := pfrac(pow)
			fmt.Printf("%s\t%0.2f\t%0.2f\t%0.2f\t%0.2f%%\n", a, expFrac, actualFrac, actualFrac-expFrac, 100*((actualFrac-expFrac)/expFrac))
		}

		return nil
	},
}

func pfrac(pow *lapi.MinerPower) float64 {
	num := big.Mul(pow.MinerPower.QualityAdjPower, big.NewInt(10000))
	perc := big.Div(num, pow.TotalPower.QualityAdjPower)
	return float64(perc.Int64()) / 100
}

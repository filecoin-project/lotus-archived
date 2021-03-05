package main

import (
	"github.com/filecoin-project/lotus/cli/util"
	_ "net/http/pprof"

	"github.com/urfave/cli/v2"
)

var stopCmd = &cli.Command{
	Name:  "stop",
	Usage: "Stop a running lotus miner",
	Flags: []cli.Flag{},
	Action: func(cctx *cli.Context) error {
		api, closer, err := cliutil.GetAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		err = api.Shutdown(cliutil.ReqContext(cctx))
		if err != nil {
			return err
		}

		return nil
	},
}

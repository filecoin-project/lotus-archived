package cli

import (
	"fmt"
	"github.com/filecoin-project/lotus/cli/util"
	"time"

	"github.com/urfave/cli/v2"
)

var waitApiCmd = &cli.Command{
	Name:  "wait-api",
	Usage: "Wait for lotus api to come online",
	Action: func(cctx *cli.Context) error {
		for i := 0; i < 30; i++ {
			api, closer, err := cliutil.GetFullNodeAPI(cctx)
			if err != nil {
				fmt.Printf("Not online yet... (%s)\n", err)
				time.Sleep(time.Second)
				continue
			}
			defer closer()

			ctx := cliutil.ReqContext(cctx)

			_, err = api.ID(ctx)
			if err != nil {
				return err
			}

			return nil
		}
		return fmt.Errorf("timed out waiting for api to come online")
	},
}

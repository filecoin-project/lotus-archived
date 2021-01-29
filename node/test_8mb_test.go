package node_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/filecoin-project/lotus/cli"
	clitest "github.com/filecoin-project/lotus/cli/test"

	"github.com/filecoin-project/lotus/chain/actors/policy"

	logging "github.com/ipfs/go-log/v2"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api/test"
	builder "github.com/filecoin-project/lotus/node/test"
)

func init() {
	_ = logging.SetLogLevel("*", "INFO")

	policy.SetConsensusMinerMinPower(abi.NewStoragePower(8388608))
	policy.SetSupportedProofTypes(abi.RegisteredSealProof_StackedDrg8MiBV1)
	policy.SetMinVerifiedDealSize(abi.NewStoragePower(256))
}

func Test8MBAPIDealFlow(t *testing.T) {
	logging.SetLogLevel("miner", "ERROR")
	logging.SetLogLevel("chainstore", "ERROR")
	logging.SetLogLevel("chain", "ERROR")
	logging.SetLogLevel("sub", "ERROR")
	logging.SetLogLevel("storageminer", "ERROR")
	logging.SetLogLevel("markets", "DEBUG")

	blockTime := 10 * time.Millisecond

	// For these tests where the block time is artificially short, just use
	// a deal start epoch that is guaranteed to be far enough in the future
	// so that the deal starts sealing in time
	dealStartEpoch := abi.ChainEpoch(2 << 12)

	t.Run("TestDealFlow", func(t *testing.T) {
		test.TestDealFlow(t, builder.MockSbBuilder, blockTime, false, false, dealStartEpoch)
	})
}

func TestClient(t *testing.T) {
	_ = os.Setenv("BELLMAN_NO_GPU", "1")
	clitest.QuietMiningLogs()

	blocktime := 5 * time.Millisecond
	ctx := context.Background()
	clientNode, _ := clitest.StartOneNodeOneMiner(ctx, t, blocktime)
	clitest.RunClientTest(t, cli.Commands, clientNode)
}

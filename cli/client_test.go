package cli

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/filecoin-project/go-state-types/abi"

	"github.com/filecoin-project/lotus/chain/actors/policy"
	clitest "github.com/filecoin-project/lotus/cli/test"
)

func init() {
	policy.SetSupportedProofTypes(abi.RegisteredSealProof_StackedDrg8MiBV1)
	policy.SetConsensusMinerMinPower(abi.NewStoragePower(8388608))
	policy.SetMinVerifiedDealSize(abi.NewStoragePower(256))
}

// TestClient does a basic test to exercise the client CLI
// commands
func TestClient(t *testing.T) {
	_ = os.Setenv("BELLMAN_NO_GPU", "1")
	clitest.QuietMiningLogs()

	blocktime := 5 * time.Millisecond
	ctx := context.Background()
	clientNode, _ := clitest.StartOneNodeOneMiner(ctx, t, blocktime)
	clitest.RunClientTest(t, Commands, clientNode)
}

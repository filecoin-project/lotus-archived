package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/filecoin-project/go-state-types/abi"
	lapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
	lcli "github.com/filecoin-project/lotus/cli"
	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

var log = logging.Logger("serve-state")

func main() {
	logging.SetLogLevel("*", "INFO")

	log.Info("Starting health agent")

	local := []*cli.Command{
		serveStateCmd,
	}

	app := &cli.App{
		Name:     "state-srv",
		Usage:    "A utility for lotus state inspection",
		Version:  build.UserVersion(),
		Commands: local,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "repo",
				EnvVars: []string{"LOTUS_PATH"},
				Value:   "~/.lotus", // TODO: Consider XDG_DATA_HOME
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
		return
	}
}

type StateServer struct {
	Dir string
	api lapi.FullNode

	procLk   sync.Mutex
	workChan map[int]chan struct{}
}

func NewStateServer(dir string, api lapi.FullNode) *StateServer {
	return &StateServer{
		Dir:      dir,
		api:      api,
		workChan: make(map[int]chan struct{}),
	}
}

func (ss *StateServer) HandleGetComputedState(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	height, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	cso, err := ss.getComputedState(context.TODO(), height)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), 500)
		return
	}

	if err := json.NewEncoder(w).Encode(cso); err != nil {
		log.Error(err)
	}
}

func (ss *StateServer) getComputedState(ctx context.Context, height int) (*lapi.ComputeStateOutput, error) {
	cv, cacheGood, err := ss.checkCache(height)
	if err != nil {
		return nil, err
	}

	if cacheGood {
		return cv, nil
	}

	ss.procLk.Lock()
	ch, ok := ss.workChan[height]
	if !ok {
		ch = make(chan struct{})
		ss.workChan[height] = ch
		ss.procLk.Unlock()
	} else {
		ss.procLk.Unlock()
		<-ch
		cv, cacheGood, err := ss.checkCache(height)
		if err != nil {
			return nil, err
		}
		if !cacheGood {
			return nil, fmt.Errorf("failed to find cache after waiting")
		}
		return cv, nil
	}

	cso, err := ss.stateCompute(context.TODO(), abi.ChainEpoch(height))
	if err != nil {
		return nil, err
	}

	if err := ss.saveCache(height, cso); err != nil {
		return nil, err // could probably not error here... idk
	}

	close(ch)

	return cso, nil
}

func (ss *StateServer) saveCache(height int, cso *lapi.ComputeStateOutput) error {
	fname := filepath.Join(ss.Dir, fmt.Sprintf("cs-%d", height))

	fi, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer fi.Close()

	if err := json.NewEncoder(fi).Encode(cso); err != nil {
		return err
	}

	return nil
}

func (ss *StateServer) stateCompute(ctx context.Context, height abi.ChainEpoch) (*lapi.ComputeStateOutput, error) {
	head, err := ss.api.ChainHead(context.TODO())
	if err != nil {
		return nil, err
	}
	if head.Height()-abi.ChainEpoch(height) < 15 {
		return nil, fmt.Errorf("height %d is too recent", height)
	}

	ts, err := ss.api.ChainGetTipSetByHeight(ctx, height, types.EmptyTSK)
	if err != nil {
		return nil, err
	}

	if ts.Height() != height {
		return &lapi.ComputeStateOutput{}, nil
	}

	return ss.api.StateCompute(context.TODO(), height, nil, ts.Key())
}

func (ss *StateServer) checkCache(height int) (*lapi.ComputeStateOutput, bool, error) {
	head, err := ss.api.ChainHead(context.TODO())
	if err != nil {
		return nil, false, err
	}
	if head.Height()-abi.ChainEpoch(height) < 15 {
		return nil, false, nil
	}

	fname := filepath.Join(ss.Dir, fmt.Sprintf("cs-%d", height))
	fi, err := os.Open(fname)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	defer fi.Close()

	var cso lapi.ComputeStateOutput
	if err := json.NewDecoder(fi).Decode(&cso); err != nil {
		return nil, false, xerrors.Errorf("failed to decode cached values for height %d: %w", height, err)
	}

	return &cso, true, nil
}

var serveStateCmd = &cli.Command{
	Name: "run",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "cache-dir",
		},
	},
	Action: func(cctx *cli.Context) error {
		logging.SetLogLevel("rpc", "ERROR")

		dir := cctx.String("cache-dir")

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		defer closer()
		ctx := lcli.ReqContext(cctx)
		_ = ctx

		ssrv := NewStateServer(dir, api)

		http.HandleFunc("/state/", ssrv.HandleGetComputedState)

		go func() {
			panic(http.ListenAndServe(":15607", nil))
		}()

		<-ctx.Done()
		return nil
	},
}

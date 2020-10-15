package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	cborutil "github.com/filecoin-project/go-cbor-util"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/gen/slashfilter"
	"github.com/filecoin-project/lotus/chain/types"
	lcli "github.com/filecoin-project/lotus/cli"
	cliutil "github.com/filecoin-project/lotus/cli/util"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
	lru "github.com/hashicorp/golang-lru"
	"github.com/ipfs/go-cid"
	levelds "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/mitchellh/go-homedir"
	"github.com/multiformats/go-multiaddr"
	ldbopts "github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/urfave/cli/v2"
	"go.uber.org/multierr"
	"golang.org/x/xerrors"
)

var consensusCmd = &cli.Command{
	Name:  "consensus",
	Usage: "tools for gathering information about consensus between nodes",
	Flags: []cli.Flag{},
	Subcommands: []*cli.Command{
		consensusCheckCmd,
		consensusWatchSlashableCmd,
		consnesusRunSlash,
	},
}

type consensusItem struct {
	multiaddr     multiaddr.Multiaddr
	genesisTipset *types.TipSet
	targetTipset  *types.TipSet
	headTipset    *types.TipSet
	peerID        peer.ID
	version       api.Version
	api           api.FullNode
}

var consensusCheckCmd = &cli.Command{
	Name:  "check",
	Usage: "verify if all nodes agree upon a common tipset for a given tipset height",
	Description: `Consensus check verifies that all nodes share a common tipset for a given
   height.

   The height flag specifies a chain height to start a comparison from. There are two special
   arguments for this flag. All other expected values should be chain tipset heights.

   @common   - Use the maximum common chain height between all nodes
   @expected - Use the current time and the genesis timestamp to determine a height

   Examples

   Find the highest common tipset and look back 10 tipsets
   lotus-shed consensus check --height @common --lookback 10

   Calculate the expected tipset height and look back 10 tipsets
   lotus-shed consensus check --height @expected --lookback 10

   Check if nodes all share a common genesis
   lotus-shed consensus check --height 0

   Check that all nodes agree upon the tipset for 1day post genesis
   lotus-shed consensus check --height 2880 --lookback 0
	`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "height",
			Value: "@common",
			Usage: "height of tipset to start check from",
		},
		&cli.IntFlag{
			Name:  "lookback",
			Value: int(build.MessageConfidence * 2),
			Usage: "number of tipsets behind to look back when comparing nodes",
		},
	},
	Action: func(cctx *cli.Context) error {
		filePath := cctx.Args().First()

		var input *bufio.Reader
		if cctx.Args().Len() == 0 {
			input = bufio.NewReader(os.Stdin)
		} else {
			var err error
			inputFile, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer inputFile.Close() //nolint:errcheck
			input = bufio.NewReader(inputFile)
		}

		var nodes []*consensusItem
		ctx := lcli.ReqContext(cctx)

		for {
			strma, errR := input.ReadString('\n')
			strma = strings.TrimSpace(strma)

			if len(strma) == 0 {
				if errR == io.EOF {
					break
				}
				continue
			}

			apima, err := multiaddr.NewMultiaddr(strma)
			if err != nil {
				return err
			}
			ainfo := cliutil.APIInfo{Addr: apima.String()}
			addr, err := ainfo.DialArgs()
			if err != nil {
				return err
			}

			api, closer, err := client.NewFullNodeRPC(cctx.Context, addr, nil)
			if err != nil {
				return err
			}
			defer closer()

			peerID, err := api.ID(ctx)
			if err != nil {
				return err
			}

			version, err := api.Version(ctx)
			if err != nil {
				return err
			}

			genesisTipset, err := api.ChainGetGenesis(ctx)
			if err != nil {
				return err
			}

			headTipset, err := api.ChainHead(ctx)
			if err != nil {
				return err
			}

			nodes = append(nodes, &consensusItem{
				genesisTipset: genesisTipset,
				headTipset:    headTipset,
				multiaddr:     apima,
				api:           api,
				peerID:        peerID,
				version:       version,
			})

			if errR != nil && errR != io.EOF {
				return err
			}

			if errR == io.EOF {
				break
			}
		}

		if len(nodes) == 0 {
			return fmt.Errorf("no nodes")
		}

		genesisBuckets := make(map[types.TipSetKey][]*consensusItem)
		for _, node := range nodes {
			genesisBuckets[node.genesisTipset.Key()] = append(genesisBuckets[node.genesisTipset.Key()], node)

		}

		if len(genesisBuckets) != 1 {
			for _, nodes := range genesisBuckets {
				for _, node := range nodes {
					log.Errorw(
						"genesis do not match",
						"genesis_tipset", node.genesisTipset.Key(),
						"peer_id", node.peerID,
						"version", node.version,
					)
				}
			}

			return fmt.Errorf("genesis does not match between all nodes")
		}

		target := abi.ChainEpoch(0)

		switch cctx.String("height") {
		case "@common":
			minTipset := nodes[0].headTipset
			for _, node := range nodes {
				if node.headTipset.Height() < minTipset.Height() {
					minTipset = node.headTipset
				}
			}

			target = minTipset.Height()
		case "@expected":
			tnow := uint64(time.Now().Unix())
			tgen := nodes[0].genesisTipset.MinTimestamp()

			target = abi.ChainEpoch((tnow - tgen) / build.BlockDelaySecs)
		default:
			h, err := strconv.Atoi(strings.TrimSpace(cctx.String("height")))
			if err != nil {
				return fmt.Errorf("failed to parse string: %s", cctx.String("height"))
			}

			target = abi.ChainEpoch(h)
		}

		lookback := abi.ChainEpoch(cctx.Int("lookback"))
		if lookback > target {
			target = abi.ChainEpoch(0)
		} else {
			target = target - lookback
		}

		for _, node := range nodes {
			targetTipset, err := node.api.ChainGetTipSetByHeight(ctx, target, types.EmptyTSK)
			if err != nil {
				log.Errorw("error checking target", "err", err)
				node.targetTipset = nil
			} else {
				node.targetTipset = targetTipset
			}

		}
		for _, node := range nodes {
			log.Debugw(
				"node info",
				"peer_id", node.peerID,
				"version", node.version,
				"genesis_tipset", node.genesisTipset.Key(),
				"head_tipset", node.headTipset.Key(),
				"target_tipset", node.targetTipset.Key(),
			)
		}

		targetBuckets := make(map[types.TipSetKey][]*consensusItem)
		for _, node := range nodes {
			if node.targetTipset == nil {
				targetBuckets[types.EmptyTSK] = append(targetBuckets[types.EmptyTSK], node)
				continue
			}

			targetBuckets[node.targetTipset.Key()] = append(targetBuckets[node.targetTipset.Key()], node)
		}

		if nodes, ok := targetBuckets[types.EmptyTSK]; ok {
			for _, node := range nodes {
				log.Errorw(
					"targeted tipset not found",
					"peer_id", node.peerID,
					"version", node.version,
					"genesis_tipset", node.genesisTipset.Key(),
					"head_tipset", node.headTipset.Key(),
					"target_tipset", node.targetTipset.Key(),
				)
			}

			return fmt.Errorf("targeted tipset not found")
		}

		if len(targetBuckets) != 1 {
			for _, nodes := range targetBuckets {
				for _, node := range nodes {
					log.Errorw(
						"targeted tipset not found",
						"peer_id", node.peerID,
						"version", node.version,
						"genesis_tipset", node.genesisTipset.Key(),
						"head_tipset", node.headTipset.Key(),
						"target_tipset", node.targetTipset.Key(),
					)
				}
			}
			return fmt.Errorf("nodes not in consensus at tipset height %d", target)
		}

		return nil
	},
}

type ToSlash struct {
	Miner address.Address
	Epoch abi.ChainEpoch
	B1    cid.Cid
	B2    cid.Cid
}

var consensusWatchSlashableCmd = &cli.Command{
	Name:  "watch-slashable",
	Usage: "Watch incoming blocks for blocks which can be slashed",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "db-dir",
			Value: "slashwatch",
		},
		&cli.StringFlag{
			Name:  "slashq",
			Value: "slashq.ndjson",
		},
	},
	Action: func(cctx *cli.Context) error {
		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)
		p, err := homedir.Expand(cctx.String("db-dir"))
		if err != nil {
			return xerrors.Errorf("expand db dir: %w", err)
		}
		ds, err := levelds.NewDatastore(p, &levelds.Options{
			Compression: ldbopts.NoCompression,
			NoSync:      false,
			Strict:      ldbopts.StrictAll,
			ReadOnly:    false,
		})
		if err != nil {
			return xerrors.Errorf("open leveldb: %w", err)
		}

		p, err = homedir.Expand(cctx.String("slashq"))
		if err != nil {
			return xerrors.Errorf("opening slashq: %w", err)
		}
		slashqFile, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		jenc := json.NewEncoder(slashqFile)

		defer ds.Close() // nolint
		sf := slashfilter.New(ds)
		sub, err := api.SyncIncomingBlocks(ctx)
		if err != nil {
			return err
		}
		arc, _ := lru.NewARC(10240)

		buf := make(chan *types.BlockHeader, 20)
		go func() {
			for bh := range sub {
				select {
				case buf <- bh:
				case <-ctx.Done():
					close(buf)
					return
				}
			}
			close(buf)
		}()
		var ok, all, dupe, slashable int
		for bh := range buf {
			if all%30 == 0 {
				log.Infow("Processed blocks", "ok-count", ok, "slashable", slashable, "all-count", all, "dupes", dupe)
			}
			all++
			_, seen := arc.Get(bh.Cid())
			if seen {
				dupe++
			}
			arc.Add(bh.Cid(), bh)

			var parH abi.ChainEpoch = -1
			{
				var merr error
				var par *types.BlockHeader
				for _, p := range bh.Parents {
					parI, ok := arc.Get(p)
					if ok {
						par = parI.(*types.BlockHeader)
						break
					}
				}

				if par == nil {
					for _, p := range bh.Parents {
						par, err = api.ChainGetBlock(ctx, p)
						if !multierr.AppendInto(&merr, err) {
							break
						}
					}
				}

				if par == nil {
					log.Infow("Getting parents failed", "block", bh.Cid(), "miner",
						bh.Miner, "error", merr)
				} else {
					parH = par.Height
				}
			}

			witness, err := sf.CheckBlock(bh, parH)
			if err != nil {
				slashable++
				log.Warnw("SLASH ERROR!", "block", bh.Cid(), "miner", bh.Miner, "error", err, "witness", witness)
				err := jenc.Encode(ToSlash{
					Miner: bh.Miner,
					Epoch: bh.Height,
					B1:    bh.Cid(),
					B2:    witness[0],
				})
				if err != nil {
					log.Errorf("got error writing slash: %+v", err)
				}
				slashqFile.Sync()

				continue
			}
			ok++
		}
		return nil
	},
}

var consnesusRunSlash = &cli.Command{
	Name:  "run-slashq",
	Usage: "Watch incoming blocks for blocks which can be slashed",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "from",
		},
		&cli.StringFlag{
			Name:  "slashq",
			Value: "slashq.ndjson",
		},
	},
	Action: func(cctx *cli.Context) error {
		if !cctx.IsSet("from") {
			return xerrors.Errorf("from has to be set")
		}
		from, err := address.NewFromString(cctx.String("from"))
		if err != nil {
			return err
		}

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		p, err := homedir.Expand(cctx.String("slashq"))
		if err != nil {
			return xerrors.Errorf("opening slashq: %w", err)
		}

		for {
			err := runSlashing(ctx, from, p, api)
			if err != nil {
				log.Errorf("got error from slashing: %+v", err)
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
		}
	},
}

func runSlashing(ctx context.Context, from address.Address, fPath string, api api.FullNode) error {
	const forward = 700

	slashqFile, err := os.OpenFile(fPath, os.O_RDONLY, 0666)
	defer slashqFile.Close()
	jenc := json.NewDecoder(slashqFile)

	ts, err := api.ChainHead(ctx)
	if err != nil {
		return xerrors.Errorf("getting tipset: %w", err)
	}

	for {
		var slashC ToSlash
		err := jenc.Decode(&slashC)
		if err != nil {
			if !xerrors.Is(err, io.EOF) {
				log.Warn("reading to slash file: %+v", err)
			}
			break
		}
		if slashC.Epoch+forward > ts.Height() {
			log.Infof("skipping %s, still waiting %d < %d (%s)", slashC.Miner, slashC.Epoch+forward, ts.Height(),
				time.Duration(slashC.Epoch+forward-ts.Height())*time.Duration(build.BlockDelaySecs)*time.Second)
			continue
		}

		minfo, err := api.StateMinerInfo(ctx, slashC.Miner, types.EmptyTSK)
		if err != nil {
			log.Warnf("MinerInfo failed: %+v", err)
			continue
		}
		if slashC.Epoch < minfo.ConsensusFaultElapsed {
			continue
		}

		b1, err := api.ChainGetBlock(ctx, slashC.B1)
		if err != nil {
			return xerrors.Errorf("getting block 1, miner %s, epoch %d: %w", slashC.Miner, slashC.Epoch, err)
		}

		b2, err := api.ChainGetBlock(ctx, slashC.B2)
		if err != nil {
			return xerrors.Errorf("getting block 2: %w", err)
		}

		bh1, err := cborutil.Dump(b1)
		if err != nil {
			return err
		}

		bh2, err := cborutil.Dump(b2)
		if err != nil {
			return err
		}

		params := miner.ReportConsensusFaultParams{
			BlockHeader1: bh1,
			BlockHeader2: bh2,
		}

		enc, err := actors.SerializeParams(&params)
		if err != nil {
			return err
		}

		msg := &types.Message{
			To:     slashC.Miner,
			From:   from,
			Value:  types.NewInt(0),
			Method: builtin.MethodsMiner.ReportConsensusFault,
			Params: enc,
		}

		log.Infow("slashing", "toSlash", slashC)
		smsg, err := api.MpoolPushMessage(ctx, msg, nil)
		if err != nil {
			return err
		}
		msgRes, err := api.StateWaitMsg(ctx, smsg.Cid(), 1)
		if err != nil {
			return err
		}
		if msgRes.Receipt.ExitCode != 0 {
			log.Warnf("got %d from msg: %s", msgRes.Receipt.ExitCode, smsg.Cid())
		}
	}

	return nil

}

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/anorth/go-dar/pkg/dar"
	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-state-types/abi"

	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	lcli "github.com/filecoin-project/lotus/cli"
	bstore "github.com/filecoin-project/lotus/lib/blockstore"
	"github.com/filecoin-project/lotus/node/repo"
)

var exportChainDarCmd = &cli.Command{
	Name:        "export-dar",
	Description: "Export chain from repo (requires node to be offline)",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "repo",
			Value: "~/.lotus",
		},
		&cli.StringFlag{
			Name:  "tipset",
			Usage: "tipset to export from",
		},
		&cli.Int64Flag{
			Name: "recent-stateroots",
		},
		&cli.BoolFlag{
			Name: "full-state",
		},
		&cli.BoolFlag{
			Name: "skip-old-msgs",
		},
	},
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return lcli.ShowHelp(cctx, fmt.Errorf("must specify file name to write export to"))
		}

		ctx := context.TODO()

		r, err := repo.NewFS(cctx.String("repo"))
		if err != nil {
			return xerrors.Errorf("opening fs repo: %w", err)
		}

		exists, err := r.Exists()
		if err != nil {
			return err
		}
		if !exists {
			return xerrors.Errorf("lotus repo doesn't exist")
		}

		lr, err := r.Lock(repo.FullNode)
		if err != nil {
			return err
		}
		defer lr.Close() //nolint:errcheck

		fi, err := os.Create(cctx.Args().First())
		if err != nil {
			return xerrors.Errorf("opening the output file: %w", err)
		}

		defer fi.Close() //nolint:errcheck

		bs, err := lr.Blockstore(ctx, repo.BlockstoreChain)
		if err != nil {
			return fmt.Errorf("failed to open blockstore: %w", err)
		}

		defer func() {
			if c, ok := bs.(io.Closer); ok {
				if err := c.Close(); err != nil {
					log.Warnf("failed to close blockstore: %s", err)
				}
			}
		}()

		mds, err := lr.Datastore(context.Background(), "/metadata")
		if err != nil {
			return err
		}

		cs := store.NewChainStore(bs, bs, mds, nil, nil)
		defer cs.Close() //nolint:errcheck

		if err := cs.Load(); err != nil {
			return err
		}

		nroots := abi.ChainEpoch(cctx.Int64("recent-stateroots"))
		fullstate := cctx.Bool("full-state")
		skipoldmsgs := cctx.Bool("skip-old-msgs")

		var ts *types.TipSet
		if tss := cctx.String("tipset"); tss != "" {
			cids, err := lcli.ParseTipSetString(tss)
			if err != nil {
				return xerrors.Errorf("failed to parse tipset (%q): %w", tss, err)
			}

			tsk := types.NewTipSetKey(cids...)

			selts, err := cs.LoadTipSet(tsk)
			if err != nil {
				return xerrors.Errorf("loading tipset: %w", err)
			}
			ts = selts
		} else {
			ts = cs.GetHeaviestTipSet()
		}

		if fullstate {
			nroots = ts.Height() + 1
		}

		if err := exportChain(ctx, bs, ts, nroots, skipoldmsgs, fi); err != nil {
			return xerrors.Errorf("export failed: %w", err)
		}

		return nil
	},
}

var exportStateDarCmd = &cli.Command{
	Name:        "export-state-dar",
	Description: "Export a state tree from repo",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "repo",
			Value: "~/.lotus",
		},
		&cli.StringFlag{
			Name:  "root",
			Usage: "state root to export",
		},
	},
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return lcli.ShowHelp(cctx, fmt.Errorf("must specify file name to write export to"))
		}

		ctx := context.TODO()

		r, err := repo.NewFS(cctx.String("repo"))
		if err != nil {
			return xerrors.Errorf("opening fs repo: %w", err)
		}

		exists, err := r.Exists()
		if err != nil {
			return err
		}
		if !exists {
			return xerrors.Errorf("lotus repo doesn't exist")
		}

		lr, err := r.Lock(repo.FullNode)
		if err != nil {
			return err
		}
		defer lr.Close() //nolint:errcheck

		fi, err := os.Create(cctx.Args().First())
		if err != nil {
			return xerrors.Errorf("opening the output file: %w", err)
		}

		defer fi.Close() //nolint:errcheck

		bs, err := lr.Blockstore(ctx, repo.BlockstoreChain)
		if err != nil {
			return fmt.Errorf("failed to open blockstore: %w", err)
		}

		defer func() {
			if c, ok := bs.(io.Closer); ok {
				if err := c.Close(); err != nil {
					log.Warnf("failed to close blockstore: %s", err)
				}
			}
		}()

		//mds, err := lr.Datastore(context.Background(), "/metadata")
		//if err != nil {
		//	return err
		//}

		//cs := store.NewChainStore(bs, bs, mds, nil, nil)
		//defer cs.Close() //nolint:errcheck
		//
		//if err := cs.Load(); err != nil {
		//	return err
		//}

		root, err := cid.Parse(strings.TrimSpace(cctx.String("root")))
		dwr, err := dar.NewWriter(fi, false)

		if err := exportStateTree(ctx, dwr, bs, cidlink.Link{root}); err != nil {
			return err
		}
		return nil
	},
}

// The following are adapted from chain.store for prototyping, and should probably be moved back there.

func exportChain(ctx context.Context, bs bstore.Blockstore, ts *types.TipSet, inclRecentRoots abi.ChainEpoch, skipOldMsgs bool, w io.Writer) error {
	dwr, err := dar.NewWriter(w, false)
	if err != nil {
		return xerrors.Errorf("failed to initialize DAR stream: %w", err)
	}

	// Each block in the head tipset is added as a DAR root.
	// The DAG connected to each include the chain and messages, but not state roots.
	// Depth-first means that all blocks are traversed before any message trees.
	// The state roots are added later as distinct DAGs.
	var stateRoots []cidlink.Link
	for _, tipCid := range ts.Cids() {
		data, err := bs.Get(tipCid)
		if err != nil {
			return xerrors.Errorf("writing object to car, bs.Get: %w", err)
		}

		if _, err := dwr.BeginDAGBlock(ctx, cidlink.Link{tipCid}, data.RawData()); err != nil {
			return xerrors.Errorf("writing object to car, bs.Get: %w", err)
		}

		var blockStack []cid.Cid
		var msgRoots []cidlink.Link // Collects message roots from the whole chain.

		enqueueBockLinks := func(data []byte) error {
			var b types.BlockHeader
			if err := b.UnmarshalCBOR(bytes.NewBuffer(data)); err != nil {
				return xerrors.Errorf("unmarshaling block header (cid=%s): %w", data, err)
			}

			// Push parent blocks on to the stack.
			for i := len(b.Parents) - 1; i >= 0; i-- {
				blockStack = append(blockStack, b.Parents[i])
			}

			// Defer state roots until after the chain, as new DAGs.
			// FIXME: DAR can't handle the genesis object's encoding yet, so pre-genesis state cannot be exported.
			if b.Height != 0 && b.Height > ts.Height()-inclRecentRoots {
				stateRoots = append(stateRoots, cidlink.Link{b.ParentStateRoot})
			}

			// Queue message trees for export after the parents.
			if !skipOldMsgs || b.Height > ts.Height()-inclRecentRoots {
				msgRoots = append(msgRoots, cidlink.Link{b.Messages})
			}
			return nil
		}

		if err := enqueueBockLinks(data.RawData()); err != nil {
			return err
		}

		//var b types.BlockHeader
		//if err := b.UnmarshalCBOR(bytes.NewBuffer(data.RawData())); err != nil {
		//	return xerrors.Errorf("unmarshaling block header (cid=%s): %w", data, err)
		//}

		// For all blocks after the first in the tipset, the traversal will halt after connecting
		// to the already-written DAGs.
		var bcid cid.Cid
		for len(blockStack) > 0 {
			top := len(blockStack) - 1
			bcid, blockStack = blockStack[top], blockStack[:top]
			data, err := bs.Get(bcid)
			if err != nil {
				return xerrors.Errorf("writing object to car, bs.Get: %w", err)
			}

			if links, err := dwr.AppendBlock(ctx, cidlink.Link{bcid}, data.RawData()); err != nil {
				return xerrors.Errorf("writing object to car, bs.Get: %w", err)
			} else if len(links) == 0 {
				// This block has already been seen, stop here.
				continue
			}

			if err := enqueueBockLinks(data.RawData()); err != nil {
				return err
			}

			//var b types.BlockHeader
			//if err := b.UnmarshalCBOR(bytes.NewBuffer(data.RawData())); err != nil {
			//	return xerrors.Errorf("unmarshaling block header (cid=%s): %w", data, err)
			//}
			//
			//// Push parent blocks on to the stack.
			//for _, p := range b.Parents {
			//	blockStack = append(blockStack, p) // FIXME stack
			//}
			//
			//// Defer state roots until after the chain, as new DAGs.
			//if b.Height == 0 || b.Height > ts.Height()-inclRecentRoots {
			//	stateRoots = append(stateRoots, cidlink.Link{b.ParentStateRoot})
			//}
			//
			//// Queue message trees for export after the parents.
			//if !skipOldMsgs || b.Height > ts.Height()-inclRecentRoots {
			//	msgRoots = append(msgRoots, cidlink.Link{b.Messages})
			//}
		}

		reverse(msgRoots)
		for _, msgRoot := range msgRoots {
			if err := appendFullDAG(ctx, dwr, bs, msgRoot); err != nil {
				return err
			}
		}
	}

	// Export all the state roots.
	var prevStateRoot cidlink.Link
	for _, sroot := range stateRoots {
		// The blocks in a tipset all reference the same parent state root, so each will appear multiple
		// times here. But the DAR writer will reject attempts to write the same root multiple times.
		if sroot == prevStateRoot {
			continue
		}
		prevStateRoot = sroot
		if err := exportStateTree(ctx, dwr, bs, sroot); err != nil {
			return err
		}
	}
	return nil
}

func exportStateTree(ctx context.Context, dwr *dar.Writer, bs bstore.Blockstore, sroot cidlink.Link) error {
	data, err := bs.Get(sroot.Cid)
	if err != nil {
		return xerrors.Errorf("writing object to car, bs.Get: %w", err)
	}
	links, err := dwr.BeginDAGBlock(ctx, sroot, data.RawData())
	if err != nil {
		return err
	}
	for _, lnk := range links {
		if err := appendFullDAG(ctx, dwr, bs, lnk); err != nil {
			return err
		}
	}
	return nil
}

// TODO: move this to dar package
func appendFullDAG(ctx context.Context, dwr *dar.Writer, bs bstore.Blockstore, top cidlink.Link) error {
	stack := []cidlink.Link{top} // LIFO from the right
	var link cidlink.Link

	for len(stack) > 0 {
		top := len(stack) - 1
		link, stack = stack[top], stack[:top]

		if link.Prefix().Codec != cid.DagCBOR {
			continue
		}

		data, err := bs.Get(link.Cid)
		if err != nil {
			return xerrors.Errorf("failed to load block %v: %w", link.Cid, err)
		}

		links, err := dwr.AppendBlock(ctx, link, data.RawData())
		if err != nil {
			return err
		}
		reverse(links)
		stack = append(stack, links...)
	}
	return nil
}

func reverse(s []cidlink.Link) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

}

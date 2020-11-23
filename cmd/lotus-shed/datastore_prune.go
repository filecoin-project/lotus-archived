package main

import (
	"context"
	"io"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/lib/blockstore"
	badgerbs "github.com/filecoin-project/lotus/lib/blockstore/badger"
	"github.com/filecoin-project/lotus/node/repo"
)

var datastorePruneFlags struct {
	srcpath string
	dstpath string
	head    string
}

var datastorePruneCmd = &cli.Command{
	Name:        "prune",
	Description: "prunes unreachable blocks from a badger blockstore, by tranferring only reachable blocks to the dest blockstore",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "src",
			Usage:       "path of the source badger datastore",
			TakesFile:   true,
			Required:    true,
			Destination: &datastorePruneFlags.srcpath,
		},
		&cli.StringFlag{
			Name:        "dst",
			Usage:       "path of the destination badger datastore",
			TakesFile:   true,
			Required:    true,
			Destination: &datastorePruneFlags.dstpath,
		},
		&cli.StringFlag{
			Name:        "head",
			Usage:       "head tipset key, used as starting traversal point",
			Required:    true,
			Destination: &datastorePruneFlags.head,
		},
	},
	Action: func(cctx *cli.Context) (err error) {
		var (
			srcBs blockstore.Blockstore
			dstBs blockstore.Blockstore
			head  *types.TipSet
		)

		opts, err := repo.BadgerBlockstoreOptions(repo.BlockstoreChain, datastorePruneFlags.srcpath, true)
		if err != nil {
			return err
		}

		srcBs, err = badgerbs.Open(opts)
		if err != nil {
			return err
		}
		defer srcBs.(io.Closer).Close() //nolint:errcheck

		opts, err = repo.BadgerBlockstoreOptions(repo.BlockstoreChain, datastorePruneFlags.dstpath, false)
		if err != nil {
			return err
		}

		dstBs, err = badgerbs.Open(opts)
		if err != nil {
			return err
		}
		defer dstBs.(io.Closer).Close() //nolint:errcheck

		cs := store.NewChainStore(srcBs, srcBs, datastore.NewMapDatastore(), nil, nil)
		cids, err := lcli.ParseTipSetString(datastorePruneFlags.head)
		if err != nil {
			return xerrors.Errorf("failed to parse tipset (%q): %w", datastorePruneFlags.head, err)
		}

		head, err = cs.LoadTipSet(types.NewTipSetKey(cids...))
		if err != nil {
			return xerrors.Errorf("loading tipset: %w", err)
		}

		viewer := srcBs.(blockstore.Viewer)
		var cnt int64
		err = cs.WalkSnapshot(context.Background(), head, head.Height(), false, func(cid cid.Cid) error {
			return viewer.View(cid, func(b []byte) error {
				cnt++
				if cnt%10000 == 0 {
					log.Infof("importing block %d", cnt)
				}
				blk, _ := blocks.NewBlockWithCid(b, cid)
				return dstBs.Put(blk)
			})
		})

		return err
	},
}

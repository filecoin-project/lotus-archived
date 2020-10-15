package slashfilter

import (
	"fmt"

	"golang.org/x/xerrors"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
)

type SlashFilter struct {
	byEpoch   ds.Datastore // double-fork mining faults, parent-grinding fault
	byParents ds.Datastore // time-offset mining faults
}

func New(dstore ds.Batching) *SlashFilter {
	return &SlashFilter{
		byEpoch:   namespace.Wrap(dstore, ds.NewKey("/slashfilter/epoch")),
		byParents: namespace.Wrap(dstore, ds.NewKey("/slashfilter/parents")),
	}
}

// A = bh
// B = res[0]
func (f *SlashFilter) CheckBlock(bh *types.BlockHeader, parentEpoch abi.ChainEpoch) ([]cid.Cid, error) {
	epochKey := ds.NewKey(fmt.Sprintf("/%s/%d", bh.Miner, bh.Height))
	{
		// double-fork mining (2 blocks at one epoch)
		if witness, err := checkFault(f.byEpoch, epochKey, bh, "double-fork mining faults"); err != nil {
			return []cid.Cid{witness}, err
		}
	}

	parentsKey := ds.NewKey(fmt.Sprintf("/%s/%x", bh.Miner, types.NewTipSetKey(bh.Parents...).Bytes()))
	{
		// time-offset mining faults (2 blocks with the same parents)
		if witness, err := checkFault(f.byParents, parentsKey, bh, "time-offset mining faults"); err != nil {
			return []cid.Cid{witness}, err
		}
	}

	if parentEpoch != -1 {
		// parent-grinding fault (didn't mine on top of our own block)

		// First check if we have mined a block on the parent epoch
		parentEpochKey := ds.NewKey(fmt.Sprintf("/%s/%d", bh.Miner, parentEpoch))
		have, err := f.byEpoch.Has(parentEpochKey)
		if err != nil {
			return nil, err
		}

		if have {
			// If we had, make sure it's in our parent tipset
			cidb, err := f.byEpoch.Get(parentEpochKey)
			if err != nil {
				return nil, xerrors.Errorf("getting other block cid: %w", err)
			}

			_, parent, err := cid.CidFromBytes(cidb)
			if err != nil {
				return nil, err
			}

			var found bool
			for _, c := range bh.Parents {
				if c.Equals(parent) {
					found = true
				}
			}

			if !found {
				return nil, xerrors.Errorf("produced block would trigger 'parent-grinding fault' consensus fault; miner: %s; bh: %s, expected parent: %s", bh.Miner, bh.Cid(), parent)
			}
		}
	}

	if err := f.byParents.Put(parentsKey, bh.Cid().Bytes()); err != nil {
		return nil, xerrors.Errorf("putting byEpoch entry: %w", err)
	}

	if err := f.byEpoch.Put(epochKey, bh.Cid().Bytes()); err != nil {
		return nil, xerrors.Errorf("putting byEpoch entry: %w", err)
	}

	return nil, nil
}

func (f *SlashFilter) MinedBlock(bh *types.BlockHeader, parentEpoch abi.ChainEpoch) error {
	_, err := f.CheckBlock(bh, parentEpoch)
	return err
}

func checkFault(t ds.Datastore, key ds.Key, bh *types.BlockHeader, faultType string) (cid.Cid, error) {
	fault, err := t.Has(key)
	if err != nil {
		return cid.Undef, err
	}

	if fault {
		cidb, err := t.Get(key)
		if err != nil {
			return cid.Undef, xerrors.Errorf("getting other block cid: %w", err)
		}

		_, other, err := cid.CidFromBytes(cidb)
		if err != nil {
			return cid.Undef, err
		}

		if other == bh.Cid() {
			return cid.Undef, nil
		}

		return other, xerrors.Errorf("produced block would trigger '%s' consensus fault; miner: %s; bh: %s, other: %s", faultType, bh.Miner, bh.Cid(), other)
	}

	return cid.Undef, nil
}

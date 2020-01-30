package storage

import (
	"bytes"
	"context"
	"math"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-sectorbuilder"
	s2 "github.com/filecoin-project/go-storage-miner"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/events"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"
)

type TicketFn func(context.Context) (*sectorbuilder.SealTicket, error)

type StorageMinerNodeAdapter struct {
	api    storageMinerApi
	events *events.Events

	maddr address.Address
	waddr address.Address

	tktFn TicketFn
}

func NewStorageMinerNodeAdapter(n storageMinerApi) *StorageMinerNodeAdapter {
	return &StorageMinerNodeAdapter{api: n}
}

func (m *StorageMinerNodeAdapter) SendSelfDeals(ctx context.Context, pieces ...s2.PieceInfo) (cid.Cid, error) {
	deals := make([]actors.StorageDealProposal, len(pieces))
	for i, p := range pieces {
		sdp := actors.StorageDealProposal{
			PieceRef:             p.CommP[:],
			PieceSize:            p.Size,
			Client:               m.waddr,
			Provider:             m.maddr,
			ProposalExpiration:   math.MaxUint64,
			Duration:             math.MaxUint64 / 2, // /2 because overflows
			StoragePricePerEpoch: types.NewInt(0),
			StorageCollateral:    types.NewInt(0),
			ProposerSignature:    nil, // nil because self dealing
		}

		deals[i] = sdp
	}

	params, aerr := actors.SerializeParams(&actors.PublishStorageDealsParams{
		Deals: deals,
	})
	if aerr != nil {
		return cid.Undef, xerrors.Errorf("serializing PublishStorageDeals params failed: ", aerr)
	}

	msg, err := m.api.MpoolPushMessage(ctx, &types.Message{
		To:       actors.StorageMarketAddress,
		From:     m.waddr,
		Value:    types.NewInt(0),
		GasPrice: types.NewInt(0),
		GasLimit: types.NewInt(1000000),
		Method:   actors.SMAMethods.PublishStorageDeals,
		Params:   params,
	})
	if err != nil {
		return cid.Undef, err
	}

	return msg.Cid(), nil
}

func (m *StorageMinerNodeAdapter) WaitForSelfDeals(ctx context.Context, publicStorageDealsMsgId cid.Cid) ([]uint64, uint8, error) {
	r, err := m.api.StateWaitMsg(ctx, publicStorageDealsMsgId)
	if err != nil {
		return nil, 0, err
	}

	if r.Receipt.ExitCode != 0 {
		return nil, 0, nil
	}

	var resp actors.PublishStorageDealResponse
	if err := resp.UnmarshalCBOR(bytes.NewReader(r.Receipt.Return)); err != nil {
		return nil, 0, err
	}

	return resp.DealIDs, r.Receipt.ExitCode, nil
}

func (m *StorageMinerNodeAdapter) SendPreCommitSector(ctx context.Context, sectorID uint64, commR []byte, ticket s2.SealTicket, pieces ...s2.Piece) (cid.Cid, error) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) WaitForPreCommitSector(context.Context, cid.Cid) (uint64, uint8, error) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) SendProveCommitSector(ctx context.Context, sectorID uint64, proof []byte, dealids ...uint64) (cid.Cid, error) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) WaitForProveCommitSector(context.Context, cid.Cid) (uint64, uint8, error) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) SendReportFaults(ctx context.Context, sectorIDs ...uint64) (cid.Cid, error) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) WaitForReportFaults(context.Context, cid.Cid) (uint8, error) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) GetSealTicket(context.Context) (s2.SealTicket, error) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) GetReplicaCommitmentByID(ctx context.Context, sectorID uint64) (commR []byte, wasFound bool, err error) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) GetSealSeed(ctx context.Context, preCommitMsg cid.Cid, interval uint64) (seed <-chan s2.SealSeed, err <-chan error, invalidated <-chan struct{}, done <-chan struct{}) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) CheckPieces(ctx context.Context, sectorID uint64, pieces []s2.Piece) *s2.CheckPiecesError {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) CheckSealing(ctx context.Context, commD []byte, dealIDs []uint64) *s2.CheckSealingError {
	panic("implement me")
}

package storage

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-sectorbuilder"
	s2 "github.com/filecoin-project/go-storage-miner"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/events"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"
	"math"
	"runtime"
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

	smsg, err := m.api.MpoolPushMessage(ctx, &types.Message{
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

	return smsg.Cid(), nil
}

func (m *StorageMinerNodeAdapter) WaitForSelfDeals(context.Context, cid.Cid) ([]uint64, uint8, error) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) SendPreCommitSector(ctx context.Context, sectorID uint64, commR []byte, ticket storage.SealTicket, pieces ...storage.Piece) (cid.Cid, error) {
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

func (m *StorageMinerNodeAdapter) GetSealTicket(context.Context) (storage.SealTicket, error) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) GetReplicaCommitmentByID(ctx context.Context, sectorID uint64) (commR []byte, wasFound bool, err error) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) GetSealSeed(ctx context.Context, preCommitMsg cid.Cid, interval uint64) (seed <-chan storage.SealSeed, err <-chan error, invalidated <-chan struct{}, done <-chan struct{}) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) CheckPieces(ctx context.Context, sectorID uint64, pieces []storage.Piece) *storage.CheckPiecesError {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) CheckSealing(ctx context.Context, commD []byte, dealIDs []uint64) *storage.CheckSealingError {
	panic("implement me")
}

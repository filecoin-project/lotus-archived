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

func (m *StorageMinerNodeAdapter) WaitForSelfDeals(ctx context.Context, publicStorageDealsMsgCid cid.Cid) ([]uint64, uint8, error) {
	r, err := m.api.StateWaitMsg(ctx, publicStorageDealsMsgCid)
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
	dealIDs := make([]uint64, len(pieces))
	for idx, p := range pieces {
		dealIDs[idx] = p.DealID
	}

	params := &actors.SectorPreCommitInfo{
		SectorNumber: sectorID,
		CommR:        commR,
		SealEpoch:    ticket.BlockHeight,
		DealIDs:      dealIDs,
	}
	enc, aerr := actors.SerializeParams(params)
	if aerr != nil {
		return cid.Undef, xerrors.Errorf("could not serialize commit sector parameters: %w", aerr)
	}

	msg := &types.Message{
		To:       m.maddr,
		From:     m.waddr,
		Method:   actors.MAMethods.PreCommitSector,
		Params:   enc,
		Value:    types.NewInt(0), // TODO: need to ensure sufficient collateral
		GasLimit: types.NewInt(1000000 /* i dont know help */),
		GasPrice: types.NewInt(1),
	}

	log.Info("submitting precommit for sector: ", sectorID)
	smsg, err := m.api.MpoolPushMessage(ctx, msg)
	if err != nil {
		return cid.Undef, xerrors.Errorf("pushing message to mpool: %w", err)
	}

	return smsg.Cid(), nil
}

func (m *StorageMinerNodeAdapter) WaitForPreCommitSector(ctx context.Context, preCommitSectorMsgCid cid.Cid) (uint64, uint8, error) {
	panic("not used - delete")

	return 0, 0, nil
}

func (m *StorageMinerNodeAdapter) SendProveCommitSector(ctx context.Context, sectorID uint64, proof []byte, dealIDs ...uint64) (cid.Cid, error) {
	// TODO: Consider splitting states and persist proof for faster recovery

	params := &actors.SectorProveCommitInfo{
		Proof:    proof,
		SectorID: sectorID,
		DealIDs:  dealIDs,
	}

	enc, aerr := actors.SerializeParams(params)
	if aerr != nil {
		return cid.Undef, xerrors.Errorf("could not serialize commit sector parameters: %w", aerr)
	}

	msg := &types.Message{
		To:       m.maddr,
		From:     m.waddr,
		Method:   actors.MAMethods.ProveCommitSector,
		Params:   enc,
		Value:    types.NewInt(0), // TODO: need to ensure sufficient collateral
		GasLimit: types.NewInt(1000000 /* i dont know help */),
		GasPrice: types.NewInt(1),
	}

	// TODO: check seed / ticket are up to date

	smsg, err := m.api.MpoolPushMessage(ctx, msg)
	if err != nil {
		return cid.Undef, xerrors.Errorf("pushing message to mpool: %w", err)
	}

	return smsg.Cid(), nil
}

func (m *StorageMinerNodeAdapter) WaitForProveCommitSector(ctx context.Context, proveCommitSectorMsgCid cid.Cid) (uint64, uint8, error) {
	mw, err := m.api.StateWaitMsg(ctx, proveCommitSectorMsgCid)
	if err != nil {
		return 0, 0, xerrors.Errorf("failed to wait for porep inclusion: %w", err)
	}

	return 0, mw.Receipt.ExitCode, nil
}

func (m *StorageMinerNodeAdapter) SendReportFaults(ctx context.Context, sectorIDs ...uint64) (cid.Cid, error) {
	// TODO: check if the fault has already been reported, and that this sector is even valid

	bf := types.NewBitField()
	for _, id := range sectorIDs {
		bf.Set(id)
	}

	enc, aerr := actors.SerializeParams(&actors.DeclareFaultsParams{bf})
	if aerr != nil {
		return cid.Undef, xerrors.Errorf("failed to serialize declare fault params: %w", aerr)
	}

	msg := &types.Message{
		To:       m.maddr,
		From:     m.waddr,
		Method:   actors.MAMethods.DeclareFaults,
		Params:   enc,
		Value:    types.NewInt(0), // TODO: need to ensure sufficient collateral
		GasLimit: types.NewInt(1000000 /* i dont know help */),
		GasPrice: types.NewInt(1),
	}

	smsg, err := m.api.MpoolPushMessage(ctx, msg)
	if err != nil {
		return cid.Undef, xerrors.Errorf("failed to push declare faults message to network: %w", err)
	}

	return smsg.Cid(), nil
}

func (m *StorageMinerNodeAdapter) WaitForReportFaults(context.Context, cid.Cid) (uint8, error) {
	panic("implement me")
}

func (m *StorageMinerNodeAdapter) GetSealTicket(ctx context.Context) (s2.SealTicket, error) {
	ticket, err := m.tktFn(ctx)
	if err != nil {
		return s2.SealTicket{}, xerrors.Errorf("getting ticket failed: %w", err)
	}

	return s2.SealTicket{
		BlockHeight: ticket.BlockHeight,
		TicketBytes: ticket.TicketBytes[:],
	}, nil
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

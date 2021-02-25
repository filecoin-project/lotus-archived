package v0

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	datatransfer "github.com/filecoin-project/go-data-transfer"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-multistore"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/dline"
	stnetwork "github.com/filecoin-project/go-state-types/network"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/actors/builtin/paych"
	"github.com/filecoin-project/lotus/chain/types"
	marketevents "github.com/filecoin-project/lotus/markets/loggers"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
)

type WrapperV1 struct {
	api.Common // todo should be v1 wrapper too

	Internal api.FullNode
}

func (c *WrapperV1) StateMinerSectorCount(ctx context.Context, addr address.Address, tsk types.TipSetKey) (api.MinerSectors, error) {
	return c.Internal.StateMinerSectorCount(ctx, addr, tsk)
}

func (c *WrapperV1) ClientListImports(ctx context.Context) ([]api.Import, error) {
	return c.Internal.ClientListImports(ctx)
}

func (c *WrapperV1) ClientRemoveImport(ctx context.Context, importID multistore.StoreID) error {
	return c.Internal.ClientRemoveImport(ctx, importID)
}

func (c *WrapperV1) ClientImport(ctx context.Context, ref api.FileRef) (*api.ImportRes, error) {
	return c.Internal.ClientImport(ctx, ref)
}

func (c *WrapperV1) ClientHasLocal(ctx context.Context, root cid.Cid) (bool, error) {
	return c.Internal.ClientHasLocal(ctx, root)
}

func (c *WrapperV1) ClientFindData(ctx context.Context, root cid.Cid, piece *cid.Cid) ([]api.QueryOffer, error) {
	return c.Internal.ClientFindData(ctx, root, piece)
}

func (c *WrapperV1) ClientMinerQueryOffer(ctx context.Context, miner address.Address, root cid.Cid, piece *cid.Cid) (api.QueryOffer, error) {
	return c.Internal.ClientMinerQueryOffer(ctx, miner, root, piece)
}

func (c *WrapperV1) ClientStartDeal(ctx context.Context, params *api.StartDealParams) (*cid.Cid, error) {
	return c.Internal.ClientStartDeal(ctx, params)
}

func (c *WrapperV1) ClientGetDealInfo(ctx context.Context, deal cid.Cid) (*api.DealInfo, error) {
	return c.Internal.ClientGetDealInfo(ctx, deal)
}

func (c *WrapperV1) ClientGetDealStatus(ctx context.Context, statusCode uint64) (string, error) {
	return c.Internal.ClientGetDealStatus(ctx, statusCode)
}

func (c *WrapperV1) ClientListDeals(ctx context.Context) ([]api.DealInfo, error) {
	return c.Internal.ClientListDeals(ctx)
}

func (c *WrapperV1) ClientGetDealUpdates(ctx context.Context) (<-chan api.DealInfo, error) {
	return c.Internal.ClientGetDealUpdates(ctx)
}

func (c *WrapperV1) ClientRetrieve(ctx context.Context, order api.RetrievalOrder, ref *api.FileRef) error {
	return c.Internal.ClientRetrieve(ctx, order, ref)
}

func (c *WrapperV1) ClientRetrieveWithEvents(ctx context.Context, order api.RetrievalOrder, ref *api.FileRef) (<-chan marketevents.RetrievalEvent, error) {
	return c.Internal.ClientRetrieveWithEvents(ctx, order, ref)
}

func (c *WrapperV1) ClientQueryAsk(ctx context.Context, p peer.ID, miner address.Address) (*storagemarket.StorageAsk, error) {
	return c.Internal.ClientQueryAsk(ctx, p, miner)
}

func (c *WrapperV1) ClientDealPieceCID(ctx context.Context, root cid.Cid) (api.DataCIDSize, error) {
	return c.Internal.ClientDealPieceCID(ctx, root)
}

func (c *WrapperV1) ClientCalcCommP(ctx context.Context, inpath string) (*api.CommPRet, error) {
	return c.Internal.ClientCalcCommP(ctx, inpath)
}

func (c *WrapperV1) ClientGenCar(ctx context.Context, ref api.FileRef, outpath string) error {
	return c.Internal.ClientGenCar(ctx, ref, outpath)
}

func (c *WrapperV1) ClientDealSize(ctx context.Context, root cid.Cid) (api.DataSize, error) {
	return c.Internal.ClientDealSize(ctx, root)
}

func (c *WrapperV1) ClientListDataTransfers(ctx context.Context) ([]api.DataTransferChannel, error) {
	return c.Internal.ClientListDataTransfers(ctx)
}

func (c *WrapperV1) ClientDataTransferUpdates(ctx context.Context) (<-chan api.DataTransferChannel, error) {
	return c.Internal.ClientDataTransferUpdates(ctx)
}

func (c *WrapperV1) ClientRestartDataTransfer(ctx context.Context, transferID datatransfer.TransferID, otherPeer peer.ID, isInitiator bool) error {
	return c.Internal.ClientRestartDataTransfer(ctx, transferID, otherPeer, isInitiator)
}

func (c *WrapperV1) ClientCancelDataTransfer(ctx context.Context, transferID datatransfer.TransferID, otherPeer peer.ID, isInitiator bool) error {
	return c.Internal.ClientCancelDataTransfer(ctx, transferID, otherPeer, isInitiator)
}

func (c *WrapperV1) ClientRetrieveTryRestartInsufficientFunds(ctx context.Context, paymentChannel address.Address) error {
	return c.Internal.ClientRetrieveTryRestartInsufficientFunds(ctx, paymentChannel)
}

func (c *WrapperV1) GasEstimateGasPremium(ctx context.Context, nblocksincl uint64, sender address.Address, gaslimit int64, tsk types.TipSetKey) (types.BigInt, error) {
	return c.Internal.GasEstimateGasPremium(ctx, nblocksincl, sender, gaslimit, tsk)
}

func (c *WrapperV1) GasEstimateFeeCap(ctx context.Context, msg *types.Message, maxqueueblks int64, tsk types.TipSetKey) (types.BigInt, error) {
	return c.Internal.GasEstimateFeeCap(ctx, msg, maxqueueblks, tsk)
}

func (c *WrapperV1) GasEstimateMessageGas(ctx context.Context, msg *types.Message, spec *api.MessageSendSpec, tsk types.TipSetKey) (*types.Message, error) {
	return c.Internal.GasEstimateMessageGas(ctx, msg, spec, tsk)
}

func (c *WrapperV1) GasEstimateGasLimit(ctx context.Context, msg *types.Message, tsk types.TipSetKey) (int64, error) {
	return c.Internal.GasEstimateGasLimit(ctx, msg, tsk)
}

func (c *WrapperV1) MpoolGetConfig(ctx context.Context) (*types.MpoolConfig, error) {
	return c.Internal.MpoolGetConfig(ctx)
}

func (c *WrapperV1) MpoolSetConfig(ctx context.Context, cfg *types.MpoolConfig) error {
	return c.Internal.MpoolSetConfig(ctx, cfg)
}

func (c *WrapperV1) MpoolSelect(ctx context.Context, tsk types.TipSetKey, tq float64) ([]*types.SignedMessage, error) {
	return c.Internal.MpoolSelect(ctx, tsk, tq)
}

func (c *WrapperV1) MpoolPending(ctx context.Context, tsk types.TipSetKey) ([]*types.SignedMessage, error) {
	return c.Internal.MpoolPending(ctx, tsk)
}

func (c *WrapperV1) MpoolClear(ctx context.Context, local bool) error {
	return c.Internal.MpoolClear(ctx, local)
}

func (c *WrapperV1) MpoolPush(ctx context.Context, smsg *types.SignedMessage) (cid.Cid, error) {
	return c.Internal.MpoolPush(ctx, smsg)
}

func (c *WrapperV1) MpoolPushUntrusted(ctx context.Context, smsg *types.SignedMessage) (cid.Cid, error) {
	return c.Internal.MpoolPushUntrusted(ctx, smsg)
}

func (c *WrapperV1) MpoolPushMessage(ctx context.Context, msg *types.Message, spec *api.MessageSendSpec) (*types.SignedMessage, error) {
	return c.Internal.MpoolPushMessage(ctx, msg, spec)
}

func (c *WrapperV1) MpoolBatchPush(ctx context.Context, smsgs []*types.SignedMessage) ([]cid.Cid, error) {
	return c.Internal.MpoolBatchPush(ctx, smsgs)
}

func (c *WrapperV1) MpoolBatchPushUntrusted(ctx context.Context, smsgs []*types.SignedMessage) ([]cid.Cid, error) {
	return c.Internal.MpoolBatchPushUntrusted(ctx, smsgs)
}

func (c *WrapperV1) MpoolBatchPushMessage(ctx context.Context, msgs []*types.Message, spec *api.MessageSendSpec) ([]*types.SignedMessage, error) {
	return c.Internal.MpoolBatchPushMessage(ctx, msgs, spec)
}

func (c *WrapperV1) MpoolSub(ctx context.Context) (<-chan api.MpoolUpdate, error) {
	return c.Internal.MpoolSub(ctx)
}

func (c *WrapperV1) MinerGetBaseInfo(ctx context.Context, maddr address.Address, epoch abi.ChainEpoch, tsk types.TipSetKey) (*api.MiningBaseInfo, error) {
	return c.Internal.MinerGetBaseInfo(ctx, maddr, epoch, tsk)
}

func (c *WrapperV1) MinerCreateBlock(ctx context.Context, bt *api.BlockTemplate) (*types.BlockMsg, error) {
	return c.Internal.MinerCreateBlock(ctx, bt)
}

func (c *WrapperV1) ChainHead(ctx context.Context) (*types.TipSet, error) {
	return c.Internal.ChainHead(ctx)
}

func (c *WrapperV1) ChainGetRandomnessFromTickets(ctx context.Context, tsk types.TipSetKey, personalization crypto.DomainSeparationTag, randEpoch abi.ChainEpoch, entropy []byte) (abi.Randomness, error) {
	return c.Internal.ChainGetRandomnessFromTickets(ctx, tsk, personalization, randEpoch, entropy)
}

func (c *WrapperV1) ChainGetRandomnessFromBeacon(ctx context.Context, tsk types.TipSetKey, personalization crypto.DomainSeparationTag, randEpoch abi.ChainEpoch, entropy []byte) (abi.Randomness, error) {
	return c.Internal.ChainGetRandomnessFromBeacon(ctx, tsk, personalization, randEpoch, entropy)
}

func (c *WrapperV1) ChainGetTipSetByHeight(ctx context.Context, h abi.ChainEpoch, tsk types.TipSetKey) (*types.TipSet, error) {
	return c.Internal.ChainGetTipSetByHeight(ctx, h, tsk)
}

func (c *WrapperV1) WalletNew(ctx context.Context, typ types.KeyType) (address.Address, error) {
	return c.Internal.WalletNew(ctx, typ)
}

func (c *WrapperV1) WalletHas(ctx context.Context, addr address.Address) (bool, error) {
	return c.Internal.WalletHas(ctx, addr)
}

func (c *WrapperV1) WalletList(ctx context.Context) ([]address.Address, error) {
	return c.Internal.WalletList(ctx)
}

func (c *WrapperV1) WalletBalance(ctx context.Context, a address.Address) (types.BigInt, error) {
	return c.Internal.WalletBalance(ctx, a)
}

func (c *WrapperV1) WalletSign(ctx context.Context, k address.Address, msg []byte) (*crypto.Signature, error) {
	return c.Internal.WalletSign(ctx, k, msg)
}

func (c *WrapperV1) WalletSignMessage(ctx context.Context, k address.Address, msg *types.Message) (*types.SignedMessage, error) {
	return c.Internal.WalletSignMessage(ctx, k, msg)
}

func (c *WrapperV1) WalletVerify(ctx context.Context, k address.Address, msg []byte, sig *crypto.Signature) (bool, error) {
	return c.Internal.WalletVerify(ctx, k, msg, sig)
}

func (c *WrapperV1) WalletDefaultAddress(ctx context.Context) (address.Address, error) {
	return c.Internal.WalletDefaultAddress(ctx)
}

func (c *WrapperV1) WalletSetDefault(ctx context.Context, a address.Address) error {
	return c.Internal.WalletSetDefault(ctx, a)
}

func (c *WrapperV1) WalletExport(ctx context.Context, a address.Address) (*types.KeyInfo, error) {
	return c.Internal.WalletExport(ctx, a)
}

func (c *WrapperV1) WalletImport(ctx context.Context, ki *types.KeyInfo) (address.Address, error) {
	return c.Internal.WalletImport(ctx, ki)
}

func (c *WrapperV1) WalletDelete(ctx context.Context, addr address.Address) error {
	return c.Internal.WalletDelete(ctx, addr)
}

func (c *WrapperV1) WalletValidateAddress(ctx context.Context, str string) (address.Address, error) {
	return c.Internal.WalletValidateAddress(ctx, str)
}

func (c *WrapperV1) MpoolGetNonce(ctx context.Context, addr address.Address) (uint64, error) {
	return c.Internal.MpoolGetNonce(ctx, addr)
}

func (c *WrapperV1) ChainGetBlock(ctx context.Context, b cid.Cid) (*types.BlockHeader, error) {
	return c.Internal.ChainGetBlock(ctx, b)
}

func (c *WrapperV1) ChainGetTipSet(ctx context.Context, key types.TipSetKey) (*types.TipSet, error) {
	return c.Internal.ChainGetTipSet(ctx, key)
}

func (c *WrapperV1) ChainGetBlockMessages(ctx context.Context, b cid.Cid) (*api.BlockMessages, error) {
	return c.Internal.ChainGetBlockMessages(ctx, b)
}

func (c *WrapperV1) ChainGetParentReceipts(ctx context.Context, b cid.Cid) ([]*types.MessageReceipt, error) {
	return c.Internal.ChainGetParentReceipts(ctx, b)
}

func (c *WrapperV1) ChainGetParentMessages(ctx context.Context, b cid.Cid) ([]api.Message, error) {
	return c.Internal.ChainGetParentMessages(ctx, b)
}

func (c *WrapperV1) ChainNotify(ctx context.Context) (<-chan []*api.HeadChange, error) {
	return c.Internal.ChainNotify(ctx)
}

func (c *WrapperV1) ChainReadObj(ctx context.Context, obj cid.Cid) ([]byte, error) {
	return c.Internal.ChainReadObj(ctx, obj)
}

func (c *WrapperV1) ChainDeleteObj(ctx context.Context, obj cid.Cid) error {
	return c.Internal.ChainDeleteObj(ctx, obj)
}

func (c *WrapperV1) ChainHasObj(ctx context.Context, o cid.Cid) (bool, error) {
	return c.Internal.ChainHasObj(ctx, o)
}

func (c *WrapperV1) ChainStatObj(ctx context.Context, obj, base cid.Cid) (api.ObjStat, error) {
	return c.Internal.ChainStatObj(ctx, obj, base)
}

func (c *WrapperV1) ChainSetHead(ctx context.Context, tsk types.TipSetKey) error {
	return c.Internal.ChainSetHead(ctx, tsk)
}

func (c *WrapperV1) ChainGetGenesis(ctx context.Context) (*types.TipSet, error) {
	return c.Internal.ChainGetGenesis(ctx)
}

func (c *WrapperV1) ChainTipSetWeight(ctx context.Context, tsk types.TipSetKey) (types.BigInt, error) {
	return c.Internal.ChainTipSetWeight(ctx, tsk)
}

func (c *WrapperV1) ChainGetNode(ctx context.Context, p string) (*api.IpldObject, error) {
	return c.Internal.ChainGetNode(ctx, p)
}

func (c *WrapperV1) ChainGetMessage(ctx context.Context, mc cid.Cid) (*types.Message, error) {
	return c.Internal.ChainGetMessage(ctx, mc)
}

func (c *WrapperV1) ChainGetPath(ctx context.Context, from types.TipSetKey, to types.TipSetKey) ([]*api.HeadChange, error) {
	return c.Internal.ChainGetPath(ctx, from, to)
}

func (c *WrapperV1) ChainExport(ctx context.Context, nroots abi.ChainEpoch, iom bool, tsk types.TipSetKey) (<-chan []byte, error) {
	return c.Internal.ChainExport(ctx, nroots, iom, tsk)
}

func (c *WrapperV1) BeaconGetEntry(ctx context.Context, epoch abi.ChainEpoch) (*types.BeaconEntry, error) {
	return c.Internal.BeaconGetEntry(ctx, epoch)
}

func (c *WrapperV1) SyncState(ctx context.Context) (*api.SyncState, error) {
	return c.Internal.SyncState(ctx)
}

func (c *WrapperV1) SyncSubmitBlock(ctx context.Context, blk *types.BlockMsg) error {
	return c.Internal.SyncSubmitBlock(ctx, blk)
}

func (c *WrapperV1) SyncIncomingBlocks(ctx context.Context) (<-chan *types.BlockHeader, error) {
	return c.Internal.SyncIncomingBlocks(ctx)
}

func (c *WrapperV1) SyncCheckpoint(ctx context.Context, tsk types.TipSetKey) error {
	return c.Internal.SyncCheckpoint(ctx, tsk)
}

func (c *WrapperV1) SyncMarkBad(ctx context.Context, bcid cid.Cid) error {
	return c.Internal.SyncMarkBad(ctx, bcid)
}

func (c *WrapperV1) SyncUnmarkBad(ctx context.Context, bcid cid.Cid) error {
	return c.Internal.SyncUnmarkBad(ctx, bcid)
}

func (c *WrapperV1) SyncUnmarkAllBad(ctx context.Context) error {
	return c.Internal.SyncUnmarkAllBad(ctx)
}

func (c *WrapperV1) SyncCheckBad(ctx context.Context, bcid cid.Cid) (string, error) {
	return c.Internal.SyncCheckBad(ctx, bcid)
}

func (c *WrapperV1) SyncValidateTipset(ctx context.Context, tsk types.TipSetKey) (bool, error) {
	return c.Internal.SyncValidateTipset(ctx, tsk)
}

func (c *WrapperV1) StateNetworkName(ctx context.Context) (dtypes.NetworkName, error) {
	return c.Internal.StateNetworkName(ctx)
}

func (c *WrapperV1) StateMinerSectors(ctx context.Context, addr address.Address, sectorNos *bitfield.BitField, tsk types.TipSetKey) ([]*miner.SectorOnChainInfo, error) {
	return c.Internal.StateMinerSectors(ctx, addr, sectorNos, tsk)
}

func (c *WrapperV1) StateMinerActiveSectors(ctx context.Context, addr address.Address, tsk types.TipSetKey) ([]*miner.SectorOnChainInfo, error) {
	return c.Internal.StateMinerActiveSectors(ctx, addr, tsk)
}

func (c *WrapperV1) StateMinerProvingDeadline(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*dline.Info, error) {
	return c.Internal.StateMinerProvingDeadline(ctx, addr, tsk)
}

func (c *WrapperV1) StateMinerPower(ctx context.Context, a address.Address, tsk types.TipSetKey) (*api.MinerPower, error) {
	return c.Internal.StateMinerPower(ctx, a, tsk)
}

func (c *WrapperV1) StateMinerInfo(ctx context.Context, actor address.Address, tsk types.TipSetKey) (miner.MinerInfo, error) {
	return c.Internal.StateMinerInfo(ctx, actor, tsk)
}

func (c *WrapperV1) StateMinerDeadlines(ctx context.Context, actor address.Address, tsk types.TipSetKey) ([]api.Deadline, error) {
	return c.Internal.StateMinerDeadlines(ctx, actor, tsk)
}

func (c *WrapperV1) StateMinerPartitions(ctx context.Context, m address.Address, dlIdx uint64, tsk types.TipSetKey) ([]api.Partition, error) {
	return c.Internal.StateMinerPartitions(ctx, m, dlIdx, tsk)
}

func (c *WrapperV1) StateMinerFaults(ctx context.Context, actor address.Address, tsk types.TipSetKey) (bitfield.BitField, error) {
	return c.Internal.StateMinerFaults(ctx, actor, tsk)
}

func (c *WrapperV1) StateAllMinerFaults(ctx context.Context, cutoff abi.ChainEpoch, endTsk types.TipSetKey) ([]*api.Fault, error) {
	return c.Internal.StateAllMinerFaults(ctx, cutoff, endTsk)
}

func (c *WrapperV1) StateMinerRecoveries(ctx context.Context, actor address.Address, tsk types.TipSetKey) (bitfield.BitField, error) {
	return c.Internal.StateMinerRecoveries(ctx, actor, tsk)
}

func (c *WrapperV1) StateMinerPreCommitDepositForPower(ctx context.Context, maddr address.Address, pci miner.SectorPreCommitInfo, tsk types.TipSetKey) (types.BigInt, error) {
	return c.Internal.StateMinerPreCommitDepositForPower(ctx, maddr, pci, tsk)
}

func (c *WrapperV1) StateMinerInitialPledgeCollateral(ctx context.Context, maddr address.Address, pci miner.SectorPreCommitInfo, tsk types.TipSetKey) (types.BigInt, error) {
	return c.Internal.StateMinerInitialPledgeCollateral(ctx, maddr, pci, tsk)
}

func (c *WrapperV1) StateMinerAvailableBalance(ctx context.Context, maddr address.Address, tsk types.TipSetKey) (types.BigInt, error) {
	return c.Internal.StateMinerAvailableBalance(ctx, maddr, tsk)
}

func (c *WrapperV1) StateMinerSectorAllocated(ctx context.Context, maddr address.Address, s abi.SectorNumber, tsk types.TipSetKey) (bool, error) {
	return c.Internal.StateMinerSectorAllocated(ctx, maddr, s, tsk)
}

func (c *WrapperV1) StateSectorPreCommitInfo(ctx context.Context, maddr address.Address, n abi.SectorNumber, tsk types.TipSetKey) (miner.SectorPreCommitOnChainInfo, error) {
	return c.Internal.StateSectorPreCommitInfo(ctx, maddr, n, tsk)
}

func (c *WrapperV1) StateSectorGetInfo(ctx context.Context, maddr address.Address, n abi.SectorNumber, tsk types.TipSetKey) (*miner.SectorOnChainInfo, error) {
	return c.Internal.StateSectorGetInfo(ctx, maddr, n, tsk)
}

func (c *WrapperV1) StateSectorExpiration(ctx context.Context, maddr address.Address, n abi.SectorNumber, tsk types.TipSetKey) (*miner.SectorExpiration, error) {
	return c.Internal.StateSectorExpiration(ctx, maddr, n, tsk)
}

func (c *WrapperV1) StateSectorPartition(ctx context.Context, maddr address.Address, sectorNumber abi.SectorNumber, tok types.TipSetKey) (*miner.SectorLocation, error) {
	return c.Internal.StateSectorPartition(ctx, maddr, sectorNumber, tok)
}

func (c *WrapperV1) StateCall(ctx context.Context, msg *types.Message, tsk types.TipSetKey) (*api.InvocResult, error) {
	return c.Internal.StateCall(ctx, msg, tsk)
}

func (c *WrapperV1) StateReplay(ctx context.Context, tsk types.TipSetKey, mc cid.Cid) (*api.InvocResult, error) {
	return c.Internal.StateReplay(ctx, tsk, mc)
}

func (c *WrapperV1) StateGetActor(ctx context.Context, actor address.Address, tsk types.TipSetKey) (*types.Actor, error) {
	return c.Internal.StateGetActor(ctx, actor, tsk)
}

func (c *WrapperV1) StateReadState(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*api.ActorState, error) {
	return c.Internal.StateReadState(ctx, addr, tsk)
}

func (c *WrapperV1) StateWaitMsg(ctx context.Context, msgc cid.Cid, confidence uint64) (*api.MsgLookup, error) {
	return c.Internal.StateWaitMsg(ctx, msgc, confidence)
}

func (c *WrapperV1) StateWaitMsgLimited(ctx context.Context, msgc cid.Cid, confidence uint64, limit abi.ChainEpoch) (*api.MsgLookup, error) {
	return c.Internal.StateWaitMsgLimited(ctx, msgc, confidence, limit)
}

func (c *WrapperV1) StateSearchMsg(ctx context.Context, msgc cid.Cid) (*api.MsgLookup, error) {
	return c.Internal.StateSearchMsg(ctx, msgc)
}

func (c *WrapperV1) StateSearchMsgLimited(ctx context.Context, msgc cid.Cid, limit abi.ChainEpoch) (*api.MsgLookup, error) {
	return c.Internal.StateSearchMsgLimited(ctx, msgc, limit)
}

func (c *WrapperV1) StateListMiners(ctx context.Context, tsk types.TipSetKey) ([]address.Address, error) {
	return c.Internal.StateListMiners(ctx, tsk)
}

func (c *WrapperV1) StateListActors(ctx context.Context, tsk types.TipSetKey) ([]address.Address, error) {
	return c.Internal.StateListActors(ctx, tsk)
}

func (c *WrapperV1) StateMarketBalance(ctx context.Context, addr address.Address, tsk types.TipSetKey) (api.MarketBalance, error) {
	return c.Internal.StateMarketBalance(ctx, addr, tsk)
}

func (c *WrapperV1) StateMarketParticipants(ctx context.Context, tsk types.TipSetKey) (map[string]api.MarketBalance, error) {
	return c.Internal.StateMarketParticipants(ctx, tsk)
}

func (c *WrapperV1) StateMarketDeals(ctx context.Context, tsk types.TipSetKey) (map[string]api.MarketDeal, error) {
	return c.Internal.StateMarketDeals(ctx, tsk)
}

func (c *WrapperV1) StateMarketStorageDeal(ctx context.Context, dealid abi.DealID, tsk types.TipSetKey) (*api.MarketDeal, error) {
	return c.Internal.StateMarketStorageDeal(ctx, dealid, tsk)
}

func (c *WrapperV1) StateLookupID(ctx context.Context, addr address.Address, tsk types.TipSetKey) (address.Address, error) {
	return c.Internal.StateLookupID(ctx, addr, tsk)
}

func (c *WrapperV1) StateAccountKey(ctx context.Context, addr address.Address, tsk types.TipSetKey) (address.Address, error) {
	return c.Internal.StateAccountKey(ctx, addr, tsk)
}

func (c *WrapperV1) StateChangedActors(ctx context.Context, olnstate cid.Cid, newstate cid.Cid) (map[string]types.Actor, error) {
	return c.Internal.StateChangedActors(ctx, olnstate, newstate)
}

func (c *WrapperV1) StateGetReceipt(ctx context.Context, msg cid.Cid, tsk types.TipSetKey) (*types.MessageReceipt, error) {
	return c.Internal.StateGetReceipt(ctx, msg, tsk)
}

func (c *WrapperV1) StateListMessages(ctx context.Context, match *api.MessageMatch, tsk types.TipSetKey, toht abi.ChainEpoch) ([]cid.Cid, error) {
	return c.Internal.StateListMessages(ctx, match, tsk, toht)
}

func (c *WrapperV1) StateDecodeParams(ctx context.Context, toAddr address.Address, method abi.MethodNum, params []byte, tsk types.TipSetKey) (interface{}, error) {
	return c.Internal.StateDecodeParams(ctx, toAddr, method, params, tsk)
}

func (c *WrapperV1) StateCompute(ctx context.Context, height abi.ChainEpoch, msgs []*types.Message, tsk types.TipSetKey) (*api.ComputeStateOutput, error) {
	return c.Internal.StateCompute(ctx, height, msgs, tsk)
}

func (c *WrapperV1) StateVerifierStatus(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*abi.StoragePower, error) {
	return c.Internal.StateVerifierStatus(ctx, addr, tsk)
}

func (c *WrapperV1) StateVerifiedClientStatus(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*abi.StoragePower, error) {
	return c.Internal.StateVerifiedClientStatus(ctx, addr, tsk)
}

func (c *WrapperV1) StateVerifiedRegistryRootKey(ctx context.Context, tsk types.TipSetKey) (address.Address, error) {
	return c.Internal.StateVerifiedRegistryRootKey(ctx, tsk)
}

func (c *WrapperV1) StateDealProviderCollateralBounds(ctx context.Context, size abi.PaddedPieceSize, verified bool, tsk types.TipSetKey) (api.DealCollateralBounds, error) {
	return c.Internal.StateDealProviderCollateralBounds(ctx, size, verified, tsk)
}

func (c *WrapperV1) StateCirculatingSupply(ctx context.Context, tsk types.TipSetKey) (abi.TokenAmount, error) {
	return c.Internal.StateCirculatingSupply(ctx, tsk)
}

func (c *WrapperV1) StateVMCirculatingSupplyInternal(ctx context.Context, tsk types.TipSetKey) (api.CirculatingSupply, error) {
	return c.Internal.StateVMCirculatingSupplyInternal(ctx, tsk)
}

func (c *WrapperV1) StateNetworkVersion(ctx context.Context, tsk types.TipSetKey) (stnetwork.Version, error) {
	return c.Internal.StateNetworkVersion(ctx, tsk)
}

func (c *WrapperV1) MsigGetAvailableBalance(ctx context.Context, a address.Address, tsk types.TipSetKey) (types.BigInt, error) {
	return c.Internal.MsigGetAvailableBalance(ctx, a, tsk)
}

func (c *WrapperV1) MsigGetVestingSchedule(ctx context.Context, a address.Address, tsk types.TipSetKey) (api.MsigVesting, error) {
	return c.Internal.MsigGetVestingSchedule(ctx, a, tsk)
}

func (c *WrapperV1) MsigGetVested(ctx context.Context, a address.Address, sTsk types.TipSetKey, eTsk types.TipSetKey) (types.BigInt, error) {
	return c.Internal.MsigGetVested(ctx, a, sTsk, eTsk)
}

func (c *WrapperV1) MsigGetPending(ctx context.Context, a address.Address, tsk types.TipSetKey) ([]*api.MsigTransaction, error) {
	return c.Internal.MsigGetPending(ctx, a, tsk)
}

func (c *WrapperV1) MsigCreate(ctx context.Context, req uint64, addrs []address.Address, duration abi.ChainEpoch, val types.BigInt, src address.Address, gp types.BigInt) (cid.Cid, error) {
	return c.Internal.MsigCreate(ctx, req, addrs, duration, val, src, gp)
}

func (c *WrapperV1) MsigPropose(ctx context.Context, msig address.Address, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (cid.Cid, error) {
	return c.Internal.MsigPropose(ctx, msig, to, amt, src, method, params)
}

func (c *WrapperV1) MsigApprove(ctx context.Context, msig address.Address, txID uint64, signer address.Address) (cid.Cid, error) {
	return c.Internal.MsigApprove(ctx, msig, txID, signer)
}

func (c *WrapperV1) MsigApproveTxnHash(ctx context.Context, msig address.Address, txID uint64, proposer address.Address, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (cid.Cid, error) {
	return c.Internal.MsigApproveTxnHash(ctx, msig, txID, proposer, to, amt, src, method, params)
}

func (c *WrapperV1) MsigCancel(ctx context.Context, msig address.Address, txID uint64, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (cid.Cid, error) {
	return c.Internal.MsigCancel(ctx, msig, txID, to, amt, src, method, params)
}

func (c *WrapperV1) MsigAddPropose(ctx context.Context, msig address.Address, src address.Address, newAdd address.Address, inc bool) (cid.Cid, error) {
	return c.Internal.MsigAddPropose(ctx, msig, src, newAdd, inc)
}

func (c *WrapperV1) MsigAddApprove(ctx context.Context, msig address.Address, src address.Address, txID uint64, proposer address.Address, newAdd address.Address, inc bool) (cid.Cid, error) {
	return c.Internal.MsigAddApprove(ctx, msig, src, txID, proposer, newAdd, inc)
}

func (c *WrapperV1) MsigAddCancel(ctx context.Context, msig address.Address, src address.Address, txID uint64, newAdd address.Address, inc bool) (cid.Cid, error) {
	return c.Internal.MsigAddCancel(ctx, msig, src, txID, newAdd, inc)
}

func (c *WrapperV1) MsigSwapPropose(ctx context.Context, msig address.Address, src address.Address, oldAdd address.Address, newAdd address.Address) (cid.Cid, error) {
	return c.Internal.MsigSwapPropose(ctx, msig, src, oldAdd, newAdd)
}

func (c *WrapperV1) MsigSwapApprove(ctx context.Context, msig address.Address, src address.Address, txID uint64, proposer address.Address, oldAdd address.Address, newAdd address.Address) (cid.Cid, error) {
	return c.Internal.MsigSwapApprove(ctx, msig, src, txID, proposer, oldAdd, newAdd)
}

func (c *WrapperV1) MsigSwapCancel(ctx context.Context, msig address.Address, src address.Address, txID uint64, oldAdd address.Address, newAdd address.Address) (cid.Cid, error) {
	return c.Internal.MsigSwapCancel(ctx, msig, src, txID, oldAdd, newAdd)
}

func (c *WrapperV1) MsigRemoveSigner(ctx context.Context, msig address.Address, proposer address.Address, toRemove address.Address, decrease bool) (cid.Cid, error) {
	return c.Internal.MsigRemoveSigner(ctx, msig, proposer, toRemove, decrease)
}

func (c *WrapperV1) MarketAddBalance(ctx context.Context, wallet address.Address, addr address.Address, amt types.BigInt) (cid.Cid, error) {
	return c.Internal.MarketAddBalance(ctx, wallet, addr, amt)
}

func (c *WrapperV1) MarketGetReserved(ctx context.Context, addr address.Address) (types.BigInt, error) {
	return c.Internal.MarketGetReserved(ctx, addr)
}

func (c *WrapperV1) MarketReserveFunds(ctx context.Context, wallet address.Address, addr address.Address, amt types.BigInt) (cid.Cid, error) {
	return c.Internal.MarketReserveFunds(ctx, wallet, addr, amt)
}

func (c *WrapperV1) MarketReleaseFunds(ctx context.Context, addr address.Address, amt types.BigInt) error {
	return c.Internal.MarketReleaseFunds(ctx, addr, amt)
}

func (c *WrapperV1) MarketWithdraw(ctx context.Context, wallet, addr address.Address, amt types.BigInt) (cid.Cid, error) {
	return c.Internal.MarketWithdraw(ctx, wallet, addr, amt)
}

func (c *WrapperV1) PaychGet(ctx context.Context, from, to address.Address, amt types.BigInt) (*api.ChannelInfo, error) {
	return c.Internal.PaychGet(ctx, from, to, amt)
}

func (c *WrapperV1) PaychGetWaitReady(ctx context.Context, sentinel cid.Cid) (address.Address, error) {
	return c.Internal.PaychGetWaitReady(ctx, sentinel)
}

func (c *WrapperV1) PaychAvailableFunds(ctx context.Context, ch address.Address) (*api.ChannelAvailableFunds, error) {
	return c.Internal.PaychAvailableFunds(ctx, ch)
}

func (c *WrapperV1) PaychAvailableFundsByFromTo(ctx context.Context, from, to address.Address) (*api.ChannelAvailableFunds, error) {
	return c.Internal.PaychAvailableFundsByFromTo(ctx, from, to)
}

func (c *WrapperV1) PaychList(ctx context.Context) ([]address.Address, error) {
	return c.Internal.PaychList(ctx)
}

func (c *WrapperV1) PaychStatus(ctx context.Context, pch address.Address) (*api.PaychStatus, error) {
	return c.Internal.PaychStatus(ctx, pch)
}

func (c *WrapperV1) PaychVoucherCheckValid(ctx context.Context, addr address.Address, sv *paych.SignedVoucher) error {
	return c.Internal.PaychVoucherCheckValid(ctx, addr, sv)
}

func (c *WrapperV1) PaychVoucherCheckSpendable(ctx context.Context, addr address.Address, sv *paych.SignedVoucher, secret []byte, proof []byte) (bool, error) {
	return c.Internal.PaychVoucherCheckSpendable(ctx, addr, sv, secret, proof)
}

func (c *WrapperV1) PaychVoucherAdd(ctx context.Context, addr address.Address, sv *paych.SignedVoucher, proof []byte, minDelta types.BigInt) (types.BigInt, error) {
	return c.Internal.PaychVoucherAdd(ctx, addr, sv, proof, minDelta)
}

func (c *WrapperV1) PaychVoucherCreate(ctx context.Context, pch address.Address, amt types.BigInt, lane uint64) (*api.VoucherCreateResult, error) {
	return c.Internal.PaychVoucherCreate(ctx, pch, amt, lane)
}

func (c *WrapperV1) PaychVoucherList(ctx context.Context, pch address.Address) ([]*paych.SignedVoucher, error) {
	return c.Internal.PaychVoucherList(ctx, pch)
}

func (c *WrapperV1) PaychSettle(ctx context.Context, a address.Address) (cid.Cid, error) {
	return c.Internal.PaychSettle(ctx, a)
}

func (c *WrapperV1) PaychCollect(ctx context.Context, a address.Address) (cid.Cid, error) {
	return c.Internal.PaychCollect(ctx, a)
}

func (c *WrapperV1) PaychAllocateLane(ctx context.Context, ch address.Address) (uint64, error) {
	return c.Internal.PaychAllocateLane(ctx, ch)
}

func (c *WrapperV1) PaychNewPayment(ctx context.Context, from, to address.Address, vouchers []api.VoucherSpec) (*api.PaymentInfo, error) {
	return c.Internal.PaychNewPayment(ctx, from, to, vouchers)
}

func (c *WrapperV1) PaychVoucherSubmit(ctx context.Context, ch address.Address, sv *paych.SignedVoucher, secret []byte, proof []byte) (cid.Cid, error) {
	return c.Internal.PaychVoucherSubmit(ctx, ch, sv, secret, proof)
}

func (c *WrapperV1) CreateBackup(ctx context.Context, fpath string) error {
	return c.Internal.CreayyyateBackup(ctx, fpath)
}

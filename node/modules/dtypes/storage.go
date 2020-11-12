package dtypes

import (
	bserv "github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-graphsync"
	exchange "github.com/ipfs/go-ipfs-exchange-interface"
	format "github.com/ipfs/go-ipld-format"

	"github.com/filecoin-project/go-fil-markets/storagemarket/impl/requestvalidation"
	"github.com/filecoin-project/go-multistore"

	datatransfer "github.com/filecoin-project/go-data-transfer"
	"github.com/filecoin-project/go-fil-markets/piecestore"
	"github.com/filecoin-project/go-statestore"

	"github.com/filecoin-project/lotus/lib/blockstore"
	"github.com/filecoin-project/lotus/node/repo/importmgr"
	"github.com/filecoin-project/lotus/node/repo/retrievalstoremgr"
)

// MetadataDS stores metadata
// dy default it's namespaced under /metadata in main repo datastore
type MetadataDS datastore.Batching

// BareLocalChainBlockstore is a blockstore that is **guaranteed** to not be
// fronted by a cache. It should be used for processes like Bitswap, to avoid
// cache thrashing by untrusted parties.
type BareLocalChainBlockstore blockstore.Blockstore

// CachedLocalChainBlockstore is a blockstore that may be fronted by a cache,
// useful during local operations (such as chain validation) that benefit from
// temporal locality and recency caching.
type CachedLocalChainBlockstore blockstore.Blockstore

// ConsensusChainBlockstore is the effective chain blockstore to be used for
// chain consensus. It may be Bitswap backed if that setting is enabled.
type ConsensusChainBlockstore blockstore.Blockstore

// ExchangeChainBlockstore is the blockstore to be used when exchanging blobs
// over the network, such as when using Bitswap, Graphsync, and possibly
// ChainExchange.
//
// Usually, this store bypasses caches to avoid cache thrashing by untrusted
// parties.
type ExchangeChainBlockstore blockstore.Blockstore

type ChainBitswap exchange.Interface
type ChainBlockService bserv.BlockService

type ClientMultiDstore *multistore.MultiStore
type ClientImportMgr *importmgr.Mgr
type ClientBlockstore blockstore.Blockstore
type ClientDealStore *statestore.StateStore
type ClientRequestValidator *requestvalidation.UnifiedRequestValidator
type ClientDatastore datastore.Batching
type ClientRetrievalStoreManager retrievalstoremgr.RetrievalStoreManager

type Graphsync graphsync.GraphExchange

// ClientDataTransfer is a data transfer manager for the client
type ClientDataTransfer datatransfer.Manager

type ProviderDealStore *statestore.StateStore
type ProviderPieceStore piecestore.PieceStore
type ProviderRequestValidator *requestvalidation.UnifiedRequestValidator

// ProviderDataTransfer is a data transfer manager for the provider
type ProviderDataTransfer datatransfer.Manager

type StagingDAG format.DAGService
type StagingBlockstore blockstore.Blockstore
type StagingGraphsync graphsync.GraphExchange
type StagingMultiDstore *multistore.MultiStore

package types

import "github.com/filecoin-project/go-address"

type MinerInfo struct {
	MinerAddr address.Address
	Power     BigInt
}

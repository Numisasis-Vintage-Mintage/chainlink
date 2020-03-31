package client

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// NewRoundEvent represents a NewRound event log that has
// been unmarshaled
type NewRoundEvent struct {
	RoundID   *big.Int
	StartedBy common.Address
}

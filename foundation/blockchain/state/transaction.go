package state

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

// UpsertWalletTransaction accepts a transaction from a wallet for inclusion.
func (s *State) UpsertWalletTransaction(signedTx database.SignedTx) error { 
	return nil
}

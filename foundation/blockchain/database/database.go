// Package database handles all the lower level support for maintaining the
// blockchain in storage and maintaining an in-memory databse of account information.
package database

import (
	"sync"

	"github.com/warlck/palladium/foundation/blockchain/genesis"
)

// Database manages data related to accounts who have transacted on the blockchain.
type Database struct {
	mu       sync.RWMutex
	genesis  genesis.Genesis
	accounts map[AccountID]Account
}

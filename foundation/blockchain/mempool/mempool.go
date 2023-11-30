package mempool

import (
	"sync"

	"github.com/warlck/palladium/foundation/blockchain/database"
)

// Mempool represents a cache of transactions organized by account:nonce.
type Mempool struct {
	mu   sync.RWMutex
	pool map[string]database.BlockTx
}

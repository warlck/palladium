package mempool

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/warlck/palladium/foundation/blockchain/database"
)

// Mempool represents a cache of transactions organized by account:nonce.
type Mempool struct {
	mu   sync.RWMutex
	pool map[string]database.BlockTx
}

// New constructs a new mempool using the default sort strategy.
func New() (*Mempool, error) {
	return &Mempool{
		pool: make(map[string]database.BlockTx),
	}, nil
}

// Upsert adds or replaces a transaction from the mempool.
func (mp *Mempool) Upsert(tx database.BlockTx) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// CORE NOTE: Different blockchains have different algorithms to limit the
	// size of the mempool. Some limit based on the amount of memory being
	// consumed and some may limit based on the number of transaction. If a limit
	// is met, then either the transaction that has the least return on investment
	// or the oldest will be dropped from the pool to make room for new the transaction.

	// For now, the Palladium blockchain in not imposing any limits.
	key, err := mapKey(tx)
	if err != nil {
		return err
	}

	// Ethereum requires a 10% bump in the tip to replace an existing
	// transaction in the mempool and so do we. We want to limit users
	// from this sort of behavior.
	if etx, exists := mp.pool[key]; exists {
		if tx.Tip < uint64(math.Round(float64(etx.Tip)*1.10)) {
			return errors.New("replacing a transaction requires a 10% bump in the tip")
		}
	}

	mp.pool[key] = tx

	return nil
}

// Delete removed a transaction from the mempool.
func (mp *Mempool) Delete(tx database.BlockTx) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	key, err := mapKey(tx)
	if err != nil {
		return err
	}

	delete(mp.pool, key)

	return nil
}

// PickBest uses the configured sort strategy to return a set of transactions.
// If 0 is passed, all transactions in the mempool will be returned.
func (mp *Mempool) PickBest(howMany ...uint16) []database.BlockTx {

	// CORE NOTE: Most blockchains do set a max block size limit and this size
	// will determined which transactions are selected. When picking the best
	// transactions for the next block, the Palladium blockchain is currently not
	// focused on block size but a max number of transactions.
	//
	// When the selection algorithm does need to consider sizing, picking the
	// right transactions that maximize profit gets really hard. On top of this,
	// today a miner gets a mining reward for each mined block. In the future
	// this could go away leaving just fees for the transactions that are
	// selected as the only form of revenue. This will change how transactions
	// need to be selected.

	// Copy all the transactions for each account into separate slices.

	mp.mu.RLock()
	defer mp.mu.RUnlock()

	var txs []database.BlockTx

	for _, tx := range mp.pool {
		txs = append(txs, tx)
	}

	return txs
}

// =============================================================================

// mapKey is used to generate the map key.
func mapKey(tx database.BlockTx) (string, error) {
	return fmt.Sprintf("%s:%d", tx.FromID, tx.Nonce), nil
}

// accountFromMapKey extracts the account information from the mapkey.
func accountFromMapKey(key string) database.AccountID {
	return database.AccountID(strings.Split(key, ":")[0])
}

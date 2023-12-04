// Package state is the core API for the blockchain and implements all the
// business rules and processing.
package state

import (
	"sync"

	"github.com/warlck/palladium/foundation/blockchain/database"
	"github.com/warlck/palladium/foundation/blockchain/genesis"
	"github.com/warlck/palladium/foundation/blockchain/mempool"
)

// =============================================================================

// EventHandler defines a function that is called when events
// occur in the processing of persisting blocks.
type EventHandler func(v string, args ...any)

// =============================================================================

// Config represents the configuration required to start
// the blockchain node.
type Config struct {
	BeneficiaryID  database.AccountID
	Host           string
	Storage        database.Storage
	Genesis        genesis.Genesis
	SelectStrategy string
	EvHandler      EventHandler
	Consensus      string
}

// State manages the blockchain database.
type State struct {
	mu          sync.RWMutex
	resyncWG    sync.WaitGroup
	allowMining bool

	beneficiaryID database.AccountID
	host          string
	evHandler     EventHandler

	genesis genesis.Genesis
	mempool *mempool.Mempool
	db      *database.Database
}

// New constructs a new blockchain for data management.
func New(cfg Config) (*State, error) {

	// Build a safe event handler function for use.
	ev := func(v string, args ...any) {
		if cfg.EvHandler != nil {
			cfg.EvHandler(v, args...)
		}
	}

	// Load the genesis file to get starting balances
	// for founders of the blockchain

	genesis, err := genesis.Load()
	if err != nil {
		return nil, err
	}

	// Access the storage for the blockchain.
	db, err := database.New(genesis, cfg.Storage, ev)
	if err != nil {
		return nil, err
	}

	// Construct a mempool with the specified sort strategy.
	mempool, err := mempool.New()
	if err != nil {
		return nil, err
	}

	// Create the State to provide support for managing the blockchain.
	state := State{
		beneficiaryID: cfg.BeneficiaryID,
		host:          cfg.Host,
		evHandler:     ev,
		allowMining:   true,

		genesis: cfg.Genesis,
		mempool: mempool,
		db:      db,
	}

	// The Worker is not set here. The call to worker.Run will assign itself
	// and start everything up and running for the node.

	return &state, nil
}

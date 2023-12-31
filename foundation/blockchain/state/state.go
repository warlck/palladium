// Package state is the core API for the blockchain and implements all the
// business rules and processing.
package state

import (
	"sync"

	"github.com/warlck/palladium/foundation/blockchain/database"
	"github.com/warlck/palladium/foundation/blockchain/genesis"
	"github.com/warlck/palladium/foundation/blockchain/mempool"
	"github.com/warlck/palladium/foundation/blockchain/peer"
)

// =============================================================================

// The set of different consensus protocols that can be used.
const (
	ConsensusPOW = "POW"
	ConsensusPOA = "POA"
)

// =============================================================================

// EventHandler defines a function that is called when events
// occur in the processing of persisting blocks.
type EventHandler func(v string, args ...any)

// Worker interface represents the behavior required to be implemented by any
// package providing support for mining, peer updates, and transaction sharing.
type Worker interface {
	Shutdown()
	SignalStartMining()
	SignalCancelMining()
}

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
	KnownPeers     *peer.PeerSet
}

// State manages the blockchain database.
type State struct {
	mu          sync.RWMutex
	resyncWG    sync.WaitGroup
	allowMining bool

	beneficiaryID database.AccountID
	host          string
	evHandler     EventHandler

	genesis    genesis.Genesis
	mempool    *mempool.Mempool
	db         *database.Database
	knownPeers *peer.PeerSet

	Worker Worker
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

		knownPeers: cfg.KnownPeers,
		genesis:    cfg.Genesis,
		mempool:    mempool,
		db:         db,
	}

	// The Worker is not set here. The call to worker.Run will assign itself
	// and start everything up and running for the node.

	return &state, nil
}

// Shutdown cleanly brings the node down.
func (s *State) Shutdown() error {
	s.evHandler("state: shutdown: started")
	defer s.evHandler("state: shutdown: completed")

	// Make sure the database file is properly closed.
	defer func() {
		s.db.Close()
	}()

	// Stop all blockchain writing activity.
	s.Worker.Shutdown()

	return nil
}

// =============================================================================

// Mempool returns a copy of the mempool.
func (s *State) Mempool() []database.BlockTx {
	return s.mempool.PickBest()
}

// MempoolLength returns the current length of the mempool.
func (s *State) MempoolLength() int {
	return s.mempool.Count()
}

// Accounts returns a copy of the database accounts.
func (s *State) Accounts() map[database.AccountID]database.Account {
	return s.db.Copy()
}

// LatestBlock returns a copy the current latest block.
func (s *State) LatestBlock() database.Block {
	return s.db.LatestBlock()
}

// KnownExternalPeers retrieves a copy of the known peer list without
// including this node.
func (s *State) KnownExternalPeers() []peer.Peer {
	return s.knownPeers.Copy(s.host)
}

// RemoveKnownPeer provides the ability to remove a peer from
// the known peer list.
func (s *State) RemoveKnownPeer(peer peer.Peer) {
	s.knownPeers.Remove(peer)
}

// AddKnownPeer provides the ability to add a new peer to
// the known peer list.
func (s *State) AddKnownPeer(peer peer.Peer) bool {
	return s.knownPeers.Add(peer)
}

// Host returns a copy of host information.
func (s *State) Host() string {
	return s.host
}

// UpsertMempool adds a new transaction to the mempool.
func (s *State) UpsertMempool(tx database.BlockTx) error {
	return s.mempool.Upsert(tx)
}

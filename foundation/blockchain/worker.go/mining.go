package worker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/warlck/palladium/foundation/blockchain/state"
)

// CORE NOTE: The POW mining operation is managed by this function which runs on
// it's own goroutine. When a startMining signal is received (mainly because a
// wallet transaction was received) a block is created and then the POW operation
// starts. This operation can be cancelled if a proposed block is received and
// is validated.

// miningOperations handles mining.
func (w *Worker) miningOperations() {
	w.evHandler("worker: powOperations: G started")
	defer w.evHandler("worker: powOperations: G completed")

	for {
		select {
		case <-w.startMining:
			if !w.isShutdown() {
				w.runMiningOperation()
			}
		case <-w.shut:
			w.evHandler("worker: miningOperations: received shut signal")
			return
		}
	}

}

// runMioningOperation takes all the transactions from the mempool and writes a
// new block to the database.
func (w *Worker) runMiningOperation() {
	w.evHandler("worker: runMiningOperation: MINING: started")
	defer w.evHandler("worker: runMiningOperation: MINING: completed")

	// Make sure there are transactions in the mempool.
	length := w.state.MempoolLength()
	if length == 0 {
		w.evHandler("worker: runMiningOperation: MINING: no transactions to mine: Txs[%d]", length)
		return
	}

	// After running a mining operation, check if a new operation should
	// be signaled again.
	defer func() {
		length := w.state.MempoolLength()
		if length > 0 {
			w.evHandler("worker: runMiningOperation: MINING: signal new mining operation: Txs[%d]", length)
			w.SignalStartMining()
		}
	}()

	select {
	case <-w.cancelMining:
		w.evHandler("worker: runMiningOperation: MINING: CANCEL: requested")
	default:
	}

	// Create a context so mining can be cancelled.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Can't return from this function until these G's are complete.
	var wg sync.WaitGroup
	wg.Add(2)

	// This G exists to cancel the mining operation.
	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()

		select {
		case <-w.cancelMining:
			w.evHandler("worker: runMiningOperation: MINING: CANCEL: requested")
		case <-ctx.Done():
		}
	}()

	// This G is performing the mining.
	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()

		t := time.Now()
		_, err := w.state.MineNewBlock(ctx)
		duration := time.Since(t)

		w.evHandler("worker: runMiningOperation: MINING: mining duration[%v]", duration)

		if err != nil {
			switch {
			case errors.Is(err, state.ErrNoTransactions):
				w.evHandler("worker: runMiningOperation: MINING: WARNING: no transactions in mempool")
			case ctx.Err() != nil:
				w.evHandler("worker: runMiningOperation: MINING: CANCEL: complete")
			default:
				w.evHandler("worker: runMiningOperation: MINING: ERROR: %s", err)
			}
			return
		}

		// WOW, we mined a block. Propose the new block to the network.
		// Log the error, but that's it.

	}()

	// Wait for both G's to terminate.
	wg.Wait()

}

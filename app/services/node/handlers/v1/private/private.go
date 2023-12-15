package private

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	v1 "github.com/warlck/palladium/business/web/v1"
	"github.com/warlck/palladium/foundation/blockchain/database"
	"github.com/warlck/palladium/foundation/blockchain/peer"
	"github.com/warlck/palladium/foundation/blockchain/state"
	"github.com/warlck/palladium/foundation/nameservice"
	"github.com/warlck/palladium/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log   *zap.SugaredLogger
	State *state.State
	NS    *nameservice.NameService
}

// SubmitPeer is called by a node so they can be added to the known peer list.
func (h Handlers) SubmitPeer(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var peer peer.Peer
	if err := web.Decode(r, &peer); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	if !h.State.AddKnownPeer(peer) {
		h.Log.Infow("adding peer", "traceid", v.TraceID, "host", peer.Host)
	}

	return web.Respond(ctx, w, nil, http.StatusOK)
}

// Status returns the current status of the node.
func (h Handlers) Status(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	h.Log.Infow("node status", "traceid", v.TraceID)
	latestBlock := h.State.LatestBlock()

	status := peer.PeerStatus{
		LatestBlockHash:   latestBlock.Hash(),
		LatestBlockNumber: latestBlock.Header.Number,
		KnownPeers:        h.State.KnownExternalPeers(),
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}

// Mempool returns the set of uncommitted transactions.
func (h Handlers) Mempool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	txs := h.State.Mempool()
	return web.Respond(ctx, w, txs, http.StatusOK)
}

// BlocksByNumber returns all the blocks based on the specified to/from values.
func (h Handlers) BlocksByNumber(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	fromStr := web.Param(r, "from")
	if fromStr == "latest" || fromStr == "" {
		fromStr = fmt.Sprintf("%d", state.QueryLastest)
	}

	toStr := web.Param(r, "to")
	if toStr == "latest" || toStr == "" {
		toStr = fmt.Sprintf("%d", state.QueryLastest)
	}

	from, err := strconv.ParseUint(fromStr, 10, 64)
	if err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}
	to, err := strconv.ParseUint(toStr, 10, 64)
	if err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	if from > to {
		return v1.NewRequestError(errors.New("from is greater than to"), http.StatusBadRequest)
	}

	blocks := h.State.QueryBlocksByNumber(from, to)
	if len(blocks) == 0 {
		return web.Respond(ctx, w, nil, http.StatusNoContent)
	}

	blockData := make([]database.BlockData, len(blocks))
	for i, block := range blocks {
		blockData[i] = database.NewBlockData(block)
	}

	return web.Respond(ctx, w, blockData, http.StatusOK)
}

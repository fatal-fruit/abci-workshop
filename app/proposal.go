package app

import (
	"cosmossdk.io/log"
	"fmt"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
)

type ProposalHandler struct {
	app     App
	logger  log.Logger
	mempool sdkmempool.Mempool
}

func (h *ProposalHandler) NewPrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		// Mempool isn't required to be set. If mempool is nil, transactions will be returned in FIFO order.
		var proposalTxs [][]byte

		numtxs := h.mempool.CountTx()
		h.logger.Info(fmt.Sprintf("This is the number of app mempool transactions : %v ", numtxs))

		counter := 0
		for _, tx := range req.Txs {
			counter++
			proposalTxs = append(proposalTxs, tx)
		}
		h.logger.Info(fmt.Sprintf("This is the number of transactions from request : %v ", counter))

		return &abci.ResponsePrepareProposal{
			Txs: proposalTxs,
		}, nil
	}
}

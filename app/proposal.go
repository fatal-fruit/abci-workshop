package app

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
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

		req.GetLocalLastCommit()

		counter := 0
		h.logger.Info(fmt.Sprintf("This is the tendermint tx length: %v ", len(req.Txs)))
		for _, tx := range req.Txs {
			counter++
			h.logger.Info(fmt.Sprintf("This is the tendermint tx: %v ", tx))
			proposalTxs = append(proposalTxs, tx)
		}

		var orderedTxs []sdk.Tx
		itr := h.mempool.Select(context.Background(), nil)
		for itr != nil {
			orderedTxs = append(orderedTxs, itr.Tx())
			itr = itr.Next()
		}

		for _, t := range orderedTxs {
			h.logger.Info(fmt.Sprintf("This is the mempool tx: %v ", t))
		}
		h.logger.Info(fmt.Sprintf("This is the mempool tx length: %v ", len(orderedTxs)))

		h.logger.Info(fmt.Sprintf("This is the number of transactions from request : %v ", counter))

		return &abci.ResponsePrepareProposal{
			Txs: proposalTxs,
		}, nil
	}
}
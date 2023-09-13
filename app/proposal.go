package app

import (
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

		for _, tx := range req.Txs {

			proposalTxs = append(proposalTxs, tx)
		}

		return &abci.ResponsePrepareProposal{
			Txs: proposalTxs,
		}, nil
	}
}

package abci

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fatal-fruit/cosmapp/mempool"
	nstypes "github.com/fatal-fruit/ns/types"
)

func NewVoteExtensionHandler(lg log.Logger, mp *mempool.ThresholdMempool, cdc codec.Codec) *VoteExtHandler {
	return &VoteExtHandler{
		logger:      lg,
		mempool:     mp,
		cdc:         cdc,
		subscribers: make(map[string]func(sdk.Context, *abci.RequestExtendVote) ([]byte, error)),
	}
}

func (h *VoteExtHandler) SetSubscriber(key string, subHandler func(sdk.Context, *abci.RequestExtendVote) ([]byte, error)) {
	h.subscribers[key] = subHandler
}

func (h *VoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		h.logger.Info(fmt.Sprintf("Extending votes at block height : %v", req.Height))

		voteExtBids := [][]byte{}
		voteExtExtra := [][]byte{}

		// Get mempool txs
		itr := h.mempool.SelectPending(context.Background(), nil)

		var txs []sdk.Tx
		for itr != nil {
			tmptx := itr.Tx()
			sdkMsgs := tmptx.GetMsgs()

			txs = append(txs, tmptx)

			// Iterate through msgs, check for any bids
			for _, msg := range sdkMsgs {
				switch msg := msg.(type) {
				case *nstypes.MsgBid:
					// Marshal sdk bids to []byte
					bz, err := h.cdc.Marshal(msg)
					if err != nil {
						h.logger.Error(fmt.Sprintf("Error marshalling VE Bid : %v", err))
						break
					}
					voteExtBids = append(voteExtBids, bz)
				default:
					for _, v := range h.subscribers {
						val, error := v(ctx, req)

						//Route to other functions.
						if error == nil {
							voteExtExtra = append(voteExtExtra, val)
						}
					}
				}
			}

			// Move tx to ready pool
			err := h.mempool.Update(context.Background(), tmptx)

			//// Remove tx from app side mempool
			//err = h.mempool.Remove(tmptx)
			if err != nil {
				h.logger.Info(fmt.Sprintf("Unable to update mempool tx: %v", err))
			}

			itr = itr.Next()
		}

		// Create vote extension
		voteExt := AppVoteExtension{
			Height:    req.Height,
			Bids:      voteExtBids,
			ExtraInfo: voteExtExtra,
		}

		// Marshal Vote Extension
		bz, err := json.Marshal(voteExt)
		if err != nil {
			return nil, fmt.Errorf("Error marshalling VE: %w", err)
		}

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

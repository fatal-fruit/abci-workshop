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
		logger:            lg,
		mempool:           mp,
		cdc:               cdc,
		extendSubscribers: make(map[string]func(sdk.Context, *abci.RequestExtendVote) ([]byte, error)),
		verifySubscribers: make(map[string]func(sdk.Context, []byte) (abci.ResponseVerifyVoteExtension_VerifyStatus, error)),
	}
}

func (h *VoteExtHandler) SetSubscriber(
	key string,
	extendSubHandler func(sdk.Context, *abci.RequestExtendVote) ([]byte, error),
	verifySubHandler func(sdk.Context, []byte) (abci.ResponseVerifyVoteExtension_VerifyStatus, error)) {
	h.extendSubscribers[key] = extendSubHandler
	h.verifySubscribers[key] = verifySubHandler
}

func (h *VoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		h.logger.Info(fmt.Sprintf("Extending votes at block height : %v", req.Height))

		voteExtBids := [][]byte{}
		voteExtExtra := map[string][]byte{}

		// Get mempool txs
		itr := h.mempool.SelectPending(context.Background(), nil)

		for itr != nil {
			tmptx := itr.Tx()
			sdkMsgs := tmptx.GetMsgs()

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

		for k, v := range h.extendSubscribers {
			//Route to other handlers
			val, error := v(ctx, req)
			if error == nil {
				voteExtExtra[k] = val
			}
		}

		// Create vote extension
		mVoteExtra, err := json.Marshal(voteExtExtra)
		if err != nil {
			h.logger.Error("Error marshalling voteExtra")
		}

		voteExt := AppVoteExtension{
			Height:    req.Height,
			Bids:      voteExtBids,
			ExtraInfo: mVoteExtra,
		}

		// Encode Vote Extension
		bz, err := json.Marshal(voteExt)
		if err != nil {
			return nil, fmt.Errorf("Error marshalling VE: %w", err)
		}

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

func (h *VoteExtHandler) VerifyVoteExtHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		h.logger.Info(fmt.Sprintf("Verifying votes at block height : %v", req.Height))
		voteExt := AppVoteExtension{}

		err := json.Unmarshal(req.VoteExtension, &voteExt)
		if err != nil {
			h.logger.Error(fmt.Sprintf("Could not unmarshal vote extension : %v", req.Height))
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		mVoteExtra := make(map[string][]byte)
		err = json.Unmarshal(voteExt.ExtraInfo, &mVoteExtra)
		if err != nil {
			h.logger.Error(fmt.Sprintf("Could not unmarshal extra info from VE : %v", req.Height))
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		for k, e := range mVoteExtra {
			//Route to other handlers
			v, ok := h.verifySubscribers[k]
			if !ok {
				h.logger.Error(fmt.Sprintf("Received extra info for module not registered: %v", k))
				return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
			}

			val, error := v(ctx, e)

			if error != nil {
				h.logger.Info(fmt.Sprintf("Verifying vote extensions at block height : %v", req.Height))
				return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_UNKNOWN}, error
			}
			if val != abci.ResponseVerifyVoteExtension_ACCEPT {
				h.logger.Info(fmt.Sprintf("Verifying vote extensions at block height : %v", req.Height))
				return &abci.ResponseVerifyVoteExtension{Status: val}, nil
			}
		}

		h.logger.Info(fmt.Sprintf("Verifying vote extensions at block height : %v", req.Height))
		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

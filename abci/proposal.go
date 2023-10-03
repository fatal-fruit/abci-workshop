package abci

import (
	"context"
	"cosmossdk.io/log"
	"encoding/base64"
	"encoding/json"
	"fmt"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fatal-fruit/cosmapp/mempool"
	"github.com/fatal-fruit/cosmapp/provider"
	nstypes "github.com/fatal-fruit/ns/types"
)

func NewPrepareProposalHandler(
	lg log.Logger,
	txCg client.TxConfig,
	cdc codec.Codec,
	mp *mempool.ThresholdMempool,
	pv provider.TxProvider,
	runProv bool,
) *PrepareProposalHandler {
	return &PrepareProposalHandler{
		logger:      lg,
		txConfig:    txCg,
		cdc:         cdc,
		mempool:     mp,
		txProvider:  pv,
		runProvider: runProv,
	}
}

func (h *PrepareProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		h.logger.Info(fmt.Sprintf("🛠️ :: Prepare Proposal"))
		var proposalTxs [][]byte

		// Get Vote Extensions
		if req.Height > 2 {

			// Get Special Transaction

			// Marshal Special Transaction

			// Append Special Transaction to proposal

		}

		var txs []sdk.Tx
		itr := h.mempool.Select(context.Background(), nil)
		for itr != nil {
			tmptx := itr.Tx()

			txs = append(txs, tmptx)
			itr = itr.Next()
		}
		h.logger.Info(fmt.Sprintf("🛠️ :: Number of Transactions available from mempool: %v", len(txs)))

		if h.runProvider {
			tmpMsgs, err := h.txProvider.BuildProposal(ctx, txs)
			if err != nil {
				h.logger.Error(fmt.Sprintf("❌️ :: Error Building Custom Proposal: %v", err))
			}
			txs = tmpMsgs
		}

		for _, sdkTxs := range txs {
			txBytes, err := h.txConfig.TxEncoder()(sdkTxs)
			if err != nil {
				h.logger.Info(fmt.Sprintf("❌~Error encoding transaction: %v", err.Error()))
			}
			proposalTxs = append(proposalTxs, txBytes)
		}

		h.logger.Info(fmt.Sprintf("🛠️ :: Number of Transactions in proposal: %v", len(proposalTxs)))

		return &abci.ResponsePrepareProposal{Txs: proposalTxs}, nil
	}
}

func (h *ProcessProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (resp *abci.ResponseProcessProposal, err error) {
		// The first transaction will always be the Special Transaction

		// But we want to first check if the proposal has any transactions

		// Double check if the first transaction is the special transaction and if it is we always want to Unmarshal

		// Check if there are any bids in the Special Transaction

		// Unmarshal the bids

		// Validate these bids

		return nil, nil
	}
}

func processVoteExtensions(req *abci.RequestPrepareProposal, log log.Logger) (SpecialTransaction, error) {
	log.Info(fmt.Sprintf("🛠️ :: Process Vote Extensions"))

	// Create empty response

	// Get Vote Ext for H-1 from Req

	// Iterate through votes

	// Unmarshal to AppExt

	// If Bids in VE, append to Special Transaction

	return SpecialTransaction{}, nil
}

func ValidateBids(txConfig client.TxConfig, veBids []nstypes.MsgBid, proposalTxs [][]byte, logger log.Logger) (bool, error) {
	var proposalBids []*nstypes.MsgBid
	for _, txBytes := range proposalTxs {
		txDecoder := txConfig.TxDecoder()
		messages, err := txDecoder(txBytes)
		if err != nil {
			logger.Error(fmt.Sprintf("❌️:: Unable to decode proposal transactions :: %v", err))

			return false, err
		}
		sdkMsgs := messages.GetMsgs()
		for _, m := range sdkMsgs {
			switch m := m.(type) {
			case *nstypes.MsgBid:
				proposalBids = append(proposalBids, m)
			}
		}
	}

	bidFreq := make(map[string]int)
	totalVotes := len(veBids)
	for _, b := range veBids {
		h, err := Hash(&b)
		if err != nil {
			logger.Error(fmt.Sprintf("❌️:: Unable to produce bid frequency map :: %v", err))

			return false, err
		}
		bidFreq[h]++
	}

	thresholdCount := int(float64(totalVotes) * 0.5)
	logger.Info(fmt.Sprintf("🛠️ :: VE Threshold: %v", thresholdCount))
	ok := true
	logger.Info(fmt.Sprintf("🛠️ :: Number of Proposal Bids: %v", len(proposalBids)))

	for _, p := range proposalBids {

		key, err := Hash(p)
		if err != nil {
			logger.Error(fmt.Sprintf("❌️:: Unable to hash proposal bid :: %v", err))

			return false, err
		}
		freq := bidFreq[key]
		logger.Info(fmt.Sprintf("🛠️ :: Frequency for Proposal Bid: %v", freq))
		if freq < thresholdCount {
			logger.Error(fmt.Sprintf("❌️:: Detected invalid proposal bid :: %v", p))

			ok = false
		}
	}
	return ok, nil
}

func Hash(m *nstypes.MsgBid) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

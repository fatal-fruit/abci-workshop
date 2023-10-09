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

		// If the block height is more than 2, get Vote Extensions from Special Transaction
		if req.Height > 2 {

			// Get Special Transaction
			ve, err := processVoteExtensions(req, h.logger)
			if err != nil {
				h.logger.Error(fmt.Sprintf("❌️ :: Unable to process Vote Extensions: %v", err))
			}

			// Marshal Special Transaction
			bz, err := json.Marshal(ve)
			if err != nil {
				h.logger.Error(fmt.Sprintf("❌️ :: Unable to marshal Vote Extensions: %v", err))
			}

			// Append Special Transaction to proposal
			proposalTxs = append(proposalTxs, bz)
		}

		// Get all the transactions from the mempool
		var txs []sdk.Tx
		itr := h.mempool.Select(context.Background(), nil)
		for itr != nil {
			tmptx := itr.Tx()

			txs = append(txs, tmptx)
			itr = itr.Next()
		}
		h.logger.Info(fmt.Sprintf("🛠️ :: Number of Transactions available from mempool: %v", len(txs)))

		// If the runProvider is set, then build a custom proposal
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
		h.Logger.Info(fmt.Sprintf("⚙️ :: Process Proposal"))

		// The first transaction will always be the Special Transaction
		numTxs := len(req.Txs)
		if numTxs == 1 {
			h.Logger.Info(fmt.Sprintf("⚙️:: Number of transactions :: %v", numTxs))
		}

		// If there are at least 2 transactions, then the first transaction is the special transaction
		if numTxs >= 1 {
			h.Logger.Info(fmt.Sprintf("⚙️:: Number of transactions :: %v", numTxs))
			var st SpecialTransaction
			err = json.Unmarshal(req.Txs[0], &st)
			if err != nil {
				h.Logger.Error(fmt.Sprintf("❌️:: Error unmarshalling special Tx in Process Proposal :: %v", err))
			}
			// If there are any bids in the special transaction, then validate them
			if len(st.Bids) > 0 {
				h.Logger.Info(fmt.Sprintf("⚙️:: There are bids in the Special Transaction"))
				var bids []nstypes.MsgBid
				for i, b := range st.Bids {
					var bid nstypes.MsgBid
					h.Codec.Unmarshal(b, &bid)
					h.Logger.Info(fmt.Sprintf("⚙️:: Special Transaction Bid No %v :: %v", i, bid))
					bids = append(bids, bid)
				}
				// Validate Bids in Tx
				txs := req.Txs[1:]
				ok, err := ValidateBids(h.TxConfig, bids, txs, h.Logger)
				if err != nil {
					h.Logger.Error(fmt.Sprintf("❌️:: Error validating bids in Process Proposal :: %v", err))
					return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
				}
				// If the bids in the proposal are not valid, then reject the Tx
				if !ok {
					h.Logger.Error(fmt.Sprintf("❌️:: Unable to validate bids in Process Proposal :: %v", err))
					return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
				}
				h.Logger.Info("⚙️:: Successfully validated bids in Process Proposal")
			}
		}

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

func processVoteExtensions(req *abci.RequestPrepareProposal, log log.Logger) (SpecialTransaction, error) {
	log.Info(fmt.Sprintf("🛠️ :: Process Vote Extensions"))

	// Create empty response
	st := SpecialTransaction{
		0,
		[][]byte{},
	}

	// Get Vote Ext for H-1 from Req
	voteExt := req.GetLocalLastCommit()
	votes := voteExt.Votes

	// Iterate through votes
	var ve AppVoteExtension
	for _, vote := range votes {
		// Unmarshal to AppExt
		err := json.Unmarshal(vote.VoteExtension, &ve)
		if err != nil {
			log.Error(fmt.Sprintf("❌ :: Error unmarshalling Vote Extension"))
		}

		st.Height = int(ve.Height)

		// If Bids in VE, append to Special Transaction
		if len(ve.Bids) > 0 {
			log.Info("🛠️ :: Bids in VE")
			for _, b := range ve.Bids {
				st.Bids = append(st.Bids, b)
			}
		}
	}

	return st, nil
}

func ValidateBids(txConfig client.TxConfig, veBids []nstypes.MsgBid, proposalTxs [][]byte, logger log.Logger) (bool, error) {
	// Create a list of bids to store in the proposal
	var proposalBids []*nstypes.MsgBid
	for _, txBytes := range proposalTxs {
		txDecoder := txConfig.TxDecoder()
		messages, err := txDecoder(txBytes)
		if err != nil {
			logger.Error(fmt.Sprintf("❌️:: Unable to decode proposal transactions :: %v", err))

			return false, err
		}
		// Get SDK messages from transactions
		sdkMsgs := messages.GetMsgs()

		// Iterate through the SDK messages and get the bids
		for _, m := range sdkMsgs {
			switch m := m.(type) {
			case *nstypes.MsgBid:
				proposalBids = append(proposalBids, m)
			}
		}
	}
	// Create map to store each bid frequency in the vote extension
	bidFreq := make(map[string]int)

	// Get the total number of votes in the vote extension
	totalVotes := len(veBids)

	// Iterate through the bids in the vote extension and calculate the frequency of each bid
	for _, b := range veBids {
		h, err := Hash(&b)
		if err != nil {
			logger.Error(fmt.Sprintf("❌️:: Unable to produce bid frequency map :: %v", err))

			return false, err
		}
		bidFreq[h]++
	}

	// Calculate the threshold count
	thresholdCount := int(float64(totalVotes) * 0.5)
	logger.Info(fmt.Sprintf("🛠️ :: VE Threshold: %v", thresholdCount))
	ok := true
	logger.Info(fmt.Sprintf("🛠️ :: Number of Proposal Bids: %v", len(proposalBids)))

	// Iterate over proposal bids and check if each bid appaears in the VE at least the threshold count number of times
	for _, p := range proposalBids {

		key, err := Hash(p)
		if err != nil {
			logger.Error(fmt.Sprintf("❌️:: Unable to hash proposal bid :: %v", err))

			return false, err
		}
		//Get the frequency of the proposal bid
		freq := bidFreq[key]
		logger.Info(fmt.Sprintf("🛠️ :: Frequency for Proposal Bid: %v", freq))
		// If the frequency of the proposal bid is less than the threshold count, then the bid is invalid
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

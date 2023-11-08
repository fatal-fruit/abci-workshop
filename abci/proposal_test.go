package abci

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fatal-fruit/cosmapp/testutils"
	nstypes "github.com/fatal-fruit/ns/types"
	"github.com/stretchr/testify/require"
)

func TestBidHash(t *testing.T) {
	bids := []*nstypes.MsgBid{
		{
			Name:           "bob.cosmos",
			ResolveAddress: "cosmos1c3f2e2d4wwhaud70h3c7rah8aede8kplevxe3j",
			Owner:          "cosmos1c3f2e2d4wwhaud70h3c7rah8aede8kplevxe3j",
			Amount:         sdk.Coins{sdk.NewCoin("uatom", math.NewInt(5))},
		},
	}

	for _, b := range bids {
		_, err := Hash(b)
		require.NoError(t, err)
	}
}

// TODO: Fix me
func TestValidateProposal(t *testing.T) {
	testEncConfig := testutils.MakeTestEncodingConfig()
	testTxConfig := testEncConfig.TxConfig
	logger := log.NewTestLogger(t)
	var proposalTxs [][]byte

	voteExtBids := []nstypes.MsgBid{
		{
			Name:           "bob.cosmos",
			ResolveAddress: "cosmos1c3f2e2d4wwhaud70h3c7rah8aede8kplevxe3j",
			Owner:          "cosmos1c3f2e2d4wwhaud70h3c7rah8aede8kplevxe3j",
			Amount:         sdk.Coins{sdk.NewCoin("uatom", math.NewInt(5))},
		},
		{
			Name:           "bob.cosmos",
			ResolveAddress: "cosmos1c3f2e2d4wwhaud70h3c7rah8aede8kplevxe3j",
			Owner:          "cosmos1c3f2e2d4wwhaud70h3c7rah8aede8kplevxe3j",
			Amount:         sdk.Coins{sdk.NewCoin("uatom", math.NewInt(5))},
		},
		{
			Name:           "bob.cosmos",
			ResolveAddress: "cosmos1c3f2e2d4wwhaud70h3c7rah8aede8kplevxe3j",
			Owner:          "cosmos1c3f2e2d4wwhaud70h3c7rah8aede8kplevxe3j",
			Amount:         sdk.Coins{sdk.NewCoin("uatom", math.NewInt(5))},
		},
	}

	builder := testTxConfig.NewTxBuilder()
	builder.SetMsgs(&voteExtBids[0])
	tx := builder.GetTx()
	bz, err := testTxConfig.TxEncoder()(tx)
	require.NoError(t, err)

	proposalTxs = append(proposalTxs, bz)

	ok, err := ValidateBids(testTxConfig, voteExtBids, proposalTxs, logger)
	require.NoError(t, err)
	require.True(t, ok)
}

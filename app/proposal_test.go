package app

import (
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	nstypes "github.com/fatal-fruit/ns/types"
	"github.com/stretchr/testify/require"

	"testing"
)

func NewPrepareProposal(ctx sdk.Context, a *abci.RequestPrepareProposal) (interface{}, interface{}) {
	return &abci.ResponsePrepareProposal{Txs: a.Txs}, nil
}

type fixture struct {
	ctx           sdk.Context
	app           App
	mockStoreKey  storetypes.StoreKey
	mockMsgURL    string
	mockclientCtx client.Context
	txBuilder     client.TxBuilder
}

func initFixture(t *testing.T) *fixture {
	mockStoreKey := storetypes.NewKVStoreKey("test")
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})
	mockclientCtx := client.Context{}.
		WithTxConfig(encCfg.TxConfig).
		WithClient(clitestutil.NewMockCometRPC(abci.ResponseQuery{}))

	return &fixture{
		ctx:           testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test")).Ctx.WithBlockHeader(cmproto.Header{}),
		mockStoreKey:  mockStoreKey,
		mockMsgURL:    "test",
		mockclientCtx: mockclientCtx,
		txBuilder:     mockclientCtx.TxConfig.NewTxBuilder(),
	}
}

func TestNewPrepareProposal(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	_, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	coins := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(100)))

	testcases := []struct {
		msg            sdk.Msg
		Name           bool
		ResolveAddress string
		Owner          string
		Amount         sdk.Coins
	}{
		{msg: &nstypes.MsgBid{
			Name:           "alice-cosmos",
			ResolveAddress: string(addr1),
			Owner:          string(addr2),
			Amount:         coins},
		},
		{msg: &nstypes.MsgBid{
			Name:           "test",
			ResolveAddress: "", // Later test that this must be address
			Owner:          "", // Later test that this must be address
			Amount:         coins},
		},
	}
	for _, tc := range testcases {
		err := f.txBuilder.SetMsgs(tc.msg)
		require.NoError(t, err)

		tx := f.txBuilder.GetTx()
		txEncoder := f.mockclientCtx.TxConfig.TxEncoder()

		encoded, errEncode := txEncoder(tx)
		require.NoError(t, errEncode)

		sdkCtx := sdk.UnwrapSDKContext(f.ctx)

		proposal, errProposal := NewPrepareProposal(sdkCtx, &abci.RequestPrepareProposal{Txs: [][]byte{encoded}})
		if errProposal != nil {
			t.Fatalf("Error from NewPrepareProposal: %v", errProposal)
		}

		require.Equal(t, proposal.(*abci.ResponsePrepareProposal).Txs[0], []byte(encoded))

	}
}

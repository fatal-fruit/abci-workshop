package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type txs struct {
	sender    sdk.AccAddress
	recipient sdk.AccAddress
	priority  uint64
}

type testTx struct {
	id       int
	priority int64
	nonce    uint64
	address  sdk.AccAddress
}

func (testTx) GetMsgsV2() ([]protoreflect.ProtoMessage, error) {
	return nil, errors.New("not implemented")
}

func (tx testTx) GetMsgs() []sdk.Msg { return nil }

func TestNewNoPrioMempool(t *testing.T) {

	_, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	_, _, addr3 := testdata.KeyTestPubAddr()
	_, _, addr4 := testdata.KeyTestPubAddr()
	_, _, addr5 := testdata.KeyTestPubAddr()
	_, _, addr6 := testdata.KeyTestPubAddr()
	_, _, addr7 := testdata.KeyTestPubAddr()
	_, _, addr8 := testdata.KeyTestPubAddr()

	testcases := []struct {
		txs []txs
	}{
		{txs: []txs{
			{sender: addr1, recipient: addr5, priority: 1},
			{sender: addr2, recipient: addr6, priority: 2},
			{sender: addr3, recipient: addr7, priority: 5},
			{sender: addr4, recipient: addr8, priority: 30},
		},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			logger := log.NewNopLogger()
			mempool := &NoPrioMempool{logger: logger}
			fmt.Println(mempool.txs)
			for i, tx := range tc.txs {
				tx := testTx{id: i, priority: int64(tx.priority), address: tx.sender, nonce: uint64(i)}
				err := mempool.Insert(context.Background(), tx)
				require.NoError(t, err)
				require.Equal(t, i+1, mempool.CountTx())

				count := mempool.CountTx()
				require.NoError(t, err)
				require.Equal(t, i+1, count)

				// not functioning
				// mempool.Remove(tx)
				// require.NoError(t, err)
				// require.Equal(t, i, count)

				// not functioning
				test := mempool.Select(context.Background(), nil)
				require.NoError(t, err)
				require.Equal(t, mempool.Count, len(tx.address))
			}
		})
	}

}

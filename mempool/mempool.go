package mempool

import (
	"bytes"
	"context"
	"cosmossdk.io/log"
	json "encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
)

var _ mempool.Mempool = (*NoPrioMempool)(nil)

func NewNoPrioMempool(logger log.Logger) *NoPrioMempool {
	return &NoPrioMempool{
		logger: logger.With("module", "noprio-mempool"),
		txs:    make([]sdk.Tx, 0),
		idx:    0,
	}
}

type NoPrioMempool struct {
	logger log.Logger
	txs    []sdk.Tx
	idx    int
}

type npmTxs struct {
	idx int
	txs []sdk.Tx
}

func (npm *npmTxs) Next() mempool.Iterator {
	// increment the index
	if npm.idx < len(npm.txs)-1 {
		npm.idx++
		return npm
	}
	return nil
}

func (npm *npmTxs) Tx() sdk.Tx {
	// return Tx
	return npm.txs[npm.idx]
}

func (npm *NoPrioMempool) Insert(_ context.Context, tx sdk.Tx) error {
	//append to txs
	npm.logger.Info(fmt.Sprintf("This is the transaction %v ", tx))
	npm.txs = append(npm.txs, tx)

	return nil
}

func (npm *NoPrioMempool) Select(_ context.Context, _ [][]byte) mempool.Iterator {
	iterator := &npmTxs{idx: 0, txs: npm.txs}
	return iterator
}

// return the length of txs
func (npm *NoPrioMempool) CountTx() int {
	return len(npm.txs)
}

func (npm *NoPrioMempool) Remove(tx sdk.Tx) error {
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	for i, mempoolTx := range npm.txs {
		mempoolTxBytes, err := json.Marshal(mempoolTx)
		if err != nil {
			return err
		}

		if bytes.Equal(txBytes, mempoolTxBytes) {
			npm.txs = append(npm.txs[:i], npm.txs[i+1:]...)
			npm.idx++
			return nil
		}
	}

	// If the transaction is not in the mempool, return an error.
	return fmt.Errorf("transaction not found in the mempool")
}

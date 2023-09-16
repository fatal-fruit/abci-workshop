package app

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
)

var _ mempool.Mempool = (*NoPrioMempool)(nil)

func NewNoPrioMempool(logger log.Logger) *NoPrioMempool {
	return &NoPrioMempool{logger: logger.With("module", "noprio-mempool")}
}

type NoPrioMempool struct {
	logger    log.Logger
	txEncoder sdk.TxEncoder
	txs       []sdk.Tx
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
	iterator := &npmTxs{idx: -1, txs: make([]sdk.Tx, len(npm.txs))}
	copy(iterator.txs, npm.txs)
	return iterator
}

// return the length of txs
func (npm *NoPrioMempool) CountTx() int {
	return len(npm.txs)
}

func (npm *NoPrioMempool) Remove(tx sdk.Tx) error {
	txBytes, err := npm.txEncoder(tx)
	if err != nil {
		return err
	}

	for i, mempoolTx := range npm.txs {
		mempoolTxBytes, err := npm.txEncoder(mempoolTx)
		if err != nil {
			return err
		}

		if bytes.Equal(txBytes, mempoolTxBytes) {
			npm.txs = append(npm.txs[:i], npm.txs[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("transaction not found in the mempool")
}

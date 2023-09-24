package mempool

import (
	"context"
	"cosmossdk.io/log"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var _ mempool.Mempool = (*ThresholdMempool)(nil)

// The ThresholdMempool guarantees transactions can only be included in a proposal if they have been seen by ExtendVote at H-1
type ThresholdMempool struct {
	logger      log.Logger
	pendingPool thTxs
	pool        thTxs
}

// Creates New ThresholdMempool
func NewThresholdMempool(logger log.Logger) *ThresholdMempool {
	return &ThresholdMempool{
		logger: logger.With("module", "threshold-mempool"),
	}
}

// Inserts a transaction into the mempool
func (t *ThresholdMempool) Insert(ctx context.Context, tx sdk.Tx) error {
	sigs, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()
	if err != nil {
		t.logger.Error("Error unable to retrieve tx signatures")
		return err
	}
	// Guarantee there is at least 1 signer
	if len(sigs) == 0 {
		t.logger.Error("Error missing tx signatures")
		return fmt.Errorf("Transaction must be signed")
	}

	sig := sigs[0]
	sender := sdk.AccAddress(sig.PubKey.Address()).String()
	t.logger.Info(fmt.Sprintf("This is the sender account address :: %v", sender))

	// Set default 0 priority
	priority := int64(0)
	appTx := thTx{
		sender,
		priority,
		tx,
	}

	t.logger.Info(fmt.Sprintf("Inserting transaction from %v with priority %v", sender, priority))
	// Add the transaction to the pending pool
	t.pendingPool.txs = append(t.pendingPool.txs, appTx)
	leng := len(t.pendingPool.txs)
	t.logger.Info(fmt.Sprintf("Transactions length %v", leng))

	return nil
}

// Returns an iterator for transactions in ThresholdMempool
func (t *ThresholdMempool) Select(ctx context.Context, i [][]byte) mempool.Iterator {
	if len(t.pool.txs) == 0 {
		return nil
	}

	return &t.pool
}

// Returns an iterator for transactions in ThresholdMempool for pending transactions
func (t *ThresholdMempool) SelectPending(ctx context.Context, i [][]byte) mempool.Iterator {
	if len(t.pendingPool.txs) == 0 {
		return nil
	}

	return &t.pendingPool
}

func (t *ThresholdMempool) Update(ctx context.Context, tx sdk.Tx) error {
	sigs, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()
	if err != nil {
		return err
	}
	if len(sigs) == 0 {
		return fmt.Errorf("tx must have at least one signer")
	}

	sig := sigs[0]
	// Get the sender address from the signature
	sender := sdk.AccAddress(sig.PubKey.Address()).String()

	txToUpdate := thTx{
		sender,
		1,
		tx,
	}

	// Search for the transaction to update in the pending pool.
	for idx, ttx := range t.pendingPool.txs {
		if ttx.Equal(txToUpdate) {
			t.pendingPool.txs = removeAtIndex(t.pendingPool.txs, idx)
			t.pool.txs = append(t.pool.txs, txToUpdate)
			return nil
		}
	}
	// Remove from pendingPool, add to
	// If the transaction is not in the pending pool, return an error.
	return mempool.ErrTxNotFound
}

// Counts the amount of transactions in the ThresholdMempool
func (t *ThresholdMempool) CountTx() int {
	return len(t.pendingPool.txs)
}

// Remove transaction from the ThresholdMempool
func (t *ThresholdMempool) Remove(tx sdk.Tx) error {
	sigs, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()
	if err != nil {
		return err
	}
	if len(sigs) == 0 {
		return fmt.Errorf("tx must have at least one signer")
	}

	sig := sigs[0]
	sender := sdk.AccAddress(sig.PubKey.Address()).String()

	txToRemove := thTx{
		sender,
		1,
		tx,
	}

	// Search for the transaction to update in the pending pool
	for idx, ttx := range t.pool.txs {
		if ttx.Equal(txToRemove) {
			t.pool.txs = removeAtIndex(t.pool.txs, idx)
			return nil
		}
	}

	return mempool.ErrTxNotFound
}

var _ mempool.Iterator = &thTxs{}

// The thTxs struct represents a slice of thTx objects
type thTxs struct {
	idx int
	txs []thTx
}

// Returns the next iterator in the sequence, or nil if there are no more iterators
func (t *thTxs) Next() mempool.Iterator {
	if len(t.txs) == 0 {
		return nil
	}

	if len(t.txs) == t.idx+1 {
		return nil
	}

	t.idx++
	return t
}

// Returns the transaction at the current index,
func (t *thTxs) Tx() sdk.Tx {
	if t.idx >= len(t.txs) {
		panic(fmt.Sprintf("index out of bound: %d, Txs: %v", t.idx, t))
	}

	return t.txs[t.idx].tx
}

// Represents a transaction in the ThresholdMempool
type thTx struct {
	address  string
	priority int64
	tx       sdk.Tx
}

// Compares two thTx objects and returns true if they are equal
func (tx thTx) Equal(other thTx) bool {
	if tx.address != other.address {
		return false
	}

	//If number of messages in the transactions are not equal, return false
	if len(tx.tx.GetMsgs()) != len(other.tx.GetMsgs()) {
		return false
	}

	// Iterate over the messages in the transactions and compare them
	for i, msg := range tx.tx.GetMsgs() {
		if msg.String() != other.tx.GetMsgs()[i].String() {
			return false
		}
	}

	return true
}

// Removes the element at the specific index
func removeAtIndex[T any](slice []T, index int) []T {
	return append(slice[:index], slice[index+1:]...)
}

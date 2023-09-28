package abci

import (
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
	"github.com/fatal-fruit/cosmapp/provider"
)

type PrepareProposalHandler struct {
	logger      log.Logger
	txConfig    client.TxConfig
	cdc         codec.Codec
	mempool     *sdkmempool.SenderNonceMempool
	txProvider  provider.TxProvider
	keyname     string
	runProvider bool
}

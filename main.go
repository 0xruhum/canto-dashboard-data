package main

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

const INITIAL_BLOCK = 2537250

func main() {

	db, err := NewDB()
	if err != nil {
		log.Err(err).Msg("failed to connect to db")
	}

	ctx := context.Background()
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Err(err).Msg("failed to connect to rpc endpoint")
		return
	}

	currBlockNum, err := client.BlockNumber(ctx)
	if err != nil {
		log.Err(err).Msg("failed to get current block number")
		return
	}

	for i := int(currBlockNum); i > INITIAL_BLOCK; i-- {
		logger := log.With().Int("block", i).Logger()
		logger.Info().Msg("pulling block")
		block, err := client.BlockByNumber(ctx, big.NewInt(int64(i)))
		if err != nil {
			logger.Err(err).Msg("failed to get block")
			return
		}
		for _, tx := range block.Transactions() {
			logger.Info().Msg("found txs")
			// we need the receipt for GasUsed
			receipt, err := client.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				logger.Err(err).Str("tx_hash", tx.Hash().Hex()).Msg("couldn't retrieve tx receipt")
			}

			from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
			if err != nil {
				logger.Err(err).Str("tx_hash", tx.Hash().Hex()).Msg("couldn't get tx sender")
			}

			txData := &TxData{
				GasUsed:   receipt.GasUsed,
				GasPrice:  tx.GasPrice().Uint64(),
				To:        tx.To().Hex(),
				From:      from.Hex(),
				Hash:      tx.Hash().Hex(),
				Timestamp: time.Unix(int64(block.Time()), 0),
			}
			if err = db.AddTx(ctx, txData); err != nil {
				logger.Err(err).Str("tx_hash", tx.Hash().Hex()).Msg("failed to insert tx")
			}

		}
	}
}

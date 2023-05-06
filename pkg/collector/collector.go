package collector

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/0xruhum/canto-dashboard-data/pkg/db"
	"github.com/0xruhum/canto-dashboard-data/pkg/models"
	"github.com/rs/zerolog"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Collector struct {
	logger zerolog.Logger
	client *ethclient.Client
	db     *db.DB
	sync.Mutex
}

func NewCollector(logger zerolog.Logger, client *ethclient.Client, db *db.DB) *Collector {
	return &Collector{
		logger: logger,
		client: client,
		db:     db,
	}
}

// collects all the relevant block & tx data in the given interval and saves it to the db
// This won't return any errors. Instead, they will be logged by the collector module.
// Otherwise we'd have to setup a channel system to send back the logs. That's to much work for now
func (c *Collector) Start(ctx context.Context, fromBlock, toBlock int64) {
	for i := fromBlock; i < toBlock; i++ {
		logger := c.logger.With().Int("block", int(i)).Logger()
		logger.Info().Msg("pulling block")
		block, err := c.client.BlockByNumber(ctx, big.NewInt(int64(i)))
		if err != nil {
			logger.Err(err).Msg("failed to get block")
			continue
		}
		blockData := &models.BlockData{
			Hash:     block.Hash().Hex(),
			Number:   block.NumberU64(),
			TxHashes: []string{},
			BaseFee:  block.BaseFee().Uint64(),
		}
		for _, tx := range block.Transactions() {
			blockData.TxHashes = append(blockData.TxHashes, tx.Hash().Hex())
			logger := logger.With().Str("tx", tx.Hash().Hex()).Logger()
			data, err := c.db.GetTx(ctx, tx.Hash())
			// if the row doesn't exist we want to add it, so we proceed
			if err != nil && err.Error() != "sql: no rows in result set" {
				logger.Err(err).Msg("failed to retrieve tx data from database")
				continue
			}
			if data != nil {
				logger.Info().Msg("tx already exists in our database, skipping")
				continue
			}
			logger.Info().Msg("pulling tx data from node")
			txData, err := c.GetTxData(block, tx)
			if err != nil {
				logger.Err(err).Msg("failed to get tx data")
				continue
			}
			if err = c.db.AddTx(ctx, txData); err != nil {
				logger.Err(err).Str("tx_hash", tx.Hash().Hex()).Msg("failed to insert tx")
			}
		}
		logger.Info().Msg("done saving tx data for block")

		if err = c.db.AddBlock(ctx, blockData); err != nil {
			logger.Err(err).Msg("failed to save block data")
		}
	}
}

func (c *Collector) GetTxData(block *types.Block, tx *types.Transaction) (*models.TxData, error) {
	ctx := context.Background()
	// we need the receipt for GasUsed
	receipt, err := c.client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve tx receipt: %+v", err)
	}

	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		return nil, fmt.Errorf("couldn't get tx sender: %+v", err)
	}

	txData := &models.TxData{
		GasUsed:   receipt.GasUsed,
		GasPrice:  tx.GasPrice().Uint64(),
		From:      from.Hex(),
		Hash:      tx.Hash().Hex(),
		Timestamp: time.Unix(int64(block.Time()), 0),
	}
	// for some reason To can be nil. We just add the 0 address then
	if tx.To() == nil {
		txData.To = "0x0000000000000000000000000000000000000000"
	} else {
		txData.To = tx.To().Hex()
		code, err := c.client.CodeAt(ctx, *tx.To(), block.Number())
		if err != nil {
			return nil, fmt.Errorf("failed to get recipient's code: %+v", err)
		}
		txData.IsContract = len(code) > 0
	}
	return txData, nil
}

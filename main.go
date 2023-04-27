package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const INITIAL_BLOCK = 2537250

func main() {
	logFileName := fmt.Sprintf("logs/%v.log", time.Now().Unix())
	logFile, err := os.Create(logFileName)
	if err != nil {
		log.Err(err).Msg("failed to initialize log file")
	}
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	multi := zerolog.MultiLevelWriter(consoleWriter, logFile)

	logger := zerolog.New(multi).With().Timestamp().Logger()

	db, err := NewDB()
	if err != nil {
		logger.Err(err).Msg("failed to connect to db")
	}

	ctx := context.Background()
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		logger.Err(err).Msg("failed to connect to rpc endpoint")
		return
	}

	collector := &Collector{
		client: client,
		db:     db,
	}

	currBlockNum, err := client.BlockNumber(ctx)
	if err != nil {
		logger.Err(err).Msg("failed to get current block number")
		return
	}

	middleBlock := (currBlockNum - INITIAL_BLOCK) / 2
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() { wg.Done() }()
		for i := int(currBlockNum); i > INITIAL_BLOCK+int(middleBlock); i-- {
			logger := logger.With().Int("block", i).Logger()
			logger.Info().Msg("pulling block")
			block, err := client.BlockByNumber(ctx, big.NewInt(int64(i)))
			if err != nil {
				logger.Err(err).Msg("failed to get block")
				continue
			}
			for _, tx := range block.Transactions() {
				logger := logger.With().Str("tx", tx.Hash().Hex()).Logger()
				logger.Info().Msg("pulling tx data")
				txData, err := collector.GetTxData(block, tx)
				if err != nil {
					logger.Err(err).Msg("failed to get tx data")
					continue
				}
				collector.Lock()
				if err = db.AddTx(ctx, txData); err != nil {
					logger.Err(err).Str("tx_hash", tx.Hash().Hex()).Msg("failed to insert tx")
				}
				collector.Unlock()
			}
			logger.Info().Msg("done saving tx data for block")
		}
	}()
	wg.Add(1)
	go func() {
		defer func() { wg.Done() }()
		for i := INITIAL_BLOCK; i <= INITIAL_BLOCK+int(middleBlock); i++ {
			logger := logger.With().Int("block", i).Logger()
			logger.Info().Msg("pulling block")
			block, err := client.BlockByNumber(ctx, big.NewInt(int64(i)))
			if err != nil {
				logger.Err(err).Msg("failed to get block")
				continue
			}
			for _, tx := range block.Transactions() {
				logger := logger.With().Str("tx", tx.Hash().Hex()).Logger()
				logger.Info().Msg("pulling tx data")
				txData, err := collector.GetTxData(block, tx)
				if err != nil {
					logger.Err(err).Msg("failed to get tx data")
					continue
				}
				collector.Lock()
				if err = db.AddTx(ctx, txData); err != nil {
					logger.Err(err).Str("tx_hash", tx.Hash().Hex()).Msg("failed to insert tx")
				}
				collector.Unlock()
			}
			logger.Info().Msg("done saving tx data for block")
		}
	}()
	wg.Wait()
}

type Collector struct {
	client *ethclient.Client
	db     *DB
	sync.Mutex
}

func (c *Collector) GetTxData(block *types.Block, tx *types.Transaction) (*TxData, error) {
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

	txData := &TxData{
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

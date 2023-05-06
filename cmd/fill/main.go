package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/0xruhum/canto-dashboard-data/pkg/collector"
	"github.com/0xruhum/canto-dashboard-data/pkg/db"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	logFileName := fmt.Sprintf("logs/fill/%v.log", time.Now().Unix())
	logFile, err := os.Create(logFileName)
	if err != nil {
		log.Err(err).Msg("failed to initialize log file")
	}
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	multi := zerolog.MultiLevelWriter(consoleWriter, logFile)

	logger := zerolog.New(multi).With().Timestamp().Logger()

	db, err := db.NewDB()
	if err != nil {
		logger.Err(err).Msg("failed to connect to db")
	}

	ctx := context.Background()
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		logger.Err(err).Msg("failed to connect to rpc endpoint")
		return
	}

	oldestKnownBlock, err := db.GetOldestBlock(ctx)
	if err != nil {
		logger.Err(err).Msg("failed to get last known block from database")
		return
	}
	// let's not get ahead of ourselfs
	oldestBlock := int64(2_000_000)
	collector := collector.NewCollector(logger, client, db)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() { wg.Done() }()
		collector.Start(ctx, oldestBlock, oldestKnownBlock)
	}()
}

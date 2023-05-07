package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/0xruhum/canto-dashboard-data/pkg/collector"
	"github.com/0xruhum/canto-dashboard-data/pkg/db"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {

}

func main() {
	regex := regexp.MustCompile(`"level":"error","block":(\d+)`)
	logFileName := fmt.Sprintf("logs/catchup/%v.log", time.Now().Unix())
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
	client, err := ethclient.Dial("http://localhost:8546")
	if err != nil {
		logger.Err(err).Msg("failed to connect to rpc endpoint")
		return
	}

	collector := collector.NewCollector(logger, client, db)

	oldLogFile, err := os.Open("logs/catchup/1683363860.log")
	if err != nil {
		logger.Err(err).Msg("failed to open old log file")
		return
	}
	defer oldLogFile.Close()
	scanner := bufio.NewScanner(oldLogFile)
	for scanner.Scan() {
		match := regex.FindStringSubmatch(scanner.Text())
		if len(match) > 1 {
			blockNumber, err := strconv.Atoi(match[1])
			if err != nil {
				logger.Err(err).Msg("failed to get block number from error message")
				return
			}
			collector.Start(ctx, int64(blockNumber), int64(blockNumber+1))
		}
	}
}

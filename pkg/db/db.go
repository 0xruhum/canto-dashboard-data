package db

import (
	"context"

	"github.com/0xruhum/canto-dashboard-data/pkg/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	CREATE_TABLE_TXS = `
		CREATE TABLE IF NOT EXISTS txs (
			hash TEXT PRIMARY KEY NOT NULL,
			sender TEXT,
			recipient TEXT,
			iscontract boolean,
			gasprice bigint,
			gasused bigint,
			timestamp TIMESTAMP 
		);	
	`

	SET_TX = `
		INSERT INTO txs (
			hash,
			sender,
			recipient,
			iscontract,
			gasprice,
			gasused,
			timestamp
		) VALUES (
			:hash,
			:sender,
			:recipient,
			:iscontract,
			:gasprice,
			:gasused,
			:timestamp
		);
	`

	GET_TX = `
		SELECT * FROM txs WHERE hash = $1;
	`

	// I should first test how long this query takes. If it's quick <0.1 secs
	// I won't have to store it
	GET_TXS_PER_DAY = `
		SELECT timestamp::date, COUNT(*) AS tx_count
		FRPOM txs
		GROUP BY timestamp::date
		ORDER BY timestamp::date ASC;
	`

	// I could probably write this using `GET_TXS_PER_DAY` but this redundant solution
	// is way easier.
	GET_TXS_PER_MONTH = `
		SELECT date_trunc('month', timestamp) AS tx_month, COUNT(*) AS tx_count
		FROM txs
		GROUP_BY tx_month
		ORDER BY tx_month ASC;
	`

	CREATE_TABLE_BLOCKS = `
		CREATE TABLE IF NOT EXISTS blocks (
			hash TEXT PRIMARY KEY NOT NULL,
			number BIGINT,
			tx_hashes TEXT[],
			base_fee BIGINT
		);
	`

	SET_BLOCK = `
		INSERT INTO blocks (
			hash,
			number,
			tx_hashes,
			base_fee
		) VALUES (
			$1,
			$2,
			$3,
			$4
		);
	`

	GET_LATEST_BLOCK = `
		SELECT MAX(number) FROM blocks;
	`

	GET_OLDEST_BLOCK = `
		SELECT MIN(number) FROM blocks;
	`
)

type DB struct {
	*sqlx.DB
}

func NewDB() (*DB, error) {
	db, err := sqlx.Connect("postgres", "dbname=mentat sslmode=disable")
	if err != nil {
		return nil, err
	}

	db.MustExec(CREATE_TABLE_TXS)
	db.MustExec(CREATE_TABLE_BLOCKS)

	return &DB{db}, nil
}

func (db *DB) AddTx(ctx context.Context, tx *models.TxData) error {
	_, err := db.NamedExecContext(ctx, SET_TX, tx)
	return err
}

func (db *DB) GetTx(ctx context.Context, hash common.Hash) (*models.TxData, error) {
	data := &models.TxData{}
	if err := db.GetContext(ctx, data, GET_TX, hash.Hex()); err != nil {
		return nil, err
	}
	return data, nil
}

func (db *DB) AddBlock(ctx context.Context, block *models.BlockData) error {
	// need to convert to pq.StringArray so we don't use NamedExecContext here
	_, err := db.ExecContext(ctx, SET_BLOCK, block.Hash, block.Number, pq.StringArray(block.TxHashes), block.BaseFee)
	return err
}

func (db *DB) GetLatestBlock(ctx context.Context) (int64, error) {
	var num int64
	if err := db.GetContext(ctx, &num, GET_LATEST_BLOCK); err != nil {
		return int64(0), err
	}
	return num, nil
}

func (db *DB) GetOldestBlock(ctx context.Context) (int64, error) {
	var num int64
	if err := db.GetContext(ctx, &num, GET_OLDEST_BLOCK); err != nil {
		return int64(0), err
	}
	return num, nil
}

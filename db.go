package main

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	CREATE_TABLE = `
		CREATE TABLE IF NOT EXISTS txs (
			hash TEXT PRIMARY KEY NOT NULL,
			sender TEXT,
			recipient TEXT,
			isContract boolean,
			gasPrice bigint,
			gasUsed bigint,
			timestamp TIMESTAMP 
		);	
	`

	SET_TX = `
		INSERT INTO txs (
			hash,
			sender,
			recipient,
			isContract,
			gasPrice,
			gasUsed,
			timestamp
		) VALUES (
			:hash,
			:sender,
			:recipient,
			:isContract,
			:gasPrice,
			:gasUsed,
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
)

type DB struct {
	*sqlx.DB
}

func NewDB() (*DB, error) {
	db, err := sqlx.Connect("postgres", "dbname=mentat sslmode=disable")
	if err != nil {
		return nil, err
	}

	db.MustExec(CREATE_TABLE)

	return &DB{db}, nil
}

func (db *DB) AddTx(ctx context.Context, tx *TxData) error {
	_, err := db.NamedExecContext(ctx, SET_TX, tx)
	return err
}

func (db *DB) GetTx(ctx context.Context, hash common.Hash) (*TxData, error) {
	data := &TxData{}
	if err := db.GetContext(ctx, data, GET_TX, hash.Hex()); err != nil {
		return nil, err
	}
	return data, nil
}

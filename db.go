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
			gasPrice,
			gasUsed,
			timestamp
		) VALUES (
			:hash,
			:sender,
			:recipient,
			:gasPrice,
			:gasUsed,
			:timestamp
		);
	`

	GET_TX = `
		SELECT * FROM txs WHERE hash = $1;
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

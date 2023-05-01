package main

import (
	"time"
)

// we don't need all the data a normal TX contains. To save DB space we filter out everything we don't need
type TxData struct {
	Hash       string    `db:"hash"`
	From       string    `db:"sender"`
	To         string    `db:"recipient"`
	IsContract bool      `db:"iscontract"`
	GasPrice   uint64    `db:"gasprice"`
	GasUsed    uint64    `db:"gasused"`
	Timestamp  time.Time `db:"timestamp"`
}

type BlockData struct {
	Hash     string   `db:"hash"`
	Number   uint64   `db:"number"`
	TxHashes []string `db:"tx_hashes"`
	BaseFee  uint64   `db:"base_fee"`
}

package models

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

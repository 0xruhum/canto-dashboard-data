package models

type BlockData struct {
	Hash     string   `db:"hash"`
	Number   uint64   `db:"number"`
	TxHashes []string `db:"tx_hashes"`
	BaseFee  uint64   `db:"base_fee"`
}

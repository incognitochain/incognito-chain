package jsonresult

import "github.com/ninjadotorg/constant/transaction"

type TransactionDetail struct {
	BlockHash 			 						string 	`json:"BlockHash"`
	Index						 						uint64 	`json:"Index"`
	Hash             						string 	`json:"Hash"`
	Tx													*transaction.Tx `json:"TxData"`
}

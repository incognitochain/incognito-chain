package jsonresult

type TransactionDetail struct {
	BlockHash 			 						string 	`json:"BlockHash"`
	Index						 						uint64 	`json:"Index"`
	Hash             						string 	`json:"Hash"`
	ValidateTransaction         bool 		`json:"ValidateTransaction"`
	Type       			 						string 	`json:"type"`
	TxVirtualSize      					uint64 	`json:"SalaryPerTx"`
	SenderAddrLastByte          byte 		`json:"BlockProducer"`
	TxFee 											uint64 	`json:"BlockProducerSig"`
}

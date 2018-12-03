package transaction

import "github.com/ninjadotorg/constant/common"

type TxTokenPrivacyData struct {
	PropertyID     common.Hash // = hash of TxTokenData data
	PropertyName   string
	PropertySymbol string

	Type   int    // action type
	Amount uint64 // init amount
	Descs  []*JoinSplitDesc `json:"Descs"`
}

package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"errors"
	"fmt"
)

type TxTokenPrivacyData struct {
	PropertyID     common.Hash // = hash of TxTokenData data
	PropertyName   string
	PropertySymbol string

	Type   int    // action type
	Amount uint64 // init amount
	Descs  []*JoinSplitDesc `json:"Descs"`
}

// Hash - return hash of token data, be used as Token ID
func (self TxTokenPrivacyData) Hash() (*common.Hash, error) {
	if self.Descs == nil {
		return nil, errors.New("Privacy data is empty")
	}
	record := self.PropertyName + self.PropertySymbol + fmt.Sprintf("%d", self.Amount)
	for _, out := range self.Descs {
		record += out.toString()
	}
	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash, nil
}

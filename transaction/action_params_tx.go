package transaction

import (
	"github.com/ninjadotorg/cash-prototype/common"
	"strconv"
	"fmt"
)

type Param struct {
	AgentID string
	NumOfIssuingCoins int
	NumOfIssuingBonds int
	Tax float64
}

type ActionParamTx struct {
	Version int
	Type string // NORMAL / ACTION_PARAMS
	Param *Param
	LockTime int
	txHash *common.Hash
}

func (self *ActionParamTx) Hash() (*common.Hash) {
	if self.txHash != nil {
		return self.txHash
	}
	record := strconv.Itoa(self.Version) + strconv.Itoa(self.Version)
	record += self.Type
	record += self.Param.AgentID
	record += fmt.Sprint(self.Param.NumOfIssuingCoins)
	record += fmt.Sprint(self.Param.NumOfIssuingBonds)
	record += fmt.Sprint(self.Param.Tax)
	hash := common.DoubleHashH([]byte(record))
	self.txHash = &hash
	return self.txHash
}

func (self *ActionParamTx) ValidateTransaction() bool {
	return true
}

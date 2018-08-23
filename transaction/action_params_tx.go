package transaction

import (
	"github.com/ninjadotorg/cash-prototype/common"
	"strconv"
	"fmt"
	//"encoding/json"
)

type Param struct {
	AgentID           string
	NumOfIssuingCoins int
	NumOfIssuingBonds int
	Tax               float64
}

//func (self *Param) MarshalJSON() ([]byte, error) {
//	m := make(map[string]interface{})
//	m["AgentID"] = self.AgentID
//	m["NumOfIssuingCoins"] = self.NumOfIssuingCoins
//	m["NumOfIssuingBonds"] = self.NumOfIssuingBonds
//	m["Tax"] = self.Tax
//	return json.Marshal(m)
//}
//
//func (self *Param) UnmarshalJSON(data []byte) (error) {
//	m := make(map[string]interface{})
//	_ = json.Unmarshal(data, &m)
//
//	self.NumOfIssuingCoins = m["NumOfIssuingCoins"].(int)
//	self.NumOfIssuingBonds = m["NumOfIssuingBonds"].(int)
//	self.Tax = m["Tax"].(float64)
//	self.AgentID = m["AgentID"].(string)
//	return nil
//}

type ActionParamTx struct {
	Version  int
	Type     string // NORMAL / ACTION_PARAMS
	Param    *Param
	LockTime int64
	txHash   *common.Hash
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

func (self *ActionParamTx) GetType() (string) {
	return self.Type
}

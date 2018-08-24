package transaction

import (
	"github.com/ninjadotorg/cash-prototype/common"
	"strconv"
	"fmt"
	//"encoding/json"
)

type Param struct {
	AgentID           string  `json:"AgentID"`
	NumOfIssuingCoins int     `json:"NumOfIssuingCoins"`
	NumOfIssuingBonds int     `json:"NumOfIssuingBonds"`
	Tax               float64 `json:"Tax"`
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
	Version  int    `json:"Version"`
	Type     string `json:"Type"` // NORMAL / ACTION_PARAMS
	Param    *Param `json:"Param"`
	LockTime int64  `json:"LockTime"`
}

func (self *ActionParamTx) Hash() (*common.Hash) {
	record := strconv.Itoa(self.Version) + strconv.Itoa(self.Version)
	record += self.Type
	record += self.Param.AgentID
	record += fmt.Sprint(self.Param.NumOfIssuingCoins)
	record += fmt.Sprint(self.Param.NumOfIssuingBonds)
	record += fmt.Sprint(self.Param.Tax)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (self *ActionParamTx) ValidateTransaction() bool {
	return true
}

func (self *ActionParamTx) GetType() (string) {
	return self.Type
}

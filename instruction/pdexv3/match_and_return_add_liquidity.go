package pdexv3

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type MatchAndReturnAddLiquidity struct {
	pairHash       string
	receiveAddress string // receive nfct
	refundAddress  string // refund pToken
	tokenID        string
	actualAmount   uint64
	returnAmount   uint64
	nfctID         string
	txReqID        string
	shardID        byte
}

func NewMatchAndReturnAddLiquidity() *MatchAndReturnAddLiquidity {
	return &MatchAndReturnAddLiquidity{}
}

func NewMatchAndReturnAddLiquidityWithValue(
	pairHash, receiveAddress,
	refundAddress, tokenID,
	nfctID, txReqID string,
	actualAmount, returnAmount uint64,
	shardID byte,
) *MatchAndReturnAddLiquidity {
	return &MatchAndReturnAddLiquidity{
		pairHash:       pairHash,
		receiveAddress: receiveAddress,
		refundAddress:  refundAddress,
		tokenID:        tokenID,
		nfctID:         nfctID,
		actualAmount:   actualAmount,
		returnAmount:   returnAmount,
		txReqID:        txReqID,
		shardID:        shardID,
	}
}

func (m *MatchAndReturnAddLiquidity) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PairHash       string `json:"PairHash"`
		ReceiveAddress string `json:"ReceiveAddress"` // receive nfct
		RefundAddress  string `json:"RefundAddress"`  // refund pToken
		TokenID        string `json:"TokenID"`
		NfctID         string `json:"NfctID"`
		ActualAmmount  uint64 `json:"ActualAmmount"`
		ReturnAmount   uint64 `json:"ReturnAmount"`
		TxReqID        string `json:"TxReqID"`
		ShardID        byte   `json:"ShardID"`
	}{
		PairHash:       m.pairHash,
		ReceiveAddress: m.receiveAddress,
		RefundAddress:  m.refundAddress,
		TokenID:        m.tokenID,
		NfctID:         m.nfctID,
		ActualAmmount:  m.actualAmount,
		ReturnAmount:   m.returnAmount,
		TxReqID:        m.txReqID,
		ShardID:        m.shardID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (m *MatchAndReturnAddLiquidity) UnmarshalJSON(data []byte) error {
	temp := struct {
		PairHash       string `json:"PairHash"`
		ReceiveAddress string `json:"ReceiveAddress"` // receive nfct
		RefundAddress  string `json:"RefundAddress"`  // refund pToken
		TokenID        string `json:"TokenID"`
		NfctID         string `json:"NfctID"`
		ActualAmmount  uint64 `json:"ActualAmmount"`
		ReturnAmount   uint64 `json:"ReturnAmount"`
		TxReqID        string `json:"TxReqID"`
		ShardID        byte   `json:"ShardID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	m.pairHash = temp.PairHash
	m.receiveAddress = temp.ReceiveAddress
	m.refundAddress = temp.RefundAddress
	m.tokenID = temp.TokenID
	m.nfctID = temp.NfctID
	m.actualAmount = temp.ActualAmmount
	m.returnAmount = temp.ReturnAmount
	m.txReqID = temp.TxReqID
	m.shardID = temp.ShardID
	return nil
}

func (m *MatchAndReturnAddLiquidity) FromStringArr(source []string) error {
	if len(source) != 11 {
		return fmt.Errorf("Receive length %v but expect %v", len(source), 11)
	}
	if source[0] != strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta) {
		return fmt.Errorf("Receive metaType %v but expect %v", source[0], metadataCommon.PDexV3AddLiquidityMeta)
	}
	if source[1] != MatchAndReturnStatus {
		return fmt.Errorf("Receive status %v but expect %v", source[1], MatchAndReturnStatus)
	}
	if source[2] == "" {
		return errors.New("Pair hash is invalid")
	}
	m.pairHash = source[2]
	receiveAddress := privacy.OTAReceiver{}
	err := receiveAddress.FromString(source[3])
	if err != nil {
		return err
	}
	if !receiveAddress.IsValid() {
		return errors.New("receive Address is invalid")
	}
	m.receiveAddress = source[3]
	refundAddress := privacy.OTAReceiver{}
	err = refundAddress.FromString(source[4])
	if err != nil {
		return err
	}
	if !refundAddress.IsValid() {
		return errors.New("refund Address is invalid")
	}
	m.refundAddress = source[4]
	tokenID, err := common.Hash{}.NewHashFromStr(source[5])
	if err != nil {
		return err
	}
	if tokenID.IsZeroValue() {
		return errors.New("TokenID is empty")
	}
	m.tokenID = source[5]
	actualAmount, err := strconv.ParseUint(source[6], 10, 32)
	if err != nil {
		return err
	}
	m.actualAmount = actualAmount
	returnAmount, err := strconv.ParseUint(source[7], 10, 32)
	if err != nil {
		return err
	}
	m.returnAmount = returnAmount
	nfctID, err := common.Hash{}.NewHashFromStr(source[8])
	if err != nil {
		return err
	}
	if nfctID.IsZeroValue() {
		return errors.New("NfctID is empty")
	}
	m.tokenID = source[8]
	m.txReqID = source[9]
	shardID, err := strconv.Atoi(source[10])
	if err != nil {
		return err
	}
	m.shardID = byte(shardID)
	return nil
}

func (m *MatchAndReturnAddLiquidity) StringArr() []string {
	metaDataType := strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta)
	res := []string{metaDataType, MatchAndReturnStatus}
	res = append(res, m.pairHash)
	res = append(res, m.receiveAddress)
	res = append(res, m.refundAddress)
	res = append(res, m.tokenID)
	actualAmount := strconv.FormatUint(m.actualAmount, 10)
	res = append(res, actualAmount)
	returnAmount := strconv.FormatUint(m.returnAmount, 10)
	res = append(res, returnAmount)
	res = append(res, m.nfctID)
	res = append(res, m.txReqID)
	shardID := strconv.Itoa(int(m.shardID))
	res = append(res, shardID)
	return res
}

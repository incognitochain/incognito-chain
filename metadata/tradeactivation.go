package metadata

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/pkg/errors"
)

// TradeActivation sends request to create a BuySellRequest or BuyBackRequest from DCB to GOV to buy or sell bonds
type TradeActivation struct {
	TradeID []byte
	MetadataBase
}

func NewTradeActivation(data map[string]interface{}) (Metadata, error) {
	result := TradeActivation{}
	s, _ := hex.DecodeString(data["TradeID"].(string))
	result.TradeID = s
	result.Type = TradeActivationMeta
	return &result, nil
}

func (act *TradeActivation) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if sender is a member of DCB Board
	if !txCreatedByDCBBoardMember(txr, bcr) {
		return false, errors.New("TradeActivation tx must be created by DCB Governor")
	}

	// Check if tradeID is in current proposal
	found := false
	for _, trade := range bcr.GetAllTrades() {
		if bytes.Equal(trade.TradeID, act.TradeID) {
			found = true
		}
	}
	if !found {
		return false, errors.New("TradeActivation id is not in current proposal")
	}

	// Check if tradeID hasn't been activated and amount is positive
	_, _, activated, amount, err := bcr.GetTradeActivation(act.TradeID)
	if err != nil {
		return false, err
	}
	if activated {
		return false, errors.New("Trade is activated")
	}
	if amount == 0 {
		return false, errors.New("Trade proposal is already done")
	}

	return true, nil
}

func (act *TradeActivation) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(act.TradeID) == 0 {
		return false, false, errors.New("Wrong TradeID")
	}
	return false, true, nil
}

func (act *TradeActivation) ValidateMetadataByItself() bool {
	return true
}

func (act *TradeActivation) Hash() *common.Hash {
	record := string(act.TradeID)

	// final hash
	record += act.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

type TradeActivationAction struct {
	TradeID []byte
}

func (act *TradeActivation) BuildReqActions(txr Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	value, err := getTradeActivationActionValue(act, txr, bcr)
	if err != nil {
		return nil, err
	}
	action := []string{strconv.Itoa(TradeActivationMeta), value}
	return [][]string{action}, nil
}

func getTradeActivationActionValue(act *TradeActivation, txr Transaction, bcr BlockchainRetriever) (string, error) {
	action := &TradeActivationAction{
		TradeID: act.TradeID,
	}
	value, err := json.Marshal(action)
	return string(value), err
}

func ParseTradeActivationActionValue(value string) ([]byte, error) {
	action := &TradeActivationAction{}
	err := json.Unmarshal([]byte(value), action)
	if err != nil {
		return nil, err
	}
	return action.TradeID, nil
}

func (act *TradeActivation) CalculateSize() uint64 {
	return calculateSize(act)
}

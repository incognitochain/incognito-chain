package metadata

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/pkg/errors"
)

// TradeActivation sends request to create a BuySellRequest or BuyBackRequest from DCB to GOV to buy or sell bonds
type TradeActivation struct {
	TradeID []byte
	Amount  uint64
	MetadataBase
}

func NewTradeActivation(data map[string]interface{}) (Metadata, error) {
	result := TradeActivation{}
	s, _ := hex.DecodeString(data["TradeID"].(string))
	amount, ok := data["Amount"].(float64)
	if !ok {
		return nil, errors.New("amount invalid")
	}
	result.TradeID = s
	result.Amount = uint64(amount)
	result.Type = TradeActivationMeta
	return &result, nil
}

func txCreatedByDCBBoardMember(txr Transaction, bcr BlockchainRetriever) bool {
	isBoard := false
	txPubKey := txr.GetSigPubKey()
	fmt.Printf("check if created by dcb board: %v\n", txPubKey)
	for _, member := range bcr.GetBoardPubKeys(common.DCBBoard) {
		fmt.Printf("member of board pubkey: %v\n", member)
		if bytes.Equal(member, txPubKey) {
			isBoard = true
		}
	}
	return isBoard
}

func (act *TradeActivation) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if sender is a member of DCB Board
	if !txCreatedByDCBBoardMember(txr, bcr) {
		return false, errors.New("TradeActivation tx must be created by DCB Governor")
	}

	// Check if tradeID is in current proposal
	var trade *component.TradeBondWithGOV
	for _, t := range bcr.GetAllTrades() {
		if bytes.Equal(t.TradeID, act.TradeID) {
			trade = t
		}
	}
	if trade == nil {
		return false, errors.New("TradeActivation id is not in current proposal")
	}

	// Check if tradeID hasn't been activated and amount left is higher than requested
	_, _, activated, amount, err := bcr.GetTradeActivation(act.TradeID)
	if err == nil {
		if activated {
			return false, errors.New("trade is activated")
		}
		if act.Amount > amount {
			return false, errors.Errorf("requested amount is too high: %d > %d", act.Amount, amount)
		}
	} else if act.Amount > trade.Amount {
		return false, errors.Errorf("requested amount is too high: %d > %d", act.Amount, trade.Amount)
	}

	return true, nil
}

func (act *TradeActivation) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(act.TradeID) == 0 {
		return false, false, errors.New("wrong TradeID")
	}
	return false, true, nil
}

func (act *TradeActivation) ValidateMetadataByItself() bool {
	return true
}

func (act *TradeActivation) Hash() *common.Hash {
	record := string(act.TradeID)
	record += strconv.FormatUint(act.Amount, 10)

	// final hash
	record += act.MetadataBase.Hash().String()
	hash := common.HashH([]byte(record))
	return &hash
}

type TradeActivationAction struct {
	TradeID []byte
	Amount  uint64
}

func (act *TradeActivation) BuildReqActions(txr Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	fmt.Printf("[db] trade act: build req act for tx: %h\n", txr.Hash())
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
		Amount:  act.Amount,
	}
	value, err := json.Marshal(action)
	return string(value), err
}

func ParseTradeActivationActionValue(value string) ([]byte, uint64, error) {
	action := &TradeActivationAction{}
	err := json.Unmarshal([]byte(value), action)
	if err != nil {
		return nil, 0, err
	}
	return action.TradeID, action.Amount, nil
}

func (act *TradeActivation) CalculateSize() uint64 {
	return calculateSize(act)
}

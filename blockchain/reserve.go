package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type IssuingReqAction struct {
	TxReqID common.Hash             `json:"txReqId"`
	Meta    metadata.IssuingRequest `json:"meta"`
}

type IssuingInfo struct {
	ReceiverAddress privacy.PaymentAddress
	Amount          uint64
	RequestedTxID   common.Hash
	TokenID         common.Hash
}

type ContractingReqAction struct {
	TxReqID common.Hash                 `json:"txReqId"`
	Meta    metadata.ContractingRequest `json:"meta"`
}

type ContractingInfo struct {
	BurnerAddress     privacy.PaymentAddress
	BurnedConstAmount uint64
	RequestedTxID     common.Hash
}

func buildInstTypeForContractingAction(
	beaconBestState *BestStateBeacon,
	md *metadata.ContractingRequest,
) string {
	if bytes.Equal(md.CurrencyType[:], common.USDAssetID[:]) {
		return "accepted"
	}
	// crypto
	stabilityInfo := beaconBestState.StabilityInfo
	oracle := stabilityInfo.Oracle
	spendReserveData := stabilityInfo.DCBConstitution.DCBParams.SpendReserveData
	bestBlockHeight := beaconBestState.BestBlock.Header.Height
	if spendReserveData == nil {
		return "refund"
	}
	reserveData, existed := spendReserveData[md.CurrencyType]
	if !existed {
		return "refund"
	}
	if bestBlockHeight+1 > reserveData.EndBlock ||
		md.BurnedConstAmount > reserveData.Amount {
		return "refund"
	}
	if bytes.Equal(md.CurrencyType[:], common.ETHAssetID[:]) &&
		oracle.ETH < reserveData.ReserveMinPrice {
		return "refund"
	}
	return "accepted"
}

func buildInstructionsForContractingReq(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return [][]string{}, err
	}
	var contractingReqAction ContractingReqAction
	err = json.Unmarshal(contentBytes, &contractingReqAction)
	if err != nil {
		return nil, err
	}
	md := contractingReqAction.Meta
	reqTxID := contractingReqAction.TxReqID
	instructions := [][]string{}
	instType := buildInstTypeForContractingAction(beaconBestState, &md)

	cInfo := ContractingInfo{
		BurnerAddress:     md.BurnerAddress,
		BurnedConstAmount: md.BurnedConstAmount,
		RequestedTxID:     reqTxID,
	}
	cInfoBytes, err := json.Marshal(cInfo)
	if err != nil {
		return nil, err
	}
	returnedInst := []string{
		strconv.Itoa(metadata.ContractingRequestMeta),
		strconv.Itoa(int(shardID)),
		instType,
		string(cInfoBytes),
	}
	instructions = append(instructions, returnedInst)
	return instructions, nil
}

func (blockgen *BlkTmplGenerator) buildContractingRes(
	instType string,
	contractingInfoStr string,
	blkProducerPrivateKey *privacy.SpendingKey,
) ([]metadata.Transaction, error) {
	var contractingInfo ContractingInfo
	err := json.Unmarshal([]byte(contractingInfoStr), &contractingInfo)
	if err != nil {
		return nil, err
	}
	txReqID := contractingInfo.RequestedTxID
	if instType == "accepted" {
		return []metadata.Transaction{}, nil
	} else if instType == "refund" {
		meta := metadata.NewResponseBase(txReqID, metadata.ContractingReponseMeta)
		tx := new(transaction.Tx)
		err := tx.InitTxSalary(
			contractingInfo.BurnedConstAmount,
			&contractingInfo.BurnerAddress,
			blkProducerPrivateKey,
			blockgen.chain.config.DataBase,
			meta,
		)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		return []metadata.Transaction{tx}, nil
	}
	return []metadata.Transaction{}, nil
}

func (blockgen *BlkTmplGenerator) buildIssuingRes(
	instType string,
	issuingInfoStr string,
	blkProducerPrivateKey *privacy.SpendingKey,
) ([]metadata.Transaction, error) {
	var issuingInfo IssuingInfo
	err := json.Unmarshal([]byte(issuingInfoStr), &issuingInfo)
	if err != nil {
		return nil, err
	}
	txReqID := issuingInfo.RequestedTxID
	if instType != "accepted" {
		return []metadata.Transaction{}, nil
	}
	// accepted
	if bytes.Equal(issuingInfo.TokenID[:], common.ConstantID[:]) {
		meta := metadata.NewIssuingResponse(txReqID, metadata.IssuingResponseMeta)
		tx := new(transaction.Tx)
		err := tx.InitTxSalary(
			issuingInfo.Amount,
			&issuingInfo.ReceiverAddress,
			blkProducerPrivateKey,
			blockgen.chain.config.DataBase,
			meta,
		)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		return []metadata.Transaction{tx}, nil
	} else if bytes.Equal(issuingInfo.TokenID[:], common.DCBTokenID[:]) {
		meta := metadata.NewIssuingResponse(txReqID, metadata.IssuingResponseMeta)
		txTokenVout := transaction.TxTokenVout{
			Value:          issuingInfo.Amount,
			PaymentAddress: issuingInfo.ReceiverAddress,
		}
		var propertyID [common.HashSize]byte
		copy(propertyID[:], issuingInfo.TokenID[:])
		txTokenData := transaction.TxTokenData{
			Type:       transaction.CustomTokenInit,
			Mintable:   true,
			Amount:     issuingInfo.Amount,
			PropertyID: common.Hash(propertyID),
			Vins:       []transaction.TxTokenVin{},
			Vouts:      []transaction.TxTokenVout{txTokenVout},
		}
		txTokenData.PropertyName = txTokenData.PropertyID.String()
		txTokenData.PropertySymbol = txTokenData.PropertyID.String()
		resTx := &transaction.TxCustomToken{
			TxTokenData: txTokenData,
		}
		resTx.Type = common.TxCustomTokenType
		resTx.SetMetadata(meta)
		return []metadata.Transaction{resTx}, nil
	}
	return []metadata.Transaction{}, nil
}

func buildInstTypeAndAmountForIssuingAction(
	beaconBestState *BestStateBeacon,
	md *metadata.IssuingRequest,
) (string, uint64) {
	stabilityInfo := beaconBestState.StabilityInfo
	oracle := stabilityInfo.Oracle
	if bytes.Equal(md.AssetType[:], common.ConstantID[:]) {
		return "accepted", md.DepositedAmount / oracle.Constant
	}
	// process for DCB token case
	raiseReserveData := stabilityInfo.DCBConstitution.DCBParams.RaiseReserveData
	bestBlockHeight := beaconBestState.BestBlock.Header.Height
	if raiseReserveData == nil {
		return "refund", 0
	}

	reqAmt := uint64(0)
	var existed bool
	var reserveData *params.RaiseReserveData
	if bytes.Equal(md.CurrencyType[:], common.USDAssetID[:]) {
		reserveData, existed = raiseReserveData[common.USDAssetID]
		reqAmt = md.DepositedAmount / oracle.DCBToken
	} else if bytes.Equal(md.CurrencyType[:], common.ETHAssetID[:]) {
		reserveData, existed = raiseReserveData[common.ETHAssetID]
		// TODO: consider the unit of ETH
		reqAmt = (md.DepositedAmount * oracle.ETH) / oracle.DCBToken
	}
	if !existed ||
		bestBlockHeight+1 > reserveData.EndBlock ||
		reserveData.Amount == 0 ||
		reserveData.Amount < reqAmt {
		return "refund", 0
	}
	return "accepted", reqAmt
}

func buildInstructionsForIssuingReq(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return [][]string{}, err
	}
	var issuingReqAction IssuingReqAction
	err = json.Unmarshal(contentBytes, &issuingReqAction)
	if err != nil {
		return nil, err
	}
	md := issuingReqAction.Meta
	reqTxID := issuingReqAction.TxReqID
	instructions := [][]string{}
	instType, reqAmt := buildInstTypeAndAmountForIssuingAction(beaconBestState, &md)

	iInfo := IssuingInfo{
		ReceiverAddress: md.ReceiverAddress,
		Amount:          reqAmt,
		RequestedTxID:   reqTxID,
		TokenID:         md.AssetType,
	}
	iInfoBytes, err := json.Marshal(iInfo)
	if err != nil {
		return nil, err
	}
	returnedInst := []string{
		strconv.Itoa(metadata.IssuingRequestMeta),
		strconv.Itoa(int(shardID)),
		instType,
		string(iInfoBytes),
	}
	instructions = append(instructions, returnedInst)
	return instructions, nil
}

package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

type IssuingReqAction struct {
	TxReqID         common.Hash             `json:"txReqId"`
	ReceiverShardID byte                    `json:"receiverShardID"`
	Meta            metadata.IssuingRequest `json:"meta"`
}

type ContractingReqAction struct {
	TxReqID common.Hash                 `json:"txReqId"`
	Meta    metadata.ContractingRequest `json:"meta"`
}

func buildInstTypeAndAmountForContractingAction(
	beaconBestState *BestStateBeacon,
	md *metadata.ContractingRequest,
	accumulativeValues *accumulativeValues,
) (string, uint64) {
	stabilityInfo := beaconBestState.StabilityInfo
	oracle := stabilityInfo.Oracle
	if bytes.Equal(md.CurrencyType[:], common.USDAssetID[:]) {
		redeemAmount := md.BurnedConstAmount * oracle.Constant
		return "accepted", redeemAmount
	}
	// crypto
	spendReserveData := stabilityInfo.DCBConstitution.DCBParams.SpendReserveData
	bestBlockHeight := beaconBestState.BestBlock.Header.Height
	fmt.Printf("[db] buildInstTypeForCont spendReserveData: %+v\n", spendReserveData)
	if spendReserveData == nil {
		return "refund", 0
	}
	reserveData, existed := spendReserveData[md.CurrencyType]
	if !existed {
		return "refund", 0
	}
	if bestBlockHeight+1 > reserveData.EndBlock ||
		md.BurnedConstAmount+accumulativeValues.constantsBurnedByETH > reserveData.Amount {
		return "refund", 0
	}
	if bytes.Equal(md.CurrencyType[:], common.ETHAssetID[:]) &&
		oracle.ETH < reserveData.ReserveMinPrice {
		return "refund", 0
	}
	// redeemAmount := md.BurnedConstAmount * oracle.Constant / oracle.ETH
	cstValue := md.BurnedConstAmount * oracle.Constant / 100
	redeemAmount := common.CentInMilliEther(cstValue, oracle.ETH)
	accumulativeValues.constantsBurnedByETH += md.BurnedConstAmount
	return "accepted", redeemAmount
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
	fmt.Printf("[db] building inst for contracting req: %s\n", contentBytes)
	var contractingReqAction ContractingReqAction
	err = json.Unmarshal(contentBytes, &contractingReqAction)
	if err != nil {
		return nil, err
	}
	md := contractingReqAction.Meta
	reqTxID := contractingReqAction.TxReqID
	instructions := [][]string{}
	instType, redeemAmount := buildInstTypeAndAmountForContractingAction(beaconBestState, &md, accumulativeValues)

	cInfo := component.ContractingInfo{
		BurnerAddress:     md.BurnerAddress,
		BurnedConstAmount: md.BurnedConstAmount,
		RedeemAmount:      redeemAmount,
		RequestedTxID:     reqTxID,
		CurrencyType:      md.CurrencyType,
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
	fmt.Printf("[db] buildInstForContReq return %+v\n", returnedInst)
	instructions = append(instructions, returnedInst)
	return instructions, nil
}

func (blockgen *BlkTmplGenerator) buildContractingRes(
	instType string,
	contractingInfoStr string,
	blkProducerPrivateKey *privacy.PrivateKey,
) ([]metadata.Transaction, error) {
	fmt.Printf("[db] buildContractingRes: %s\n", contractingInfoStr)
	var contractingInfo component.ContractingInfo
	err := json.Unmarshal([]byte(contractingInfoStr), &contractingInfo)
	if err != nil {
		return nil, err
	}
	txReqID := contractingInfo.RequestedTxID
	if instType == "accepted" {
		return []metadata.Transaction{}, nil
	} else if instType == "refund" {
		meta := metadata.NewContractingResponse(txReqID, metadata.ContractingResponseMeta)
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
	blkProducerPrivateKey *privacy.PrivateKey,
	shardID byte,
) ([]metadata.Transaction, error) {
	var issuingInfo component.IssuingInfo
	fmt.Printf("[db] buildIssuingRes %s\n", issuingInfoStr)
	err := json.Unmarshal([]byte(issuingInfoStr), &issuingInfo)
	if err != nil {
		return nil, err
	}
	txReqID := issuingInfo.RequestedTxID
	if instType != "accepted" {
		return []metadata.Transaction{}, nil
	}

	db := blockgen.chain.config.DataBase
	if bytes.Equal(issuingInfo.TokenID[:], common.ConstantID[:]) { // accepted
		meta := metadata.NewIssuingResponse(txReqID, metadata.IssuingResponseMeta)
		tx := new(transaction.Tx)
		err := tx.InitTxSalary(
			issuingInfo.Amount,
			&issuingInfo.ReceiverAddress,
			blkProducerPrivateKey,
			db,
			meta,
		)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		return []metadata.Transaction{tx}, nil
	} else if bytes.Equal(issuingInfo.TokenID[:], common.DCBTokenID[:]) {
		meta := metadata.NewIssuingResponse(txReqID, metadata.IssuingResponseMeta)
		txCustom := &transaction.TxCustomToken{}
		customTokenParamTx := &transaction.CustomTokenParamTx{
			PropertyID:  common.DCBTokenID.String(),
			TokenTxType: transaction.CustomTokenMint,
			Amount:      issuingInfo.Amount,
			Receiver: []transaction.TxTokenVout{
				transaction.TxTokenVout{
					Value:          issuingInfo.Amount,
					PaymentAddress: issuingInfo.ReceiverAddress,
				},
			},
		}
		err = txCustom.Init(
			blkProducerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			customTokenParamTx,
			db,
			meta,
			false,
			shardID,
		)
		if err.(*transaction.TransactionError) != nil {
			fmt.Printf("[db] build issuing resp err: %v\n", err)
			return nil, err
		}
		fmt.Printf("[db] build issuing resp success: %h\n", txCustom.Hash())
		return []metadata.Transaction{txCustom}, nil
	}
	// TODO(@0xbunyip): fail to issue/refund
	return []metadata.Transaction{}, nil
}

func buildInstTypeAndAmountForIssuingAction(
	beaconBestState *BestStateBeacon,
	md *metadata.IssuingRequest,
	accumulativeValues *accumulativeValues,
) (string, uint64) {
	stabilityInfo := beaconBestState.StabilityInfo
	oracle := stabilityInfo.Oracle
	if bytes.Equal(md.AssetType[:], common.ConstantID[:]) {
		return "accepted", (md.DepositedAmount * 100) / oracle.Constant
	}
	// process for case of DCB token
	raiseReserveData := stabilityInfo.DCBConstitution.DCBParams.RaiseReserveData
	bestBlockHeight := beaconBestState.BestBlock.Header.Height
	if raiseReserveData == nil {
		return "refund", 0
	}

	dcbTokensNeeded := uint64(0)
	reqAmt := uint64(0)
	var existed bool
	var reserveData *component.RaiseReserveData
	isOnUSD := bytes.Equal(md.CurrencyType[:], common.USDAssetID[:])
	isOnETH := bytes.Equal(md.CurrencyType[:], common.ETHAssetID[:])
	if isOnUSD {
		reserveData, existed = raiseReserveData[common.USDAssetID]
		reqAmt = md.DepositedAmount / oracle.DCBToken
		dcbTokensNeeded = reqAmt + accumulativeValues.dcbTokensSoldByUSD
	} else if isOnETH {
		reserveData, existed = raiseReserveData[common.ETHAssetID]
		reqAmt = common.MilliEtherValue(md.DepositedAmount, oracle.ETH) / oracle.DCBToken
		// reqAmt = (md.DepositedAmount * oracle.ETH) / oracle.DCBToken
		dcbTokensNeeded = reqAmt + accumulativeValues.dcbTokensSoldByETH
		fmt.Printf("[db] isOnETH: %+v %d %d %d\n", reserveData, reqAmt, dcbTokensNeeded, bestBlockHeight)
	}
	if !existed ||
		bestBlockHeight+1 > reserveData.EndBlock ||
		reserveData.Amount == 0 ||
		reserveData.Amount < dcbTokensNeeded {
		return "refund", 0
	}
	if isOnUSD {
		accumulativeValues.dcbTokensSoldByUSD += reqAmt
	} else if isOnETH {
		accumulativeValues.dcbTokensSoldByETH += reqAmt
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
	fmt.Printf("[db] building inst for issuing req: %s\n", contentBytes)
	err = json.Unmarshal(contentBytes, &issuingReqAction)
	if err != nil {
		return nil, err
	}
	md := issuingReqAction.Meta
	reqTxID := issuingReqAction.TxReqID
	instructions := [][]string{}
	instType, reqAmt := buildInstTypeAndAmountForIssuingAction(beaconBestState, &md, accumulativeValues)

	iInfo := component.IssuingInfo{
		ReceiverAddress: md.ReceiverAddress,
		Amount:          reqAmt,
		RequestedTxID:   reqTxID,
		TokenID:         md.AssetType,
		CurrencyType:    md.CurrencyType,
	}
	iInfoBytes, err := json.Marshal(iInfo)
	if err != nil {
		return nil, err
	}
	returnedInst := []string{
		strconv.Itoa(metadata.IssuingRequestMeta),
		strconv.Itoa(int(issuingReqAction.ReceiverShardID)),
		instType,
		string(iInfoBytes),
	}
	fmt.Printf("[db] buildInstForIssuingReq return %+v\n", returnedInst)
	instructions = append(instructions, returnedInst)
	return instructions, nil
}

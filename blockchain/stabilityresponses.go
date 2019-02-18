package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/ninjadotorg/constant/metadata/toshardins"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/pkg/errors"
)

type BuyBackInfo struct {
	PaymentAddress privacy.PaymentAddress
	BuyBackPrice   uint64
	Value          uint64
	RequestedTxID  common.Hash
	TokenID        common.Hash
}

type IssuingInfo struct {
	ReceiverAddress privacy.PaymentAddress
	Amount          uint64
	RequestedTxID   common.Hash
	TokenID         common.Hash
}

type BuyGOVTokenReqAction struct {
	TxReqID common.Hash                 `json:"txReqId"`
	Meta    metadata.BuyGOVTokenRequest `json:"meta"`
}

type BuySellReqAction struct {
	TxReqID common.Hash             `json:"txReqId"`
	Meta    metadata.BuySellRequest `json:"meta"`
}

type IssuingReqAction struct {
	TxReqID common.Hash             `json:"txReqId"`
	Meta    metadata.IssuingRequest `json:"meta"`
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

func buildInstructionsForBuyBackBondsReq(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
	bc *BlockChain,
) ([][]string, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return [][]string{}, err
	}
	var buyBackReqTx transaction.TxCustomToken
	err = json.Unmarshal(contentBytes, &buyBackReqTx)
	if err != nil {
		return nil, err
	}
	meta := buyBackReqTx.GetMetadata()
	buyBackReqMeta, ok := meta.(*metadata.BuyBackRequest)
	if !ok {
		return nil, errors.New("Could not parse BuyBackRequest metadata.")
	}
	vins := buyBackReqTx.TxTokenData.Vins
	if len(vins) == 0 {
		return nil, errors.New("No existed Vins from BuyBackRequest tx")
	}
	priorTxID := vins[0].TxCustomTokenID
	_, _, _, priorTx, err := bc.GetTransactionByHash(&priorTxID)
	if err != nil {
		return nil, err
	}
	priorCustomTokenTx, ok := priorTx.(*transaction.TxCustomToken)
	if !ok {
		return nil, errors.New("Could not parse prior TxCustomToken.")
	}
	priorMeta := priorCustomTokenTx.GetMetadata()
	if priorMeta == nil {
		return nil, errors.New("Not existed metadata in priorCustomTokenTx")
	}
	buySellResMeta, ok := priorMeta.(*metadata.BuySellResponse)
	if !ok {
		return nil, errors.New("Could not parse BuySellResponse metadata.")
	}

	instType := ""
	bestBlockHeight := beaconBestState.BestBlock.Header.Height
	if (buySellResMeta.StartSellingAt+buySellResMeta.Maturity > bestBlockHeight+1) ||
		(buyBackReqMeta.Amount*buySellResMeta.BuyBackPrice > beaconBestState.StabilityInfo.SalaryFund) {
		instType = "refund"
	} else {
		instType = "accepted"
		beaconBestState.StabilityInfo.SalaryFund -= buyBackReqMeta.Amount * buySellResMeta.BuyBackPrice
	}

	buyBackInfo := BuyBackInfo{
		PaymentAddress: buyBackReqMeta.PaymentAddress,
		BuyBackPrice:   buySellResMeta.BuyBackPrice,
		Value:          buyBackReqMeta.Amount,
		RequestedTxID:  *(buyBackReqTx.Hash()),
		TokenID:        buyBackReqMeta.TokenID,
	}
	buyBackInfoBytes, err := json.Marshal(buyBackInfo)
	if err != nil {
		return nil, err
	}
	returnedInst := []string{
		strconv.Itoa(metadata.BuyBackRequestMeta),
		strconv.Itoa(int(shardID)),
		instType,
		string(buyBackInfoBytes),
	}
	return [][]string{returnedInst}, nil
}

func buildInstructionsForBuyBondsFromGOVReq(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return [][]string{}, err
	}
	var buySellReqAction BuySellReqAction
	err = json.Unmarshal(contentBytes, &buySellReqAction)
	if err != nil {
		return nil, err
	}
	md := buySellReqAction.Meta
	instructions := [][]string{}
	stabilityInfo := beaconBestState.StabilityInfo
	sellingBondsParams := stabilityInfo.GOVConstitution.GOVParams.SellingBonds
	bestBlockHeight := beaconBestState.BestBlock.Header.Height
	instType := ""
	if (sellingBondsParams == nil) ||
		(bestBlockHeight+1 < sellingBondsParams.StartSellingAt) ||
		(bestBlockHeight+1 > sellingBondsParams.StartSellingAt+sellingBondsParams.SellingWithin) ||
		(accumulativeValues.bondsSold+md.Amount > sellingBondsParams.BondsToSell) {
		instType = "refund"
	} else {
		accumulativeValues.incomeFromBonds += (md.Amount + md.BuyPrice)
		accumulativeValues.bondsSold += md.Amount
		instType = "accepted"
	}
	sellingBondsParamsBytes, err := json.Marshal(sellingBondsParams)
	if err != nil {
		return nil, err
	}
	returnedInst := []string{
		strconv.Itoa(metadata.BuyFromGOVRequestMeta),
		strconv.Itoa(int(shardID)),
		instType,
		contentStr,
		string(sellingBondsParamsBytes),
	}
	instructions = append(instructions, returnedInst)
	return instructions, nil
}

func buildInstructionsForBuyGOVTokensReq(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return [][]string{}, err
	}
	var buyGOVTokenReqAction BuyGOVTokenReqAction
	err = json.Unmarshal(contentBytes, &buyGOVTokenReqAction)
	if err != nil {
		return nil, err
	}
	md := buyGOVTokenReqAction.Meta
	instructions := [][]string{}
	stabilityInfo := beaconBestState.StabilityInfo
	sellingGOVTokensParams := stabilityInfo.GOVConstitution.GOVParams.SellingGOVTokens
	bestBlockHeight := beaconBestState.BestBlock.Header.Height
	instType := ""
	if (sellingGOVTokensParams == nil) ||
		(bestBlockHeight+1 < sellingGOVTokensParams.StartSellingAt) ||
		(bestBlockHeight+1 > sellingGOVTokensParams.StartSellingAt+sellingGOVTokensParams.SellingWithin) ||
		(accumulativeValues.govTokensSold+md.Amount > sellingGOVTokensParams.GOVTokensToSell) {
		instType = "refund"
	} else {
		accumulativeValues.incomeFromGOVTokens += (md.Amount + md.BuyPrice)
		accumulativeValues.govTokensSold += md.Amount
		instType = "accepted"
	}
	returnedInst := []string{
		strconv.Itoa(metadata.BuyGOVTokenRequestMeta),
		strconv.Itoa(int(shardID)),
		instType,
		contentStr,
	}
	instructions = append(instructions, returnedInst)
	return instructions, nil
}

// build instructions at beacon chain before syncing to shards
func (blkTmpGen *BlkTmplGenerator) buildStabilityInstructions(
	shardID byte,
	shardBlockInstructions [][]string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	instructions := [][]string{}
	for _, inst := range shardBlockInstructions {
		// TODO: will improve the condition later
		if inst[0] == "stake" || inst[0] == "swap" || inst[0] == "random" {
			continue
		}
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			return [][]string{}, err
		}
		contentStr := inst[1]
		switch metaType {
		case metadata.BuyFromGOVRequestMeta:
			buyBondsInst, err := buildInstructionsForBuyBondsFromGOVReq(shardID, contentStr, beaconBestState, accumulativeValues)
			if err != nil {
				return [][]string{}, err
			}
			instructions = append(instructions, buyBondsInst...)

		case metadata.BuyGOVTokenRequestMeta:
			buyGOVTokensInst, err := buildInstructionsForBuyGOVTokensReq(shardID, contentStr, beaconBestState, accumulativeValues)
			if err != nil {
				return [][]string{}, err
			}
			instructions = append(instructions, buyGOVTokensInst...)

		case metadata.CrowdsaleRequestMeta:
			saleInst, err := buildInstructionsForCrowdsaleRequest(shardID, contentStr, beaconBestState, accumulativeValues)
			if err != nil {
				return [][]string{}, err
			}
			instructions = append(instructions, saleInst...)

		case metadata.BuyBackRequestMeta:
			buyBackInst, err := buildInstructionsForBuyBackBondsReq(shardID, contentStr, beaconBestState, blkTmpGen.chain)
			if err != nil {
				return [][]string{}, err
			}
			instructions = append(instructions, buyBackInst...)

		case metadata.IssuingRequestMeta:
			issuingInst, err := buildInstructionsForIssuingReq(shardID, contentStr, beaconBestState, accumulativeValues)
			if err != nil {
				return [][]string{}, err
			}
			instructions = append(instructions, issuingInst...)

		default:
			continue
		}
	}
	// update params in beststate
	updateParamsFromBeaconBestState(beaconBestState, accumulativeValues)
	return instructions, nil
}

func updateParamsFromBeaconBestState(
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) {
	beaconBestState.StabilityInfo.SalaryFund += (accumulativeValues.incomeFromBonds + accumulativeValues.incomeFromGOVTokens)
	if beaconBestState.StabilityInfo.GOVConstitution.GOVParams.SellingBonds != nil {
		beaconBestState.StabilityInfo.GOVConstitution.GOVParams.SellingBonds.BondsToSell -= accumulativeValues.bondsSold
	}
	if beaconBestState.StabilityInfo.GOVConstitution.GOVParams.SellingGOVTokens != nil {
		beaconBestState.StabilityInfo.GOVConstitution.GOVParams.SellingGOVTokens.GOVTokensToSell -= accumulativeValues.govTokensSold
	}

	// reset gov values
	accumulativeValues.govTokensSold = 0
	accumulativeValues.bondsSold = 0
	accumulativeValues.incomeFromBonds = 0
	accumulativeValues.incomeFromGOVTokens = 0
}

func (blockgen *BlkTmplGenerator) buildLoanResponseTx(tx metadata.Transaction, producerPrivateKey *privacy.SpendingKey) (metadata.Transaction, error) {
	// Get loan request
	withdrawMeta := tx.GetMetadata().(*metadata.LoanWithdraw)
	meta, err := blockgen.chain.GetLoanRequestMeta(withdrawMeta.LoanID)
	if err != nil {
		return nil, err
	}

	// Build loan unlock tx
	unlockMeta := &metadata.LoanUnlock{
		LoanID:       make([]byte, len(withdrawMeta.LoanID)),
		MetadataBase: metadata.MetadataBase{Type: metadata.LoanUnlockMeta},
	}
	copy(unlockMeta.LoanID, withdrawMeta.LoanID)
	unlockMetaList := []metadata.Metadata{unlockMeta}
	amounts := []uint64{meta.LoanAmount}
	txNormals, err := transaction.BuildCoinbaseTxs([]*privacy.PaymentAddress{meta.ReceiveAddress}, amounts, producerPrivateKey, blockgen.chain.GetDatabase(), unlockMetaList)
	if err != nil {
		return nil, errors.Errorf("Error building unlock tx for loan id %x", withdrawMeta.LoanID)
	}
	return txNormals[0], nil
}

func (blockgen *BlkTmplGenerator) buildBuyBackRes(
	instType string,
	buyBackInfoStr string,
	blkProducerPrivateKey *privacy.SpendingKey,
) ([]metadata.Transaction, error) {
	var buyBackInfo BuyBackInfo
	err := json.Unmarshal([]byte(buyBackInfoStr), &buyBackInfo)
	if err != nil {
		return nil, err
	}

	if instType == "refund" {
		bondID := buyBackInfo.TokenID
		buyBackRes := metadata.NewResponseBase(buyBackInfo.RequestedTxID, metadata.ResponseBaseMeta)
		txTokenVout := transaction.TxTokenVout{
			Value:          buyBackInfo.Value,
			PaymentAddress: buyBackInfo.PaymentAddress,
		}
		var propertyID [common.HashSize]byte
		copy(propertyID[:], bondID[:])
		txTokenData := transaction.TxTokenData{
			Type:       transaction.CustomTokenInit,
			Mintable:   true,
			Amount:     buyBackInfo.Value,
			PropertyID: common.Hash(propertyID),
			Vins:       []transaction.TxTokenVin{},
			Vouts:      []transaction.TxTokenVout{txTokenVout},
		}
		txTokenData.PropertyName = txTokenData.PropertyID.String()
		txTokenData.PropertySymbol = txTokenData.PropertyID.String()

		refundTx := &transaction.TxCustomToken{
			TxTokenData: txTokenData,
		}
		refundTx.Type = common.TxCustomTokenType
		refundTx.SetMetadata(buyBackRes)
		return []metadata.Transaction{refundTx}, nil

	} else if instType == "accepted" {
		buyBackAmount := buyBackInfo.Value * buyBackInfo.BuyBackPrice
		buyBackRes := metadata.NewBuyBackResponse(buyBackInfo.RequestedTxID, metadata.BuyBackResponseMeta)
		buyBackResTx := new(transaction.Tx)
		err := buyBackResTx.InitTxSalary(
			buyBackAmount,
			&buyBackInfo.PaymentAddress,
			blkProducerPrivateKey,
			blockgen.chain.GetDatabase(),
			buyBackRes,
		)
		if err != nil {
			return nil, err
		}
		return []metadata.Transaction{buyBackResTx}, nil
	}
	return nil, nil
}

func (blockgen *BlkTmplGenerator) buildBuyGOVTokensRes(
	instType string,
	contentStr string,
	blkProducerPrivateKey *privacy.SpendingKey,
) ([]metadata.Transaction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return nil, err
	}
	var buyGOVTokenReqAction BuyGOVTokenReqAction
	err = json.Unmarshal(contentBytes, &buyGOVTokenReqAction)
	if err != nil {
		return nil, err
	}
	txReqID := buyGOVTokenReqAction.TxReqID
	reqMeta := buyGOVTokenReqAction.Meta
	if instType == "refund" {
		refundMeta := metadata.NewResponseBase(txReqID, metadata.ResponseBaseMeta)
		refundTx := new(transaction.Tx)
		err := refundTx.InitTxSalary(
			reqMeta.Amount*reqMeta.BuyPrice,
			&reqMeta.BuyerAddress,
			blkProducerPrivateKey,
			blockgen.chain.config.DataBase,
			refundMeta,
		)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		return []metadata.Transaction{refundTx}, nil

	} else if instType == "accepted" {
		govTokenID := reqMeta.TokenID
		buyGOVTokensRes := metadata.NewResponseBase(txReqID, metadata.ResponseBaseMeta)
		txTokenVout := transaction.TxTokenVout{
			Value:          reqMeta.Amount,
			PaymentAddress: reqMeta.BuyerAddress,
		}
		var propertyID [common.HashSize]byte
		copy(propertyID[:], govTokenID[:])
		txTokenData := transaction.TxTokenData{
			Type:       transaction.CustomTokenInit,
			Mintable:   true,
			Amount:     reqMeta.Amount,
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
		resTx.SetMetadata(buyGOVTokensRes)
		return []metadata.Transaction{resTx}, nil
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
	if instType == "accepted" {
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
	}
	return []metadata.Transaction{}, nil
}

func (blockgen *BlkTmplGenerator) buildBuyBondsFromGOVRes(
	instType string,
	contentStr string,
	sellingBondsParamsStr string,
	blkProducerPrivateKey *privacy.SpendingKey,
) ([]metadata.Transaction, error) {
	sellingBondsParamsBytes := []byte(sellingBondsParamsStr)
	var sellingBondsParams params.SellingBonds
	err := json.Unmarshal(sellingBondsParamsBytes, &sellingBondsParams)
	if err != nil {
		return nil, err
	}

	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return nil, err
	}
	var buySellReqAction BuySellReqAction
	err = json.Unmarshal(contentBytes, &buySellReqAction)
	if err != nil {
		return nil, err
	}
	txReqID := buySellReqAction.TxReqID
	reqMeta := buySellReqAction.Meta
	if instType == "refund" {
		refundMeta := metadata.NewResponseBase(txReqID, metadata.ResponseBaseMeta)
		refundTx := new(transaction.Tx)
		err := refundTx.InitTxSalary(
			reqMeta.Amount*reqMeta.BuyPrice,
			&reqMeta.PaymentAddress,
			blkProducerPrivateKey,
			blockgen.chain.config.DataBase,
			refundMeta,
		)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		return []metadata.Transaction{refundTx}, nil
	} else if instType == "accepted" {
		bondID := reqMeta.TokenID
		buySellRes := metadata.NewBuySellResponse(
			txReqID,
			sellingBondsParams.StartSellingAt,
			sellingBondsParams.Maturity,
			sellingBondsParams.BuyBackPrice,
			bondID[:],
			metadata.BuyFromGOVResponseMeta,
		)
		txTokenVout := transaction.TxTokenVout{
			Value:          reqMeta.Amount,
			PaymentAddress: reqMeta.PaymentAddress,
		}
		var propertyID [common.HashSize]byte
		copy(propertyID[:], bondID[:])
		txTokenData := transaction.TxTokenData{
			Type:       transaction.CustomTokenInit,
			Mintable:   true,
			Amount:     reqMeta.Amount,
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
		resTx.SetMetadata(buySellRes)
		return []metadata.Transaction{resTx}, nil
	}
	return []metadata.Transaction{}, nil
}

func (blockgen *BlkTmplGenerator) buildStabilityResponseTxsFromInstructions(
	beaconBlocks []*BeaconBlock,
	producerPrivateKey *privacy.SpendingKey,
	shardID byte,
) ([]metadata.Transaction, error) {
	// TODO(@0xbunyip): refund bonds in multiple blocks since many refund instructions might come at once and UTXO picking order is not perfect
	unspentTokenMap := map[string]([]transaction.TxTokenVout){}
	resTxs := []metadata.Transaction{}
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			// TODO: will improve the condition later
			if l[0] == "stake" || l[0] == "swap" || l[0] == "random" {
				continue
			}
			continue
			if len(l) <= 2 {
			}
			shardToProcess, err := strconv.Atoi(l[1])
			if err == nil && shardToProcess == int(shardID) {
				metaType, err := strconv.Atoi(l[0])
				if err != nil {
					return nil, err
				}
				switch metaType {
				case metadata.CrowdsalePaymentMeta:
					paymentInst, err := ParseCrowdsalePaymentInstruction(l[2])
					if err != nil {
						return nil, err
					}
					tx, err := blockgen.buildPaymentForCrowdsale(paymentInst, unspentTokenMap, producerPrivateKey)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, tx)

				case metadata.BuyFromGOVRequestMeta:
					contentStr := l[3]
					sellingBondsParamsStr := l[4]
					txs, err := blockgen.buildBuyBondsFromGOVRes(l[2], contentStr, sellingBondsParamsStr, producerPrivateKey)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, txs...)

				case metadata.BuyGOVTokenRequestMeta:
					contentStr := l[3]
					txs, err := blockgen.buildBuyGOVTokensRes(l[2], contentStr, producerPrivateKey)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, txs...)

				case metadata.BuyBackRequestMeta:
					buyBackInfoStr := l[3]
					txs, err := blockgen.buildBuyBackRes(l[2], buyBackInfoStr, producerPrivateKey)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, txs...)

				case metadata.AcceptDCBBoardMeta:
					acceptDCBBoardIns := toshardins.TxAcceptDCBBoardIns{}
					err := json.Unmarshal([]byte(l[2]), &acceptDCBBoardIns)
					if err != nil {
						return nil, err
					}
					txs := acceptDCBBoardIns.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase)
					resTxs = append(resTxs, txs)
				case metadata.AcceptGOVBoardMeta:
					acceptGOVBoardIns := toshardins.TxAcceptGOVBoardIns{}
					err := json.Unmarshal([]byte(l[2]), &acceptGOVBoardIns)
					if err != nil {
						return nil, err
					}
					txs := acceptGOVBoardIns.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase)
					resTxs = append(resTxs, txs)
				case metadata.SendBackTokenVoteFailMeta:
					sendBackTokenVoteFail := toshardins.TxSendBackTokenVoteFailIns{}
					err := json.Unmarshal([]byte(l[2]), &sendBackTokenVoteFail)
					if err != nil {
						return nil, err
					}
					txs := sendBackTokenVoteFail.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase)
					resTxs = append(resTxs, txs)
				case metadata.SendInitDCBVoteTokenMeta:
					sendInitDCBVoteToken := toshardins.TxSendInitDCBVoteTokenMetadataIns{}
					err := json.Unmarshal([]byte(l[2]), &sendInitDCBVoteToken)
					if err != nil {
						return nil, err
					}
					txs := sendInitDCBVoteToken.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase)
					resTxs = append(resTxs, txs)
				case metadata.SendInitGOVVoteTokenMeta:
					sendInitGOVVoteToken := toshardins.TxSendInitGOVVoteTokenMetadataIns{}
					err := json.Unmarshal([]byte(l[2]), &sendInitGOVVoteToken)
					if err != nil {
						return nil, err
					}
					txs := sendInitGOVVoteToken.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase)
					resTxs = append(resTxs, txs)
				case metadata.ShareRewardOldDCBBoardMeta, metadata.ShareRewardOldGOVBoardMeta:
					shareRewardOldBoard := toshardins.TxShareRewardOldBoardMetadataIns{}
					err := json.Unmarshal([]byte(l[2]), &shareRewardOldBoard)
					if err != nil {
						return nil, err
					}
					txs := shareRewardOldBoard.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase)
					resTxs = append(resTxs, txs)
				case metadata.IssuingRequestMeta:
					issuingInfoStr := l[3]
					txs, err := blockgen.buildIssuingRes(l[2], issuingInfoStr, producerPrivateKey)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, txs...)
				}
			}
		}
	}
	return resTxs, nil
}

func (blockgen *BlkTmplGenerator) buildStabilityResponseTxsAtShardOnly(txs []metadata.Transaction, producerPrivateKey *privacy.SpendingKey) ([]metadata.Transaction, error) {
	respTxs := []metadata.Transaction{}
	removeIds := []int{}
	for i, tx := range txs {
		var respTx metadata.Transaction
		var err error

		switch tx.GetMetadataType() {
		case metadata.LoanWithdrawMeta:
			respTx, err = blockgen.buildLoanResponseTx(tx, producerPrivateKey)
		}

		if err != nil {
			// Remove this tx if cannot create corresponding response
			removeIds = append(removeIds, i)
		} else if respTx != nil {
			respTxs = append(respTxs, respTx)
		}
	}

	// TODO(@0xbunyip): remove tx from txsToAdd?
	return respTxs, nil
}

// func (blockgen *BlkTmplGenerator) buildIssuingResTxs(
// 	chainID byte,
// 	issuingReqTxs []metadata.Transaction,
// 	privatekey *privacy.SpendingKey,
// ) ([]metadata.Transaction, error) {
// 	prevBlock := blockgen.chain.BestState[chainID].BestBlock
// 	oracleParams := prevBlock.Header.Oracle

// 	issuingResTxs := []metadata.Transaction{}
// 	for _, issuingReqTx := range issuingReqTxs {
// 		meta := issuingReqTx.GetMetadata()
// 		issuingReq, ok := meta.(*metadata.IssuingRequest)
// 		if !ok {
// 			return []metadata.Transaction{}, errors.New("Could not parse IssuingRequest metadata.")
// 		}
// 		if issuingReq.AssetType == common.DCBTokenID {
// 			issuingRes := metadata.NewIssuingResponse(*issuingReqTx.Hash(), metadata.IssuingResponseMeta)
// 			dcbTokenPrice := uint64(1)
// 			if oracleParams.DCBToken != 0 {
// 				dcbTokenPrice = oracleParams.DCBToken
// 			}
// 			issuingAmt := issuingReq.DepositedAmount / dcbTokenPrice
// 			txTokenVout := transaction.TxTokenVout{
// 				Value:          issuingAmt,
// 				PaymentAddress: issuingReq.ReceiverAddress,
// 			}
// 			txTokenData := transaction.TxTokenData{
// 				Type:       transaction.CustomTokenInit,
// 				Amount:     issuingAmt,
// 				PropertyID: common.Hash(common.DCBTokenID),
// 				Vins:       []transaction.TxTokenVin{},
// 				Vouts:      []transaction.TxTokenVout{txTokenVout},
// 				// PropertyName:   "",
// 				// PropertySymbol: coinbaseTxType,
// 			}
// 			resTx := &transaction.TxCustomToken{
// 				TxTokenData: txTokenData,
// 			}
// 			resTx.Type = common.TxCustomTokenType
// 			resTx.SetMetadata(issuingRes)
// 			issuingResTxs = append(issuingResTxs, resTx)
// 			continue
// 		}
// 		if issuingReq.AssetType == common.ConstantID {
// 			constantPrice := uint64(1)
// 			if oracleParams.Constant != 0 {
// 				constantPrice = oracleParams.Constant
// 			}
// 			issuingAmt := issuingReq.DepositedAmount / constantPrice
// 			issuingRes := metadata.NewIssuingResponse(*issuingReqTx.Hash(), metadata.IssuingResponseMeta)
// 			resTx := new(transaction.Tx)
// 			err := resTx.InitTxSalary(issuingAmt, &issuingReq.ReceiverAddress, privatekey, blockgen.chain.GetDatabase(), issuingRes)
// 			if err != nil {
// 				return []metadata.Transaction{}, err
// 			}
// 			issuingResTxs = append(issuingResTxs, resTx)
// 		}
// 	}
// 	return issuingResTxs, nil
// }

// func calculateAmountOfRefundTxs(
// 	smallTxHashes []*common.Hash,
// 	addresses []*privacy.PaymentAddress,
// 	estimatedRefundAmt uint64,
// 	remainingFund uint64,
// 	db database.DatabaseInterface,
// 	privatekey *privacy.SpendingKey,
// ) ([]metadata.Transaction, uint64) {
// 	amt := uint64(0)
// 	if estimatedRefundAmt <= remainingFund {
// 		amt = estimatedRefundAmt
// 	} else {
// 		amt = remainingFund
// 	}
// 	actualRefundAmt := amt / uint64(len(addresses))
// 	var refundTxs []metadata.Transaction
// 	for i := 0; i < len(addresses); i++ {
// 		addr := addresses[i]
// 		refundMeta := metadata.NewRefund(*smallTxHashes[i], metadata.RefundMeta)
// 		refundTx := new(transaction.Tx)
// 		err := refundTx.InitTxSalary(actualRefundAmt, addr, privatekey, db, refundMeta)
// 		if err != nil {
// 			Logger.log.Error(err)
// 			continue
// 		}
// 		refundTxs = append(refundTxs, refundTx)
// 	}
// 	return refundTxs, amt
// }

// func (blockgen *BlkTmplGenerator) buildRefundTxs(
// 	chainID byte,
// 	remainingFund uint64,
// 	privatekey *privacy.SpendingKey,
// ) ([]metadata.Transaction, uint64) {
// 	if remainingFund <= 0 {
// 		Logger.log.Info("GOV fund is not enough for refund.")
// 		return []metadata.Transaction{}, 0
// 	}
// 	prevBlock := blockgen.chain.BestState[chainID].BestBlock
// 	header := prevBlock.Header
// 	govParams := header.GOVConstitution.GOVParams
// 	refundInfo := govParams.RefundInfo
// 	if refundInfo == nil {
// 		Logger.log.Info("Refund info is not existed.")
// 		return []metadata.Transaction{}, 0
// 	}
// 	lookbackBlockHeight := header.Height - common.RefundPeriod
// 	if lookbackBlockHeight < 0 {
// 		return []metadata.Transaction{}, 0
// 	}
// 	lookbackBlock, err := blockgen.chain.GetBlockByBlockHeight(lookbackBlockHeight, chainID)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return []metadata.Transaction{}, 0
// 	}
// 	addresses := []*privacy.PaymentAddress{}
// 	smallTxHashes := []*common.Hash{}
// 	estimatedRefundAmt := uint64(0)
// 	for _, tx := range lookbackBlock.Transactions {
// 		if tx.GetType() != common.TxNormalType {
// 			continue
// 		}
// 		lookbackTx, ok := tx.(*transaction.Tx)
// 		if !ok {
// 			continue
// 		}
// 		addr := lookbackTx.GetSenderAddress()
// 		if addr == nil {
// 			continue
// 		}
// 		txValue := lookbackTx.CalculateTxValue()
// 		if txValue > refundInfo.ThresholdToLargeTx {
// 			continue
// 		}
// 		addresses = append(addresses, addr)
// 		smallTxHashes = append(smallTxHashes, tx.Hash())
// 		estimatedRefundAmt += refundInfo.RefundAmount
// 	}
// 	if len(addresses) == 0 {
// 		return []metadata.Transaction{}, 0
// 	}
// 	refundTxs, totalRefundAmt := calculateAmountOfRefundTxs(
// 		smallTxHashes,
// 		addresses,
// 		estimatedRefundAmt,
// 		remainingFund,
// 		blockgen.chain.GetDatabase(),
// 		privatekey,
// 	)
// 	return refundTxs, totalRefundAmt
// }

// func (blockgen *BlkTmplGenerator) processLoan(sourceTxns []*metadata.TxDesc, producerPrivateKey *privacy.SpendingKey) (uint64, []metadata.Transaction, []metadata.Transaction) {
// 	amount := uint64(0)
// 	loanUnlockTxs := []metadata.Transaction{}
// 	removableTxs := []metadata.Transaction{}
// 	for _, txDesc := range sourceTxns {
// 		if txDesc.Tx.GetMetadataType() == metadata.LoanPaymentMeta {
// 			paymentAmount, err := blockgen.calculateInterestPaid(txDesc.Tx)
// 			if err != nil {
// 				removableTxs = append(removableTxs, txDesc.Tx)
// 				continue
// 			}
// 			amount += paymentAmount
// 		} else if txDesc.Tx.GetMetadataType() == metadata.LoanWithdrawMeta {
//          DONE
// 			withdrawMeta := txDesc.Tx.GetMetadata().(*metadata.LoanWithdraw)
// 			meta, err := blockgen.chain.GetLoanRequestMeta(withdrawMeta.LoanID)
// 			if err != nil {
// 				removableTxs = append(removableTxs, txDesc.Tx)
// 				continue
// 			}

// 			unlockMeta := &metadata.LoanUnlock{
// 				LoanID:       make([]byte, len(withdrawMeta.LoanID)),
// 				MetadataBase: metadata.MetadataBase{Type: metadata.LoanUnlockMeta},
// 			}
// 			copy(unlockMeta.LoanID, withdrawMeta.LoanID)
// 			uplockMetaList := []metadata.Metadata{unlockMeta}
// 			amounts := []uint64{meta.LoanAmount}
// 			txNormals, err := transaction.BuildCoinbaseTxs([]*privacy.PaymentAddress{meta.ReceiveAddress}, amounts, producerPrivateKey, blockgen.chain.GetDatabase(), uplockMetaList)
// 			if err != nil {
// 				removableTxs = append(removableTxs, txDesc.Tx)
// 				continue
// 			}
// 			loanUnlockTxs = append(loanUnlockTxs, txNormals[0]) // There's only one tx
// 		}
// 	}
// 	return amount, loanUnlockTxs, removableTxs
// }

// func (blockgen *BlkTmplGenerator) buildBuyBackResponseTxs(
// 	buyBackFromInfos []*buyBackFromInfo,
// 	chainID byte,
// 	privatekey *privacy.SpendingKey,
// ) ([]metadata.Transaction, error) {
// 	if len(buyBackFromInfos) == 0 {
// 		return []metadata.Transaction{}, nil
// 	}

// 	// prevBlock := blockgen.chain.BestState[chainID].BestBlock
// 	var buyBackResTxs []metadata.Transaction
// 	for _, buyBackFromInfo := range buyBackFromInfos {
// 		buyBackAmount := buyBackFromInfo.value * buyBackFromInfo.buyBackPrice
// 		buyBackRes := metadata.NewBuyBackResponse(*buyBackFromInfo.requestedTxID, metadata.BuyBackResponseMeta)
// 		buyBackResTx := new(transaction.Tx)
// 		err := buyBackResTx.InitTxSalary(buyBackAmount, &buyBackFromInfo.paymentAddress, privatekey, blockgen.chain.GetDatabase(), buyBackRes)
// 		if err != nil {
// 			return []metadata.Transaction{}, err
// 		}
// 		buyBackResTxs = append(buyBackResTxs, buyBackResTx)
// 	}
// 	return buyBackResTxs, nil
// }

// // buildBuySellResponsesTx
// // the tx is to distribute tokens (bond, gov, ...) to token requesters
// func (blockgen *BlkTmplGenerator) buildBuySellResponsesTx(
// 	buySellReqTxs []metadata.Transaction,
// 	sellingBondsParam *params.SellingBonds,
// ) ([]metadata.Transaction, error) {
// 	if len(buySellReqTxs) == 0 {
// 		return []metadata.Transaction{}, nil
// 	}
// 	var resTxs []metadata.Transaction
// 	for _, reqTx := range buySellReqTxs {
// 		resTx, err := buildSingleBuySellResponseTx(reqTx, sellingBondsParam)
// 		if err != nil {
// 			return []metadata.Transaction{}, err
// 		}
// 		resTxs = append(resTxs, resTx)
// 	}
// 	return resTxs, nil
// }

// func buildSingleBuySellResponseTx(
// 	buySellReqTx metadata.Transaction,
// 	sellingBondsParam *params.SellingBonds,
// ) (*transaction.TxCustomToken, error) {
// 	bondID := sellingBondsParam.GetID()
// 	buySellRes := metadata.NewBuySellResponse(
// 		*buySellReqTx.Hash(),
// 		sellingBondsParam.StartSellingAt,
// 		sellingBondsParam.Maturity,
// 		sellingBondsParam.BuyBackPrice,
// 		bondID[:],
// 		metadata.BuyFromGOVResponseMeta,
// 	)

// 	buySellReqMeta := buySellReqTx.GetMetadata()
// 	buySellReq, ok := buySellReqMeta.(*metadata.BuySellRequest)
// 	if !ok {
// 		return nil, errors.New("Could not assert BuySellRequest metadata.")
// 	}
// 	txTokenVout := transaction.TxTokenVout{
// 		Value:          buySellReq.Amount,
// 		PaymentAddress: buySellReq.PaymentAddress,
// 	}

// 	var propertyID [common.HashSize]byte
// 	copy(propertyID[:], bondID[:])
// 	txTokenData := transaction.TxTokenData{
// 		Type:       transaction.CustomTokenInit,
// 		Mintable:   true,
// 		Amount:     buySellReq.Amount,
// 		PropertyID: common.Hash(propertyID),
// 		Vins:       []transaction.TxTokenVin{},
// 		Vouts:      []transaction.TxTokenVout{txTokenVout},
// 	}
// 	txTokenData.PropertyName = txTokenData.PropertyID.String()
// 	txTokenData.PropertySymbol = txTokenData.PropertyID.String()

// 	resTx := &transaction.TxCustomToken{
// 		TxTokenData: txTokenData,
// 	}
// 	resTx.Type = common.TxCustomTokenType
// 	resTx.SetMetadata(buySellRes)
// 	return resTx, nil
// }

// func buildSingleBuyGOVTokensResTx(
// 	buyGOVTokensReqTx metadata.Transaction,
// 	sellingGOVTokensParams *params.SellingGOVTokens,
// ) (*transaction.TxCustomToken, error) {
// 	buyGOVTokensRes := metadata.NewResponseBase(
// 		*buyGOVTokensReqTx.Hash(),
// 		metadata.ResponseBaseMeta,
// 	)

// 	buyGOVTokensReqMeta := buyGOVTokensReqTx.GetMetadata()
// 	buyGOVTokensReq, ok := buyGOVTokensReqMeta.(*metadata.BuyGOVTokenRequest)
// 	if !ok {
// 		return nil, errors.New("Could not assert BuyGOVTokenRequest metadata.")
// 	}
// 	txTokenVout := transaction.TxTokenVout{
// 		Value:          buyGOVTokensReq.Amount,
// 		PaymentAddress: buyGOVTokensReq.BuyerAddress,
// 	}

// 	var propertyID [common.HashSize]byte
// 	copy(propertyID[:], buyGOVTokensReq.TokenID[:])
// 	txTokenData := transaction.TxTokenData{
// 		Type:       transaction.CustomTokenInit,
// 		Mintable:   true,
// 		Amount:     buyGOVTokensReq.Amount,
// 		PropertyID: common.Hash(propertyID),
// 		Vins:       []transaction.TxTokenVin{},
// 		Vouts:      []transaction.TxTokenVout{txTokenVout},
// 	}
// 	txTokenData.PropertyName = txTokenData.PropertyID.String()
// 	txTokenData.PropertySymbol = txTokenData.PropertyID.String()

// 	resTx := &transaction.TxCustomToken{
// 		TxTokenData: txTokenData,
// 	}
// 	resTx.Type = common.TxCustomTokenType
// 	resTx.SetMetadata(buyGOVTokensRes)
// 	return resTx, nil
// }

// func (blockgen *BlkTmplGenerator) buildBuyGOVTokensResTxs(
// 	buyGOVTokensReqTxs []metadata.Transaction,
// 	sellingGOVTokensParams *params.SellingGOVTokens,
// ) ([]metadata.Transaction, error) {
// 	if len(buyGOVTokensReqTxs) == 0 {
// 		return []metadata.Transaction{}, nil
// 	}
// 	var resTxs []metadata.Transaction
// 	for _, reqTx := range buyGOVTokensReqTxs {
// 		resTx, err := buildSingleBuyGOVTokensResTx(reqTx, sellingGOVTokensParams)
// 		if err != nil {
// 			return []metadata.Transaction{}, err
// 		}
// 		resTxs = append(resTxs, resTx)
// 	}
// 	return resTxs, nil
// }

// func (blockgen *BlkTmplGenerator) buildResponseTxs(
// 	chainID byte,
// 	sourceTxns []*metadata.TxDesc,
// 	privatekey *privacy.SpendingKey,
// 	txGroups *txGroups,
// 	accumulativeValues *accumulativeValues,
// 	buyBackFromInfos []*buyBackFromInfo,
// ) (*txGroups, *accumulativeValues, map[string]uint64, error) {
// 	prevBlock := blockgen.chain.BestState[chainID].BestBlock
// 	// create buy/sell response txs to distribute bonds/govs to requesters
// 	buySellResTxs, err := blockgen.buildBuySellResponsesTx(
// 		txGroups.buySellReqTxs,
// 		blockgen.chain.BestState[14].BestBlock.Header.GOVConstitution.GOVParams.SellingBonds,
// 	)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return nil, nil, nil, err
// 	}
// 	buyGOVTokensResTxs, err := blockgen.buildBuyGOVTokensResTxs(
// 		txGroups.buyGOVTokensReqTxs,
// 		blockgen.chain.BestState[14].BestBlock.Header.GOVConstitution.GOVParams.SellingGOVTokens,
// 	)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return nil, nil, nil, err
// 	}

// 	// create buy-back response txs to distribute constants to buy-back requesters
// 	buyBackResTxs, err := blockgen.buildBuyBackResponseTxs(buyBackFromInfos, chainID, privatekey)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return nil, nil, nil, err
// 	}

// 	oracleRewardTxs, totalOracleRewards, updatedOracleValues, err := blockgen.buildOracleRewardTxs(chainID, privatekey)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return nil, nil, nil, err
// 	}

// 	// create refund txs
// 	currentSalaryFund := prevBlock.Header.SalaryFund
// 	remainingFund := currentSalaryFund + accumulativeValues.totalFee + accumulativeValues.incomeFromBonds + accumulativeValues.incomeFromGOVTokens - (accumulativeValues.totalSalary + accumulativeValues.buyBackCoins + totalOracleRewards)
// 	refundTxs, totalRefundAmt := blockgen.buildRefundTxs(chainID, remainingFund, privatekey)

// 	issuingResTxs, err := blockgen.buildIssuingResTxs(chainID, txGroups.issuingReqTxs, privatekey)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return nil, nil, nil, err
// 	}

// 	// Get loan payment amount to add to DCB fund
// 	loanPaymentAmount, unlockTxs, removableTxs := blockgen.processLoan(sourceTxns, privatekey)
// 	for _, tx := range removableTxs {
// 		txGroups.txToRemove = append(txGroups.txToRemove, tx)
// 	}
// 	txGroups.buySellResTxs = buySellResTxs
// 	txGroups.buyGOVTokensResTxs = buyGOVTokensResTxs
// 	txGroups.buyBackResTxs = buyBackResTxs
// 	txGroups.oracleRewardTxs = oracleRewardTxs
// 	txGroups.refundTxs = refundTxs
// 	txGroups.issuingResTxs = issuingResTxs
// 	txGroups.unlockTxs = unlockTxs

// 	accumulativeValues.totalOracleRewards = totalOracleRewards
// 	accumulativeValues.totalRefundAmt = totalRefundAmt
// 	accumulativeValues.loanPaymentAmount = loanPaymentAmount
// 	accumulativeValues.currentSalaryFund = currentSalaryFund
// 	return txGroups, accumulativeValues, updatedOracleValues, nil
// }

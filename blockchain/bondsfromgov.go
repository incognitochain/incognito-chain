package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/ninjadotorg/constant/blockchain/component"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type BuySellReqAction struct {
	TxReqID common.Hash             `json:"txReqId"`
	Meta    metadata.BuySellRequest `json:"meta"`
}

type BuyBackInfo struct {
	PaymentAddress privacy.PaymentAddress
	BuyBackPrice   uint64
	Value          uint64
	RequestedTxID  common.Hash
	TokenID        common.Hash
}

type BuyBackReqAction struct {
	TxReqID        common.Hash               `json:"txReqId"`
	BuyBacReqkMeta *metadata.BuyBackRequest  `json:"buyBackReqMeta"`
	PrevMeta       *metadata.BuySellResponse `json:"prevMeta"`
}

func buildInstructionsForBuyBackBondsReq(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
	bc *BlockChain,
) ([][]string, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return [][]string{}, err
	}
	var buyBackReqAction BuyBackReqAction
	err = json.Unmarshal(contentBytes, &buyBackReqAction)
	if err != nil {
		return nil, err
	}
	buyBackReqMeta := buyBackReqAction.BuyBacReqkMeta
	buySellResMeta := buyBackReqAction.PrevMeta

	instType := ""
	bestBlockHeight := beaconBestState.BestBlock.Header.Height
	if (buySellResMeta.StartSellingAt+buySellResMeta.Maturity > bestBlockHeight+1) ||
		!isGOVFundEnough(beaconBestState, accumulativeValues, buyBackReqMeta.Amount*buySellResMeta.BuyBackPrice) {
		instType = "refund"
	} else {
		instType = "accepted"
		accumulativeValues.buyBackCoins += buyBackReqMeta.Amount * buySellResMeta.BuyBackPrice
	}

	buyBackInfo := BuyBackInfo{
		PaymentAddress: buyBackReqMeta.PaymentAddress,
		BuyBackPrice:   buySellResMeta.BuyBackPrice,
		Value:          buyBackReqMeta.Amount,
		RequestedTxID:  buyBackReqAction.TxReqID,
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

func (blockgen *BlkTmplGenerator) buildBuyBondsFromGOVRes(
	instType string,
	contentStr string,
	sellingBondsParamsStr string,
	blkProducerPrivateKey *privacy.SpendingKey,
) ([]metadata.Transaction, error) {
	sellingBondsParamsBytes := []byte(sellingBondsParamsStr)
	var sellingBondsParams component.SellingBonds
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

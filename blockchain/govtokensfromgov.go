package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type BuyGOVTokenReqAction struct {
	TxReqID common.Hash                 `json:"txReqId"`
	Meta    metadata.BuyGOVTokenRequest `json:"meta"`
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

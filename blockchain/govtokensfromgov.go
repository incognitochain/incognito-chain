package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
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
		(!bytes.Equal(common.GOVTokenID[:], md.TokenID[:])) ||
		(bestBlockHeight+1 < sellingGOVTokensParams.StartSellingAt) ||
		(bestBlockHeight+1 > sellingGOVTokensParams.StartSellingAt+sellingGOVTokensParams.SellingWithin) ||
		(accumulativeValues.govTokensSold+md.Amount > sellingGOVTokensParams.GOVTokensToSell) {
		instType = "refund"
	} else {
		buyPrice := uint64(0)
		govTokenPriceFromOracle := stabilityInfo.Oracle.GOVToken
		if govTokenPriceFromOracle == 0 {
			buyPrice = sellingGOVTokensParams.GOVTokenPrice
		} else {
			buyPrice = govTokenPriceFromOracle
		}

		if md.BuyPrice < buyPrice {
			instType = "refund"
		} else {
			accumulativeValues.incomeFromGOVTokens += (md.Amount + md.BuyPrice)
			accumulativeValues.govTokensSold += md.Amount
			instType = "accepted"
		}
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
	blkProducerPrivateKey *privacy.PrivateKey,
	shardID byte,
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
		propID := common.Hash(propertyID)
		tokenParams := &transaction.CustomTokenParamTx{
			PropertyID:     propID.String(),
			PropertyName:   propID.String(),
			PropertySymbol: propID.String(),
			Amount:         reqMeta.Amount,
			TokenTxType:    transaction.CustomTokenInit,
			Receiver:       []transaction.TxTokenVout{txTokenVout},
			Mintable:       true,
		}

		resTx := &transaction.TxCustomToken{}
		initErr := resTx.Init(
			blkProducerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			blockgen.chain.config.DataBase,
			buyGOVTokensRes,
			false,
			shardID,
		)
		if initErr != nil {
			Logger.log.Error(err)
			return nil, initErr
		}
		return []metadata.Transaction{resTx}, nil
	}
	return []metadata.Transaction{}, nil
}

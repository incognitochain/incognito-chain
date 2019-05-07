package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
)

type BuySellReqAction struct {
	TxReqID common.Hash             `json:"txReqId"`
	Meta    metadata.BuySellRequest `json:"meta"`
}

type BuyBackInfo struct {
	BondID         common.Hash
	PaymentAddress privacy.PaymentAddress
	BuyBackPrice   uint64
	Value          uint64
	RequestedTxID  common.Hash
	TradeID        []byte
}

type BuyBackReqAction struct {
	TxReqID        common.Hash              `json:"txReqId"`
	BuyBackReqMeta *metadata.BuyBackRequest `json:"buyBackReqMeta"`
	BondID         *common.Hash             `json:"bondId"`
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
	buyBackReqMeta := buyBackReqAction.BuyBackReqMeta
	bondID := buyBackReqAction.BondID
	bondTypeBytes, err := bc.config.DataBase.GetSoldBondTypeByID(bondID)
	if err != nil {
		return nil, err
	}
	bestBlockHeight := beaconBestState.BestBlock.Header.Height
	var bondTypeInfo component.SellingBonds
	instType := ""
	if len(bondTypeBytes) == 0 {
		instType = "refund"
	} else {
		err = json.Unmarshal(bondTypeBytes, &bondTypeInfo)
		if err != nil {
			return nil, err
		}
		if (bondTypeInfo.StartSellingAt+bondTypeInfo.Maturity > bestBlockHeight+1) ||
			!isGOVFundEnough(beaconBestState, accumulativeValues, buyBackReqMeta.Amount*bondTypeInfo.BuyBackPrice) {
			instType = "refund"
		} else {
			instType = "accepted"
			accumulativeValues.buyBackCoins += buyBackReqMeta.Amount * bondTypeInfo.BuyBackPrice
		}
	}

	buyBackInfo := BuyBackInfo{
		BondID:         *bondID,
		PaymentAddress: buyBackReqMeta.PaymentAddress,
		BuyBackPrice:   bondTypeInfo.BuyBackPrice,
		Value:          buyBackReqMeta.Amount,
		RequestedTxID:  buyBackReqAction.TxReqID,
		TradeID:        buyBackReqMeta.TradeID,
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
		(!bytes.Equal(sellingBondsParams.GetID()[:], md.TokenID[:])) ||
		(bestBlockHeight+1 < sellingBondsParams.StartSellingAt) ||
		(bestBlockHeight+1 > sellingBondsParams.StartSellingAt+sellingBondsParams.SellingWithin) ||
		(accumulativeValues.bondsSold+md.Amount > sellingBondsParams.BondsToSell) ||
		(md.BuyPrice < sellingBondsParams.BondPrice) {
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
	blkProducerPrivateKey *privacy.PrivateKey,
	shardID byte,
) ([]metadata.Transaction, error) {
	var buyBackInfo BuyBackInfo
	err := json.Unmarshal([]byte(buyBackInfoStr), &buyBackInfo)
	if err != nil {
		return nil, err
	}

	if instType == "refund" {
		bondID := buyBackInfo.BondID
		txTokenVout := transaction.TxTokenVout{
			Value:          buyBackInfo.Value,
			PaymentAddress: buyBackInfo.PaymentAddress,
		}
		var propertyID [common.HashSize]byte
		copy(propertyID[:], bondID[:])
		propID := common.Hash(propertyID)
		tokenParams := &transaction.CustomTokenParamTx{
			PropertyID:     propID.String(),
			PropertyName:   propID.String(),
			PropertySymbol: propID.String(),
			Amount:         buyBackInfo.Value,
			// TokenTxType:    transaction.CustomTokenMint,
			TokenTxType: transaction.CustomTokenInit,
			Receiver:    []transaction.TxTokenVout{txTokenVout},
			Mintable:    true,
		}

		buyBackRes := metadata.NewBuyBackResponse(buyBackInfo.RequestedTxID, metadata.BuyBackResponseMeta)

		resTx := &transaction.TxCustomToken{}
		err = resTx.Init(
			blkProducerPrivateKey,
			nil,
			nil,
			0,
			tokenParams,
			blockgen.chain.config.DataBase,
			buyBackRes,
			false,
			shardID,
		)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		return []metadata.Transaction{resTx}, nil

	} else if instType == "accepted" {
		dcbKey, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
		dcbAddr := dcbKey.KeySet.PaymentAddress
		burningKey, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
		burningAddr := burningKey.KeySet.PaymentAddress

		receiverAddr := buyBackInfo.PaymentAddress
		if bytes.Equal(buyBackInfo.PaymentAddress.Pk[:], dcbAddr.Pk[:]) {
			receiverAddr = burningAddr
		}

		buyBackAmount := buyBackInfo.Value * buyBackInfo.BuyBackPrice
		buyBackRes := metadata.NewBuyBackResponse(buyBackInfo.RequestedTxID, metadata.BuyBackResponseMeta)
		buyBackResTx := new(transaction.Tx)
		err := buyBackResTx.InitTxSalary(
			buyBackAmount,
			&receiverAddr,
			blkProducerPrivateKey,
			blockgen.chain.GetDatabase(),
			buyBackRes,
		)
		// buyBackResTx.Type = common.TxNormalType // NOTE: keep it as salary/coinbase type in order to bypass validation
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
	blkProducerPrivateKey *privacy.PrivateKey,
	shardID byte,
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
	key, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	dcbAddr := key.KeySet.PaymentAddress
	if instType == "refund" && !bytes.Equal(dcbAddr.Pk[:], reqMeta.PaymentAddress.Pk[:]) {
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
		propID := common.Hash(propertyID)
		tokenParams := &transaction.CustomTokenParamTx{
			PropertyID:     propID.String(),
			PropertyName:   propID.String(),
			PropertySymbol: propID.String(),
			Amount:         reqMeta.Amount,
			// TokenTxType:    transaction.CustomTokenMint,
			TokenTxType: transaction.CustomTokenInit,
			Receiver:    []transaction.TxTokenVout{txTokenVout},
			Mintable:    true,
		}

		resTx := &transaction.TxCustomToken{}
		initErr := resTx.Init(
			blkProducerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			blockgen.chain.config.DataBase,
			buySellRes,
			false,
			shardID,
		)

		if initErr != nil {
			Logger.log.Error(initErr)
			return nil, initErr
		}
		return []metadata.Transaction{resTx}, nil
	}
	return []metadata.Transaction{}, nil
}

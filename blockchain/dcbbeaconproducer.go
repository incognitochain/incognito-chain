package blockchain

import (
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
)

// buildPassThroughInstruction converts shard instruction to beacon instruction in order to update BeaconBestState later on in beaconprocess
func buildPassThroughInstruction(receivedType int, contentStr string) ([][]string, error) {
	metaType := strconv.Itoa(receivedType)
	shardID := strconv.Itoa(component.BeaconOnly)
	return [][]string{[]string{metaType, shardID, contentStr}}, nil
}

func buildInstructionsForCrowdsaleRequest(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
	bc *BlockChain,
) ([][]string, error) {
	saleID, paymentAddress, sentAmount, err := metadata.ParseCrowdsaleRequestActionValue(contentStr)
	if err != nil {
		// fmt.Printf("[db] error parsing action: %+v\n", err)
		return nil, err
	}

	// Get data of current crowdsale
	key := string(saleID)
	var saleData *component.SaleData
	ok := false
	if saleData, ok = accumulativeValues.saleDataMap[key]; !ok {
		saleData, err = bc.GetSaleData(saleID)
		if err != nil {
			// fmt.Printf("[db] saleid not exist: %x\n", saleID)
			return nil, fmt.Errorf("saleID not exist: %x", saleID)
		}
	}
	accumulativeValues.saleDataMap[key] = saleData

	inst, err := buildPaymentInstructionForCrowdsale(
		paymentAddress,
		sentAmount,
		beaconBestState,
		saleData,
		bc,
	)
	fmt.Println("[db] built crowdsale payment inst:", inst, err)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func buildPaymentInstructionForCrowdsale(
	paymentAddress privacy.PaymentAddress,
	sentAmount uint64,
	beaconBestState *BestStateBeacon,
	saleData *component.SaleData,
	bc *BlockChain,
) ([][]string, error) {
	bondID := saleData.BondID
	price := saleData.Price
	buyingAsset := &common.ConstantID
	sellingAsset := bondID
	if saleData.Buy {
		buyingAsset = bondID
		sellingAsset = &common.ConstantID
	}

	// Check if sale is on-going
	if bestStateBeacon.BeaconHeight >= saleData.EndBlock {
		return generateCrowdsalePaymentInstruction(paymentAddress, sentAmount, buyingAsset, saleData.SaleID, 0, false) // refund
	}

	paymentAmount := uint64(0)
	if saleData.Buy {
		// Number of Constant must send to user
		paymentAmount = sentAmount * price

		// Check if there's still enough bond to buy
		if sentAmount > saleData.Amount {
			// fmt.Printf("[db] Crowdsale reached limit\n")
			return generateCrowdsalePaymentInstruction(paymentAddress, sentAmount, buyingAsset, saleData.SaleID, 0, false) // refund
		}

		// Update amount of buying/selling asset of the crowdsale
		saleData.Amount -= sentAmount

	} else {
		// Number of Bond must send to user
		paymentAmount = sentAmount / price

		// Check if there's still enough asset to trade
		dcbBondAmount, _ := bc.GetDCBBondInfo(sellingAsset)
		if paymentAmount > saleData.Amount || paymentAmount > dcbBondAmount {
			// fmt.Printf("[db] Crowdsale reached limit\n")
			return generateCrowdsalePaymentInstruction(paymentAddress, sentAmount, buyingAsset, saleData.SaleID, 0, false) // refund
		}

		// Update amount of buying/selling asset of the crowdsale
		saleData.Amount -= paymentAmount
	}

	// fmt.Printf("[db] sentValue, payAmount, buyLeft, sellLeft: %d %d %d %d\n", sentAssetValue, paymentAmount, saleData.BuyingAmount, saleData.SellingAmount)
	// Build instructions
	return generateCrowdsalePaymentInstruction(paymentAddress, paymentAmount, sellingAsset, saleData.SaleID, sentAmount, true)
}

func buildInstructionsForTradeActivation(
	shardID byte,
	contentStr string,
) ([][]string, error) {
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	dcbPk := keyWalletDCBAccount.KeySet.PaymentAddress.Pk
	dcbShardID := common.GetShardIDFromLastByte(dcbPk[len(dcbPk)-1])
	inst := []string{
		strconv.Itoa(metadata.TradeActivationMeta),
		strconv.Itoa(int(dcbShardID)),
		contentStr,
	}
	fmt.Printf("[db] beacon built inst: %v\n", inst)
	return [][]string{inst}, nil
}

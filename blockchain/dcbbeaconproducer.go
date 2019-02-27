package blockchain

import (
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
)

// buildPassthroughInstruction converts shard instruction to beacon instruction in order to update BeaconBestState later on in beaconprocess
func buildPassthroughInstruction(receivedType int, contentStr string) ([][]string, error) {
	metaType := strconv.Itoa(receivedType)
	shardID := strconv.Itoa(metadata.BeaconOnly)
	return [][]string{[]string{metaType, shardID, contentStr}}, nil
}

func buildInstructionsForCrowdsaleRequest(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	saleID, priceLimit, limitSell, paymentAddress, sentAmount, err := metadata.ParseCrowdsaleRequestActionValue(contentStr)
	if err != nil {
		fmt.Printf("[db] error parsing action: %+v\n", err)
		return nil, err
	}

	// Get data of current crowdsale
	key := getSaleDataKeyBeacon(saleID)
	var saleData *params.SaleData
	ok := false
	if saleData, ok = accumulativeValues.saleDataMap[key]; !ok {
		if value, ok := beaconBestState.Params[key]; ok {
			saleData, err = parseSaleDataValueBeacon(value)
		} else {
			fmt.Printf("[db] saleid not exist: %x\n", saleID)
			return nil, errors.Errorf("SaleID not exist: %x", saleID)
		}
	}
	accumulativeValues.saleDataMap[key] = saleData

	// Skip payment if either selling or buying asset is offchain (needs confirmation)
	if common.IsOffChainAsset(&saleData.SellingAsset) || common.IsOffChainAsset(&saleData.BuyingAsset) {
		fmt.Println("[db] crowdsale offchain asset")
		return nil, nil
	}

	inst, err := buildPaymentInstructionForCrowdsale(
		priceLimit,
		limitSell,
		paymentAddress,
		sentAmount,
		beaconBestState,
		saleData,
	)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func buildPaymentInstructionForCrowdsale(
	priceLimit uint64,
	limitSell bool,
	paymentAddress privacy.PaymentAddress,
	sentAmount uint64,
	beaconBestState *BestStateBeacon,
	saleData *params.SaleData,
) ([][]string, error) {
	// Get price for asset
	buyingAsset := saleData.BuyingAsset
	sellingAsset := saleData.SellingAsset
	buyPrice := beaconBestState.getAssetPrice(buyingAsset)
	sellPrice := beaconBestState.getAssetPrice(sellingAsset)
	if buyPrice == 0 {
		buyPrice = saleData.DefaultBuyPrice
	}
	if sellPrice == 0 {
		sellPrice = saleData.DefaultSellPrice
	}
	if buyPrice == 0 || sellPrice == 0 {
		fmt.Printf("[db] asset price is 0: %d %d\n", buyPrice, sellPrice)
		return generateCrowdsalePaymentInstruction(paymentAddress, sentAmount, buyingAsset, saleData.SaleID, 0, false) // refund
	}
	fmt.Printf("[db] buy and sell price: %d %d\n", buyPrice, sellPrice)

	// Check if price limit is not violated
	priceLimit *= beaconBestState.StabilityInfo.Oracle.Constant // Convert from Nano to Millicent
	if limitSell && sellPrice > priceLimit {
		fmt.Printf("Price limit violated: %d %d\n", sellPrice, priceLimit)
		return generateCrowdsalePaymentInstruction(paymentAddress, sentAmount, buyingAsset, saleData.SaleID, 0, false) // refund
	} else if !limitSell && buyPrice < priceLimit {
		fmt.Printf("Price limit violated: %d %d\n", buyPrice, priceLimit)
		return generateCrowdsalePaymentInstruction(paymentAddress, sentAmount, buyingAsset, saleData.SaleID, 0, false) // refund
	}

	// Calculate value of asset sent in request tx
	sentAssetValue := sentAmount * buyPrice // in USD

	// Number of asset must pay to user
	paymentAmount := sentAssetValue / sellPrice

	// Check if there's still enough asset to trade
	if sentAmount > saleData.BuyingAmount || paymentAmount > saleData.SellingAmount {
		fmt.Printf("Crowdsale reached limit\n")
		return generateCrowdsalePaymentInstruction(paymentAddress, sentAmount, buyingAsset, saleData.SaleID, 0, false) // refund
	}

	// Update amount of buying/selling asset of the crowdsale
	saleData.BuyingAmount -= sentAmount
	saleData.SellingAmount -= paymentAmount

	fmt.Printf("[db] sentValue, payAmount, buyLeft, sellLeft: %d %d %d %d\n", sentAssetValue, paymentAmount, saleData.BuyingAmount, saleData.SellingAmount)

	// Build instructions
	return generateCrowdsalePaymentInstruction(paymentAddress, paymentAmount, sellingAsset, saleData.SaleID, sentAmount, true)
}

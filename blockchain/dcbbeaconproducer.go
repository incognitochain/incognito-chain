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

// buildInstructionsForLoanRequest converts shard LoanRequest inst to beacon inst to update BeaconBestState later on
func buildInstructionsForLoanRequest(contentStr string) ([][]string, error) {
	// Pass through
	metaType := strconv.Itoa(metadata.LoanRequestMeta)
	shardID := strconv.Itoa(metadata.BeaconOnly)
	return [][]string{[]string{metaType, shardID, contentStr}}, nil
}

// buildInstructionsForLoanResponse converts shard LoanResponse inst to beacon inst to update BeaconBestState later on
func buildInstructionsForLoanResponse(contentStr string) ([][]string, error) {
	// Pass through
	metaType := strconv.Itoa(metadata.LoanResponseMeta)
	shardID := strconv.Itoa(metadata.BeaconOnly)
	return [][]string{[]string{metaType, shardID, contentStr}}, nil
}

// buildInstructionsForLoanPayment converts shard LoanPayment inst to beacon inst to update BeaconBestState later on
func buildInstructionsForLoanPayment(contentStr string) ([][]string, error) {
	// Pass through
	metaType := strconv.Itoa(metadata.LoanPaymentMeta)
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
		return nil, err
	}
	key := getSaleDataKeyBeacon(saleID)

	// Get data of current crowdsale
	var saleData *params.SaleData
	ok := false
	if saleData, ok = accumulativeValues.saleDataMap[key]; !ok {
		if value, ok := beaconBestState.Params[key]; ok {
			saleData, _ = parseSaleDataValueBeacon(value)
		} else {
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
	if buyPrice == 0 || sellPrice == 0 {
		buyPrice = saleData.DefaultBuyPrice
		sellPrice = saleData.DefaultSellPrice
		if buyPrice == 0 || sellPrice == 0 {
			return generateCrowdsalePaymentInstruction(paymentAddress, sentAmount, buyingAsset, saleData.SaleID, 0, false) // refund
		}
	}
	fmt.Printf("[db] buy and sell price: %d %d\n", buyPrice, sellPrice)

	// Check if price limit is not violated
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

	// Build instructions
	return generateCrowdsalePaymentInstruction(paymentAddress, paymentAmount, sellingAsset, saleData.SaleID, sentAmount, true)
}

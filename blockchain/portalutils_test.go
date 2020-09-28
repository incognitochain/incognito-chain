package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"testing"
)

func TestPickUpCustodianForPorting(t *testing.T) {
	// 1BTC
	portingAmount := uint64(1200000000)
	portalTokenID := common.PortalBTCIDStr

	finalExchangeRate := statedb.NewFinalExchangeRatesStateWithValue(
		map[string]statedb.FinalExchangeRatesDetail{
			common.PRVIDStr:       {Amount: 1000000},
			common.PortalBNBIDStr: {Amount: 20000000},
			common.PortalBTCIDStr: {Amount: 10000000000},
			common.EthAddrStr:     {Amount: 400000000},
		})

	portalParam := PortalParams{
		MinPercentLockedCollateral: 150,
	}

	supportCollateral := []PortalCollateral{
		{common.EthAddrStr, 9},
	}

	custodianKey1 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress1").String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress2").String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress3").String()
	custodianKey4 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress4").String()

	custodian1 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress1", 7000*1e9, 7000*1e9,
		map[string]uint64{}, map[string]uint64{},
		map[string]string{
			common.PortalBNBIDStr: "bnbAddress1",
			common.PortalBTCIDStr: "btcAddress1",
		}, map[string]uint64{},
		map[string]uint64{
			common.EthAddrStr: 1e9,
		},
		map[string]uint64{
			common.EthAddrStr: 1e9,
		}, map[string]map[string]uint64{})

	custodian2 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress2", 1000*1e9, 1000*1e9,
		map[string]uint64{}, map[string]uint64{},
		map[string]string{
			common.PortalBNBIDStr: "bnbAddress2",
			common.PortalBTCIDStr: "btcAddress2",
		}, map[string]uint64{},
		map[string]uint64{
			common.EthAddrStr: 1e9,
		}, map[string]uint64{
			common.EthAddrStr: 1e9,
		}, map[string]map[string]uint64{})


	custodian3 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress3", 10000*1e9, 10000*1e9,
		map[string]uint64{}, map[string]uint64{},
		map[string]string{
			common.PortalBTCIDStr: "btcAddress3",
		}, map[string]uint64{},
		map[string]uint64{
			common.EthAddrStr: 1e9,
		}, map[string]uint64{
			common.EthAddrStr: 1e9,
		}, map[string]map[string]uint64{})

	custodian4 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress4", 1200*1e9, 1200*1e9,
		map[string]uint64{}, map[string]uint64{},
		map[string]string{
			common.PortalBNBIDStr: "bnbAddress4",
			common.PortalBTCIDStr: "btcAddress4",
		}, map[string]uint64{},
		map[string]uint64{
			common.EthAddrStr: 1e9,
		}, map[string]uint64{
			common.EthAddrStr: 1e9,
		}, map[string]map[string]uint64{})

	custodianPool := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
		custodianKey4: custodian4,
	}

	matchCustodians, err := pickUpCustodianForPorting(portingAmount, portalTokenID, custodianPool, finalExchangeRate, portalParam, supportCollateral)

	fmt.Println("Err: ", err)
	fmt.Printf("Result: %+v", matchCustodians)
	for i, cus := range matchCustodians {
		fmt.Printf("Custodian %v ***** \n", i)
		fmt.Printf("cus.IncAddress %v\n", cus.IncAddress)
		fmt.Printf("cus.Amount %v\n", cus.Amount)
		fmt.Printf("cus.LockedAmountCollateral %v\n", cus.LockedAmountCollateral)
		fmt.Printf("cus.LockedTokenCollaterals %v\n", cus.LockedTokenCollaterals)
	}

}

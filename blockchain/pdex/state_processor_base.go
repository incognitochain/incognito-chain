package pdex

import "github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"

type stateProcessorBase struct {
}

func (stateProcessorBase) deductSharesForWithdrawal(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	withdrawerAddressStr string,
	amt uint64,
	shares map[string]uint64,
) {
	pdeShareKeyBytes, err := rawdbv2.BuildPDESharesKeyV2(beaconHeight, token1IDStr, token2IDStr, withdrawerAddressStr)
	if err != nil {
		Logger.log.Errorf("cannot find pdeShareKey for address: %v. Error: %v\n", withdrawerAddressStr, err)
		return
	}
	pdeShareKey := string(pdeShareKeyBytes)
	adjustingAmt := uint64(0)
	currentAmt, found := shares[pdeShareKey]
	if found && amt <= currentAmt {
		adjustingAmt = currentAmt - amt
	}
	shares[pdeShareKey] = adjustingAmt
}

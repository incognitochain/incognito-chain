package committeestate

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
)

func (s *BeaconCommitteeStateV4) ProcessBeaconSharePrice(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	for _, inst := range env.BeaconInstructions {
		//share price update instruction
		if inst[0] == instruction.SHARE_PRICE {
			sharePrice, err := instruction.ValidateAndImportSharePriceInstructionFromString(inst)
			if err != nil {
				return nil, err
			}
			newSharePrice := sharePrice.GetValue()
			for k, v := range newSharePrice {
				statedb.StoreBeaconSharePrice(s.stateDB, k, v)
			}

		}
	}
	return nil, nil
}

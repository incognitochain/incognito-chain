package committeestate

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

func (s *BeaconCommitteeStateV4) ProcessBeaconSharePrice(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	for _, inst := range env.BeaconInstructions {
		//share price update instruction
		if inst[0] == instruction.... {
			beaconStakeInst := instruction.ImportBeaconStakeInstructionFromString(inst)

		}
	}
	return nil, nil
}

package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processResetFeatureReward(
	stateDB *statedb.StateDB,
	instructions []string) error {

	// unmarshal instructions content
	var actionData metadata.ResetFeatureRewardRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == "resetFeatureReward" {
		// store total custodian reward into db
		_, err = statedb.ResetRewardFeatureStateByTokenID(
			stateDB,
			actionData.TokenID.String(),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while reset reward feature: %+v", err)
			return nil
		}
	} else {
		Logger.log.Errorf("ERROR: Invalid status of instruction: %+v", reqStatus)
		return nil
	}

	return nil
}
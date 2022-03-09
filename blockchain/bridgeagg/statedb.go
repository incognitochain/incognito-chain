package bridgeagg

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func CheckTokenIDExisted(sDBs map[int]*statedb.StateDB, tokenID common.Hash) error {
	for _, sDB := range sDBs {
		if statedb.PrivacyTokenIDExisted(sDB, tokenID) {
			return fmt.Errorf("Cannot find tokenID %s in network", tokenID.String())
		}
	}
	return nil
}

package blockchain

import (
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"testing"
)

func TestBeaconBestState_storeAllShardSubstitutesValidatorV3(t *testing.T) {
	type fields struct {
		beaconCommitteeState committeestate.BeaconCommitteeState
		consensusStateDB     *statedb.StateDB
	}
	type args struct {
		allAddedValidators map[byte][]incognitokey.CommitteePublicKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//TODO: @hung Add test cases.
		// 1. no change
		// 2. change but remain pos
		// 3. change pos
		// 4. add new at first
		// 5. add new at last
		// 6. add new at middle
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beaconCurView := &BeaconBestState{
				beaconCommitteeState: tt.fields.beaconCommitteeState,
				consensusStateDB:     tt.fields.consensusStateDB,
			}
			if err := beaconCurView.storeAllShardSubstitutesValidatorV3(tt.args.allAddedValidators); (err != nil) != tt.wantErr {
				t.Errorf("storeAllShardSubstitutesValidatorV3() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

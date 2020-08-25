package committeestate

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

func TestBeaconCommitteeEngine_BuildIncurredInstructions(t *testing.T) {
	type fields struct {
		beaconHeight                      uint64
		beaconHash                        common.Hash
		beaconCommitteeStateV1            *BeaconCommitteeStateV1
		uncommittedBeaconCommitteeStateV1 *BeaconCommitteeStateV1
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		wantErr bool
	}{
		{},
		{},
		{},
		{},
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := BeaconCommitteeEngine{
				beaconHeight:                      tt.fields.beaconHeight,
				beaconHash:                        tt.fields.beaconHash,
				beaconCommitteeStateV1:            tt.fields.beaconCommitteeStateV1,
				uncommittedBeaconCommitteeStateV1: tt.fields.uncommittedBeaconCommitteeStateV1,
			}
			got, err := engine.BuildIncurredInstructions(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngine.BuildIncurredInstructions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeEngine.BuildIncurredInstructions() = %v, want %v", got, tt.want)
			}
		})
	}
}

package committeestate

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

func TestBeaconCommitteeEngine_AssignShardsPoolUsingRandomInstruction(t *testing.T) {
	type fields struct {
		beaconHeight                      uint64
		beaconHash                        common.Hash
		beaconCommitteeStateV1            *BeaconCommitteeStateV1
		uncommittedBeaconCommitteeStateV1 *BeaconCommitteeStateV1
	}
	type args struct {
		seed           int64
		numShards      int
		candidatesList []string
		subsSizeMap    map[byte]int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[byte][]string
	}{
		{
			name:   "imbalance",
			fields: fields{},
			args: args{
				seed:           500,
				numShards:      4,
				candidatesList: []string{"1", "2", "3", "4", "5"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeEngine{
				beaconHeight:                      tt.fields.beaconHeight,
				beaconHash:                        tt.fields.beaconHash,
				beaconCommitteeStateV1:            tt.fields.beaconCommitteeStateV1,
				uncommittedBeaconCommitteeStateV1: tt.fields.uncommittedBeaconCommitteeStateV1,
			}
			if got := b.AssignShardsPoolUsingRandomInstruction(tt.args.seed, tt.args.numShards, tt.args.candidatesList, tt.args.subsSizeMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeEngine.AssignShardsPoolUsingRandomInstruction() = %v, want %v", got, tt.want)
			}
		})
	}
}

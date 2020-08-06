package committeestate

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

func SampleCandidateList(len int) []string {
	res := []string{}
	for i := 0; i < len; i++ {
		res = append(res, fmt.Sprintf("committeepubkey%v", i))
	}
	return res
}

func GetMinMaxRange(sizeMap map[byte]int) int {
	min := -1
	max := -1
	for _, v := range sizeMap {
		if min == -1 {
			min = v
		}
		if min > v {
			min = v
		}
		if max < v {
			max = v
		}
	}
	return max - min
}

func TestBeaconCommitteeEngine_AssignShardsPoolUsingRandomInstruction(t *testing.T) {
	type fields struct {
		beaconHeight                      uint64
		beaconHash                        common.Hash
		beaconCommitteeStateV1            *BeaconCommitteeStateV1
		uncommittedBeaconCommitteeStateV1 *BeaconCommitteeStateV1
	}
	type args struct {
		seed        int64
		numShards   int
		subsSizeMap map[byte]int
		epoches     int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name:   "imbalance",
			fields: fields{},
			args: args{
				seed:      500,
				numShards: 4,
				subsSizeMap: map[byte]int{
					0: 5,
					1: 5,
					2: 5,
					3: 5,
				},
				epoches: 100,
			},
			want: 2,
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
			for i := 0; i < tt.args.epoches; i++ {
				_ = b.AssignShardsPoolUsingRandomInstruction(rand.Int63(), tt.args.numShards, SampleCandidateList(12), tt.args.subsSizeMap)

				fmt.Println(tt.args.subsSizeMap)
				diff := GetMinMaxRange(tt.args.subsSizeMap)
				if diff > tt.want {
					t.Errorf("BeaconCommitteeEngine.AssignShardsPoolUsingRandomInstruction() = %v, want %v", diff, tt.want)
				}
			}

		})
	}
}

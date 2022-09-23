package committeestate

import (
	"fmt"
	"sync"
	"testing"

	"github.com/tendermint/tendermint/libs/common"
)

func TestBeaconCommitteeStateV4_UpdateBeaconReputation(t *testing.T) {
	type fields struct {
		BeaconCommitteeStateV3 *BeaconCommitteeStateV3
		bDelegateState         *BeaconDelegateState
		bReputation            map[string]uint64
	}
	type args struct {
		bCommittee []string
		listVotes  []int
	}
	setUpTest := func(testName string, f fields, a args) (fields, args) {
		switch testName {
		case "Test 1":
			keyLength := 10
			committeeSize := 8
			// fmt.Println(f.BeaconCommitteeStateV3)
			for i := 0; i < committeeSize; i++ {
				f.BeaconCommitteeStateV3.beaconCommittee = append(f.BeaconCommitteeStateV3.beaconCommittee, common.RandStr(keyLength))
			}
			f.bDelegateState.DelegateInfo = map[string]*BeaconDelegatorInfo{}
			for _, bPK := range f.BeaconCommitteeStateV3.beaconCommittee {
				f.bDelegateState.DelegateInfo[bPK] = &BeaconDelegatorInfo{
					CurrentDelegators:        0,
					CurrentDelegatorsDetails: map[string]interface{}{},
					locker:                   &sync.RWMutex{},
				}
				f.bDelegateState.NextEpochDelegate = map[string]struct {
					Old string
					New string
				}{}
				f.bDelegateState.locker = &sync.RWMutex{}
			}
			for _, bPK := range f.BeaconCommitteeStateV3.beaconCommittee {
				f.bReputation[bPK] = 500
			}
			a.bCommittee = f.BeaconCommitteeStateV3.beaconCommittee
			a.listVotes = []int{0, 1, 2, 5, 6, 7, 8}
		}
		return f, a
	}
	printReputation := func(f fields) {
		for k, v := range f.bReputation {
			fmt.Printf("[%v]: %v\n", k, v)
		}
		fmt.Println("-------------------------")
	}
	tests := []struct {
		name    string
		loop    int
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test 1",
			fields: fields{
				BeaconCommitteeStateV3: NewBeaconCommitteeStateV3(),
				bDelegateState:         &BeaconDelegateState{},
				bReputation:            map[string]uint64{},
			},
			loop:    50,
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields, tt.args = setUpTest(tt.name, tt.fields, tt.args)
			b := &BeaconCommitteeStateV4{
				BeaconCommitteeStateV3: tt.fields.BeaconCommitteeStateV3,
				bDelegateState:         tt.fields.bDelegateState,
				bReputation:            tt.fields.bReputation,
			}
			for i := 0; i < tt.loop; i++ {
				if err := b.updateBeaconReputation(tt.args.bCommittee, tt.args.listVotes); (err != nil) != tt.wantErr {
					t.Errorf("BeaconCommitteeStateV4.UpdateBeaconReputation() error = %v, wantErr %v", err, tt.wantErr)
				}
				printReputation(tt.fields)
			}
		})
	}
}

package committeestate

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
)

func Test_beacon_swap_v1(t *testing.T) {

	type args struct {
		pendingList         []CandidateInfo
		committeeList       []CandidateInfo
		fixNodeStakingPower int64
		maxCommitteeSize    int
	}

	type testcase struct {
		name  string
		args  args
		want  []CandidateInfo
		want1 []CandidateInfo
	}

	testcase1 := func() testcase {
		name := "swap in all until max"
		args := args{
			[]CandidateInfo{
				{incognitokey.CommitteePublicKey{}, "1", 500, 50, "pending"},
				{incognitokey.CommitteePublicKey{}, "2", 500, 1500, "pending"},
				{incognitokey.CommitteePublicKey{}, "3", 500, 500, "pending"},
			},
			[]CandidateInfo{},
			3000,
			2,
		}
		newpendingList := []CandidateInfo{
			{incognitokey.CommitteePublicKey{}, "2", 500, 1500, "pending"},
		}
		newCommitteeList := []CandidateInfo{
			{incognitokey.CommitteePublicKey{}, "1", 500, 50, "pending"},
			{incognitokey.CommitteePublicKey{}, "3", 500, 500, "pending"},
		}
		return testcase{
			name, args, newpendingList, newCommitteeList,
		}
	}
	testcase2 := func() testcase {
		name := "swap in all until max 2"
		args := args{
			[]CandidateInfo{
				{incognitokey.CommitteePublicKey{}, "1", 500, 50, "pending"},
				{incognitokey.CommitteePublicKey{}, "2", 500, 1500, "pending"},
				{incognitokey.CommitteePublicKey{}, "3", 500, 500, "pending"},
			},
			[]CandidateInfo{
				{incognitokey.CommitteePublicKey{}, "4", 500, 500, "committee"},
			},
			3000,
			2,
		}
		newpendingList := []CandidateInfo{
			{incognitokey.CommitteePublicKey{}, "2", 500, 1500, "pending"},
			{incognitokey.CommitteePublicKey{}, "3", 500, 500, "pending"},
		}
		newCommitteeList := []CandidateInfo{
			{incognitokey.CommitteePublicKey{}, "4", 500, 500, "committee"},
			{incognitokey.CommitteePublicKey{}, "1", 500, 50, "pending"},
		}
		return testcase{
			name, args, newpendingList, newCommitteeList,
		}
	}
	testcase3 := func() testcase {
		name := "no swap (no higher score)"
		args := args{
			[]CandidateInfo{
				{incognitokey.CommitteePublicKey{}, "1", 400, 100, "pending"},
				{incognitokey.CommitteePublicKey{}, "2", 500, 200, "pending"},
				{incognitokey.CommitteePublicKey{}, "3", 450, 500, "pending"},
			},
			[]CandidateInfo{
				{incognitokey.CommitteePublicKey{}, "4", 500, 50, "committee"},
				{incognitokey.CommitteePublicKey{}, "5", 500, 50, "committee"},
			},
			3000,
			2,
		}
		newpendingList := []CandidateInfo{
			{incognitokey.CommitteePublicKey{}, "2", 500, 200, "pending"},
			{incognitokey.CommitteePublicKey{}, "3", 450, 500, "pending"},
			{incognitokey.CommitteePublicKey{}, "1", 400, 100, "pending"},
		}
		newCommitteeList := []CandidateInfo{
			{incognitokey.CommitteePublicKey{}, "4", 500, 50, "committee"},
			{incognitokey.CommitteePublicKey{}, "5", 500, 50, "committee"},
		}
		return testcase{
			name, args, newpendingList, newCommitteeList,
		}
	}

	testcase4 := func() testcase {
		name := "swap normal"
		args := args{
			[]CandidateInfo{
				{incognitokey.CommitteePublicKey{}, "1", 600, 100, "pending"},
				{incognitokey.CommitteePublicKey{}, "2", 700, 200, "pending"},
				{incognitokey.CommitteePublicKey{}, "3", 450, 500, "pending"},
			},
			[]CandidateInfo{
				{incognitokey.CommitteePublicKey{}, "4", 550, 50, "committee"},
				{incognitokey.CommitteePublicKey{}, "5", 500, 50, "committee"},
			},
			3000,
			2,
		}
		newpendingList := []CandidateInfo{
			{incognitokey.CommitteePublicKey{}, "4", 550, 50, "committee"},
			{incognitokey.CommitteePublicKey{}, "5", 500, 50, "committee"},
			{incognitokey.CommitteePublicKey{}, "3", 450, 500, "pending"},
		}
		newCommitteeList := []CandidateInfo{
			{incognitokey.CommitteePublicKey{}, "2", 700, 200, "pending"},
			{incognitokey.CommitteePublicKey{}, "1", 600, 100, "pending"},
		}
		return testcase{
			name, args, newpendingList, newCommitteeList,
		}
	}
	tests := []testcase{
		testcase1(),
		testcase2(),
		testcase3(),
		testcase4(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := beacon_swap_v1(tt.args.pendingList, tt.args.committeeList, tt.args.fixNodeStakingPower, tt.args.maxCommitteeSize)
			fmt.Println("pending", got)
			fmt.Println("committee", got1)
			assert.Equalf(t, tt.want, got, "beacon_swap_v1(%v, %v, %v, %v)", tt.args.pendingList, tt.args.committeeList, tt.args.fixNodeStakingPower, tt.args.maxCommitteeSize)
			assert.Equalf(t, tt.want1, got1, "beacon_swap_v1(%v, %v, %v, %v)", tt.args.pendingList, tt.args.committeeList, tt.args.fixNodeStakingPower, tt.args.maxCommitteeSize)
		})
	}
}

func dummyKey() incognitokey.CommitteePublicKey {
	randIncPK := privacy.RandomPoint().GetKey()
	newCPK, err := incognitokey.NewCommitteeKeyFromSeed(common.RandBytes(10), randIncPK[:])
	if err != nil {
		panic(err)
	}
	return newCPK
}

func dummyBeaconCommittee(fixedSize, cSize int, nonFixedPerf, nonFixedSAmount []uint64) ([]string, map[string]*StakerInfo) {
	res := map[string]*StakerInfo{}
	StakingAmountShard := uint64(1750 * 10e9)
	keys := []string{}
	for i := 0; i < fixedSize; i++ {
		dK := dummyKey()
		dKString, _ := dK.ToBase58()
		keys = append(keys, dKString)
		res[dKString] = &StakerInfo{
			cpkStr:        dK,
			StakingAmount: 120 * StakingAmountShard,
			Unstake:       true,
			Performance:   1000,
			EpochScore:    StakingAmountShard * 120,
			FixedNode:     true,
			EnterTime:     time.Now(),
		}
	}
	for i := fixedSize; i < fixedSize+cSize; i++ {
		dK := dummyKey()
		dKString, _ := dK.ToBase58()
		keys = append(keys, dKString)
		res[dKString] = &StakerInfo{
			cpkStr:        dK,
			StakingAmount: nonFixedSAmount[i-fixedSize],
			Unstake:       false,
			Performance:   nonFixedPerf[i-fixedSize],
			EpochScore:    nonFixedPerf[i-fixedSize] * nonFixedSAmount[i-fixedSize] / 1000,
			FixedNode:     false,
			EnterTime:     time.Now(),
		}
	}
	return keys, res
}

func dummyBeaconCandidate(cSize int, perf, sAmount []uint64) ([]string, map[string]*StakerInfo) {
	res := map[string]*StakerInfo{}
	keys := []string{}
	for i := 0; i < cSize; i++ {
		dK := dummyKey()
		dKString, _ := dK.ToBase58()
		keys = append(keys, dKString)
		res[dKString] = &StakerInfo{
			cpkStr:        dK,
			StakingAmount: sAmount[i],
			Unstake:       false,
			Performance:   perf[i],
			EpochScore:    perf[i] * sAmount[i] / 1000,
			FixedNode:     false,
			EnterTime:     time.Now(),
		}
	}
	return keys, res
}

func TestBeaconCommitteeStateV4_beaconRemoveAndSwap(t *testing.T) {
	type fields struct {
		config                 BeaconCommitteeStateV4Config
		BeaconCommitteeStateV3 *BeaconCommitteeStateV3
		beaconCommittee        map[string]*StakerInfo
		beaconPending          map[string]*StakerInfo
		beaconWaiting          map[string]*StakerInfo
		beaconLocking          map[string]*LockingInfo
		stateDB                *statedb.StateDB
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]bool
		want1   map[string]bool
		want2   map[string]incognitokey.CommitteePublicKey
		want3   map[string]incognitokey.CommitteePublicKey
		wantErr bool
	}{
		{
			name: "Case swap in all until max",
			fields: fields{
				config: BeaconCommitteeStateV4Config{
					MIN_PERFORMANCE:    100,
					DEFAULT_PERFORMING: 500,
				},
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{},
				beaconLocking:          map[string]*LockingInfo{},
				stateDB:                &statedb.StateDB{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					MaxBeaconCommitteeSize: 8,
				},
			},
		},
	}
	StakingAmountShard := uint64(1750 * 10e9)

	buildTestEnv := func(testCase *struct {
		name    string
		fields  fields
		args    args
		want    map[string]bool
		want1   map[string]bool
		want2   map[string]incognitokey.CommitteePublicKey
		want3   map[string]incognitokey.CommitteePublicKey
		wantErr bool
	}) {
		switch testCase.name {
		case "Case swap in all until max":
			keysCommittee, committee := dummyBeaconCommittee(
				4,
				0,
				[]uint64{
					800,
					700,
					600,
					500,
				},
				[]uint64{
					55 * StakingAmountShard,
					54 * StakingAmountShard,
					53 * StakingAmountShard,
					52 * StakingAmountShard,
				},
			)
			keysPending, pending := dummyBeaconCandidate(
				8,
				[]uint64{
					500,
					500,
					500,
					500,
					500,
					500,
					500,
					500,
				},
				[]uint64{
					65 * StakingAmountShard,
					64 * StakingAmountShard,
					63 * StakingAmountShard,
					62 * StakingAmountShard,
					55 * StakingAmountShard,
					54 * StakingAmountShard,
					53 * StakingAmountShard,
					52 * StakingAmountShard,
				},
			)
			keysWaiting, waiting := dummyBeaconCandidate(
				8,
				[]uint64{
					500,
					500,
					500,
					500,
					500,
					500,
					500,
					500,
				},
				[]uint64{
					50 * StakingAmountShard,
					50 * StakingAmountShard,
					50 * StakingAmountShard,
					50 * StakingAmountShard,
					50 * StakingAmountShard,
					50 * StakingAmountShard,
					50 * StakingAmountShard,
					50 * StakingAmountShard,
				},
			)
			testCase.fields.beaconCommittee = committee
			testCase.fields.beaconPending = pending
			testCase.fields.beaconWaiting = waiting
			slashPK := map[string]bool{}
			unstake := map[string]bool{}
			newCommittee := map[string]incognitokey.CommitteePublicKey{}
			for _, k := range keysCommittee {
				newCommittee[k] = committee[k].cpkStr
			}
			fmt.Println(len(keysCommittee))
			newPending := map[string]incognitokey.CommitteePublicKey{}
			for idx, k := range keysPending {
				if idx < 3 {
					newCommittee[k] = pending[k].cpkStr
				} else {
					newPending[k] = pending[k].cpkStr
				}
			}
			_ = keysWaiting
			testCase.want = slashPK
			testCase.want1 = unstake
			testCase.want2 = newCommittee
			testCase.want3 = newPending
			testCase.wantErr = false
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildTestEnv(&tt)
			s := &BeaconCommitteeStateV4{
				config:                 tt.fields.config,
				BeaconCommitteeStateV3: tt.fields.BeaconCommitteeStateV3,
				beaconCommittee:        tt.fields.beaconCommittee,
				beaconPending:          tt.fields.beaconPending,
				beaconWaiting:          tt.fields.beaconWaiting,
				beaconLocking:          tt.fields.beaconLocking,
				stateDB:                tt.fields.stateDB,
			}
			got, got1, got2, got3, err := s.beacon_swap_v1(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV4.beaconRemoveAndSwap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV4.beaconRemoveAndSwap() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeStateV4.beaconRemoveAndSwap() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeStateV4.beaconRemoveAndSwap() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("BeaconCommitteeStateV4.beaconRemoveAndSwap() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}

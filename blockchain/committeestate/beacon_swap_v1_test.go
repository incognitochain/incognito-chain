package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/stretchr/testify/assert"
	"testing"
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
		name := "only fixnode"
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

	tests := []testcase{
		testcase1(),
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

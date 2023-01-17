package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_beacon_swap_v1(t *testing.T) {

	type args struct {
		pendingList      []CandidateInfoV1
		committeeList    []CandidateInfoV1
		numberFixNode    int
		maxCommitteeSize int
	}

	type testcase struct {
		name  string
		args  args
		want  []CandidateInfoV1
		want1 []CandidateInfoV1
	}

	testcase1 := func() testcase {
		name := "swap in all until max"
		args := args{
			[]CandidateInfoV1{
				{incognitokey.CommitteePublicKey{}, "1", 500, "pending"},
				{incognitokey.CommitteePublicKey{}, "2", 500, "pending"},
				{incognitokey.CommitteePublicKey{}, "3", 500, "pending"},
			},
			[]CandidateInfoV1{},
			4,
			2,
		}
		newpendingList := []CandidateInfoV1{
			{incognitokey.CommitteePublicKey{}, "2", 500, "pending"},
			{incognitokey.CommitteePublicKey{}, "3", 500, "pending"},
		}
		newCommitteeList := []CandidateInfoV1{
			{incognitokey.CommitteePublicKey{}, "1", 500, "pending"},
		}
		return testcase{
			name, args, newpendingList, newCommitteeList,
		}
	}
	testcase2 := func() testcase {
		name := "swap in all until max 2"
		args := args{
			[]CandidateInfoV1{
				{incognitokey.CommitteePublicKey{}, "1", 500, "pending"},
				{incognitokey.CommitteePublicKey{}, "2", 500, "pending"},
				{incognitokey.CommitteePublicKey{}, "3", 500, "pending"},
			},
			[]CandidateInfoV1{
				{incognitokey.CommitteePublicKey{}, "4", 500, "committee"},
			},
			4,
			2,
		}
		newpendingList := []CandidateInfoV1{
			{incognitokey.CommitteePublicKey{}, "2", 500, "pending"},
			{incognitokey.CommitteePublicKey{}, "3", 500, "pending"},
		}
		newCommitteeList := []CandidateInfoV1{
			{incognitokey.CommitteePublicKey{}, "4", 500, "committee"},
			{incognitokey.CommitteePublicKey{}, "1", 500, "pending"},
		}
		return testcase{
			name, args, newpendingList, newCommitteeList,
		}
	}
	testcase3 := func() testcase {
		name := "no swap (no higher score)"
		args := args{
			[]CandidateInfoV1{
				{incognitokey.CommitteePublicKey{}, "1", 400, "pending"},
				{incognitokey.CommitteePublicKey{}, "2", 500, "pending"},
				{incognitokey.CommitteePublicKey{}, "3", 450, "pending"},
			},
			[]CandidateInfoV1{
				{incognitokey.CommitteePublicKey{}, "4", 500, "committee"},
				{incognitokey.CommitteePublicKey{}, "5", 500, "committee"},
			},
			4,
			2,
		}
		newpendingList := []CandidateInfoV1{
			{incognitokey.CommitteePublicKey{}, "2", 500, "pending"},
			{incognitokey.CommitteePublicKey{}, "3", 450, "pending"},
			{incognitokey.CommitteePublicKey{}, "1", 400, "pending"},
		}
		newCommitteeList := []CandidateInfoV1{
			{incognitokey.CommitteePublicKey{}, "4", 500, "committee"},
			{incognitokey.CommitteePublicKey{}, "5", 500, "committee"},
		}
		return testcase{
			name, args, newpendingList, newCommitteeList,
		}
	}

	testcase4 := func() testcase {
		name := "swap normal 1"
		args := args{
			[]CandidateInfoV1{
				{incognitokey.CommitteePublicKey{}, "1", 600, "pending"},
				{incognitokey.CommitteePublicKey{}, "2", 700, "pending"},
				{incognitokey.CommitteePublicKey{}, "3", 450, "pending"},
			},
			[]CandidateInfoV1{
				{incognitokey.CommitteePublicKey{}, "5", 500, "committee"},
				{incognitokey.CommitteePublicKey{}, "4", 550, "committee"},
			},
			4,
			2,
		}
		newpendingList := []CandidateInfoV1{
			{incognitokey.CommitteePublicKey{}, "1", 600, "pending"},
			{incognitokey.CommitteePublicKey{}, "5", 500, "committee"},
			{incognitokey.CommitteePublicKey{}, "3", 450, "pending"},
		}
		newCommitteeList := []CandidateInfoV1{
			{incognitokey.CommitteePublicKey{}, "2", 700, "pending"},
			{incognitokey.CommitteePublicKey{}, "4", 550, "committee"},
		}
		return testcase{
			name, args, newpendingList, newCommitteeList,
		}
	}

	testcase5 := func() testcase {
		name := "swap normal 1"
		args := args{
			[]CandidateInfoV1{
				{incognitokey.CommitteePublicKey{}, "1", 600, "pending"},
				{incognitokey.CommitteePublicKey{}, "2", 700, "pending"},
				{incognitokey.CommitteePublicKey{}, "3", 450, "pending"},
			},
			[]CandidateInfoV1{
				{incognitokey.CommitteePublicKey{}, "5", 500, "committee"},
				{incognitokey.CommitteePublicKey{}, "4", 550, "committee"},
			},
			5,
			2,
		}
		newpendingList := []CandidateInfoV1{
			{incognitokey.CommitteePublicKey{}, "4", 550, "committee"},
			{incognitokey.CommitteePublicKey{}, "5", 500, "committee"},
			{incognitokey.CommitteePublicKey{}, "3", 450, "pending"},
		}
		newCommitteeList := []CandidateInfoV1{
			{incognitokey.CommitteePublicKey{}, "2", 700, "pending"},
			{incognitokey.CommitteePublicKey{}, "1", 600, "pending"},
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
		testcase5(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := beacon_swap_v1(tt.args.pendingList, tt.args.committeeList, tt.args.numberFixNode, tt.args.maxCommitteeSize)
			fmt.Println("pending", got)
			fmt.Println("committee", got1)
			assert.Equalf(t, tt.want, got, "beacon_swap_v1(%v, %v, %v, %v)", tt.args.pendingList, tt.args.committeeList, tt.args.numberFixNode, tt.args.maxCommitteeSize)
			assert.Equalf(t, tt.want1, got1, "beacon_swap_v1(%v, %v, %v, %v)", tt.args.pendingList, tt.args.committeeList, tt.args.numberFixNode, tt.args.maxCommitteeSize)
		})
	}
}

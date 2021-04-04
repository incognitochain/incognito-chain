package committeestate

import (
	"reflect"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

func Test_beaconCommitteeStateSlashingBase_clone(t *testing.T) {

	initTestParams()
	initLog()

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	hash, _ := common.Hash{}.NewHashFromStr("123")
	hash6, _ := common.Hash{}.NewHashFromStr("456")

	mutex := &sync.RWMutex{}

	type fields struct {
		beaconCommitteeStateBase   beaconCommitteeStateBase
		shardCommonPool            []string
		numberOfAssignedCandidates int
		swapRule                   SwapRuleProcessor
	}
	tests := []struct {
		name   string
		fields fields
		want   *beaconCommitteeStateSlashingBase
	}{
		{
			name: "[valid input] full data",
			fields: fields{
				beaconCommitteeStateBase: beaconCommitteeStateBase{
					beaconCommittee: []string{
						key6, key7, key8,
					},
					shardCommittee: map[byte][]string{
						0: []string{
							key3, key4, key5,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key,
						},
					},
					autoStake: map[string]bool{
						key:  true,
						key8: false,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey.GetIncKeyBase58(): paymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key:  *hash,
						key6: *hash6,
					},
					mu: mutex,
				},
				numberOfAssignedCandidates: 1,
				shardCommonPool: []string{
					key2,
				},
			},
			want: &beaconCommitteeStateSlashingBase{
				beaconCommitteeStateBase: beaconCommitteeStateBase{
					beaconCommittee: []string{
						key6, key7, key8,
					},
					shardCommittee: map[byte][]string{
						0: []string{
							key3, key4, key5,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key,
						},
					},
					autoStake: map[string]bool{
						key:  true,
						key8: false,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey.GetIncKeyBase58(): paymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key:  *hash,
						key6: *hash6,
					},
					mu: mutex,
				},
				numberOfAssignedCandidates: 1,
				shardCommonPool: []string{
					key2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := beaconCommitteeStateSlashingBase{
				beaconCommitteeStateBase:   tt.fields.beaconCommitteeStateBase,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				swapRule:                   tt.fields.swapRule,
			}
			if got := b.clone(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("beaconCommitteeStateV2.clone() = %v, want %v", got, tt.want)
			}
		})
	}
}

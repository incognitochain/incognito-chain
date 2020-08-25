package committeestate

import (
	"reflect"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

func TestBeaconCommitteeStateV1_processUnstakeInstruction(t *testing.T) {
	type fields struct {
		beaconCommittee             []incognitokey.CommitteePublicKey
		beaconSubstitute            []incognitokey.CommitteePublicKey
		nextEpochShardCandidate     []incognitokey.CommitteePublicKey
		currentEpochShardCandidate  []incognitokey.CommitteePublicKey
		nextEpochBeaconCandidate    []incognitokey.CommitteePublicKey
		currentEpochBeaconCandidate []incognitokey.CommitteePublicKey
		shardCommittee              map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute             map[byte][]incognitokey.CommitteePublicKey
		autoStake                   map[string]bool
		rewardReceiver              map[string]privacy.PaymentAddress
		stakingTx                   map[string]common.Hash
		mu                          *sync.RWMutex
	}
	type args struct {
		unstakeInstruction *instruction.UnstakeInstruction
		env                *BeaconCommitteeStateEnvironment
		committeeChange    *CommitteeChange
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		want1   [][]string
		wantErr bool
	}{
		{},
		{},
		{},
		{},
		{},
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV1{
				beaconCommittee:             tt.fields.beaconCommittee,
				beaconSubstitute:            tt.fields.beaconSubstitute,
				nextEpochShardCandidate:     tt.fields.nextEpochShardCandidate,
				currentEpochShardCandidate:  tt.fields.currentEpochShardCandidate,
				nextEpochBeaconCandidate:    tt.fields.nextEpochBeaconCandidate,
				currentEpochBeaconCandidate: tt.fields.currentEpochBeaconCandidate,
				shardCommittee:              tt.fields.shardCommittee,
				shardSubstitute:             tt.fields.shardSubstitute,
				autoStake:                   tt.fields.autoStake,
				rewardReceiver:              tt.fields.rewardReceiver,
				stakingTx:                   tt.fields.stakingTx,
				mu:                          tt.fields.mu,
			}
			got, got1, err := b.processUnstakeInstruction(tt.args.unstakeInstruction, tt.args.env, tt.args.committeeChange)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeInstruction() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeStateV1.processUnstakeInstruction() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

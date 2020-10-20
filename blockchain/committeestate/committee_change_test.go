package committeestate

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

func TestCommitteeChange_GetStakerKeys(t *testing.T) {

	initPublicKey()

	type fields struct {
		NextEpochBeaconCandidateAdded      []incognitokey.CommitteePublicKey
		NextEpochBeaconCandidateRemoved    []incognitokey.CommitteePublicKey
		CurrentEpochBeaconCandidateAdded   []incognitokey.CommitteePublicKey
		CurrentEpochBeaconCandidateRemoved []incognitokey.CommitteePublicKey
		NextEpochShardCandidateAdded       []incognitokey.CommitteePublicKey
		NextEpochShardCandidateRemoved     []incognitokey.CommitteePublicKey
		CurrentEpochShardCandidateAdded    []incognitokey.CommitteePublicKey
		CurrentEpochShardCandidateRemoved  []incognitokey.CommitteePublicKey
		ShardSubstituteAdded               map[byte][]incognitokey.CommitteePublicKey
		ShardSubstituteRemoved             map[byte][]incognitokey.CommitteePublicKey
		ShardCommitteeAdded                map[byte][]incognitokey.CommitteePublicKey
		ShardCommitteeRemoved              map[byte][]incognitokey.CommitteePublicKey
		BeaconSubstituteAdded              []incognitokey.CommitteePublicKey
		BeaconSubstituteRemoved            []incognitokey.CommitteePublicKey
		BeaconCommitteeAdded               []incognitokey.CommitteePublicKey
		BeaconCommitteeRemoved             []incognitokey.CommitteePublicKey
		BeaconCommitteeReplaced            [2][]incognitokey.CommitteePublicKey
		ShardCommitteeReplaced             map[byte][2][]incognitokey.CommitteePublicKey
		StopAutoStake                      []string
		Unstake                            []string
	}
	tests := []struct {
		name   string
		fields fields
		want   []incognitokey.CommitteePublicKey
	}{
		{
			name: "Valid Input",
			fields: fields{
				NextEpochShardCandidateAdded: []incognitokey.CommitteePublicKey{
					*incKey, *incKey2, *incKey3, *incKey4,
				},
			},
			want: []incognitokey.CommitteePublicKey{
				*incKey, *incKey2, *incKey3, *incKey4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			committeeChange := &CommitteeChange{
				NextEpochBeaconCandidateAdded:      tt.fields.NextEpochBeaconCandidateAdded,
				NextEpochBeaconCandidateRemoved:    tt.fields.NextEpochBeaconCandidateRemoved,
				CurrentEpochBeaconCandidateAdded:   tt.fields.CurrentEpochBeaconCandidateAdded,
				CurrentEpochBeaconCandidateRemoved: tt.fields.CurrentEpochBeaconCandidateRemoved,
				NextEpochShardCandidateAdded:       tt.fields.NextEpochShardCandidateAdded,
				NextEpochShardCandidateRemoved:     tt.fields.NextEpochShardCandidateRemoved,
				CurrentEpochShardCandidateAdded:    tt.fields.CurrentEpochShardCandidateAdded,
				CurrentEpochShardCandidateRemoved:  tt.fields.CurrentEpochShardCandidateRemoved,
				ShardSubstituteAdded:               tt.fields.ShardSubstituteAdded,
				ShardSubstituteRemoved:             tt.fields.ShardSubstituteRemoved,
				ShardCommitteeAdded:                tt.fields.ShardCommitteeAdded,
				ShardCommitteeRemoved:              tt.fields.ShardCommitteeRemoved,
				BeaconSubstituteAdded:              tt.fields.BeaconSubstituteAdded,
				BeaconSubstituteRemoved:            tt.fields.BeaconSubstituteRemoved,
				BeaconCommitteeAdded:               tt.fields.BeaconCommitteeAdded,
				BeaconCommitteeRemoved:             tt.fields.BeaconCommitteeRemoved,
				BeaconCommitteeReplaced:            tt.fields.BeaconCommitteeReplaced,
				ShardCommitteeReplaced:             tt.fields.ShardCommitteeReplaced,
				StopAutoStake:                      tt.fields.StopAutoStake,
				Unstake:                            tt.fields.Unstake,
			}
			if got := committeeChange.GetStakerKeys(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CommitteeChange.GetStakerKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommitteeChange_GetUnstakerKeys(t *testing.T) {

	initPublicKey()

	type fields struct {
		NextEpochBeaconCandidateAdded      []incognitokey.CommitteePublicKey
		NextEpochBeaconCandidateRemoved    []incognitokey.CommitteePublicKey
		CurrentEpochBeaconCandidateAdded   []incognitokey.CommitteePublicKey
		CurrentEpochBeaconCandidateRemoved []incognitokey.CommitteePublicKey
		NextEpochShardCandidateAdded       []incognitokey.CommitteePublicKey
		NextEpochShardCandidateRemoved     []incognitokey.CommitteePublicKey
		CurrentEpochShardCandidateAdded    []incognitokey.CommitteePublicKey
		CurrentEpochShardCandidateRemoved  []incognitokey.CommitteePublicKey
		ShardSubstituteAdded               map[byte][]incognitokey.CommitteePublicKey
		ShardSubstituteRemoved             map[byte][]incognitokey.CommitteePublicKey
		ShardCommitteeAdded                map[byte][]incognitokey.CommitteePublicKey
		ShardCommitteeRemoved              map[byte][]incognitokey.CommitteePublicKey
		BeaconSubstituteAdded              []incognitokey.CommitteePublicKey
		BeaconSubstituteRemoved            []incognitokey.CommitteePublicKey
		BeaconCommitteeAdded               []incognitokey.CommitteePublicKey
		BeaconCommitteeRemoved             []incognitokey.CommitteePublicKey
		BeaconCommitteeReplaced            [2][]incognitokey.CommitteePublicKey
		ShardCommitteeReplaced             map[byte][2][]incognitokey.CommitteePublicKey
		StopAutoStake                      []string
		Unstake                            []string
	}
	type args struct {
		autoStake map[string]bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []incognitokey.CommitteePublicKey
		want1  []incognitokey.CommitteePublicKey
	}{
		{
			name: "Valid Input",
			fields: fields{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{
					*incKey,
				},
				Unstake: []string{
					key, key2,
				},
				ShardCommitteeRemoved: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey3,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey4,
					},
				},
			},
			args: args{
				autoStake: map[string]bool{
					key2: false,
					key3: false,
					key4: true,
				},
			},
			want: []incognitokey.CommitteePublicKey{
				*incKey, *incKey3,
			},
			want1: []incognitokey.CommitteePublicKey{
				*incKey2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			committeeChange := &CommitteeChange{
				NextEpochBeaconCandidateAdded:      tt.fields.NextEpochBeaconCandidateAdded,
				NextEpochBeaconCandidateRemoved:    tt.fields.NextEpochBeaconCandidateRemoved,
				CurrentEpochBeaconCandidateAdded:   tt.fields.CurrentEpochBeaconCandidateAdded,
				CurrentEpochBeaconCandidateRemoved: tt.fields.CurrentEpochBeaconCandidateRemoved,
				NextEpochShardCandidateAdded:       tt.fields.NextEpochShardCandidateAdded,
				NextEpochShardCandidateRemoved:     tt.fields.NextEpochShardCandidateRemoved,
				CurrentEpochShardCandidateAdded:    tt.fields.CurrentEpochShardCandidateAdded,
				CurrentEpochShardCandidateRemoved:  tt.fields.CurrentEpochShardCandidateRemoved,
				ShardSubstituteAdded:               tt.fields.ShardSubstituteAdded,
				ShardSubstituteRemoved:             tt.fields.ShardSubstituteRemoved,
				ShardCommitteeAdded:                tt.fields.ShardCommitteeAdded,
				ShardCommitteeRemoved:              tt.fields.ShardCommitteeRemoved,
				BeaconSubstituteAdded:              tt.fields.BeaconSubstituteAdded,
				BeaconSubstituteRemoved:            tt.fields.BeaconSubstituteRemoved,
				BeaconCommitteeAdded:               tt.fields.BeaconCommitteeAdded,
				BeaconCommitteeRemoved:             tt.fields.BeaconCommitteeRemoved,
				BeaconCommitteeReplaced:            tt.fields.BeaconCommitteeReplaced,
				ShardCommitteeReplaced:             tt.fields.ShardCommitteeReplaced,
				StopAutoStake:                      tt.fields.StopAutoStake,
				Unstake:                            tt.fields.Unstake,
			}
			got, got1 := committeeChange.GetUnstakerKeys(tt.args.autoStake)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CommitteeChange.GetUnstakerKeys() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("CommitteeChange.GetUnstakerKeys() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

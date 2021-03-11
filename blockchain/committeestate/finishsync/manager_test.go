package finishsync_test

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate/finishsync"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func TestManager_AddFinishedSyncValidators(t *testing.T) {
	initTestParams()
	type fields struct {
		validators map[byte][]incognitokey.CommitteePublicKey
	}
	type args struct {
		validators        []string
		syncingValidators []incognitokey.CommitteePublicKey
		shardID           byte
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
	}{
		{
			name: "Valid Input",
			fields: fields{
				validators: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey,
					},
				},
			},
			fieldsAfterProcess: fields{
				validators: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
				},
			},
			args: args{
				validators: []string{
					key2, key3, key4,
				},
				syncingValidators: []incognitokey.CommitteePublicKey{
					*incKey0, *incKey, *incKey2, *incKey3,
				},
				shardID: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := finishsync.NewManagerWithValue(tt.fields.validators)
			manager.AddFinishedSyncValidators(tt.args.validators, tt.args.syncingValidators, tt.args.shardID)
			if !reflect.DeepEqual(manager.Validators(tt.args.shardID), tt.fieldsAfterProcess.validators[tt.args.shardID]) {
				t.Errorf("validators = %v, want %v", manager.Validators(tt.args.shardID), tt.fieldsAfterProcess.validators[tt.args.shardID])
			}
		})
	}
}

func TestManager_RemoveValidators(t *testing.T) {
	initTestParams()
	type fields struct {
		validators map[byte][]incognitokey.CommitteePublicKey
	}
	type args struct {
		validators []incognitokey.CommitteePublicKey
		shardID    byte
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
	}{
		{
			name: "Valid Input",
			fields: fields{
				validators: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3},
				},
			},
			fieldsAfterProcess: fields{
				validators: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{*incKey, *incKey3},
				},
			},
			args: args{
				validators: []incognitokey.CommitteePublicKey{*incKey0, *incKey2},
				shardID:    1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := finishsync.NewManagerWithValue(tt.fields.validators)
			manager.RemoveValidators(tt.args.validators, tt.args.shardID)
			if !reflect.DeepEqual(manager.Validators(tt.args.shardID), tt.fieldsAfterProcess.validators[tt.args.shardID]) {
				t.Errorf("validators = %v, want %v", manager.Validators(tt.args.shardID), tt.fieldsAfterProcess.validators[tt.args.shardID])
			}
		})
	}
}

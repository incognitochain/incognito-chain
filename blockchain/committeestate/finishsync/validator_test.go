package finishsync_test

import (
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

func TestManager_AddFinishedSyncValidators(t *testing.T) {
	type fields struct {
		validators map[byte][]incognitokey.CommitteePublicKey
		mu         *sync.RWMutex
	}
	type args struct {
		validators        []string
		syncingValidators []incognitokey.CommitteePublicKey
		shardID           byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{
				validators: tt.fields.validators,
				mu:         tt.fields.mu,
			}
			manager.AddFinishedSyncValidators(tt.args.validators, tt.args.syncingValidators, tt.args.shardID)
		})
	}
}

func TestManager_RemoveValidators(t *testing.T) {
	type fields struct {
		validators map[byte][]incognitokey.CommitteePublicKey
		mu         *sync.RWMutex
	}
	type args struct {
		validators []incognitokey.CommitteePublicKey
		shardID    byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{
				validators: tt.fields.validators,
				mu:         tt.fields.mu,
			}
			manager.RemoveValidators(tt.args.validators, tt.args.shardID)
		})
	}
}

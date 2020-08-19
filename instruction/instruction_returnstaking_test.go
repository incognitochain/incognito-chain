package instruction

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func TestValidateReturnStakingInstructionSanity(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateReturnStakingInstructionSanity(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("ValidateReturnStakingInstructionSanity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImportReturnStakingInstructionFromString(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		want    *ReturnStakeIns
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ImportReturnStakingInstructionFromString(tt.args.instruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("ImportReturnStakingInstructionFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ImportReturnStakingInstructionFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAndImportReturnStakingInstructionFromString(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		want    *ReturnStakeIns
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndImportReturnStakingInstructionFromString(tt.args.instruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndImportReturnStakingInstructionFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateAndImportReturnStakingInstructionFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReturnStakeIns_ToString(t *testing.T) {
	type fields struct {
		PublicKeys       []string
		PublicKeysStruct []incognitokey.CommitteePublicKey
		ShardID          byte
		StakingTXIDs     []string
		StakingTxHashes  []common.Hash
		PercentReturns   []uint
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsI := &ReturnStakeIns{
				PublicKeys:       tt.fields.PublicKeys,
				PublicKeysStruct: tt.fields.PublicKeysStruct,
				ShardID:          tt.fields.ShardID,
				StakingTXIDs:     tt.fields.StakingTXIDs,
				StakingTxHashes:  tt.fields.StakingTxHashes,
				PercentReturns:   tt.fields.PercentReturns,
			}
			if got := rsI.ToString(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReturnStakeIns.ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

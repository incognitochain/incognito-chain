package instruction

import (
	"reflect"
	"strings"
	"testing"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

func TestValidateAndImportStakeInstructionFromString(t *testing.T) {

	initPublicKey()

	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		want    *StakeInstruction
		wantErr bool
	}{
		{
			name: "len(instruction) != 6",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "instruction[0] != STAKE_ACTION",
			args: args{
				instruction: []string{ASSIGN_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Invalid chain id",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					"test",
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Invalid public key type",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{"key1", "key2", "key3", "key4"}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Length of tx stakes and public key is not similar",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Length of public key and reward address is not similar",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Length of public key and tx stop auto staking before is not similar",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"}, SPLITTER),
					strings.Join([]string{"true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Length of tx stakes and reward address is not similar",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Length of tx stakes and stop auto staking request is not similar",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"}, SPLITTER),
					strings.Join([]string{"true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Length of reward address and stop auto staking request is not similar",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Valid Input",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			want: &StakeInstruction{
				PublicKeys: []string{key1, key2, key3, key4},
				PublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey1,
					*incKey2,
					*incKey3,
					*incKey4,
				},
				Chain:           SHARD_INST,
				TxStakes:        []string{"tx1", "tx2", "tx3", "tx4"},
				RewardReceivers: []string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"},
				AutoStakingFlag: []bool{true, true, true, true},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndImportStakeInstructionFromString(tt.args.instruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndImportStakeInstructionFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateAndImportStakeInstructionFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

//StakeInstruction Format:
// ["STAKE_ACTION", list_public_keys, chain or beacon, list_txs, list_reward_addresses, list_autostaking_status(boolean)]

func TestValidateStakeInstructionSanity(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "len(instruction) != 6",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "instruction[0] != STAKE_ACTION",
			args: args{
				instruction: []string{ASSIGN_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Invalid chain id",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					"test",
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Invalid public key type",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{"key1", "key2", "key3", "key4"}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Length of tx stakes and public key is not similar",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Length of public key and reward address is not similar",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Length of public key and tx stop auto staking before is not similar",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"}, SPLITTER),
					strings.Join([]string{"true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Length of tx stakes and reward address is not similar",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Length of tx stakes and stop auto staking request is not similar",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"}, SPLITTER),
					strings.Join([]string{"true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Length of reward address and stop auto staking request is not similar",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Valid Input",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateStakeInstructionSanity(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("ValidateStakeInstructionSanity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImportStakeInstructionFromString(t *testing.T) {

	initPublicKey()

	type args struct {
		instruction []string
	}
	tests := []struct {
		name string
		args args
		want *StakeInstruction
	}{
		{
			name: "Valid Input",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					SHARD_INST,
					strings.Join([]string{"tx1", "tx2", "tx3", "tx4"}, SPLITTER),
					strings.Join([]string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			want: &StakeInstruction{
				PublicKeys: []string{key1, key2, key3, key4},
				PublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey1,
					*incKey2,
					*incKey3,
					*incKey4,
				},
				Chain:           SHARD_INST,
				TxStakes:        []string{"tx1", "tx2", "tx3", "tx4"},
				RewardReceivers: []string{"reward-addr1", "reward-addr2", "reward-addr3", "reward-addr4"},
				AutoStakingFlag: []bool{true, true, true, true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ImportStakeInstructionFromString(tt.args.instruction); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ImportStakeInstructionFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

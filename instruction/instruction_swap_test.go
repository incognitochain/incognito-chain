package instruction

import (
	"reflect"
	"strings"
	"testing"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

//Swap instruction format:
//["swap-action", list-keys-in, list-keys-out, shard or beacon chain, shard_id(optional), "punished public key", "new reward receivers"]

func TestValidateSwapInstructionSanity(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Length of instructions is between from 5 to 7",
			args: args{
				instruction: []string{},
			},
			wantErr: true,
		},
		{
			name: "instruction[0] is not swap action",
			args: args{
				instruction: []string{
					ASSIGN_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					SHARD_INST,
					"0",
					"punished public keys",
					"new reward receivers"},
			},
			wantErr: true,
		},
		{
			name: "len(instruction) == 5 && instruction[3] != BEACON_INST",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					SHARD_INST,
					"0"},
			},
			wantErr: true,
		},
		{
			name: "len(instruction) == 6 & instruction[3] != SHARD_INST",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					"test",
					"0",
					"new reward receivers"},
			},
			wantErr: true,
		},
		{
			name: "len(instruction) == 6 & instruction[3] == SHARD_INST & instruction[4] is not integer (shard_id)",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					SHARD_INST,
					"ads",
					"new reward receivers"},
			},
			wantErr: true,
		},
		{
			name: "instruction[1] is not type of incognito public key",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{"key1", "key2"}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					SHARD_INST,
					"0",
					"punished public keys",
					"new reward receivers"},
			},
			wantErr: true,
		},
		{
			name: "instruction[2] is not type of incognito public key",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{"key3", "key4"}, SPLITTER),
					SHARD_INST,
					"0",
					"punished public keys",
					"new reward receivers"},
			},
			wantErr: true,
		},
		{
			name: "Valid Input",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					SHARD_INST,
					"0",
					"punished public keys",
					"new reward receivers"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateSwapInstructionSanity(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("ValidateSwapInstructionSanity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAndImportSwapInstructionFromString(t *testing.T) {

	initPublicKey()

	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		want    *SwapInstruction
		wantErr bool
	}{
		{
			name: "Length of instructions is between from 5 to 7",
			args: args{
				instruction: []string{},
			},
			wantErr: true,
		},
		{
			name: "instruction[0] is not swap action",
			args: args{
				instruction: []string{
					ASSIGN_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					SHARD_INST,
					"0",
					"punished public keys",
					strings.Join([]string{key3, key4}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "len(instruction) == 5 && instruction[3] != BEACON_INST",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					SHARD_INST,
					"0"},
			},
			wantErr: true,
		},
		{
			name: "len(instruction) == 6 & instruction[3] != SHARD_INST",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					"test",
					"0",
					strings.Join([]string{key3, key4}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "len(instruction) == 6 & instruction[3] == SHARD_INST & instruction[4] is not integer (shard_id)",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					SHARD_INST,
					"ads",
					strings.Join([]string{key3, key4}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "instruction[1] is not type of incognito public key",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{"key1", "key2"}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					SHARD_INST,
					"0",
					"punished public keys",
					strings.Join([]string{key3, key4}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "instruction[2] is not type of incognito public key",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{"key3", "key4"}, SPLITTER),
					SHARD_INST,
					"0",
					"punished public keys",
					strings.Join([]string{key3, key4}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Valid Input_shard",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					SHARD_INST,
					"0",
					"punished public keys",
					strings.Join([]string{key3, key4}, SPLITTER)},
			},
			want: &SwapInstruction{
				InPublicKeys:  []string{key1, key2},
				OutPublicKeys: []string{key3, key4},
				InPublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey1,
					*incKey2,
				},
				OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey3,
					*incKey4,
				},
				ChainID:            0,
				PunishedPublicKeys: "punished public keys",
				NewRewardReceivers: []string{key3, key4},
				IsReplace:          true,
			},
			wantErr: false,
		},
		{
			name: "Valid Input_beacon",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					BEACON_INST,
					"punished public keys",
					"",
					strings.Join([]string{key3, key4}, SPLITTER)},
			},
			want: &SwapInstruction{
				InPublicKeys:  []string{key1, key2},
				OutPublicKeys: []string{key3, key4},
				InPublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey1,
					*incKey2,
				},
				OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey3,
					*incKey4,
				},
				ChainID:            -1,
				PunishedPublicKeys: "punished public keys",
				NewRewardReceivers: []string{key3, key4},
				IsReplace:          true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndImportSwapInstructionFromString(tt.args.instruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndImportSwapInstructionFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateAndImportSwapInstructionFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportSwapInstructionFromString(t *testing.T) {

	initPublicKey()

	type args struct {
		instruction []string
	}
	tests := []struct {
		name string
		args args
		want *SwapInstruction
	}{
		{
			name: "Valid Input_beacon",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					BEACON_INST,
					"",
					"",
					strings.Join([]string{key3, key4}, SPLITTER)},
			},
			want: &SwapInstruction{
				InPublicKeys:  []string{key1, key2},
				OutPublicKeys: []string{key3, key4},
				InPublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey1,
					*incKey2,
				},
				OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey3,
					*incKey4,
				},
				ChainID:            -1,
				PunishedPublicKeys: "",
				NewRewardReceivers: []string{key3, key4},
				IsReplace:          true,
			},
		},
		{
			name: "Valid Input_shard",
			args: args{
				instruction: []string{
					SWAP_ACTION,
					strings.Join([]string{key1, key2}, SPLITTER),
					strings.Join([]string{key3, key4}, SPLITTER),
					SHARD_INST,
					"0",
					"",
					strings.Join([]string{key3, key4}, SPLITTER)},
			},
			want: &SwapInstruction{
				InPublicKeys:  []string{key1, key2},
				OutPublicKeys: []string{key3, key4},
				InPublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey1,
					*incKey2,
				},
				OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
					*incKey3,
					*incKey4,
				},
				ChainID:            0,
				PunishedPublicKeys: "",
				NewRewardReceivers: []string{key3, key4},
				IsReplace:          true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ImportSwapInstructionFromString(tt.args.instruction); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ImportSwapInstructionFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

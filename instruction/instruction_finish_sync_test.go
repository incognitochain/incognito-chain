package instruction

import (
	"reflect"
	"strings"
	"testing"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

func TestFinishSyncInstruction_ToString(t *testing.T) {

	initPublicKey()

	type fields struct {
		ChainID          int
		PublicKeys       []string
		PublicKeysStruct []incognitokey.CommitteePublicKey
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "Valid Input",
			fields: fields{
				ChainID:          1,
				PublicKeys:       []string{key1, key2},
				PublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey1, *incKey2},
			},
			want: []string{
				FINISH_SYNC_ACTION,
				"1",
				strings.Join([]string{key1, key2}, SPLITTER),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FinishSyncInstruction{
				ChainID:          tt.fields.ChainID,
				PublicKeys:       tt.fields.PublicKeys,
				PublicKeysStruct: tt.fields.PublicKeysStruct,
			}
			if got := f.ToString(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FinishSyncInstruction.ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFinishSyncInstruction_SetPublicKeys(t *testing.T) {

	initPublicKey()

	type fields struct {
		ChainID          int
		PublicKeys       []string
		PublicKeysStruct []incognitokey.CommitteePublicKey
	}
	type args struct {
		publicKeys []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *FinishSyncInstruction
	}{
		{
			name: "Valid Input",
			fields: fields{
				ChainID: 1,
			},
			args: args{
				publicKeys: []string{key1, key2},
			},
			want: &FinishSyncInstruction{
				ChainID:          1,
				PublicKeys:       []string{key1, key2},
				PublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey1, *incKey2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FinishSyncInstruction{
				ChainID:          tt.fields.ChainID,
				PublicKeys:       tt.fields.PublicKeys,
				PublicKeysStruct: tt.fields.PublicKeysStruct,
			}
			if got := f.SetPublicKeys(tt.args.publicKeys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FinishSyncInstruction.SetPublicKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAndImportFinishSyncInstructionFromString(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		want    *FinishSyncInstruction
		wantErr bool
	}{
		{
			name: "Length of instruction != 3",
			args: args{
				instruction: []string{
					FINISH_SYNC_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER), SHARD_INST},
			},
			wantErr: true,
		},
		{
			name: "instruction[0] != FINISH_SYNC_ACTION",
			args: args{
				instruction: []string{RANDOM_ACTION, strings.Join([]string{key1, key2, key3, key4}, SPLITTER), "1"},
			},
			wantErr: true,
		},
		{
			name: "instruction[1] is not integer",
			args: args{
				instruction: []string{
					FINISH_SYNC_ACTION,
					"asd",
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
				},
			},
			wantErr: true,
		},
		{
			name: "Valid Input",
			args: args{
				instruction: []string{
					FINISH_SYNC_ACTION,
					"1",
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
				},
			},
			want: &FinishSyncInstruction{
				ChainID:          1,
				PublicKeys:       []string{key1, key2, key3, key4},
				PublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey1, *incKey2, *incKey3, *incKey4},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndImportFinishSyncInstructionFromString(tt.args.instruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndImportFinishSyncInstructionFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateAndImportFinishSyncInstructionFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportFinishSyncInstructionFromString(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		want    *FinishSyncInstruction
		wantErr bool
	}{
		{
			name: "Valid Input",
			args: args{
				instruction: []string{
					FINISH_SYNC_ACTION,
					"1",
					strings.Join([]string{key1, key2}, SPLITTER),
				},
			},
			want: &FinishSyncInstruction{
				ChainID:          1,
				PublicKeys:       []string{key1, key2},
				PublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey1, *incKey2},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ImportFinishSyncInstructionFromString(tt.args.instruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("ImportFinishSyncInstructionFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ImportFinishSyncInstructionFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateFinishSyncInstructionSanity(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Length of instruction != 3",
			args: args{
				instruction: []string{
					FINISH_SYNC_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER), SHARD_INST},
			},
			wantErr: true,
		},
		{
			name: "instruction[0] != FINISH_SYNC_ACTION",
			args: args{
				instruction: []string{RANDOM_ACTION, strings.Join([]string{key1, key2, key3, key4}, SPLITTER), "1"},
			},
			wantErr: true,
		},
		{
			name: "instruction[1] is not integer",
			args: args{
				instruction: []string{
					FINISH_SYNC_ACTION,
					"asd",
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
				},
			},
			wantErr: true,
		},
		{
			name: "Valid Input",
			args: args{
				instruction: []string{
					FINISH_SYNC_ACTION,
					"1",
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateFinishSyncInstructionSanity(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("ValidateFinishSyncInstructionSanity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

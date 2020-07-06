package instruction

import (
	"reflect"
	"strings"
	"testing"
)

func TestValidateAndImportStakeInstructionFromString(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		want    *StakeInstruction
		wantErr bool
	}{
		{},
		{},
		{},
		{},
		{},
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
// ["STAKE_ACTION", "key1,key2,key3,key4", "chainID", "txstake1,txstake2,txstake3,txstake4", "rewardaddr1,rewadaddr2", "true,false,true"]

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
					"0",
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "instruction[0] != STAKE_ACTION",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					"0",
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "len(instruction) != 6",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					"0",
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "len(instruction) != 6",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					"0",
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "len(instruction) != 6",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					"0",
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{"true", "true", "true", "true"}, SPLITTER)},
			},
			wantErr: true,
		},
		{
			name: "Valid Input",
			args: args{
				instruction: []string{STAKE_ACTION,
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					"0",
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
					strings.Join([]string{key1, key2, key3, key4}, SPLITTER),
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

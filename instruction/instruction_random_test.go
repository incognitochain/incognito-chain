package instruction

import (
	"reflect"
	"testing"
)

func TestValidateRandomInstructionSanity(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid Input",
			args: args{
				instruction: []string{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateRandomInstructionSanity(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("ValidateRandomInstructionSanity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAndImportRandomInstructionFromString(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		want    *RandomInstruction
		wantErr bool
	}{
		{
			name: "Valid Input",
			args: args{
				instruction: []string{},
			},
		},
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndImportRandomInstructionFromString(tt.args.instruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndImportRandomInstructionFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateAndImportRandomInstructionFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

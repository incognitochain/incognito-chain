package instruction

import (
	"reflect"
	"testing"
)

func TestUnstakeInstruction_ToString(t *testing.T) {
	type fields struct {
		PublicKeys []string
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
			unstakeIns := &UnstakeInstruction{
				PublicKeys: tt.fields.PublicKeys,
			}
			if got := unstakeIns.ToString(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnstakeInstruction.ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAndImportUnstakeInstructionFromString(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name    string
		args    args
		want    *UnstakeInstruction
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndImportUnstakeInstructionFromString(tt.args.instruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndImportUnstakeInstructionFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateAndImportUnstakeInstructionFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportUnstakeInstructionFromString(t *testing.T) {
	type args struct {
		instruction []string
	}
	tests := []struct {
		name string
		args args
		want *UnstakeInstruction
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ImportUnstakeInstructionFromString(tt.args.instruction); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ImportUnstakeInstructionFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateUnstakeInstructionSanity(t *testing.T) {
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
			if err := ValidateUnstakeInstructionSanity(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUnstakeInstructionSanity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

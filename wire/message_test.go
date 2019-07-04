package wire

import (
	"reflect"
	"testing"
)

func TestMakeEmptyMessage(t *testing.T) {
	type args struct {
		messageType string
	}
	tests := []struct {
		name    string
		args    args
		want    Message
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeEmptyMessage(tt.args.messageType)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeEmptyMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeEmptyMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCmdType(t *testing.T) {
	type args struct {
		msgType reflect.Type
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCmdType(tt.args.msgType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCmdType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetCmdType() = %v, want %v", got, tt.want)
			}
		})
	}
}

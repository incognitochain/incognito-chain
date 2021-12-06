package v2utils

import (
	"reflect"
	"testing"
)

func TestGetChangeOptionBy2Map(t *testing.T) {
	type args struct {
		map0 interface{}
		map1 interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]*ChangeOption
		wantErr bool
	}{
		{
			name: "Test",
			args: args{
				map0: map[string]uint64{
					"123": 123,
				},
				map1: map[string]uint64{},
			},
			want:    map[string]*ChangeOption{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetChangeOptionBy2Map(tt.args.map0, tt.args.map1)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChangeOptionBy2Map() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChangeOptionBy2Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

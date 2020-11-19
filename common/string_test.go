package common

import (
	"reflect"
	"testing"
)

func TestDifferentElementStrings(t *testing.T) {
	type args struct {
		src1 []string
		src2 []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "",
			args: args{},
			want: []string{},
		},
		{
			name: "",
			args: args{},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DifferentElementStrings(tt.args.src1, tt.args.src2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DifferentElementStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

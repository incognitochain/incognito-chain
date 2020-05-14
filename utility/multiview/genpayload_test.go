package multiview

import (
	"testing"
)

func TestLoadKeyList(t *testing.T) {
	LoadKeyList("keylist.json")
}

func TestGenPayload(t *testing.T) {
	type args struct {
		filePriKey string
		pubGroup   map[int]map[int][]int
		fileOut    string
		From       map[int]uint64
		To         map[int]uint64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test1",
			args: args{
				filePriKey: "keylist.json",
				pubGroup: map[int]map[int][]int{
					255: map[int][]int{
						1: []int{
							2, 3, 4,
						},
						2: []int{
							1, 3, 4,
						},
					},
				},
				From: map[int]uint64{
					255: 0,
				},
				To: map[int]uint64{
					255: 1000,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GenPayload(tt.args.filePriKey, tt.args.pubGroup, tt.args.From, tt.args.To, tt.args.fileOut)
		})
	}
}

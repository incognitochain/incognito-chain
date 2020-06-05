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
			name: "1",
			args: args{
				filePriKey: "keylist.json",
				pubGroup: map[int]map[int][]int{
					255: map[int][]int{
						3: []int{
							5,
						},
						5: []int{
							3,
						},
					},
				},
				From: map[int]uint64{
					255: 2,
				},
				To: map[int]uint64{
					255: 2,
				},
			},
		},
		{
			name: "1",
			args: args{
				filePriKey: "keylist.json",
				pubGroup: map[int]map[int][]int{
					255: map[int][]int{
						1: []int{
							6,
						},
						2: []int{
							6,
						},
						4: []int{
							1, 2, 6, 7,
						},

						7: []int{
							6,
						},
					},
				},
				From: map[int]uint64{
					255: 3,
				},
				To: map[int]uint64{
					255: 3,
				},
			},
		},
		{
			name: "1",
			args: args{
				filePriKey: "keylist.json",
				pubGroup: map[int]map[int][]int{
					255: map[int][]int{
						1: []int{
							2, 3, 4,
						},
						2: []int{
							1, 3, 4, 5, 7,
						},
						3: []int{
							1, 2, 4, 5, 7,
						},
						4: []int{
							1, 2, 3, 5, 7,
						},
						5: []int{
							1, 2, 3, 4, 7,
						},
						7: []int{
							1, 2, 3, 4, 5,
						},
					},
				},
				From: map[int]uint64{
					255: 4,
				},
				To: map[int]uint64{
					255: 4,
				},
			},
		},
		{
			name: "1",
			args: args{
				filePriKey: "keylist.json",
				pubGroup: map[int]map[int][]int{
					255: map[int][]int{},
				},
				From: map[int]uint64{
					255: 7,
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

package peerv2

import (
	"reflect"
	"testing"
)

func TestBatchingBlkHeightsForSync(t *testing.T) {
	type args struct {
		batchlen int
		height   []uint64
	}
	tests := []struct {
		name string
		args args
		want [][]uint64
	}{
		{
			name: "normal",
			args: args{
				batchlen: 5,
				height:   []uint64{5, 6, 7, 8, 9, 10, 12, 14, 15, 18, 30},
			},
			want: [][]uint64{
				[]uint64{5, 6, 7, 8, 9},
				[]uint64{10, 12, 14, 15, 18},
				[]uint64{30},
			},
		},
		{
			name: "normal 2",
			args: args{
				batchlen: 5,
				height:   []uint64{5, 6, 7, 8, 9, 10, 12, 14, 15},
			},
			want: [][]uint64{
				[]uint64{5, 6, 7, 8, 9},
				[]uint64{10, 12, 14, 15},
			},
		},
		{
			name: "normal 3",
			args: args{
				batchlen: 5,
				height:   []uint64{5},
			},
			want: [][]uint64{
				[]uint64{5},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := batchingBlkHeightsForSync(tt.args.batchlen, tt.args.height); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BatchingBlkHeightsForSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchingRangeBlkForSync(t *testing.T) {
	type args struct {
		batchlen uint64
		from     uint64
		to       uint64
	}
	tests := []struct {
		name string
		args args
		want []uint64
	}{
		{
			name: "normal",
			args: args{
				batchlen: 5,
				from:     5,
				to:       19,
			},
			want: []uint64{
				5,
				10,
				15,
				19,
			},
		},
		{
			name: "normal 2",
			args: args{
				batchlen: 5,
				from:     5,
				to:       15,
			},
			want: []uint64{
				5,
				10,
				15,
			},
		},
		{
			name: "normal 3",
			args: args{
				batchlen: 5,
				from:     5,
				to:       8,
			},
			want: []uint64{
				5,
				8,
			},
		},
		{
			name: "normal 4",
			args: args{
				batchlen: 5,
				from:     5,
				to:       5,
			},
			want: []uint64{
				5,
				5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := batchingRangeBlkForSync(tt.args.batchlen, tt.args.from, tt.args.to); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BatchingRangeBlkForSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchingBlkForSync(t *testing.T) {
	type args struct {
		batchlen int
		info     syncBlkInfo
	}
	tests := []struct {
		name string
		args args
		want []syncBlkInfo
	}{
		{
			name: "By heights",
			args: args{
				batchlen: 5,
				info: syncBlkInfo{
					bySpecHeights: true,
					byHash:        false,
					from:          0,
					to:            0,
					heights:       []uint64{3, 5, 6, 7, 8, 25, 70, 77, 79},
					hashes:        [][]byte{},
				},
			},
			want: []syncBlkInfo{
				syncBlkInfo{
					bySpecHeights: true,
					byHash:        false,
					from:          0,
					to:            0,
					heights:       []uint64{3, 5, 6, 7, 8},
					hashes:        [][]byte{},
				},
				syncBlkInfo{
					bySpecHeights: true,
					byHash:        false,
					from:          0,
					to:            0,
					heights:       []uint64{25, 70, 77, 79},
					hashes:        [][]byte{},
				},
			},
		},
		{
			name: "By range 1",
			args: args{
				batchlen: 5,
				info: syncBlkInfo{
					bySpecHeights: false,
					byHash:        false,
					from:          5,
					to:            19,
					heights:       []uint64{},
					hashes:        [][]byte{},
				},
			},
			want: []syncBlkInfo{
				syncBlkInfo{
					bySpecHeights: false,
					byHash:        false,
					from:          5,
					to:            10,
					heights:       []uint64{},
					hashes:        [][]byte{},
				},
				syncBlkInfo{
					bySpecHeights: false,
					byHash:        false,
					from:          10,
					to:            15,
					heights:       []uint64{},
					hashes:        [][]byte{},
				},
				syncBlkInfo{
					bySpecHeights: false,
					byHash:        false,
					from:          15,
					to:            19,
					heights:       []uint64{},
					hashes:        [][]byte{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := batchingBlkForSync(tt.args.batchlen, tt.args.info); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BatchingBlkForSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

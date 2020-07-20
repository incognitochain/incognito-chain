package blockchain

import (
	"sync"
	"testing"
)

func Test_getNoBlkPerYear(t *testing.T) {
	type args struct {
		blockCreationTimeSeconds uint64
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "40s",
			args: args{blockCreationTimeSeconds: 40},
			want: 788940,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNoBlkPerYear(tt.args.blockCreationTimeSeconds); got != tt.want {
				t.Errorf("getNoBlkPerYear() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_getRewardAmount(t *testing.T) {
	type fields struct {
		Chains           map[string]ChainInterface
		BestState        *BestState
		config           Config
		chainLock        sync.Mutex
		cQuitSync        chan struct{}
		Synker           Synker
		ConsensusOngoing bool
		IsTest           bool
	}
	type args struct {
		blkHeight uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchain := &BlockChain{
				Chains:           tt.fields.Chains,
				BestState:        tt.fields.BestState,
				config:           tt.fields.config,
				chainLock:        tt.fields.chainLock,
				cQuitSync:        tt.fields.cQuitSync,
				Synker:           tt.fields.Synker,
				ConsensusOngoing: tt.fields.ConsensusOngoing,
				IsTest:           tt.fields.IsTest,
			}
			if got := blockchain.getRewardAmount(tt.args.blkHeight); got != tt.want {
				t.Errorf("getRewardAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

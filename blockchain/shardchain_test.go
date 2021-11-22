package blockchain

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/txpool"
	"reflect"
	"sync"
	"testing"
)

//TODO: @hung unit-test
//case 1: block version multiview
//case 2: block version slashing
//case 3: block version dcs
func TestShardChain_GetSigningCommittees(t *testing.T) {
	type fields struct {
		shardID     int
		multiView   *multiview.MultiView
		BlockGen    *BlockGenerator
		Blockchain  *BlockChain
		hashHistory *lru.Cache
		ChainName   string
		Ready       bool
		TxPool      txpool.TxPool
		TxsVerifier txpool.TxVerifier
		insertLock  sync.Mutex
	}
	type args struct {
		proposerIndex int
		committees    []incognitokey.CommitteePublicKey
		blockVersion  int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []incognitokey.CommitteePublicKey
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := &ShardChain{
				shardID:     tt.fields.shardID,
				multiView:   tt.fields.multiView,
				BlockGen:    tt.fields.BlockGen,
				Blockchain:  tt.fields.Blockchain,
				hashHistory: tt.fields.hashHistory,
				ChainName:   tt.fields.ChainName,
				Ready:       tt.fields.Ready,
				TxPool:      tt.fields.TxPool,
				TxsVerifier: tt.fields.TxsVerifier,
				insertLock:  tt.fields.insertLock,
			}
			if got := chain.GetSigningCommittees(tt.args.proposerIndex, tt.args.committees, tt.args.blockVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigningCommittees() = %v, want %v", got, tt.want)
			}
		})
	}
}

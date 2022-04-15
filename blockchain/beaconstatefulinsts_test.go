package blockchain

import (
	"reflect"
	"sync"
	"testing"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
)

func TestBlockChain_buildStatefulInstructions(t *testing.T) {
	config.AbortParam()
	config.Param().BCHeightBreakPointPrivacyV2 = 15
	type fields struct {
		BeaconChain                 *BeaconChain
		ShardChain                  []*ShardChain
		config                      Config
		cQuitSync                   chan struct{}
		IsTest                      bool
		beaconViewCache             *lru.Cache
		committeeByEpochCache       *lru.Cache
		committeeByEpochProcessLock sync.Mutex
	}
	type args struct {
		beaconBestState           *BeaconBestState
		featureStateDB            *statedb.StateDB
		statefulActionsByShardID  map[byte][][]string
		beaconHeight              uint64
		rewardForCustodianByEpoch map[common.Hash]uint64
		portalParams              portal.PortalParams
		shardStates               map[byte][]types.ShardState
		allPdexv3Txs              map[byte][]metadata.Transaction
		pdexReward                uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		wantErr bool
	}{
		{
			name:   "addToken + shield pToken",
			fields: fields{},
			args: args{
				beaconHeight: 10,
				beaconBestState: &BeaconBestState{
					pdeStates:      map[uint]pdex.State{},
					featureStateDB: &statedb.StateDB{},
					portalStateV3:  &portalprocessv3.CurrentPortalState{},
					portalStateV4:  &portalprocessv4.CurrentPortalStateV4{},
					bridgeAggState: bridgeagg.NewState(),
				},
				portalParams: portal.PortalParams{
					RelayingParam:  portalrelaying.RelayingParams{},
					PortalParamsV3: map[uint64]portalv3.PortalParams{},
					PortalParamsV4: map[uint64]portalv4.PortalParams{
						20: portalv4.PortalParams{},
					},
				},
			},
			want:    [][]string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchain := &BlockChain{
				BeaconChain:                 tt.fields.BeaconChain,
				ShardChain:                  tt.fields.ShardChain,
				config:                      tt.fields.config,
				cQuitSync:                   tt.fields.cQuitSync,
				IsTest:                      tt.fields.IsTest,
				beaconViewCache:             tt.fields.beaconViewCache,
				committeeByEpochCache:       tt.fields.committeeByEpochCache,
				committeeByEpochProcessLock: tt.fields.committeeByEpochProcessLock,
			}
			got, err := blockchain.buildStatefulInstructions(tt.args.beaconBestState, tt.args.featureStateDB, tt.args.statefulActionsByShardID, tt.args.beaconHeight, tt.args.rewardForCustodianByEpoch, tt.args.portalParams, tt.args.shardStates, tt.args.allPdexv3Txs, tt.args.pdexReward)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockChain.buildStatefulInstructions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockChain.buildStatefulInstructions() = %v, want %v", got, tt.want)
			}
		})
	}
}

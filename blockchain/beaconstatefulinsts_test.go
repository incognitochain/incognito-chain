package blockchain

import (
	"io/ioutil"
	"os"
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
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
)

func TestBlockChain_buildStatefulInstructions(t *testing.T) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "bridgeagg_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)

	bridgeagg.Logger.Init(common.NewBackend(nil).Logger("test", true))
	config.AbortParam()
	config.Param().BCHeightBreakPointPrivacyV2 = 15
	config.Param().EpochParam.NumberOfBlockInEpoch = 10
	config.Param().EthContractAddressStr = "0x7bebc8445c6223b41b7bb4b0ae9742e2fd2f47f3"
	config.AbortUnifiedToken()
	common.MaxShardNumber = 8
	view := multiview.NewMultiView()
	view.AddView(&BeaconBestState{
		BestBlock: types.BeaconBlock{
			Header: types.BeaconHeader{
				Height: 10,
			},
		},
	})
	temp := map[uint64]map[common.Hash]map[uint]config.Vault{
		10: {
			common.PRVCoinID: map[uint]config.Vault{
				1: {
					ExternalDecimal: 9,
					ExternalTokenID: "0x0000000000000000000000000000000000000001",
					IncTokenID:      "375825bf838527610102c6943282642826901937679d8df5b5634d43d54a5769",
				},
			},
		},
	}
	config.SetUnifiedToken(temp)
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
			name: "addToken + shield pToken",
			fields: fields{
				ShardChain: []*ShardChain{
					{
						multiView: view,
					},
					{
						multiView: view,
					},
					{
						multiView: view,
					},
					{
						multiView: view,
					},
					{
						multiView: view,
					},
					{
						multiView: view,
					},
					{
						multiView: view,
					},
					{
						multiView: view,
					},
				},
			},
			args: args{
				beaconHeight: 10,
				beaconBestState: &BeaconBestState{
					pdeStates:      map[uint]pdex.State{},
					featureStateDB: sDB,
					portalStateV3:  &portalprocessv3.CurrentPortalState{},
					portalStateV4:  &portalprocessv4.CurrentPortalStateV4{},
					bridgeAggState: bridgeagg.NewState(),
				},
				portalParams: portal.PortalParams{
					RelayingParam:  portalrelaying.RelayingParams{},
					PortalParamsV3: map[uint64]portalv3.PortalParams{},
					PortalParamsV4: map[uint64]portalv4.PortalParams{
						20: {},
					},
				},
				statefulActionsByShardID: map[byte][][]string{
					0: {
						{
							"80",
							"eyJldGhSZWNlaXB0Ijp7InJvb3QiOiIweCIsInN0YXR1cyI6IjB4MSIsImN1bXVsYXRpdmVHYXNVc2VkIjoiMHgxZDZjMGEiLCJsb2dzQmxvb20iOiIweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAxMDAwMDAwMDAwMDIwMDAwMDAwMDAwMDQwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA0MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMjAwMDAwMDAwMDAwMDAwMDAwMDAxMDAwIiwibG9ncyI6W3siYWRkcmVzcyI6IjB4N2JlYmM4NDQ1YzYyMjNiNDFiN2JiNGIwYWU5NzQyZTJmZDJmNDdmMyIsInRvcGljcyI6WyIweDJkNGI1OTc5MzVmM2NkNjdmYjJlZWJmMWRiNGRlYmM5MzRjZWU1YzdiYWE3MTUzZjk4MGZkYmViMmU3NDA4NGUiXSwiZGF0YSI6IjB4MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwNjAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMmQ3OTg4M2QyMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA5NDMxMzI3Mzc2NjY2YjUwMzY3NzM1NTU0NDRhNDQ1MzQzNzc3MTQ4MzkzNzM4NTA3NjcxNjk3MTQyNzg0YjZkNTU2ZTQxMzk2NTZkMzk3OTQxNTk1NzU5NGE1NjUyNzYzNzc3NzU1ODU5MzE3MTY4Njg1OTcwNTA0MTZkMzQ0MjQ0N2EzMjZkNGM2MjQ2NzI1MjZkNjQ0YjMzNzk1MjY4NmU1NDcxNGE0MzVhNTg0YjQ4NTU2ZDZmNjkzNzRlNTYzODMzNDg0MzQ4MzI1OTQ2NzA2Mzc0NDg0ZTYxNDQ2NDZiNTM2OTUxNzM2ODczNmE3NzMyNTU0NjU1NzU3NzY0NDU3NjYzNjk2NDY3NjE0YjZkNDYzMzU2NGE3MDU5MzU2NjM4NTI2NDRlMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwIiwiYmxvY2tOdW1iZXIiOiIweDAiLCJ0cmFuc2FjdGlvbkhhc2giOiIweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAiLCJ0cmFuc2FjdGlvbkluZGV4IjoiMHgwIiwiYmxvY2tIYXNoIjoiMHgwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwIiwibG9nSW5kZXgiOiIweDAiLCJyZW1vdmVkIjpmYWxzZX1dLCJ0cmFuc2FjdGlvbkhhc2giOiIweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAiLCJjb250cmFjdEFkZHJlc3MiOiIweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAiLCJnYXNVc2VkIjoiMHgwIiwiYmxvY2tIYXNoIjoiMHgwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwIiwidHJhbnNhY3Rpb25JbmRleCI6IjB4MCJ9LCJtZXRhIjp7IkJsb2NrSGFzaCI6IjB4Mjg4ZjBhOGU5YzI2YjNmN2EwOTVjZmFhMWYzOWE5NjE1MDk3NDBmMThiNmVkODMxOWJlNjE3ZDViMTgwNjQ0MCIsIlR4SW5kZXgiOjMsIlByb29mU3RycyI6WyIrRkdnRnA2NU55eVRkeTIyWUpaYTJQejlpWUpWNEZTMDkweXAvS2YrZENVa21HV0FnSUNBZ0lDQW9LTHUrd2tvZlJ3a0MrbFI4MWIvbjhtRml2a254RmV4ZGw3YzlwRmUzdGgxZ0lDQWdJQ0FnSUE9IiwiK0pHQW9DY1h2d096Mi95K2NxS29Vcit3bGxvMXpnRDJ5UG9teDBUUVlTYTBocC9Ub0xKNkZ2c1J3ZjFtdmVTRStZZ3kxOW9EUkVUNllQakhodkwvNVhvcGVOUmFvUHlrN2I5R295bnA3N2tWT0ZGaUJjb3BvVXM4WHdhc0JnYmRzRnFEUkYvMG9Kb3kxWFAxODhCdWVaUzFwaE1LM3ZURW95Ukt3OWl0S2FrU2pJL3JDV0phZ0lDQWdJQ0FnSUNBZ0lDQSIsIitRSndJTGtDYkFMNUFtZ0JneDFzQ3JrQkFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFFQUFBQUFBQ0FBQUFBQUFFQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQVFBQUFBQUFBQUFBQUFBQUFBQUFBQWdBQUFBQUFBQUFBQUFFQUQ1QVYzNUFWcVVlK3ZJUkZ4aUk3UWJlN1N3cnBkQzR2MHZSL1Bob0MxTFdYazE4ODFuK3k3cjhkdE42OGswenVYSHVxY1ZQNWdQMitzdWRBaE91UUVnQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBWUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBTFhtSVBTQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUpReE1uTjJabXRRTm5jMVZVUktSRk5EZDNGSU9UYzRVSFp4YVhGQ2VFdHRWVzVCT1dWdE9YbEJXVmRaU2xaU2RqZDNkVmhaTVhGb2FGbHdVRUZ0TkVKRWVqSnRUR0pHY2xKdFpFc3plVkpvYmxSeFNrTmFXRXRJVlcxdmFUZE9Wamd6U0VOSU1sbEdjR04wU0U1aFJHUnJVMmxSYzJoemFuY3lWVVpWZFhka1JYWmphV1JuWVV0dFJqTldTbkJaTldZNFVtUk9BQUFBQUFBQUFBQUFBQUFBIl0sIkluY1Rva2VuSUQiOiIzNzU4MjViZjgzODUyNzYxMDEwMmM2OTQzMjgyNjQyODI2OTAxOTM3Njc5ZDhkZjViNTYzNGQ0M2Q1NGE1NzY5IiwiVHlwZSI6ODB9LCJ0eFJlcUlkIjoiMTNlNTZhNDU3MmFhN2VlYjVlN2Y2NThhOGFmMDM2ZWQzZmY4YTEyODU0NTJmMzBjNzQxZjQ3MzNiYzhjN2Y5YyJ9",
						},
					},
				},
			},
			want: [][]string{
				{
					"348",
					"eyJOZXdMaXN0VG9rZW5zIjp7IjAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDQiOnsiMSI6eyJFeHRlcm5hbERlY2ltYWwiOjksIkV4dGVybmFsVG9rZW5JRCI6IjB4MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMSIsIkluY1Rva2VuSUQiOiIzNzU4MjViZjgzODUyNzYxMDEwMmM2OTQzMjgyNjQyODI2OTAxOTM3Njc5ZDhkZjViNTYzNGQ0M2Q1NGE1NzY5In19fX0=",
				},
				{
					"80", "0", "rejected", "13e56a4572aa7eeb5e7f658a8af036ed3ff8a1285452f30c741f4733bc8c7f9c",
				},
			},
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

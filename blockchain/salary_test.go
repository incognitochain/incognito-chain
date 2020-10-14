package blockchain

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/trie"
)

var (
	wrarperDB      statedb.DatabaseAccessWarper
	diskDB         incdb.Database
	committeesKeys []incognitokey.CommitteePublicKey
	rewardReceiver map[string]string
)

var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_reward")
	if err != nil {
		panic(err)
	}
	diskDB, _ = incdb.Open("leveldb", dbPath)
	wrarperDB = statedb.NewDatabaseAccessWarper(diskDB)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	dataaccessobject.Logger.Init(common.NewBackend(nil).Logger("test", true))
	committeesKeysStr := []string{
		"121VhftSAygpEJZ6i9jGkGco4dFKpqVXZA6nmGjRKYWR7Q5NngQSX1adAfYY3EGtS32c846sAxYSKGCpqouqmJghfjtYfHEPZTRXctAcc6bYhR3d1YpB6m3nNjEdTYWf85agBq5QnVShMjBRFf54dK25MAazxBSYmpowxwiaEnEikpQah2W4LY9P9vF9HJuLUZ4BnknoXXK3BVkGHsimy5RXtvNet2LqXZgZWHX5CDj31q7kQ2jUGJHr862MgsaHfT4Qq8o4u71nhgtzKBYgw9fvXqJUU6EVynqJCVdqaDXmUvjanGkaZb9vQjaXVoHyf6XRxVSbQBTS5G7eb4D4V3RucXRLQp34KTadmmNQUxnCoPQztVcuDQwNqy9zRXPPAdw7pWvv7P7p4HuQVAHKqvJskMNk3v971WBH5VpZA1XMkmtu",
		"121VhftSAygpEJZ6i9jGk4diwdFxA6whUVx3P9GmT35Lw6txpbDmeVgSJ4qUwSHPAep8FedvNrZfGB1eoXZXnCwwHVQs7htn7XigUSowaRJyXVf9n42Auhk65GJbxnE7C2t8HWjW3N97m4TejbAQoR5WoWSeaixXRSimadBeWVF4cgZxPUvLuPsSfGYWi4DQ4GwJhpSLNEbite3NseJBDM5N7DGas6mn9roe2jcSYSVyFRR87fqHMfPhhyMQ7k21up58RtMa3tRsEBDBRmKZgeaKr67MuBbEFKJw1Hh8fwbRVaFKeD38EAG9oykANrTmBvZXk4gU8Dvm3uJEJLX7iwDLVxgSDaNYtaYAoePD4dbgWmvotELQW2kJaQ7DEmttV7ZgukQCVPg36pHbDF8oijr5bobgLhft3ajJy5x8mMpuRDYy",
		"121VhftSAygpEJZ6i9jGkMJDUd7LzasxGKrkhLZ2s9mw9YtWaeRVURg9TNxrsxhYLHLSmj1Vh6bpwtpu1i9A7kAM1DzYpTPDws7PftpqS7h9nyHGAu1GLEha57mEEv9U8AWopiQa1Bp5AUq4dQd7o65Ub4S6zMb7aj1Rc7xZ8HRfXScXNdG2cvUEGZXrqS9dF6VwxfdjQa6PKymVHagd2MTL4vaWVcHp9SyKHQqhzkkafKZcb2MruPLPodLnm6Zvd4EqRJYGEZoWWSfSoxsmZsY63W6dsWwpJdZ7sMjuHmqvo1ousSHXnNcGV7LYsuhuwqaMshDsMHdJsFsDErRzTWhwi6jCRUwmaGV82c8JVP3L5HLPdMEJSCa79GwZQNCnkzyHv8DDdX4ptgkzfwQd6bCaNE7DbUUz7yTDa9quVTuYjRWU",
	}
	rewardReceiver = make(map[string]string)
	rewardReceiver["1jTb2CjHbu6dZbuPrumUvtfMkE2qkk39F3M3cnAmsm1kvTYhEY"] = "12RvZy9CVoYK9CWt6JSJhPV2WQjp21UjywYf3YZMoUUaDgGMWeQ1qF9WHZmuPHTyZ28B4Za9hDmxdVLDpy5nJXgTYXANqenbY7Kt4zy"
	rewardReceiver["12bGNd9ofTJSbZYB2BXtaAQpsRV4a3KJ3xP5kLQmoxYBMaqhtXw"] = "12S3wJTUwb8RjHPheQFwer9UPJgs3k1puFnuyAodokcJ7zEcPn6qL72kWCACuDEh5NYrKmz3ctdzd4W2L5P1rbwP75H2D117PVjuS7x"
	rewardReceiver["122yPg6oriAu4uqYeRNr1VN2DScMQAY6xnc1eLaJLr7MPsAsSoJ"] = "12RyAEaUz4sErApu1f23PEydvotxDnC5gHoWDy5Th7JQuoT57oUowk8eSQN44ojPj3wZ5sEYFcLeFU5R8zgiXkSbAuY367Tek31gM1z"
	committeesKeys, _ = incognitokey.CommitteeBase58KeyListToStruct(committeesKeysStr)
	return
}()

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
		{
			name: "10s",
			args: args{blockCreationTimeSeconds: 10},
			want: 3155760,
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
	numberOfBlockPerYear := getNoBlkPerYear(40)
	type fields struct {
		config Config
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
		{
			name: "Mainnet year 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				1,
			},
			want: 1386666000,
		},
		{
			name: "Mainnet year 1",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear,
			},
			want: 1386666000,
		},
		{
			name: "Mainnet year 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear + 1,
			},
			want: 1261866060,
		},
		{
			name: "Mainnet year 2",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 2,
			},
			want: 1261866060,
		},
		{
			name: "Mainnet year 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear*2 + 1,
			},
			want: 1148298114,
		},
		{
			name: "Mainnet year 3",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 3,
			},
			want: 1148298114,
		},
		{
			name: "Mainnet year 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear*3 + 1,
			},
			want: 1044951283,
		},
		{
			name: "Mainnet year 4",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 4,
			},
			want: 1044951283,
		},
		{
			name: "Mainnet year 5",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear*4 + 1,
			},
			want: 950905667,
		},
		{
			name: "Mainnet year 5",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 5,
			},
			want: 950905667,
		},
		{
			name: "Mainnet year 6",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear*5 + 1,
			},
			want: 865324156,
		},
		{
			name: "Mainnet year 6",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 6,
			},
			want: 865324156,
		},
		{
			name: "Mainnet year 7",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear*6 + 1,
			},
			want: 787444981,
		},
		{
			name: "Mainnet year 7",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 7,
			},
			want: 787444981,
		},
		{
			name: "Mainnet year 8",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear*7 + 1,
			},
			want: 716574932,
		},
		{
			name: "Mainnet year 8",
			fields: fields{
				config: Config{
					ChainParams: &Params{
						MinBeaconBlockInterval: MainnetMinBeaconBlkInterval,
						BasicReward:            MainnetBasicReward,
					},
				},
			},
			args: args{
				numberOfBlockPerYear * 8,
			},
			want: 716574932,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchain := &BlockChain{
				config: tt.fields.config,
			}
			if got := blockchain.getRewardAmount(tt.args.blkHeight); got != tt.want {
				t.Errorf("getRewardAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPercentForIncognitoDAO(t *testing.T) {
	type args struct {
		blockHeight uint64
		blkPerYear  uint64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "year 1",
			args: args{
				blockHeight: 788940,
				blkPerYear:  788940,
			},
			want: 10,
		},
		{
			name: "year 2-1",
			args: args{
				blockHeight: 788941,
				blkPerYear:  788940,
			},
			want: 9,
		}, {
			name: "year 2-2",
			args: args{
				blockHeight: 1577880,
				blkPerYear:  788940,
			},
			want: 9,
		},
		{
			name: "year 3-1",
			args: args{
				blockHeight: 1577881,
				blkPerYear:  788940,
			},
			want: 8,
		},
		{
			name: "year 3-2",
			args: args{
				blockHeight: 2366820,
				blkPerYear:  788940,
			},
			want: 8,
		},
		{
			name: "year 4-1",
			args: args{
				blockHeight: 2366821,
				blkPerYear:  788940,
			},
			want: 7,
		},
		{
			name: "year 4-2",
			args: args{
				blockHeight: 3155760,
				blkPerYear:  788940,
			},
			want: 7,
		},
		{
			name: "year 5-1",
			args: args{
				blockHeight: 3155761,
				blkPerYear:  788940,
			},
			want: 6,
		},
		{
			name: "year 5-2",
			args: args{
				blockHeight: 3944700,
				blkPerYear:  788940,
			},
			want: 6,
		},
		{
			name: "year 6-1",
			args: args{
				blockHeight: 3944701,
				blkPerYear:  788940,
			},
			want: 5,
		},
		{
			name: "year 6-2",
			args: args{
				blockHeight: 4733640,
				blkPerYear:  788940,
			},
			want: 5,
		},
		{
			name: "year 7",
			args: args{
				blockHeight: 5522580,
				blkPerYear:  788940,
			},
			want: 4,
		},
		{
			name: "year 8",
			args: args{
				blockHeight: 6311520,
				blkPerYear:  788940,
			},
			want: 3,
		},
		{
			name: "year 9",
			args: args{
				blockHeight: 7100460,
				blkPerYear:  788940,
			},
			want: 3,
		},
		{
			name: "year 10",
			args: args{
				blockHeight: 7889400,
				blkPerYear:  788940,
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPercentForIncognitoDAO(tt.args.blockHeight, tt.args.blkPerYear); got != tt.want {
				t.Errorf("getPercentForIncognitoDAO() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_addShardRewardRequestToBeacon(t *testing.T) {
	config := Config{}
	config.ChainParams = &ChainMainParam
	sDB, _ := statedb.NewWithPrefixTrie(common.EmptyRoot, wrarperDB)
	acceptedBlockRewardInfoBase := metadata.NewAcceptedBlockRewardInfo(0, make(map[common.Hash]uint64), 2)
	acceptedBlockRewardInfoBaseInst, _ := acceptedBlockRewardInfoBase.GetStringFormat()
	txFee := make(map[common.Hash]uint64)
	txFee1 := uint64(10000)
	txFee[common.PRVCoinID] = txFee1
	acceptedBlockRewardInfo1 := metadata.NewAcceptedBlockRewardInfo(0, txFee, 2)
	acceptedBlockRewardInfo1Inst, _ := acceptedBlockRewardInfo1.GetStringFormat()
	type fields struct {
		config Config
	}
	type args struct {
		beaconBlock   *types.BeaconBlock
		rewardStateDB *statedb.StateDB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "add base reward",
			fields: fields{
				config: config,
			},
			args: args{
				beaconBlock: &types.BeaconBlock{
					Body: types.BeaconBody{
						Instructions: [][]string{acceptedBlockRewardInfoBaseInst},
					},
					Header: types.BeaconHeader{
						Epoch: 1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "add base reward + 10000",
			fields: fields{
				config: config,
			},
			args: args{
				beaconBlock: &types.BeaconBlock{
					Body: types.BeaconBody{
						Instructions: [][]string{acceptedBlockRewardInfo1Inst},
					},
					Header: types.BeaconHeader{
						Epoch: 1,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchain := &BlockChain{
				config: tt.fields.config,
			}
			if err := blockchain.addShardRewardRequestToBeacon(tt.args.beaconBlock, sDB); (err != nil) != tt.wantErr {
				t.Errorf("addShardRewardRequestToBeacon() error = %v, wantErr %v", err, tt.wantErr)
			}
			rootHash, _ := sDB.Commit(true)
			_ = sDB.Database().TrieDB().Commit(rootHash, false)
		})
	}
	reward, err := statedb.GetRewardOfShardByEpoch(sDB, 1, 0, common.PRVCoinID)
	if err != nil {
		t.Error(err)
	}
	wantReward := 1386666000*2 + txFee1
	if reward != wantReward {
		t.Errorf("addShardRewardRequestToBeacon() got base reward = %v, want %v", reward, wantReward)
	}
}

// func TestBlockChain_buildInstRewardForBeacons(t *testing.T) {
// 	type fields struct {
// 		BestState *BestState
// 	}
// 	fields1 := fields{
// 		BestState: &BestState{Beacon: &BeaconBestState{BeaconCommittee: committeesKeys}},
// 	}
// 	totalReward1 := make(map[common.Hash]uint64)
// 	totalReward1_1 := make(map[common.Hash]uint64)
// 	totalReward1[common.PRVCoinID] = 900
// 	totalReward1_1[common.PRVCoinID] = 300
// 	rewardInst1_1, _ := metadata.BuildInstForBeaconReward(totalReward1_1, committeesKeys[0].GetNormalKey())
// 	rewardInst1_2, _ := metadata.BuildInstForBeaconReward(totalReward1_1, committeesKeys[1].GetNormalKey())
// 	rewardInst1_3, _ := metadata.BuildInstForBeaconReward(totalReward1_1, committeesKeys[2].GetNormalKey())
// 	type args struct {
// 		epoch       uint64
// 		totalReward map[common.Hash]uint64
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    [][]string
// 		wantErr bool
// 	}{
// 		{
// 			name:   "committee len 3",
// 			fields: fields1,
// 			args: args{
// 				epoch:       1,
// 				totalReward: totalReward1,
// 			},
// 			want:    [][]string{rewardInst1_1, rewardInst1_2, rewardInst1_3},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			blockchain := &BlockChain{
// 				BestState: tt.fields.BestState,
// 			}
// 			got, err := blockchain.buildInstRewardForBeacons(tt.args.epoch, tt.args.totalReward)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("buildInstRewardForBeacons() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("buildInstRewardForBeacons() got = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestBlockChain_buildInstRewardForIncDAO(t *testing.T) {
	type fields struct {
		config Config
	}
	fields1 := fields{
		config: Config{
			ChainParams: &ChainMainParam,
		},
	}
	totalReward1 := make(map[common.Hash]uint64)
	wantReward1 := uint64(256)
	totalReward1[common.PRVCoinID] = wantReward1
	rewardInst1_DAO, _ := metadata.BuildInstForIncDAOReward(totalReward1, ChainMainParam.IncognitoDAOAddress)
	type args struct {
		epoch       uint64
		totalReward map[common.Hash]uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		wantErr bool
	}{
		{
			name:   "build DAO mainnet",
			fields: fields1,
			args: args{
				epoch:       1,
				totalReward: totalReward1,
			},
			want:    [][]string{rewardInst1_DAO},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchain := &BlockChain{
				config: tt.fields.config,
			}
			got, err := blockchain.buildInstRewardForIncDAO(tt.args.epoch, tt.args.totalReward)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildInstRewardForIncDAO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildInstRewardForIncDAO() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_buildInstRewardForShards(t *testing.T) {
	type args struct {
		epoch        uint64
		totalRewards []map[common.Hash]uint64
	}
	totalRewardShard0_1 := make(map[common.Hash]uint64)
	totalRewardShard0_1[common.PRVCoinID] = 1000
	totalRewardShard1_1 := make(map[common.Hash]uint64)
	totalRewardShard1_1[common.PRVCoinID] = 1123
	rewardInstShard0_1, _ := metadata.BuildInstForShardReward(totalRewardShard0_1, 1, 0)
	rewardInstShard1_1, _ := metadata.BuildInstForShardReward(totalRewardShard1_1, 1, 1)
	tests := []struct {
		name    string
		args    args
		want    [][]string
		wantErr bool
	}{
		{
			name: "shard 0,1",
			args: args{
				1,
				[]map[common.Hash]uint64{totalRewardShard0_1, totalRewardShard1_1},
			},
			want:    [][]string{rewardInstShard0_1[0], rewardInstShard1_1[0]},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchain := &BlockChain{}
			got, err := blockchain.buildInstRewardForShards(tt.args.epoch, tt.args.totalRewards)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildInstRewardForShards() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildInstRewardForShards() got = %v, want %v", got, tt.want)
			}
		})
	}
}

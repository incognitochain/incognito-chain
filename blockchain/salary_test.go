package blockchain

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/trie"
	"github.com/incognitochain/incognito-chain/wallet"
)

var (
	wrarperDB      statedb.DatabaseAccessWarper
	diskDB         incdb.Database
	committeesKeys []incognitokey.CommitteePublicKey
	rewardReceiver map[string]string
	emptyRoot      = common.HexToHash(common.HexEmptyRoot)
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

func initStateDB() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "data")
	if err != nil {
		panic(err)
	}
	diskDB, _ = incdb.Open("leveldb", dbPath)
	wrarperDB = statedb.NewDatabaseAccessWarper(diskDB)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	dataaccessobject.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}

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
	type param struct {
		BasicReward            uint64
		MinBeaconBlockInterval time.Duration
	}
	type fields struct {
	}
	type args struct {
		blkHeight uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
		param  param
	}{
		{
			name:   "Mainnet year 1",
			fields: fields{},
			args: args{
				1,
			},
			want: 1386666000,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 1",
			fields: fields{},
			args: args{
				numberOfBlockPerYear,
			},
			want: 1386666000,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 2",
			fields: fields{},
			args: args{
				numberOfBlockPerYear + 1,
			},
			want: 1261866060,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 2",
			fields: fields{},
			args: args{
				numberOfBlockPerYear * 2,
			},
			want: 1261866060,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 3",
			fields: fields{},
			args: args{
				numberOfBlockPerYear*2 + 1,
			},
			want: 1148298114,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 3",
			fields: fields{},
			args: args{
				numberOfBlockPerYear * 3,
			},
			want: 1148298114,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 4",
			fields: fields{},
			args: args{
				numberOfBlockPerYear*3 + 1,
			},
			want: 1044951283,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 4",
			fields: fields{},
			args: args{
				numberOfBlockPerYear * 4,
			},
			want: 1044951283,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 5",
			fields: fields{},
			args: args{
				numberOfBlockPerYear*4 + 1,
			},
			want: 950905667,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 5",
			fields: fields{},
			args: args{
				numberOfBlockPerYear * 5,
			},
			want: 950905667,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 6",
			fields: fields{},
			args: args{
				numberOfBlockPerYear*5 + 1,
			},
			want: 865324156,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 6",
			fields: fields{},
			args: args{
				numberOfBlockPerYear * 6,
			},
			want: 865324156,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 7",
			fields: fields{},
			args: args{
				numberOfBlockPerYear*6 + 1,
			},
			want: 787444981,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 7",
			fields: fields{},
			args: args{
				numberOfBlockPerYear * 7,
			},
			want: 787444981,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 8",
			fields: fields{},
			args: args{
				numberOfBlockPerYear*7 + 1,
			},
			want: 716574932,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
		{
			name:   "Mainnet year 8",
			fields: fields{},
			args: args{
				numberOfBlockPerYear * 8,
			},
			want: 716574932,
			param: param{
				BasicReward:            1386666000,
				MinBeaconBlockInterval: time.Second * 40,
			},
		},
	}

	setupParam := func(param param) {
		config.Param().BasicReward = param.BasicReward
		config.Param().BlockTime.MinBeaconBlockInterval = param.MinBeaconBlockInterval
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchain := &BlockChain{}
			config.AbortParam()
			setupParam(tt.param)
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
	config.AbortParam()
	config.Param().BasicReward = 1386666000
	config.Param().BlockTime.MinBeaconBlockInterval = 40 * time.Second

	sDB, _ := statedb.NewWithPrefixTrie(common.EmptyRoot, wrarperDB)
	acceptedBlockRewardInfoBase := instruction.NewAcceptBlockRewardV1WithValue(0, make(map[common.Hash]uint64), 2)
	acceptedBlockRewardInfoBaseInst, _ := acceptedBlockRewardInfoBase.String()
	txFee := make(map[common.Hash]uint64)
	txFee1 := uint64(10000)
	txFee[common.PRVCoinID] = txFee1
	acceptedBlockRewardInfo1 := instruction.NewAcceptBlockRewardV1WithValue(0, txFee, 2)
	acceptedBlockRewardInfo1Inst, _ := acceptedBlockRewardInfo1.String()
	type fields struct {
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
			name:   "add base reward",
			fields: fields{},
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
			name:   "add base reward + 10000",
			fields: fields{},
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
			blockchain := &BlockChain{}
			if err := blockchain.addShardRewardRequestToBeacon(tt.args.beaconBlock, sDB, &BeaconBestState{}); (err != nil) != tt.wantErr {
				t.Errorf("addShardRewardRequestToBeacon() error = %v, wantErr %v", err, tt.wantErr)
			}
			rootHash, _, _ := sDB.Commit(true)
			_ = sDB.Database().TrieDB().Commit(rootHash, false, nil)
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

func TestBlockChain_buildInstRewardForIncDAO(t *testing.T) {
	config.AbortParam()
	config.Param().IncognitoDAOAddress = "12S32fSyF4h8VxFHt4HfHvU1m9KHvBQsab5zp4TpQctmMdWuveXFH9KYWNemo7DRKvaBEvMgqm4XAuq1a1R4cNk2kfUfvXR3DdxCho3"
	common.MaxShardNumber = 8
	totalReward1 := make(map[common.Hash]uint64)
	wantReward1 := uint64(256)
	totalReward1[common.PRVCoinID] = wantReward1
	rewardInst1_DAO, _ := metadata.BuildInstForIncDAOReward(totalReward1, "12S32fSyF4h8VxFHt4HfHvU1m9KHvBQsab5zp4TpQctmMdWuveXFH9KYWNemo7DRKvaBEvMgqm4XAuq1a1R4cNk2kfUfvXR3DdxCho3")
	type args struct {
		epoch       uint64
		totalReward map[common.Hash]uint64
	}
	tests := []struct {
		name    string
		args    args
		want    [][]string
		wantErr bool
	}{
		{
			name: "build DAO mainnet",
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
			blockchain := &BlockChain{}
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
	rewardInstShard0_1, _ := instruction.NewShardReceiveRewardV1WithValue(totalRewardShard0_1, 1, 0)
	rewardInstShard1_1, _ := instruction.NewShardReceiveRewardV1WithValue(totalRewardShard1_1, 1, 1)
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
			got, err := blockchain.buildInstructionRewardForShards(tt.args.epoch, tt.args.totalRewards)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildInstructionRewardForShards() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildInstructionRewardForShards() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestBeaconBestState_calculateReward(t *testing.T) {

// 	config.AbortParam()
// 	config.Param().BlockTime.MaxBeaconBlockCreation = 8 * time.Second

// 	initStateDB()
// 	initPublicKey()

// 	hash, _ := common.Hash{}.NewHashFromStr("123")

// 	rewards := []uint64{1093995, 1093995}
// 	beaconReward := []uint64{196919, 51054}
// 	shardReward := []uint64{787677, 933543}
// 	daoReward := []uint64{109399, 109399}
// 	sDBs := []*statedb.StateDB{}
// 	splitRewardRuleProcessors := []*mocks.SplitRewardRuleProcessor{}
// 	for i := 0; i < 2; i++ {
// 		sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
// 		assert.Nil(t, err)
// 		for j := 0; j < 8; j++ {
// 			statedb.AddShardRewardRequest(
// 				sDB, 0, byte(j), *hash, rewards[i],
// 			)
// 		}
// 		sDBs = append(sDBs, sDB)
// 		splitRewardRuleProcessor := &mocks.SplitRewardRuleProcessor{}
// 		for j := 0; j < 8; j++ {
// 			splitRewardRuleProcessor.On("SplitReward", &committeestate.SplitRewardEnvironment{
// 				ActiveShards:           8,
// 				DAOPercent:             10,
// 				PercentCustodianReward: 0,
// 				ShardID:                byte(j),
// 				TotalReward:            make(map[common.Hash]uint64),
// 				BeaconHeight:           20,
// 			}).Return(
// 				map[common.Hash]uint64{
// 					*hash: beaconReward[i],
// 				},
// 				map[common.Hash]uint64{
// 					*hash: shardReward[i],
// 				},
// 				map[common.Hash]uint64{
// 					*hash: daoReward[i],
// 				},
// 				map[common.Hash]uint64{},
// 				nil,
// 			)
// 		}
// 		splitRewardRuleProcessors = append(splitRewardRuleProcessors, splitRewardRuleProcessor)
// 	}

// 	type args struct {
// 		maxBeaconBlockCreation    uint64
// 		splitRewardRuleProcessor  committeestate.SplitRewardRuleProcessor
// 		numberOfActiveShards      int
// 		beaconHeight              uint64
// 		epoch                     uint64
// 		rewardStateDB             *statedb.StateDB
// 		isSplitRewardForCustodian bool
// 		percentCustodianRewards   uint64
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    map[common.Hash]uint64
// 		want1   []map[common.Hash]uint64
// 		want2   map[common.Hash]uint64
// 		want3   map[common.Hash]uint64
// 		wantErr bool
// 	}{
// 		{
// 			name: "Year 1 - V1",
// 			args: args{
// 				beaconHeight:             20,
// 				epoch:                    1,
// 				rewardStateDB:            sDBs[1],
// 				numberOfActiveShards:     8,
// 				splitRewardRuleProcessor: splitRewardRuleProcessors[0],
// 			},
// 			want: map[common.Hash]uint64{
// 				*hash: 1575352,
// 			},
// 			want1: []map[common.Hash]uint64{
// 				map[common.Hash]uint64{
// 					*hash: 787677,
// 				},
// 				map[common.Hash]uint64{
// 					*hash: 787677,
// 				},
// 				map[common.Hash]uint64{
// 					*hash: 787677,
// 				},
// 				map[common.Hash]uint64{
// 					*hash: 787677,
// 				},
// 				map[common.Hash]uint64{
// 					*hash: 787677,
// 				},
// 				map[common.Hash]uint64{
// 					*hash: 787677,
// 				},
// 				map[common.Hash]uint64{
// 					*hash: 787677,
// 				},
// 				map[common.Hash]uint64{
// 					*hash: 787677,
// 				},
// 			},
// 			want2: map[common.Hash]uint64{
// 				*hash: 109399 * 8,
// 			},
// 			want3:   map[common.Hash]uint64{},
// 			wantErr: false,
// 		},
// 		// @NOICE: No use split rule reward v2
// 		/*
// 			{
// 				name: "Year 1 - V2",
// 				args: args{
// 					beaconHeight:  20,
// 					epoch:         1,
// 					rewardStateDB: sDBs[0],
// 				},
// 				want: map[common.Hash]uint64{
// 					*hash: 51054 * 8,
// 				},
// 				want1: []map[common.Hash]uint64{
// 					map[common.Hash]uint64{
// 						*hash: 933543,
// 					},
// 					map[common.Hash]uint64{
// 						*hash: 933543,
// 					},
// 					map[common.Hash]uint64{
// 						*hash: 933543,
// 					},
// 					map[common.Hash]uint64{
// 						*hash: 933543,
// 					},
// 					map[common.Hash]uint64{
// 						*hash: 933543,
// 					},
// 					map[common.Hash]uint64{
// 						*hash: 933543,
// 					},
// 					map[common.Hash]uint64{
// 						*hash: 933543,
// 					},
// 					map[common.Hash]uint64{
// 						*hash: 933543,
// 					},
// 				},
// 				want2: map[common.Hash]uint64{
// 					*hash: 109399 * 8,
// 				},
// 				want3:   map[common.Hash]uint64{},
// 				wantErr: false,
// 			},
// 		*/
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, got1, got2, got3, err := calculateReward(tt.args.splitRewardRuleProcessor, tt.args.numberOfActiveShards, tt.args.beaconHeight, tt.args.epoch, tt.args.rewardStateDB, tt.args.isSplitRewardForCustodian, tt.args.percentCustodianRewards)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("calculateReward() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("calculateReward() got = %v, want %v", got, tt.want)
// 			}
// 			if !reflect.DeepEqual(got1, tt.want1) {
// 				t.Errorf("calculateReward() got1 = %v, want %v", got1, tt.want1)
// 			}
// 			if !reflect.DeepEqual(got2, tt.want2) {
// 				t.Errorf("calculateReward() got2 = %v, want %v", got2, tt.want2)
// 			}
// 			if !reflect.DeepEqual(got3, tt.want3) {
// 				t.Errorf("calculateReward() got3 = %v, want %v", got3, tt.want3)
// 			}
// 		})
// 	}
// }

func Test_getCommitteeToPayRewardMultiset(t *testing.T) {

	initPublicKey()
	initLog()
	wl0, _ := wallet.Base58CheckDeserialize("12RpJeW7vnRPekywx9nhSraaAfANDYaWYYYE7iVhivr5chCo5S9g1YHQFJ6TxHMJz6okHXTmmMEyqaov912ZkxJuyLicEU4QXJeRysY")
	wl1, _ := wallet.Base58CheckDeserialize("12RxxmjTTzEGm6eCvHhXmNoCjZt9WMAAKECyGu9wpWMKHXD1Pgvh2TZRVM47CiVFG3XjzjdzMRSsSdWE4p1evcyJ6pCS5QCWWNtn5Bk")
	wl2, _ := wallet.Base58CheckDeserialize("12Ry1WG32b6D2PXnA2AfUfwnfuKJdtExZeZbwszKxTbUaeH8BJBpaHibnJmHTzehKduPDVbRUbv38wrZ4snsX6sKZH5PmUoys7r8KJA")
	wl3, _ := wallet.Base58CheckDeserialize("12Ry21dYZX8qmrSq9JJkSydwtKrSpvGGJGUqJyJVP491gWtiKA9ya7XSsJ5gYfTSjALMudEn3M5iXTVw3sQ1rouZ3EBnFMf9ojfJwpG")
	wl4, _ := wallet.Base58CheckDeserialize("12Ry2Tej7dFBy4TrfGeu6hkTDMSTsSSExCiyutSSex4y28stcWLqbFMP2pEVFvgQzuLdPaYS2Gcwwv4wmUtecMYxNkBneEMK1BFELvJ")
	wl5, _ := wallet.Base58CheckDeserialize("12Ry4s7jtiJ9C61iUmLToDYgVAsc2gA7GsihqRgdyqq9u7DY3qTe3Ns3UPpuhts6qkfXbtqtzRxZp1cE7ULZA2sCttsBYjFVwNcNRw3")
	stakerInfos := make([]*statedb.StakerInfo, 6)
	stakerInfos[0] = &statedb.StakerInfo{}
	stakerInfos[0].SetRewardReceiver(wl0.KeySet.PaymentAddress)
	stakerInfos[0].SetTxStakingID(common.HashH([]byte{0}))
	stakerInfos[0].SetAutoStaking(true)
	stakerInfos[1] = &statedb.StakerInfo{}
	stakerInfos[1].SetRewardReceiver(wl1.KeySet.PaymentAddress)
	stakerInfos[1].SetTxStakingID(common.HashH([]byte{1}))
	stakerInfos[1].SetAutoStaking(false)
	stakerInfos[2] = &statedb.StakerInfo{}
	stakerInfos[2].SetRewardReceiver(wl2.KeySet.PaymentAddress)
	stakerInfos[2].SetTxStakingID(common.HashH([]byte{2}))
	stakerInfos[2].SetAutoStaking(true)
	stakerInfos[3] = &statedb.StakerInfo{}
	stakerInfos[3].SetRewardReceiver(wl3.KeySet.PaymentAddress)
	stakerInfos[3].SetTxStakingID(common.HashH([]byte{3}))
	stakerInfos[3].SetAutoStaking(false)
	stakerInfos[4] = &statedb.StakerInfo{}
	stakerInfos[4].SetRewardReceiver(wl4.KeySet.PaymentAddress)
	stakerInfos[4].SetTxStakingID(common.HashH([]byte{4}))
	stakerInfos[4].SetAutoStaking(true)
	stakerInfos[5] = &statedb.StakerInfo{}
	stakerInfos[5].SetRewardReceiver(wl5.KeySet.PaymentAddress)
	stakerInfos[5].SetTxStakingID(common.HashH([]byte{5}))
	stakerInfos[5].SetAutoStaking(true)
	type args struct {
		committees           []*statedb.StakerInfo
		shardReceiveRewardV3 *instruction.ShardReceiveRewardV3
	}
	tests := []struct {
		name string
		args args
		want []*statedb.StakerInfo
	}{
		{
			name: "subset 0",
			args: args{
				committees:           stakerInfos,
				shardReceiveRewardV3: instruction.NewShardReceiveRewardV3().SetSubsetID(0),
			},
			want: []*statedb.StakerInfo{
				stakerInfos[0],
				stakerInfos[2],
				stakerInfos[4],
			},
		},
		{
			name: "subset 1",
			args: args{
				committees:           stakerInfos,
				shardReceiveRewardV3: instruction.NewShardReceiveRewardV3().SetSubsetID(1),
			},
			want: []*statedb.StakerInfo{
				stakerInfos[1],
				stakerInfos[3],
				stakerInfos[5],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCommitteeToPayRewardMultiset(tt.args.committees, tt.args.shardReceiveRewardV3); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCommitteeToPayRewardV3() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getYearOfBlockChain(t *testing.T) {
	config.LoadParam()
	config.Param().BlockTimeParam = make(map[string]int64)
	config.Param().BlockTimeParam = map[string]int64{
		BLOCKTIME_DEFAULT: 40,
		BLOCKTIME_20:      20,
		BLOCKTIME_10:      10,
	}
	type args struct {
		featuresManager *TSManager
		blkHeight       uint64
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "Test 1",
			args: args{
				featuresManager: &TSManager{
					Anchors: []anchorTime{
						{
							PreviousEndTime: 100,
							StartTime:       101,
							StartTimeslot:   10,
							Feature:         BLOCKTIME_20,
							BlockHeight:     1000,
						},
						{
							PreviousEndTime: 100,
							StartTime:       101,
							StartTimeslot:   10,
							Feature:         BLOCKTIME_10,
							BlockHeight:     2000,
						},
					},
					CurrentBlockVersion: 1,
					CurrentBlockTS:      1,
					CurrentProposeTime:  time.Now().Unix(),
				},
				blkHeight: 300,
			},
			want: 0,
		},
		{
			name: "Test 2",
			args: args{
				featuresManager: &TSManager{
					Anchors: []anchorTime{
						{
							PreviousEndTime: 100,
							StartTime:       101,
							StartTimeslot:   10,
							Feature:         BLOCKTIME_20,
							BlockHeight:     788000,
						},
						{
							PreviousEndTime: 100,
							StartTime:       101,
							StartTimeslot:   10,
							Feature:         BLOCKTIME_10,
							BlockHeight:     790820,
						},
					},
					CurrentBlockVersion: 1,
					CurrentBlockTS:      1,
					CurrentProposeTime:  time.Now().Unix(),
				},
				blkHeight: 789879,
			},
			want: 0,
		},
		{
			name: "Test 3",
			args: args{
				featuresManager: &TSManager{
					Anchors: []anchorTime{
						{
							PreviousEndTime: 100,
							StartTime:       101,
							StartTimeslot:   10,
							Feature:         BLOCKTIME_20,
							BlockHeight:     788000,
						},
						{
							PreviousEndTime: 100,
							StartTime:       101,
							StartTimeslot:   10,
							Feature:         BLOCKTIME_10,
							BlockHeight:     790818,
						},
					},
					CurrentBlockVersion: 1,
					CurrentBlockTS:      1,
					CurrentProposeTime:  time.Now().Unix(),
				},
				blkHeight: 790822,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getYearOfBlockChain(tt.args.featuresManager, tt.args.blkHeight); got != tt.want {
				t.Errorf("getYearOfBlockChain() = %v, want %v", got, tt.want)
			}
		})
	}
}

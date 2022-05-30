package blockchain

import (
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate/externalmocks"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/common/consensus"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/trie"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var (
	keys                []string
	blsKeys             []string
	validators          []*consensus.Validator
	incognitoKeys       []incognitokey.CommitteePublicKey
	warperDBStatedbTest statedb.DatabaseAccessWarper
)

var _ = func() (_ struct{}) {
	keys = []string{
		"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",
	}
	blsKeys = []string{
		"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",
	}
	for _, blsKey := range blsKeys {
		validator := new(consensus.Validator)
		temp, _, _ := base58.Base58Check{}.Decode(blsKey)
		validator.MiningKey = signatureschemes.MiningKey{}
		validator.MiningKey.PubKey = make(map[string][]byte)
		validator.MiningKey.PubKey[common.BlsConsensus] = temp
		validator.MiningKey.PubKey[common.BridgeConsensus] = []byte{}
		validators = append(validators, validator)
	}

	incognitoKeys, _ = incognitokey.CommitteeBase58KeyListToStruct(keys)
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	dbPath, err := ioutil.TempDir(os.TempDir(), "test")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestBeaconBestState_CalculateExpectedTotalBlock(t *testing.T) {
	type fields struct {
		NumberOfShardBlock map[byte]uint
	}
	type args struct {
		blockVersion int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[byte]uint
	}{
		{
			name: "moderate different between shards",
			args: args{
				blockVersion: types.SHARD_SFV3_VERSION,
			},
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 100,
					1: 150,
					2: 200,
					3: 200,
					4: 250,
					5: 300,
					6: 330,
					7: 340,
				},
			},
			want: map[byte]uint{
				0: 233,
				1: 233,
				2: 233,
				3: 233,
				4: 250,
				5: 300,
				6: 330,
				7: 340,
			},
		},
		{
			name: "big different between shards",
			args: args{
				blockVersion: types.SHARD_SFV3_VERSION,
			},
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 10,
					1: 0,
					2: 20,
					3: 200,
					4: 250,
					5: 300,
					6: 330,
					7: 340,
				},
			},
			want: map[byte]uint{
				0: 181,
				1: 181,
				2: 181,
				3: 200,
				4: 250,
				5: 300,
				6: 330,
				7: 340,
			},
		},
		{
			name: "only one shard big different compare to shards",
			args: args{
				blockVersion: types.SHARD_SFV3_VERSION,
			},
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 10,
					1: 300,
					2: 280,
					3: 290,
					4: 310,
					5: 300,
					6: 330,
					7: 340,
				},
			},
			want: map[byte]uint{
				0: 270,
				1: 300,
				2: 280,
				3: 290,
				4: 310,
				5: 300,
				6: 330,
				7: 340,
			},
		},
		{
			name: "0 all shard",
			args: args{
				blockVersion: types.SHARD_SFV3_VERSION,
			},
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 0,
					1: 0,
					2: 0,
					3: 0,
					4: 0,
					5: 0,
					6: 0,
					7: 0,
				},
			},
			want: map[byte]uint{
				0: 0,
				1: 0,
				2: 0,
				3: 0,
				4: 0,
				5: 0,
				6: 0,
				7: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconBestState{
				NumberOfShardBlock: tt.fields.NumberOfShardBlock,
			}
			if got := b.CalculateExpectedTotalBlock(tt.args.blockVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalculateExpectedTotalBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconBestState_GetExpectedTotalBlock(t *testing.T) {
	type fields struct {
		NumberOfShardBlock    map[byte]uint
		beaconCommitteeEngine committeestate.BeaconCommitteeState
	}

	type args struct {
		blockVersion int
	}

	mockCommittee := map[byte][]incognitokey.CommitteePublicKey{
		0: incognitoKeys[0:2],
		1: incognitoKeys[2:4],
		2: incognitoKeys[4:6],
		3: incognitoKeys[6:8],
		4: incognitoKeys[8:10],
		5: incognitoKeys[10:12],
		6: incognitoKeys[12:14],
		7: incognitoKeys[14:16],
	}

	mockState1 := &externalmocks.BeaconCommitteeState{}
	mockState1.On("GetShardCommittee").Return(mockCommittee).Times(4)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]uint
	}{
		{
			name: "moderate different between shards",
			args: args{
				blockVersion: types.SHARD_SFV3_VERSION,
			},
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 100,
					1: 150,
					2: 200,
					3: 200,
					4: 250,
					5: 300,
					6: 330,
					7: 340,
				},
				beaconCommitteeEngine: mockState1,
			},
			want: map[string]uint{
				keys[0]:  233,
				keys[1]:  233,
				keys[2]:  233,
				keys[3]:  233,
				keys[4]:  233,
				keys[5]:  233,
				keys[6]:  233,
				keys[7]:  233,
				keys[8]:  250,
				keys[9]:  250,
				keys[10]: 300,
				keys[11]: 300,
				keys[12]: 330,
				keys[13]: 330,
				keys[14]: 340,
				keys[15]: 340,
			},
		},
		{
			name: "big different between shards",
			args: args{
				blockVersion: types.SHARD_SFV3_VERSION,
			},
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 10,
					1: 0,
					2: 20,
					3: 200,
					4: 250,
					5: 300,
					6: 330,
					7: 340,
				},
				beaconCommitteeEngine: mockState1,
			},
			want: map[string]uint{
				keys[0]:  181,
				keys[1]:  181,
				keys[2]:  181,
				keys[3]:  181,
				keys[4]:  181,
				keys[5]:  181,
				keys[6]:  200,
				keys[7]:  200,
				keys[8]:  250,
				keys[9]:  250,
				keys[10]: 300,
				keys[11]: 300,
				keys[12]: 330,
				keys[13]: 330,
				keys[14]: 340,
				keys[15]: 340,
			},
		},
		{
			name: "only one shard big different compare to shards",
			args: args{
				blockVersion: types.SHARD_SFV3_VERSION,
			},
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 10,
					1: 300,
					2: 280,
					3: 290,
					4: 310,
					5: 300,
					6: 330,
					7: 340,
				},
				beaconCommitteeEngine: mockState1,
			},
			want: map[string]uint{
				keys[0]:  270,
				keys[1]:  270,
				keys[2]:  300,
				keys[3]:  300,
				keys[4]:  280,
				keys[5]:  280,
				keys[6]:  290,
				keys[7]:  290,
				keys[8]:  310,
				keys[9]:  310,
				keys[10]: 300,
				keys[11]: 300,
				keys[12]: 330,
				keys[13]: 330,
				keys[14]: 340,
				keys[15]: 340,
			},
		},
		{
			name: "0 all shard",
			args: args{
				blockVersion: types.SHARD_SFV3_VERSION,
			},
			fields: fields{
				NumberOfShardBlock: map[byte]uint{
					0: 0,
					1: 0,
					2: 0,
					3: 0,
					4: 0,
					5: 0,
					6: 0,
					7: 0,
				},
				beaconCommitteeEngine: mockState1,
			},
			want: map[string]uint{
				keys[0]:  0,
				keys[1]:  0,
				keys[2]:  0,
				keys[3]:  0,
				keys[4]:  0,
				keys[5]:  0,
				keys[6]:  0,
				keys[7]:  0,
				keys[8]:  0,
				keys[9]:  0,
				keys[10]: 0,
				keys[11]: 0,
				keys[12]: 0,
				keys[13]: 0,
				keys[14]: 0,
				keys[15]: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconBestState{
				NumberOfShardBlock:   tt.fields.NumberOfShardBlock,
				beaconCommitteeState: tt.fields.beaconCommitteeEngine,
			}
			if got := b.GetExpectedTotalBlock(tt.args.blockVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetExpectedTotalBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_filterNonSlashingCommittee(t *testing.T) {
	type args struct {
		committees         []*statedb.StakerInfoSlashingVersion
		slashingCommittees []string
	}
	tests := []struct {
		name string
		args args
		want []*statedb.StakerInfoSlashingVersion
	}{
		{
			name: "no slashing committee",
			args: args{
				committees: []*statedb.StakerInfoSlashingVersion{
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
				},
				slashingCommittees: []string{},
			},
			want: []*statedb.StakerInfoSlashingVersion{
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
			},
		},
		{
			name: "no slashing committee 1",
			args: args{
				committees: []*statedb.StakerInfoSlashingVersion{
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
				},
				slashingCommittees: []string{
					keys[9],
				},
			},
			want: []*statedb.StakerInfoSlashingVersion{
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
			},
		},
		{
			name: "slash some committee",
			args: args{
				committees: []*statedb.StakerInfoSlashingVersion{
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
				},
				slashingCommittees: []string{
					keys[0],
					keys[1],
					keys[2],
				},
			},
			want: []*statedb.StakerInfoSlashingVersion{
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
			},
		},
		{
			name: "slash some committee 1",
			args: args{
				committees: []*statedb.StakerInfoSlashingVersion{
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
				},
				slashingCommittees: []string{
					keys[0],
					keys[3],
					keys[5],
				},
			},
			want: []*statedb.StakerInfoSlashingVersion{
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
			},
		},
		{
			name: "slash all committee",
			args: args{
				committees: []*statedb.StakerInfoSlashingVersion{
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
				},
				slashingCommittees: []string{
					keys[0],
					keys[1],
					keys[2],
					keys[3],
					keys[4],
					keys[5],
					keys[6],
					keys[7],
					keys[8],
				},
			},
			want: []*statedb.StakerInfoSlashingVersion{},
		},
		{
			name: "some slashing committee not in committees list",
			args: args{
				committees: []*statedb.StakerInfoSlashingVersion{
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
				},
				slashingCommittees: []string{
					keys[0],
					keys[2],
					keys[5],
					keys[10],
				},
			},
			want: []*statedb.StakerInfoSlashingVersion{
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterNonSlashingCommittee(tt.args.committees, tt.args.slashingCommittees); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterNonSlashingCommittee() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconBestState_GetNonSlashingCommittee(t *testing.T) {
	sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	statedb.StoreSlashingCommittee(sDB, 1, map[byte][]string{
		0: []string{},
	})
	statedb.StoreSlashingCommittee(sDB, 2, map[byte][]string{
		0: []string{
			keys[0],
			keys[1],
			keys[2],
		},
	})
	statedb.StoreSlashingCommittee(sDB, 3, map[byte][]string{
		0: []string{
			keys[0],
			keys[3],
			keys[5],
		},
	})
	statedb.StoreSlashingCommittee(sDB, 4, map[byte][]string{
		0: []string{
			keys[0],
			keys[1],
			keys[2],
			keys[3],
			keys[4],
			keys[5],
			keys[6],
			keys[7],
			keys[8],
		},
	})
	rootHash, _ := sDB.Commit(true)
	sDB.Database().TrieDB().Commit(rootHash, false)

	sDB2, _ := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)

	type fields struct {
		Epoch        uint64
		slashStateDB *statedb.StateDB
	}
	type args struct {
		committees []*statedb.StakerInfoSlashingVersion
		epoch      uint64
		shardID    byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*statedb.StakerInfoSlashingVersion
		wantErr bool
	}{
		{
			name: "input epoch higher than best state epoch",
			fields: fields{
				Epoch:        5,
				slashStateDB: sDB2,
			},
			args: args{
				committees: []*statedb.StakerInfoSlashingVersion{
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
				},
				epoch:   5,
				shardID: 0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no slashing committee",
			fields: fields{
				Epoch:        5,
				slashStateDB: sDB2,
			},
			args: args{
				committees: []*statedb.StakerInfoSlashingVersion{
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
				},
				epoch:   1,
				shardID: 0,
			},
			want: []*statedb.StakerInfoSlashingVersion{
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
			},
			wantErr: false,
		},
		{
			name: "slash some committee",
			fields: fields{
				Epoch:        5,
				slashStateDB: sDB2,
			},
			args: args{
				committees: []*statedb.StakerInfoSlashingVersion{
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
				},
				epoch:   2,
				shardID: 0,
			},
			want: []*statedb.StakerInfoSlashingVersion{
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
			},
			wantErr: false,
		},
		{
			name: "slash some committee 1",
			fields: fields{
				Epoch:        5,
				slashStateDB: sDB2,
			},
			args: args{
				committees: []*statedb.StakerInfoSlashingVersion{
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
				},
				epoch:   3,
				shardID: 0,
			},
			want: []*statedb.StakerInfoSlashingVersion{
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
				statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
			},
			wantErr: false,
		},
		{
			name: "slash all committee",
			fields: fields{
				Epoch:        5,
				slashStateDB: sDB2,
			},
			args: args{
				committees: []*statedb.StakerInfoSlashingVersion{
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[0]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[1]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[2]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[3]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[4]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[5]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[6]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[7]),
					statedb.NewStakerInfoSlashingVersionWithCommittee(keys[8]),
				},
				epoch:   4,
				shardID: 0,
			},
			want:    []*statedb.StakerInfoSlashingVersion{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beaconBestState := &BeaconBestState{
				Epoch:        tt.fields.Epoch,
				slashStateDB: tt.fields.slashStateDB,
			}
			got, err := beaconBestState.GetNonSlashingCommittee(tt.args.committees, tt.args.epoch, tt.args.shardID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNonSlashingCommittee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNonSlashingCommittee() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconBestState_GetFinishSyncingValidators(t *testing.T) {

	beaconCommitteeStateMocks1 := &externalmocks.BeaconCommitteeState{}
	beaconCommitteeStateMocks1.On("GetSyncingValidators").Return(
		map[byte][]incognitokey.CommitteePublicKey{
			0: []incognitokey.CommitteePublicKey{},
		},
	)

	beaconCommitteeStateMocks2 := &externalmocks.BeaconCommitteeState{}
	beaconCommitteeStateMocks2.On("GetSyncingValidators").Return(
		map[byte][]incognitokey.CommitteePublicKey{
			0: incognitoKeys[:5],
		},
	)

	beaconCommitteeStateMocks3 := &externalmocks.BeaconCommitteeState{}
	beaconCommitteeStateMocks3.On("GetSyncingValidators").Return(
		map[byte][]incognitokey.CommitteePublicKey{
			0: incognitoKeys[3:5],
		},
	)

	beaconCommitteeStateMocks4 := &externalmocks.BeaconCommitteeState{}
	beaconCommitteeStateMocks4.On("GetSyncingValidators").Return(
		map[byte][]incognitokey.CommitteePublicKey{
			0: incognitoKeys,
		},
	)

	beaconCommitteeStateMocks5 := &externalmocks.BeaconCommitteeState{}
	beaconCommitteeStateMocks5.On("GetSyncingValidators").Return(
		map[byte][]incognitokey.CommitteePublicKey{
			0: incognitoKeys,
		},
	)

	type fields struct {
		beaconCommitteeState committeestate.BeaconCommitteeState
	}
	type args struct {
		validatorFromUserKeys []*consensus.Validator
		shardID               byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "beaconBestState.SyncValidator = 0, SyncValidatorFromUser > 0",
			fields: fields{
				beaconCommitteeState: beaconCommitteeStateMocks1,
			},
			args: args{
				validatorFromUserKeys: validators[:5],
				shardID:               0,
			},
			want: []string{},
		},
		{
			name: "beaconBestState.SyncValidator > 0, SyncValidatorFromUser > 0, the same 1",
			fields: fields{
				beaconCommitteeState: beaconCommitteeStateMocks2,
			},
			args: args{
				validatorFromUserKeys: validators[:5],
				shardID:               0,
			},
			want: keys[:5],
		},
		{
			name: "beaconBestState.SyncValidator > 0, SyncValidatorFromUser > 0, different",
			fields: fields{
				beaconCommitteeState: beaconCommitteeStateMocks3,
			},
			args: args{
				validatorFromUserKeys: validators[:4],
				shardID:               0,
			},
			want: keys[3:4],
		},
		{
			name: "beaconBestState.SyncValidator > 0, SyncValidatorFromUser = 0",
			fields: fields{
				beaconCommitteeState: beaconCommitteeStateMocks4,
			},
			args: args{
				validatorFromUserKeys: []*consensus.Validator{},
				shardID:               0,
			},
			want: []string{},
		},
		{
			name: "beaconBestState.SyncValidator > 0, SyncValidatorFromUser > 0, the same 2",
			fields: fields{
				beaconCommitteeState: beaconCommitteeStateMocks5,
			},
			args: args{
				validatorFromUserKeys: validators[2:10],
				shardID:               0,
			},
			want: keys[2:10],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beaconBestState := &BeaconBestState{
				beaconCommitteeState: tt.fields.beaconCommitteeState,
			}
			if _, got := beaconBestState.ExtractFinishSyncingValidators(tt.args.validatorFromUserKeys, tt.args.shardID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractFinishSyncingValidators() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMaxCommitteeSize(t *testing.T) {
	type args struct {
		currentMaxCommitteeSize int
		triggerFeature          map[string]uint64
		beaconHeight            uint64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "empty increase max committee size",
			args: args{
				currentMaxCommitteeSize: 10,
				triggerFeature:          make(map[string]uint64),
				beaconHeight:            100,
			},
			want: 10,
		},
		{
			name: "beacon height < increase max committee size break point",
			args: args{
				currentMaxCommitteeSize: 10,
				triggerFeature: map[string]uint64{
					MAX_COMMITTEE_SIZE_48_FEATURE: 103,
				},
				beaconHeight: 101,
			},
			want: 10,
		},
		{
			name: "beacon height = increase max committee size break point",
			args: args{
				currentMaxCommitteeSize: 10,
				triggerFeature: map[string]uint64{
					MAX_COMMITTEE_SIZE_48_FEATURE: 102,
				},
				beaconHeight: 102,
			},
			want: 48,
		},
		{
			name: "beacon height > increase max committee size break point",
			args: args{
				currentMaxCommitteeSize: 10,
				triggerFeature: map[string]uint64{
					MAX_COMMITTEE_SIZE_48_FEATURE: 102,
				},
				beaconHeight: 103,
			},
			want: 48,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMaxCommitteeSize(tt.args.currentMaxCommitteeSize, tt.args.triggerFeature, tt.args.beaconHeight); got != tt.want {
				t.Errorf("GetMaxCommitteeSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

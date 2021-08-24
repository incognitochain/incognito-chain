package blockchain

import (
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate/externalmocks"
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
	tests := []struct {
		name   string
		fields fields
		want   map[byte]uint
	}{
		{
			name: "moderate different between shards",
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
			if got := b.CalculateExpectedTotalBlock(); !reflect.DeepEqual(got, tt.want) {
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
		want   map[string]uint
	}{
		{
			name: "moderate different between shards",
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
			if got := b.GetExpectedTotalBlock(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetExpectedTotalBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_filterNonSlashingCommittee(t *testing.T) {
	type args struct {
		committees         []*statedb.StakerInfoV2
		slashingCommittees []string
	}
	tests := []struct {
		name string
		args args
		want []*statedb.StakerInfoV2
	}{
		{
			name: "no slashing committee",
			args: args{
				committees: []*statedb.StakerInfoV2{
					statedb.NewStakerInfoV2WithCommittee(keys[0]),
					statedb.NewStakerInfoV2WithCommittee(keys[1]),
					statedb.NewStakerInfoV2WithCommittee(keys[2]),
					statedb.NewStakerInfoV2WithCommittee(keys[3]),
					statedb.NewStakerInfoV2WithCommittee(keys[4]),
					statedb.NewStakerInfoV2WithCommittee(keys[5]),
					statedb.NewStakerInfoV2WithCommittee(keys[6]),
					statedb.NewStakerInfoV2WithCommittee(keys[7]),
					statedb.NewStakerInfoV2WithCommittee(keys[8]),
				},
				slashingCommittees: []string{},
			},
			want: []*statedb.StakerInfoV2{
				statedb.NewStakerInfoV2WithCommittee(keys[0]),
				statedb.NewStakerInfoV2WithCommittee(keys[1]),
				statedb.NewStakerInfoV2WithCommittee(keys[2]),
				statedb.NewStakerInfoV2WithCommittee(keys[3]),
				statedb.NewStakerInfoV2WithCommittee(keys[4]),
				statedb.NewStakerInfoV2WithCommittee(keys[5]),
				statedb.NewStakerInfoV2WithCommittee(keys[6]),
				statedb.NewStakerInfoV2WithCommittee(keys[7]),
				statedb.NewStakerInfoV2WithCommittee(keys[8]),
			},
		},
		{
			name: "no slashing committee 1",
			args: args{
				committees: []*statedb.StakerInfoV2{
					statedb.NewStakerInfoV2WithCommittee(keys[0]),
					statedb.NewStakerInfoV2WithCommittee(keys[1]),
					statedb.NewStakerInfoV2WithCommittee(keys[2]),
					statedb.NewStakerInfoV2WithCommittee(keys[3]),
					statedb.NewStakerInfoV2WithCommittee(keys[4]),
					statedb.NewStakerInfoV2WithCommittee(keys[5]),
					statedb.NewStakerInfoV2WithCommittee(keys[6]),
					statedb.NewStakerInfoV2WithCommittee(keys[7]),
					statedb.NewStakerInfoV2WithCommittee(keys[8]),
				},
				slashingCommittees: []string{
					keys[9],
				},
			},
			want: []*statedb.StakerInfoV2{
				statedb.NewStakerInfoV2WithCommittee(keys[0]),
				statedb.NewStakerInfoV2WithCommittee(keys[1]),
				statedb.NewStakerInfoV2WithCommittee(keys[2]),
				statedb.NewStakerInfoV2WithCommittee(keys[3]),
				statedb.NewStakerInfoV2WithCommittee(keys[4]),
				statedb.NewStakerInfoV2WithCommittee(keys[5]),
				statedb.NewStakerInfoV2WithCommittee(keys[6]),
				statedb.NewStakerInfoV2WithCommittee(keys[7]),
				statedb.NewStakerInfoV2WithCommittee(keys[8]),
			},
		},
		{
			name: "slash some committee",
			args: args{
				committees: []*statedb.StakerInfoV2{
					statedb.NewStakerInfoV2WithCommittee(keys[0]),
					statedb.NewStakerInfoV2WithCommittee(keys[1]),
					statedb.NewStakerInfoV2WithCommittee(keys[2]),
					statedb.NewStakerInfoV2WithCommittee(keys[3]),
					statedb.NewStakerInfoV2WithCommittee(keys[4]),
					statedb.NewStakerInfoV2WithCommittee(keys[5]),
					statedb.NewStakerInfoV2WithCommittee(keys[6]),
					statedb.NewStakerInfoV2WithCommittee(keys[7]),
					statedb.NewStakerInfoV2WithCommittee(keys[8]),
				},
				slashingCommittees: []string{
					keys[0],
					keys[1],
					keys[2],
				},
			},
			want: []*statedb.StakerInfoV2{
				statedb.NewStakerInfoV2WithCommittee(keys[3]),
				statedb.NewStakerInfoV2WithCommittee(keys[4]),
				statedb.NewStakerInfoV2WithCommittee(keys[5]),
				statedb.NewStakerInfoV2WithCommittee(keys[6]),
				statedb.NewStakerInfoV2WithCommittee(keys[7]),
				statedb.NewStakerInfoV2WithCommittee(keys[8]),
			},
		},
		{
			name: "slash some committee 1",
			args: args{
				committees: []*statedb.StakerInfoV2{
					statedb.NewStakerInfoV2WithCommittee(keys[0]),
					statedb.NewStakerInfoV2WithCommittee(keys[1]),
					statedb.NewStakerInfoV2WithCommittee(keys[2]),
					statedb.NewStakerInfoV2WithCommittee(keys[3]),
					statedb.NewStakerInfoV2WithCommittee(keys[4]),
					statedb.NewStakerInfoV2WithCommittee(keys[5]),
					statedb.NewStakerInfoV2WithCommittee(keys[6]),
					statedb.NewStakerInfoV2WithCommittee(keys[7]),
					statedb.NewStakerInfoV2WithCommittee(keys[8]),
				},
				slashingCommittees: []string{
					keys[0],
					keys[3],
					keys[5],
				},
			},
			want: []*statedb.StakerInfoV2{
				statedb.NewStakerInfoV2WithCommittee(keys[1]),
				statedb.NewStakerInfoV2WithCommittee(keys[2]),
				statedb.NewStakerInfoV2WithCommittee(keys[4]),
				statedb.NewStakerInfoV2WithCommittee(keys[6]),
				statedb.NewStakerInfoV2WithCommittee(keys[7]),
				statedb.NewStakerInfoV2WithCommittee(keys[8]),
			},
		},
		{
			name: "slash all committee",
			args: args{
				committees: []*statedb.StakerInfoV2{
					statedb.NewStakerInfoV2WithCommittee(keys[0]),
					statedb.NewStakerInfoV2WithCommittee(keys[1]),
					statedb.NewStakerInfoV2WithCommittee(keys[2]),
					statedb.NewStakerInfoV2WithCommittee(keys[3]),
					statedb.NewStakerInfoV2WithCommittee(keys[4]),
					statedb.NewStakerInfoV2WithCommittee(keys[5]),
					statedb.NewStakerInfoV2WithCommittee(keys[6]),
					statedb.NewStakerInfoV2WithCommittee(keys[7]),
					statedb.NewStakerInfoV2WithCommittee(keys[8]),
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
			want: []*statedb.StakerInfoV2{},
		},
		{
			name: "some slashing committee not in committees list",
			args: args{
				committees: []*statedb.StakerInfoV2{
					statedb.NewStakerInfoV2WithCommittee(keys[0]),
					statedb.NewStakerInfoV2WithCommittee(keys[1]),
					statedb.NewStakerInfoV2WithCommittee(keys[2]),
					statedb.NewStakerInfoV2WithCommittee(keys[3]),
					statedb.NewStakerInfoV2WithCommittee(keys[4]),
					statedb.NewStakerInfoV2WithCommittee(keys[5]),
					statedb.NewStakerInfoV2WithCommittee(keys[6]),
					statedb.NewStakerInfoV2WithCommittee(keys[7]),
					statedb.NewStakerInfoV2WithCommittee(keys[8]),
				},
				slashingCommittees: []string{
					keys[0],
					keys[2],
					keys[5],
					keys[10],
				},
			},
			want: []*statedb.StakerInfoV2{
				statedb.NewStakerInfoV2WithCommittee(keys[1]),
				statedb.NewStakerInfoV2WithCommittee(keys[3]),
				statedb.NewStakerInfoV2WithCommittee(keys[4]),
				statedb.NewStakerInfoV2WithCommittee(keys[6]),
				statedb.NewStakerInfoV2WithCommittee(keys[7]),
				statedb.NewStakerInfoV2WithCommittee(keys[8]),
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
		committees []*statedb.StakerInfoV2
		epoch      uint64
		shardID    byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*statedb.StakerInfoV2
		wantErr bool
	}{
		{
			name: "input epoch higher than best state epoch",
			fields: fields{
				Epoch:        5,
				slashStateDB: sDB2,
			},
			args: args{
				committees: []*statedb.StakerInfoV2{
					statedb.NewStakerInfoV2WithCommittee(keys[0]),
					statedb.NewStakerInfoV2WithCommittee(keys[1]),
					statedb.NewStakerInfoV2WithCommittee(keys[2]),
					statedb.NewStakerInfoV2WithCommittee(keys[3]),
					statedb.NewStakerInfoV2WithCommittee(keys[4]),
					statedb.NewStakerInfoV2WithCommittee(keys[5]),
					statedb.NewStakerInfoV2WithCommittee(keys[6]),
					statedb.NewStakerInfoV2WithCommittee(keys[7]),
					statedb.NewStakerInfoV2WithCommittee(keys[8]),
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
				committees: []*statedb.StakerInfoV2{
					statedb.NewStakerInfoV2WithCommittee(keys[0]),
					statedb.NewStakerInfoV2WithCommittee(keys[1]),
					statedb.NewStakerInfoV2WithCommittee(keys[2]),
					statedb.NewStakerInfoV2WithCommittee(keys[3]),
					statedb.NewStakerInfoV2WithCommittee(keys[4]),
					statedb.NewStakerInfoV2WithCommittee(keys[5]),
					statedb.NewStakerInfoV2WithCommittee(keys[6]),
					statedb.NewStakerInfoV2WithCommittee(keys[7]),
					statedb.NewStakerInfoV2WithCommittee(keys[8]),
				},
				epoch:   1,
				shardID: 0,
			},
			want: []*statedb.StakerInfoV2{
				statedb.NewStakerInfoV2WithCommittee(keys[0]),
				statedb.NewStakerInfoV2WithCommittee(keys[1]),
				statedb.NewStakerInfoV2WithCommittee(keys[2]),
				statedb.NewStakerInfoV2WithCommittee(keys[3]),
				statedb.NewStakerInfoV2WithCommittee(keys[4]),
				statedb.NewStakerInfoV2WithCommittee(keys[5]),
				statedb.NewStakerInfoV2WithCommittee(keys[6]),
				statedb.NewStakerInfoV2WithCommittee(keys[7]),
				statedb.NewStakerInfoV2WithCommittee(keys[8]),
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
				committees: []*statedb.StakerInfoV2{
					statedb.NewStakerInfoV2WithCommittee(keys[0]),
					statedb.NewStakerInfoV2WithCommittee(keys[1]),
					statedb.NewStakerInfoV2WithCommittee(keys[2]),
					statedb.NewStakerInfoV2WithCommittee(keys[3]),
					statedb.NewStakerInfoV2WithCommittee(keys[4]),
					statedb.NewStakerInfoV2WithCommittee(keys[5]),
					statedb.NewStakerInfoV2WithCommittee(keys[6]),
					statedb.NewStakerInfoV2WithCommittee(keys[7]),
					statedb.NewStakerInfoV2WithCommittee(keys[8]),
				},
				epoch:   2,
				shardID: 0,
			},
			want: []*statedb.StakerInfoV2{
				statedb.NewStakerInfoV2WithCommittee(keys[3]),
				statedb.NewStakerInfoV2WithCommittee(keys[4]),
				statedb.NewStakerInfoV2WithCommittee(keys[5]),
				statedb.NewStakerInfoV2WithCommittee(keys[6]),
				statedb.NewStakerInfoV2WithCommittee(keys[7]),
				statedb.NewStakerInfoV2WithCommittee(keys[8]),
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
				committees: []*statedb.StakerInfoV2{
					statedb.NewStakerInfoV2WithCommittee(keys[0]),
					statedb.NewStakerInfoV2WithCommittee(keys[1]),
					statedb.NewStakerInfoV2WithCommittee(keys[2]),
					statedb.NewStakerInfoV2WithCommittee(keys[3]),
					statedb.NewStakerInfoV2WithCommittee(keys[4]),
					statedb.NewStakerInfoV2WithCommittee(keys[5]),
					statedb.NewStakerInfoV2WithCommittee(keys[6]),
					statedb.NewStakerInfoV2WithCommittee(keys[7]),
					statedb.NewStakerInfoV2WithCommittee(keys[8]),
				},
				epoch:   3,
				shardID: 0,
			},
			want: []*statedb.StakerInfoV2{
				statedb.NewStakerInfoV2WithCommittee(keys[1]),
				statedb.NewStakerInfoV2WithCommittee(keys[2]),
				statedb.NewStakerInfoV2WithCommittee(keys[4]),
				statedb.NewStakerInfoV2WithCommittee(keys[6]),
				statedb.NewStakerInfoV2WithCommittee(keys[7]),
				statedb.NewStakerInfoV2WithCommittee(keys[8]),
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
				committees: []*statedb.StakerInfoV2{
					statedb.NewStakerInfoV2WithCommittee(keys[0]),
					statedb.NewStakerInfoV2WithCommittee(keys[1]),
					statedb.NewStakerInfoV2WithCommittee(keys[2]),
					statedb.NewStakerInfoV2WithCommittee(keys[3]),
					statedb.NewStakerInfoV2WithCommittee(keys[4]),
					statedb.NewStakerInfoV2WithCommittee(keys[5]),
					statedb.NewStakerInfoV2WithCommittee(keys[6]),
					statedb.NewStakerInfoV2WithCommittee(keys[7]),
					statedb.NewStakerInfoV2WithCommittee(keys[8]),
				},
				epoch:   4,
				shardID: 0,
			},
			want:    []*statedb.StakerInfoV2{},
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

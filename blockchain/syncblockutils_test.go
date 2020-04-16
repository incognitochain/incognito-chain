package blockchain

import (
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
)

// TODO add more test case

func setupTestGetMissingCrossShardBlock(
	fromShard byte,
	toShard byte,
	start uint64,
	listWantedBlock []uint64,
	db incdb.Database,
) {
	if len(listWantedBlock) == 0 {
		return
	}
	rawdbv2.StoreCrossShardNextHeight(db, fromShard, toShard, start, listWantedBlock[0])
	for i := 0; i < len(listWantedBlock)-1; i++ {
		rawdbv2.StoreCrossShardNextHeight(db, fromShard, toShard, listWantedBlock[i], listWantedBlock[i+1])
	}
	return
}

func TestGetMissingCrossShardBlock(t *testing.T) {
	type args struct {
		db                  incdb.Database
		bestCrossShardState map[byte]map[byte]uint64
		latestValidHeight   map[byte]uint64
		userShardID         byte
	}
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Errorf("failed to create temp dir: %+v", err)
	}
	t.Log(dbPath)
	db, err := incdb.Open("leveldb", dbPath)

	tests := []struct {
		name string
		args args
		want map[byte][]uint64
	}{
		{
			name: "Case 1",
			args: args{
				db:          db,
				userShardID: 1,
				bestCrossShardState: map[byte]map[byte]uint64{
					0: {
						1: 10,
					},
					1: {
						1: 10,
					},
					2: {
						1: 3,
					},
					4: {
						1: 5,
					},
				},
				latestValidHeight: map[byte]uint64{
					0: 2,
					1: 1,
					2: 1,
					3: 1,
					4: 1,
					5: 1,
					6: 1,
					7: 1,
				},
			},
			want: map[byte][]uint64{
				0: []uint64{3, 5, 7, 9, 10},
				1: []uint64{},
				2: []uint64{2, 3},
				3: []uint64{},
				4: []uint64{2, 3, 5},
				5: []uint64{},
				6: []uint64{},
				7: []uint64{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for fromShard, listWanted := range tt.want {
				setupTestGetMissingCrossShardBlock(fromShard, tt.args.userShardID, uint64(tt.args.latestValidHeight[fromShard]), listWanted, tt.args.db)
			}
			if got := GetMissingCrossShardBlock(tt.args.db, tt.args.bestCrossShardState, tt.args.latestValidHeight, tt.args.userShardID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMissingCrossShardBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func setupTestGetMissingBlockHashesFromPeersState(
	want map[int][]common.Hash,
) (
	map[string]*PeerState,
	map[byte]struct{},
	map[byte]ShardBestState,
	BeaconBestState,
) {
	bcBestState := new(BeaconBestState)
	shards := map[byte]struct{}{}
	shsBestState := map[byte]ShardBestState{}
	peersState := map[string]*PeerState{}
	for cID, hashes := range want {
		if cID == -1 {
			bcBestState.BeaconHeight = 100
			bcBestState.BestBlockHash = common.HashH([]byte{0, 1, 2})
			for _, blkHash := range hashes {
				peerState := &PeerState{
					Beacon: &ChainState{
						BlockHash: blkHash,
						Height:    100,
					},
				}
				peersState[blkHash.String()] = peerState
			}
		} else {
			shards[byte(cID)] = struct{}{}
			shBestState := new(ShardBestState)
			shBestState.ShardHeight = 100
			shBestState.BestBlockHash = common.HashH([]byte{0, 1, 2, byte(cID)})
			shsBestState[byte(cID)] = *shBestState
			for _, blkHash := range hashes {
				peerState := &PeerState{
					Beacon: &ChainState{
						BlockHash: common.HashH([]byte{0, 1, 2}),
						Height:    100,
					},
					Shard: map[byte]*ChainState{
						byte(cID): &ChainState{
							BlockHash: blkHash,
							Height:    100,
						},
					},
				}
				peersState[blkHash.String()] = peerState
			}
		}
	}
	return peersState, shards, shsBestState, *bcBestState
}

func TestGetMissingBlockHashesFromPeersState(t *testing.T) {
	wantedShard := 8
	numOfBlkHash := 10
	wantedBlkHash := map[int][]common.Hash{}
	for i := -1; i < wantedShard; i++ {
		hashes := []common.Hash{}
		for j := 0; j < numOfBlkHash; j++ {
			hashes = append(hashes, common.HashH([]byte{0, 1, 2, byte(i), byte(j)}))
		}
		sort.Slice(hashes, func(i int, j int) bool {
			res, _ := hashes[i].Cmp(&hashes[j])
			return res == -1
		})
		wantedBlkHash[i] = hashes
	}

	peersState, shards, shsBestState, bcBestState := setupTestGetMissingBlockHashesFromPeersState(wantedBlkHash)
	shardBestStateGetter := func(shardID byte) *ShardBestState {
		shBestState := shsBestState[shardID]
		return &shBestState
	}

	beaconBestStateGetter := func() *BeaconBestState {
		return &bcBestState
	}

	got := GetMissingBlockHashesFromPeersState(
		peersState,
		shards,
		beaconBestStateGetter,
		shardBestStateGetter,
	)
	for key, value := range got {
		sort.Slice(value, func(i int, j int) bool {
			res, _ := value[i].Cmp(&value[j])
			return res == -1
		})
		got[key] = value
	}
	if !reflect.DeepEqual(got, wantedBlkHash) {
		t.Errorf("GetMissingBlockHashesFromPeersState() = %v, want %v", got, wantedBlkHash)
	}

}

func TestGetMissingBlockInPool(t *testing.T) {
	type args struct {
		latestHeight    uint64
		listPendingBlks []uint64
	}
	tests := []struct {
		name string
		args args
		want []uint64
	}{
		{
			name: "Happy case",
			args: args{
				latestHeight:    10,
				listPendingBlks: []uint64{3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 15, 17, 21, 22, 24},
			},
			want: []uint64{12, 14, 16, 18, 19, 20, 23},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMissingBlockInPool(tt.args.latestHeight, tt.args.listPendingBlks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMissingBlockInPool() = %v, want %v", got, tt.want)
			}
		})
	}
}

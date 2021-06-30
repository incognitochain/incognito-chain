package stats

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

var (
	IsEnableBPV3Stats bool = false
)
var (
	shardEpochBPV3StatsPrefix  = []byte("s-b-bpv3-stats-epoch")
	shardHeightBPV3StatsPrefix = []byte("s-b-bpv3-stats-detail-height")
	split                      = []byte("=[=]=")
)

type NumberOfBlockInOneEpochStats struct {
	Epoch uint64
	Odd   int
	Even  int
}

func NewBlockInOneEpochStats(epoch uint64) *NumberOfBlockInOneEpochStats {
	return &NumberOfBlockInOneEpochStats{
		Epoch: epoch,
	}
}

type DetailBlockInOneEpochStats struct {
	Height     uint64
	Producer   string
	ProducerID int
	Proposer   string
	ProposerID int
	VotersID   []int
}

func NewDetailBlockInOneEpochStats(height uint64) *DetailBlockInOneEpochStats {
	return &DetailBlockInOneEpochStats{
		Height: height,
	}
}

func getShardEpochBPV3StatsKey(shardID byte, epoch uint64) []byte {
	key := append([]byte{}, shardEpochBPV3StatsPrefix...)
	key = append(key, split...)
	key = append(key, shardID)
	key = append(key, split...)
	key = append(key, common.Uint64ToBytes(epoch)...)
	return key
}

func getShardHeightBPV3StatsKey(shardID byte, epoch uint64, height uint64) []byte {
	key := append([]byte{}, shardHeightBPV3StatsPrefix...)
	key = append(key, split...)
	key = append(key, shardID)
	key = append(key, split...)
	key = append(key, common.Uint64ToBytes(epoch)...)
	key = append(key, split...)
	key = append(key, common.Uint64ToBytes(height)...)
	return key
}

func getShardHeightBPV3StatsPrefix(shardID byte, epoch uint64) []byte {
	key := append([]byte{}, shardHeightBPV3StatsPrefix...)
	key = append(key, split...)
	key = append(key, shardID)
	key = append(key, split...)
	key = append(key, common.Uint64ToBytes(epoch)...)
	key = append(key, split...)
	return key
}

func GetShardHeightBPV3Stats(db incdb.Database, shardID byte, epoch uint64) (map[uint64]*DetailBlockInOneEpochStats, error) {
	m := make(map[uint64]*DetailBlockInOneEpochStats)
	prefix := getShardHeightBPV3StatsPrefix(shardID, epoch)
	iterator := db.NewIteratorWithPrefix(prefix)
	for iterator.Next() {
		temp := make([]byte, len(iterator.Value()))
		copy(temp, iterator.Value())
		detailBlockInOneEpochStats := &DetailBlockInOneEpochStats{}
		err := json.Unmarshal(temp, detailBlockInOneEpochStats)
		if err != nil {
			return m, err
		}
		m[detailBlockInOneEpochStats.Height] = detailBlockInOneEpochStats
	}
	return m, nil
}

func GetShardEpochBPV3Stats(db incdb.Database, shardID byte, epoch uint64) (*NumberOfBlockInOneEpochStats, error) {
	key := getShardEpochBPV3StatsKey(shardID, epoch)
	numberOfBlockInOneEpochStats := NewBlockInOneEpochStats(epoch)
	temp, err := db.Get(key)
	if err != nil {
		return numberOfBlockInOneEpochStats, err
	}
	err2 := json.Unmarshal(temp, numberOfBlockInOneEpochStats)
	if err2 != nil {
		return numberOfBlockInOneEpochStats, err
	}
	return numberOfBlockInOneEpochStats, nil
}

func UpdateBPV3Stats(db incdb.Database, shardBlock *types.ShardBlock, subsetID int, totalCommittees []incognitokey.CommitteePublicKey) error {
	if err := addNumberOfBlockInOneEpochStats(db, shardBlock.Header.ShardID, shardBlock.Header.Epoch, subsetID); err != nil {
		return err
	}
	if err := addShardBlockDetailStats(db, shardBlock, totalCommittees); err != nil {
		return err
	}
	return nil
}

func addShardBlockDetailStats(db incdb.Database, shardBlock *types.ShardBlock, totalCommittees []incognitokey.CommitteePublicKey) error {
	decodedValidationData, err := consensustypes.DecodeValidationData(shardBlock.ValidationData)
	if err != nil {
		return err
	}
	detailBlockInOneEpochStats := NewDetailBlockInOneEpochStats(shardBlock.Header.Height)
	detailBlockInOneEpochStats.VotersID = append([]int{}, decodedValidationData.ValidatiorsIdx...)
	detailBlockInOneEpochStats.Producer = shardBlock.Header.Producer
	detailBlockInOneEpochStats.Proposer = shardBlock.Header.Proposer
	for idx, committee := range totalCommittees {
		v, _ := committee.ToBase58()
		if v == shardBlock.Header.Producer {
			detailBlockInOneEpochStats.ProducerID = idx
		}
		if v == shardBlock.Header.Proposer {
			detailBlockInOneEpochStats.ProposerID = idx
		}
	}

	key := getShardHeightBPV3StatsKey(
		shardBlock.Header.ShardID,
		shardBlock.Header.Epoch,
		shardBlock.Header.Height,
	)
	value, err := json.Marshal(detailBlockInOneEpochStats)
	if err != nil {
		return err
	}
	err2 := db.Put(key, value)
	if err2 != nil {
		return err2
	}

	return nil
}

func addNumberOfBlockInOneEpochStats(db incdb.Database, shardID byte, epoch uint64, subsetID int) error {
	blockInOneEpochStats := NewBlockInOneEpochStats(epoch)
	key := getShardEpochBPV3StatsKey(shardID, epoch)
	has, err := db.Has(key)
	if err != nil {
		return err
	}
	if has {
		tempBlockInOneEpochStats, err := db.Get(key)
		if err != nil {
			return err
		}
		err2 := json.Unmarshal(tempBlockInOneEpochStats, blockInOneEpochStats)
		if err2 != nil {
			return err2
		}
	}
	if subsetID%2 == 0 {
		blockInOneEpochStats.Even += 1
	} else {
		blockInOneEpochStats.Odd += 1
	}
	tempBlockInOneEpochStats2, err := json.Marshal(blockInOneEpochStats)
	if err != nil {
		return err
	}
	if err := db.Put(key, tempBlockInOneEpochStats2); err != nil {
		return err
	}
	return nil
}

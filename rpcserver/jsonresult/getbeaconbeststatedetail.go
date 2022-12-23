package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type committeeKeySetAutoStake struct {
	IncPubKey    string
	MiningPubKey map[string]string
	IsAutoStake  bool
}

type GetBeaconBestStateDetail struct {
	BestBlockHash                          common.Hash                                  `json:"BestBlockHash"`         // The hash of the block.
	PreviousBestBlockHash                  common.Hash                                  `json:"PreviousBestBlockHash"` // The hash of the block.
	BestShardHash                          map[byte]common.Hash                         `json:"BestShardHash"`
	BestShardHeight                        map[byte]uint64                              `json:"BestShardHeight"`
	Epoch                                  uint64                                       `json:"Epoch"`
	BeaconHeight                           uint64                                       `json:"BeaconHeight"`
	BeaconProposerIndex                    int                                          `json:"BeaconProposerIndex"`
	BeaconCommittee                        []incognitokey.CommitteeKeyString            `json:"BeaconCommittee"`
	BeaconPendingValidator                 []incognitokey.CommitteeKeyString            `json:"BeaconPendingValidator"`
	BeaconWaiting                          []incognitokey.CommitteeKeyString            `json:"BeaconWaiting"`
	BeaconLocking                          []incognitokey.CommitteeKeyString            `json:"BeaconLocking"`
	CandidateShardWaitingForCurrentRandom  []incognitokey.CommitteeKeyString            `json:"CandidateShardWaitingForCurrentRandom"` // snapshot shard candidate list, waiting to be shuffled in this current epoch
	CandidateBeaconWaitingForCurrentRandom []incognitokey.CommitteeKeyString            `json:"CandidateBeaconWaitingForCurrentRandom"`
	CandidateShardWaitingForNextRandom     []incognitokey.CommitteeKeyString            `json:"CandidateShardWaitingForNextRandom"` // shard candidate list, waiting to be shuffled in next epoch
	CandidateBeaconWaitingForNextRandom    []incognitokey.CommitteeKeyString            `json:"CandidateBeaconWaitingForNextRandom"`
	RewardReceiver                         map[string]string                            `json:"RewardReceiver"`        // key: incognito public key of committee, value: payment address reward receiver
	ShardCommittee                         map[byte][]incognitokey.CommitteeKeyString   `json:"ShardCommittee"`        // current committee and validator of all shard
	ShardPendingValidator                  map[byte][]incognitokey.CommitteeKeyString   `json:"ShardPendingValidator"` // pending candidate waiting for swap to get in committee of all shard
	SyncingValidators                      map[byte][]incognitokey.CommitteeKeyString   `json:"SyncingValidator"`      // current syncing validators of all shard
	AutoStaking                            []committeeKeySetAutoStake                   `json:"AutoStaking"`
	StakingTx                              map[string]common.Hash                       `json:"StakingTx"`
	CurrentRandomNumber                    int64                                        `json:"CurrentRandomNumber"`
	CurrentRandomTimeStamp                 int64                                        `json:"CurrentRandomTimeStamp"` // random timestamp for this epoch
	IsGetRandomNumber                      bool                                         `json:"IsGetRandomNumber"`
	MaxBeaconCommitteeSize                 int                                          `json:"MaxBeaconCommitteeSize"`
	MinBeaconCommitteeSize                 int                                          `json:"MinBeaconCommitteeSize"`
	MaxShardCommitteeSize                  int                                          `json:"MaxShardCommitteeSize"`
	MinShardCommitteeSize                  int                                          `json:"MinShardCommitteeSize"`
	ActiveShards                           int                                          `json:"ActiveShards"`
	LastCrossShardState                    map[byte]map[byte]uint64                     `json:"LastCrossShardState"`
	ShardHandle                            map[byte]bool                                `json:"ShardHandle"`             // lock sync.RWMutex
	NumberOfMissingSignature               map[string]signaturecounter.MissingSignature `json:"MissingSignature"`        // lock sync.RWMutex
	MissingSignaturePenalty                map[string]signaturecounter.Penalty          `json:"MissingSignaturePenalty"` // lock sync.RWMutex
}

func NewGetBeaconBestStateDetail(data *blockchain.BeaconBestState) *GetBeaconBestStateDetail {
	result := &GetBeaconBestStateDetail{
		BestBlockHash:          data.BestBlockHash,
		PreviousBestBlockHash:  data.PreviousBestBlockHash,
		Epoch:                  data.Epoch,
		BeaconHeight:           data.BeaconHeight,
		BeaconProposerIndex:    data.BeaconProposerIndex,
		CurrentRandomNumber:    data.CurrentRandomNumber,
		CurrentRandomTimeStamp: data.CurrentRandomTimeStamp,
		IsGetRandomNumber:      data.IsGetRandomNumber,
		MaxShardCommitteeSize:  data.MaxShardCommitteeSize,
		MinShardCommitteeSize:  data.MinShardCommitteeSize,
		MaxBeaconCommitteeSize: data.MaxBeaconCommitteeSize,
		MinBeaconCommitteeSize: data.MinBeaconCommitteeSize,
		ActiveShards:           data.ActiveShards,
	}
	result.BestShardHash = make(map[byte]common.Hash)
	for k, v := range data.BestShardHash {
		result.BestShardHash[k] = v
	}

	result.BestShardHeight = make(map[byte]uint64)
	for k, v := range data.BestShardHeight {
		result.BestShardHeight[k] = v
	}

	tempBeaconCommittee := incognitokey.CommitteeKeyListToStringList(data.GetBeaconCommittee())

	result.BeaconCommittee = make([]incognitokey.CommitteeKeyString, len(data.GetBeaconCommittee()))
	copy(result.BeaconCommittee, tempBeaconCommittee)

	tempBeaconPendingValidator := incognitokey.CommitteeKeyListToStringList(data.GetBeaconPendingValidator())

	result.BeaconPendingValidator = make([]incognitokey.CommitteeKeyString, len(data.GetBeaconPendingValidator()))
	copy(result.BeaconPendingValidator, tempBeaconPendingValidator)

	tempBeaconLocking := incognitokey.CommitteeKeyListToStringList(data.GetBeaconLocking())
	result.BeaconLocking = make([]incognitokey.CommitteeKeyString, len(data.GetBeaconLocking()))
	copy(result.BeaconLocking, tempBeaconLocking)

	tempBeaconWaiting := incognitokey.CommitteeKeyListToStringList(data.GetBeaconWaiting())
	result.BeaconWaiting = make([]incognitokey.CommitteeKeyString, len(data.GetBeaconWaiting()))
	copy(result.BeaconWaiting, tempBeaconWaiting)

	tempCandidateShardWaitingForCurrentRandom := incognitokey.CommitteeKeyListToStringList(data.GetCandidateShardWaitingForCurrentRandom())

	result.CandidateShardWaitingForCurrentRandom = make([]incognitokey.CommitteeKeyString, len(data.GetCandidateShardWaitingForCurrentRandom()))
	copy(result.CandidateShardWaitingForCurrentRandom, tempCandidateShardWaitingForCurrentRandom)

	tempCandidateBeaconWaitingForCurrentRandom := incognitokey.CommitteeKeyListToStringList(data.GetCandidateBeaconWaitingForCurrentRandom())

	result.CandidateBeaconWaitingForCurrentRandom = make([]incognitokey.CommitteeKeyString, len(data.GetCandidateBeaconWaitingForCurrentRandom()))
	copy(result.CandidateBeaconWaitingForCurrentRandom, tempCandidateBeaconWaitingForCurrentRandom)

	tempCandidateShardWaitingForNextRandom := incognitokey.CommitteeKeyListToStringList(data.GetCandidateShardWaitingForNextRandom())

	result.CandidateShardWaitingForNextRandom = make([]incognitokey.CommitteeKeyString, len(data.GetCandidateShardWaitingForNextRandom()))
	copy(result.CandidateShardWaitingForNextRandom, tempCandidateShardWaitingForNextRandom)

	tempCandidateBeaconWaitingForNextRandom := incognitokey.CommitteeKeyListToStringList(data.GetCandidateBeaconWaitingForNextRandom())

	result.CandidateBeaconWaitingForNextRandom = make([]incognitokey.CommitteeKeyString, len(data.GetCandidateBeaconWaitingForNextRandom()))
	copy(result.CandidateBeaconWaitingForNextRandom, tempCandidateBeaconWaitingForNextRandom)

	result.RewardReceiver = make(map[string]string)
	for k, v := range data.GetRewardReceiver() {
		tempV := base58.Base58Check{}.Encode(v.Bytes(), common.Base58Version)
		result.RewardReceiver[k] = tempV
	}

	result.ShardCommittee = make(map[byte][]incognitokey.CommitteeKeyString)
	for k, v := range data.GetShardCommittee() {
		result.ShardCommittee[k] = make([]incognitokey.CommitteeKeyString, len(v))
		tempV := incognitokey.CommitteeKeyListToStringList(v)
		copy(result.ShardCommittee[k], tempV)
	}

	result.ShardPendingValidator = make(map[byte][]incognitokey.CommitteeKeyString)
	for k, v := range data.GetShardPendingValidator() {
		result.ShardPendingValidator[k] = make([]incognitokey.CommitteeKeyString, len(v))
		tempV := incognitokey.CommitteeKeyListToStringList(v)
		copy(result.ShardPendingValidator[k], tempV)
	}

	result.LastCrossShardState = make(map[byte]map[byte]uint64)
	for k1, v1 := range data.LastCrossShardState {
		result.LastCrossShardState[k1] = make(map[byte]uint64)
		for k2, v2 := range v1 {
			result.LastCrossShardState[k1][k2] = v2
		}
	}

	for k, v := range data.GetAutoStaking() {
		var keySet incognitokey.CommitteePublicKey
		keySet.FromString(k)

		var keyMap committeeKeySetAutoStake
		keyMap.IncPubKey = keySet.GetIncKeyBase58()
		keyMap.MiningPubKey = make(map[string]string)
		for keyType := range keySet.MiningPubKey {
			keyMap.MiningPubKey[keyType] = keySet.GetMiningKeyBase58(keyType)
		}
		keyMap.IsAutoStake = v
		result.AutoStaking = append(result.AutoStaking, keyMap)

	}

	result.StakingTx = make(map[string]common.Hash)
	for k, v := range data.GetStakingTx() {
		committeePublicKey := incognitokey.CommitteePublicKey{}
		committeePublicKey.FromString(k)
		result.StakingTx[committeePublicKey.GetIncKeyBase58()] = v
	}

	result.NumberOfMissingSignature = make(map[string]signaturecounter.MissingSignature)
	for k, v := range data.GetNumberOfMissingSignature() {
		res, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{k})
		incKey := res[0].GetIncKeyBase58()
		result.NumberOfMissingSignature[incKey] = v
	}

	result.MissingSignaturePenalty = make(map[string]signaturecounter.Penalty)
	for k, v := range data.GetMissingSignaturePenalty() {
		res, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{k})
		incKey := res[0].GetIncKeyBase58()
		result.MissingSignaturePenalty[incKey] = v
	}

	result.SyncingValidators = make(map[byte][]incognitokey.CommitteeKeyString)
	for k, v := range data.GetSyncingValidators() {
		result.SyncingValidators[k] = make([]incognitokey.CommitteeKeyString, len(v))
		tempV := incognitokey.CommitteeKeyListToStringList(v)
		copy(result.SyncingValidators[k], tempV)
	}

	return result
}

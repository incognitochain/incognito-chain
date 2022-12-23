package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/syncker/finishsync"
	"github.com/incognitochain/incognito-chain/wallet"
)

type GetBeaconBestState struct {
	BlockVersion                           int                  `json:"BlockVersion"`
	BestBlockHash                          common.Hash          `json:"BestBlockHash"`         // The hash of the block.
	PreviousBestBlockHash                  common.Hash          `json:"PreviousBestBlockHash"` // The hash of the block.
	BestShardHash                          map[byte]common.Hash `json:"BestShardHash"`
	BestShardHeight                        map[byte]uint64      `json:"BestShardHeight"`
	Epoch                                  uint64               `json:"Epoch"`
	BeaconHeight                           uint64               `json:"BeaconHeight"`
	BeaconProposerIndex                    int                  `json:"BeaconProposerIndex"`
	BeaconCommittee                        []string             `json:"BeaconCommittee"`
	BeaconPendingValidator                 []string             `json:"BeaconPendingValidator"`
	CandidateShardWaitingForCurrentRandom  []string             `json:"CandidateShardWaitingForCurrentRandom"` // snapshot shard candidate list, waiting to be shuffled in this current epoch
	CandidateBeaconWaitingForCurrentRandom []string             `json:"CandidateBeaconWaitingForCurrentRandom"`
	CandidateBeaconWaitingForNextRandom    []string             `json:"CandidateBeaconWaitingForNextRandom"`
	BeaconWaiting                          []string             `json:"BeaconWaiting"`
	BeaconLocking                          []string             `json:"BeaconLocking"`
	CandidateShardWaitingForNextRandom     []string             `json:"CandidateShardWaitingForNextRandom"` // shard candidate list, waiting to be shuffled in next epoch

	RewardReceiver           map[string]string                            `json:"RewardReceiver"`        // key: incognito public key of committee, value: payment address reward receiver
	ShardCommittee           map[byte][]string                            `json:"ShardCommittee"`        // current committee and validator of all shard
	ShardPendingValidator    map[byte][]string                            `json:"ShardPendingValidator"` // pending candidate waiting for swap to get in committee of all shard
	SyncingValidators        map[byte][]string                            `json:"SyncingValidator"`      // current syncing validators of all shard
	AutoStaking              map[string]bool                              `json:"AutoStaking"`
	StakingTx                map[string]common.Hash                       `json:"StakingTx"`
	CurrentRandomNumber      int64                                        `json:"CurrentRandomNumber"`
	CurrentRandomTimeStamp   int64                                        `json:"CurrentRandomTimeStamp"` // random timestamp for this epoch
	IsGetRandomNumber        bool                                         `json:"IsGetRandomNumber"`
	MaxBeaconCommitteeSize   int                                          `json:"MaxBeaconCommitteeSize"`
	MinBeaconCommitteeSize   int                                          `json:"MinBeaconCommitteeSize"`
	MaxShardCommitteeSize    int                                          `json:"MaxShardCommitteeSize"`
	MinShardCommitteeSize    int                                          `json:"MinShardCommitteeSize"`
	ActiveShards             int                                          `json:"ActiveShards"`
	LastCrossShardState      map[byte]map[byte]uint64                     `json:"LastCrossShardState"`
	ShardHandle              map[byte]bool                                `json:"ShardHandle"` // lock sync.RWMutex
	CommitteeEngineVersion   int                                          `json:"CommitteeStateVersion"`
	NumberOfShardBlock       map[byte]uint                                `json:"NumberOfShardBlock"` // lock sync.RWMutex
	FinishSyncManager        map[byte][]string                            `json:"FinishSyncManager"`
	NumberOfMissingSignature map[string]signaturecounter.MissingSignature `json:"MissingSignature"`        // lock sync.RWMutex
	MissingSignaturePenalty  map[string]signaturecounter.Penalty          `json:"MissingSignaturePenalty"` // lock sync.RWMutex
	Config                   map[string]interface{}                       `json:"Config"`
	TriggeredFeature         map[string]uint64                            `json:"TriggeredFeature"`
	RewardMinted             uint64                                       `json:"RewardMinted"`
}

func NewGetBeaconBestState(data *blockchain.BeaconBestState) *GetBeaconBestState {
	result := &GetBeaconBestState{
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
		NumberOfShardBlock:     data.NumberOfShardBlock,
		BlockVersion:           data.BestBlock.GetVersion(),
		RewardMinted:           data.RewardMinted,
	}

	result.TriggeredFeature = make(map[string]uint64)
	for k, v := range data.TriggeredFeature {
		result.TriggeredFeature[k] = v
	}

	result.BestShardHash = make(map[byte]common.Hash)
	for k, v := range data.BestShardHash {
		result.BestShardHash[k] = v
	}

	result.BestShardHeight = make(map[byte]uint64)
	for k, v := range data.BestShardHeight {
		result.BestShardHeight[k] = v
	}
	tempBeaconCommittee, err := incognitokey.CommitteeKeyListToString(data.GetBeaconCommittee())
	if err != nil {
		panic(err)
	}
	result.BeaconCommittee = make([]string, len(data.GetBeaconCommittee()))
	copy(result.BeaconCommittee, tempBeaconCommittee)

	tempBeaconPendingValidator, err := incognitokey.CommitteeKeyListToString(data.GetBeaconPendingValidator())
	if err != nil {
		panic(err)
	}
	result.BeaconPendingValidator = make([]string, len(data.GetBeaconPendingValidator()))
	copy(result.BeaconPendingValidator, tempBeaconPendingValidator)

	tempCandidateShardWaitingForCurrentRandom, err := incognitokey.CommitteeKeyListToString(data.GetCandidateShardWaitingForCurrentRandom())
	if err != nil {
		panic(err)
	}
	result.CandidateShardWaitingForCurrentRandom = make([]string, len(data.GetCandidateShardWaitingForCurrentRandom()))
	copy(result.CandidateShardWaitingForCurrentRandom, tempCandidateShardWaitingForCurrentRandom)

	tempBeaconWaiting, err := incognitokey.CommitteeKeyListToString(data.GetBeaconWaiting())
	if err != nil {
		panic(err)
	}
	result.BeaconWaiting = make([]string, len(tempBeaconWaiting))
	copy(result.BeaconWaiting, tempBeaconWaiting)

	tempCandidateShardWaitingForNextRandom, err := incognitokey.CommitteeKeyListToString(data.GetCandidateShardWaitingForNextRandom())
	if err != nil {
		panic(err)
	}
	result.CandidateShardWaitingForNextRandom = make([]string, len(data.GetCandidateShardWaitingForNextRandom()))
	copy(result.CandidateShardWaitingForNextRandom, tempCandidateShardWaitingForNextRandom)

	tempBeaconLocking, err := incognitokey.CommitteeKeyListToString(data.GetBeaconLocking())
	if err != nil {
		panic(err)
	}
	result.BeaconLocking = make([]string, len(tempBeaconLocking))
	copy(result.BeaconLocking, tempBeaconLocking)

	result.RewardReceiver = make(map[string]string)
	for k, v := range data.GetRewardReceiver() {
		wl := wallet.KeyWallet{}
		wl.KeySet.PaymentAddress = v
		paymentAddress := wl.Base58CheckSerialize(wallet.PaymentAddressType)
		result.RewardReceiver[k] = paymentAddress
	}

	result.ShardCommittee = make(map[byte][]string)
	for k, v := range data.GetShardCommittee() {
		result.ShardCommittee[k] = make([]string, len(v))
		tempV, err := incognitokey.CommitteeKeyListToString(v)
		if err != nil {
			panic(err)
		}
		copy(result.ShardCommittee[k], tempV)
	}

	result.ShardPendingValidator = make(map[byte][]string)
	for k, v := range data.GetShardPendingValidator() {
		result.ShardPendingValidator[k] = make([]string, len(v))
		tempV, err := incognitokey.CommitteeKeyListToString(v)
		if err != nil {
			panic(err)
		}
		copy(result.ShardPendingValidator[k], tempV)
	}

	result.LastCrossShardState = make(map[byte]map[byte]uint64)
	for k1, v1 := range data.LastCrossShardState {
		result.LastCrossShardState[k1] = make(map[byte]uint64)
		for k2, v2 := range v1 {
			result.LastCrossShardState[k1][k2] = v2
		}
	}

	result.AutoStaking = make(map[string]bool)
	for k, v := range data.GetAutoStaking() {
		result.AutoStaking[k] = v
	}

	result.StakingTx = make(map[string]common.Hash)
	for k, v := range data.GetStakingTx() {
		result.StakingTx[k] = v
	}

	result.NumberOfMissingSignature = make(map[string]signaturecounter.MissingSignature)
	for k, v := range data.GetNumberOfMissingSignature() {
		result.NumberOfMissingSignature[k] = v
	}
	result.MissingSignaturePenalty = make(map[string]signaturecounter.Penalty)
	for k, v := range data.GetMissingSignaturePenalty() {
		result.MissingSignaturePenalty[k] = v
	}
	result.CommitteeEngineVersion = data.CommitteeStateVersion()
	result.SyncingValidators = make(map[byte][]string)
	for k, v := range data.GetSyncingValidators() {
		result.SyncingValidators[k] = make([]string, len(v))
		tempV, err := incognitokey.CommitteeKeyListToString(v)
		if err != nil {
			panic(err)
		}
		copy(result.SyncingValidators[k], tempV)
	}

	result.FinishSyncManager = finishsync.DefaultFinishSyncMsgPool.GetFinishedSyncValidators()
	result.Config = make(map[string]interface{})
	result.Config["EnableSlashingHeightV1"] = config.Param().ConsensusParam.EnableSlashingHeight
	result.Config["InitShardCommitteeSize"] = config.Param().CommitteeSize.InitShardCommitteeSize
	result.Config["NumberOfBlockInEpoch"] = config.Param().EpochParam.NumberOfBlockInEpoch
	result.Config["RandomTime"] = config.Param().EpochParam.RandomTime
	result.Config["EnableSlashingHeightV2"] = config.Param().ConsensusParam.EnableSlashingHeightV2
	result.Config["StakingFlowV3"] = config.Param().ConsensusParam.StakingFlowV3Height
	result.Config["StakingFlowV2"] = config.Param().ConsensusParam.StakingFlowV2Height
	result.Config["BlockProducingV3"] = config.Param().ConsensusParam.BlockProducingV3Height
	result.Config["BlockProducingV3Height"] = config.Param().ConsensusParam.BlockProducingV3Height
	return result
}

type GetTotalStaker struct {
	TotalStaker int `json:"TotalStaker"`
}

func NewGetTotalStaker(total int) *GetTotalStaker {
	return &GetTotalStaker{
		TotalStaker: total,
	}
}

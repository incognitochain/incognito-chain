package blockchain

import (
	"testing"
)

// func TestShardBestState_restoreCommittee(t *testing.T) {
// 	type fields struct {
// 		BestBlockHash              common.Hash
// 		BestBlock                  *ShardBlock
// 		BestBeaconHash             common.Hash
// 		BeaconHeight               uint64
// 		ShardID                    byte
// 		Epoch                      uint64
// 		ShardHeight                uint64
// 		MaxShardCommitteeSize      int
// 		MinShardCommitteeSize      int
// 		ShardProposerIdx           int
// 		ShardCommittee             []incognitokey.CommitteePublicKey
// 		ShardPendingValidator      []incognitokey.CommitteePublicKey
// 		BestCrossShard             map[byte]uint64
// 		StakingTx                  map[string]string
// 		NumTxns                    uint64
// 		TotalTxns                  uint64
// 		TotalTxnsExcludeSalary     uint64
// 		ActiveShards               int
// 		ConsensusAlgorithm         string
// 		NumOfBlocksByProducers     map[string]uint64
// 		BlockInterval              time.Duration
// 		BlockMaxCreateTime         time.Duration
// 		MetricBlockHeight          uint64
// 		consensusStateDB           *statedb.StateDB
// 		ConsensusStateDBRootHash   common.Hash
// 		transactionStateDB         *statedb.StateDB
// 		TransactionStateDBRootHash common.Hash
// 		featureStateDB             *statedb.StateDB
// 		FeatureStateDBRootHash     common.Hash
// 		rewardStateDB              *statedb.StateDB
// 		RewardStateDBRootHash      common.Hash
// 		slashStateDB               *statedb.StateDB
// 		SlashStateDBRootHash       common.Hash
// 		lock                       sync.RWMutex
// 	}
// 	type args struct {
// 		shardID byte
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			shardBestState := &ShardBestState{
// 				BestBlockHash:              tt.fields.BestBlockHash,
// 				BestBlock:                  tt.fields.BestBlock,
// 				BestBeaconHash:             tt.fields.BestBeaconHash,
// 				BeaconHeight:               tt.fields.BeaconHeight,
// 				ShardID:                    tt.fields.ShardID,
// 				Epoch:                      tt.fields.Epoch,
// 				ShardHeight:                tt.fields.ShardHeight,
// 				MaxShardCommitteeSize:      tt.fields.MaxShardCommitteeSize,
// 				MinShardCommitteeSize:      tt.fields.MinShardCommitteeSize,
// 				ShardProposerIdx:           tt.fields.ShardProposerIdx,
// 				ShardCommittee:             tt.fields.ShardCommittee,
// 				ShardPendingValidator:      tt.fields.ShardPendingValidator,
// 				BestCrossShard:             tt.fields.BestCrossShard,
// 				StakingTx:                  tt.fields.StakingTx,
// 				NumTxns:                    tt.fields.NumTxns,
// 				TotalTxns:                  tt.fields.TotalTxns,
// 				TotalTxnsExcludeSalary:     tt.fields.TotalTxnsExcludeSalary,
// 				ActiveShards:               tt.fields.ActiveShards,
// 				ConsensusAlgorithm:         tt.fields.ConsensusAlgorithm,
// 				NumOfBlocksByProducers:     tt.fields.NumOfBlocksByProducers,
// 				BlockInterval:              tt.fields.BlockInterval,
// 				BlockMaxCreateTime:         tt.fields.BlockMaxCreateTime,
// 				MetricBlockHeight:          tt.fields.MetricBlockHeight,
// 				consensusStateDB:           tt.fields.consensusStateDB,
// 				ConsensusStateDBRootHash:   tt.fields.ConsensusStateDBRootHash,
// 				transactionStateDB:         tt.fields.transactionStateDB,
// 				TransactionStateDBRootHash: tt.fields.TransactionStateDBRootHash,
// 				featureStateDB:             tt.fields.featureStateDB,
// 				FeatureStateDBRootHash:     tt.fields.FeatureStateDBRootHash,
// 				rewardStateDB:              tt.fields.rewardStateDB,
// 				RewardStateDBRootHash:      tt.fields.RewardStateDBRootHash,
// 				slashStateDB:               tt.fields.slashStateDB,
// 				SlashStateDBRootHash:       tt.fields.SlashStateDBRootHash,
// 				lock:                       tt.fields.lock,
// 			}
// 			if err := shardBestState.restoreCommittee(tt.args.shardID); (err != nil) != tt.wantErr {
// 				t.Errorf("ShardBestState.restoreCommittee() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestBeaconBestState_restoreShardCommittee(t *testing.T) {
// 	type fields struct {
// 		BestBlockHash                          common.Hash
// 		PreviousBestBlockHash                  common.Hash
// 		BestBlock                              BeaconBlock
// 		BestShardHash                          map[byte]common.Hash
// 		BestShardHeight                        map[byte]uint64
// 		Epoch                                  uint64
// 		BeaconHeight                           uint64
// 		BeaconProposerIndex                    int
// 		BeaconCommittee                        []incognitokey.CommitteePublicKey
// 		BeaconPendingValidator                 []incognitokey.CommitteePublicKey
// 		CandidateShardWaitingForCurrentRandom  []incognitokey.CommitteePublicKey
// 		CandidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePublicKey
// 		CandidateShardWaitingForNextRandom     []incognitokey.CommitteePublicKey
// 		CandidateBeaconWaitingForNextRandom    []incognitokey.CommitteePublicKey
// 		ShardCommittee                         map[byte][]incognitokey.CommitteePublicKey
// 		ShardPendingValidator                  map[byte][]incognitokey.CommitteePublicKey
// 		AutoStaking                            map[string]bool
// 		CurrentRandomNumber                    int64
// 		CurrentRandomTimeStamp                 int64
// 		IsGetRandomNumber                      bool
// 		Params                                 map[string]string
// 		MaxBeaconCommitteeSize                 int
// 		MinBeaconCommitteeSize                 int
// 		MaxShardCommitteeSize                  int
// 		MinShardCommitteeSize                  int
// 		ActiveShards                           int
// 		ConsensusAlgorithm                     string
// 		ShardConsensusAlgorithm                map[byte]string
// 		RewardReceiver                         map[string]string
// 		LastCrossShardState                    map[byte]map[byte]uint64
// 		ShardHandle                            map[byte]bool
// 		NumOfBlocksByProducers                 map[string]uint64
// 		BlockInterval                          time.Duration
// 		BlockMaxCreateTime                     time.Duration
// 		consensusStateDB                       *statedb.StateDB
// 		ConsensusStateDBRootHash               common.Hash
// 		rewardStateDB                          *statedb.StateDB
// 		RewardStateDBRootHash                  common.Hash
// 		featureStateDB                         *statedb.StateDB
// 		FeatureStateDBRootHash                 common.Hash
// 		slashStateDB                           *statedb.StateDB
// 		SlashStateDBRootHash                   common.Hash
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			beaconBestState := &BeaconBestState{
// 				BestBlockHash:                          tt.fields.BestBlockHash,
// 				PreviousBestBlockHash:                  tt.fields.PreviousBestBlockHash,
// 				BestBlock:                              tt.fields.BestBlock,
// 				BestShardHash:                          tt.fields.BestShardHash,
// 				BestShardHeight:                        tt.fields.BestShardHeight,
// 				Epoch:                                  tt.fields.Epoch,
// 				BeaconHeight:                           tt.fields.BeaconHeight,
// 				BeaconProposerIndex:                    tt.fields.BeaconProposerIndex,
// 				BeaconCommittee:                        tt.fields.BeaconCommittee,
// 				BeaconPendingValidator:                 tt.fields.BeaconPendingValidator,
// 				CandidateShardWaitingForCurrentRandom:  tt.fields.CandidateShardWaitingForCurrentRandom,
// 				CandidateBeaconWaitingForCurrentRandom: tt.fields.CandidateBeaconWaitingForCurrentRandom,
// 				CandidateShardWaitingForNextRandom:     tt.fields.CandidateShardWaitingForNextRandom,
// 				CandidateBeaconWaitingForNextRandom:    tt.fields.CandidateBeaconWaitingForNextRandom,
// 				ShardCommittee:                         tt.fields.ShardCommittee,
// 				ShardPendingValidator:                  tt.fields.ShardPendingValidator,
// 				AutoStaking:                            tt.fields.AutoStaking,
// 				CurrentRandomNumber:                    tt.fields.CurrentRandomNumber,
// 				CurrentRandomTimeStamp:                 tt.fields.CurrentRandomTimeStamp,
// 				IsGetRandomNumber:                      tt.fields.IsGetRandomNumber,
// 				Params:                                 tt.fields.Params,
// 				MaxBeaconCommitteeSize:                 tt.fields.MaxBeaconCommitteeSize,
// 				MinBeaconCommitteeSize:                 tt.fields.MinBeaconCommitteeSize,
// 				MaxShardCommitteeSize:                  tt.fields.MaxShardCommitteeSize,
// 				MinShardCommitteeSize:                  tt.fields.MinShardCommitteeSize,
// 				ActiveShards:                           tt.fields.ActiveShards,
// 				ConsensusAlgorithm:                     tt.fields.ConsensusAlgorithm,
// 				ShardConsensusAlgorithm:                tt.fields.ShardConsensusAlgorithm,
// 				RewardReceiver:                         tt.fields.RewardReceiver,
// 				LastCrossShardState:                    tt.fields.LastCrossShardState,
// 				ShardHandle:                            tt.fields.ShardHandle,
// 				NumOfBlocksByProducers:                 tt.fields.NumOfBlocksByProducers,
// 				BlockInterval:                          tt.fields.BlockInterval,
// 				BlockMaxCreateTime:                     tt.fields.BlockMaxCreateTime,
// 				consensusStateDB:                       tt.fields.consensusStateDB,
// 				ConsensusStateDBRootHash:               tt.fields.ConsensusStateDBRootHash,
// 				rewardStateDB:                          tt.fields.rewardStateDB,
// 				RewardStateDBRootHash:                  tt.fields.RewardStateDBRootHash,
// 				featureStateDB:                         tt.fields.featureStateDB,
// 				FeatureStateDBRootHash:                 tt.fields.FeatureStateDBRootHash,
// 				slashStateDB:                           tt.fields.slashStateDB,
// 				SlashStateDBRootHash:                   tt.fields.SlashStateDBRootHash,
// 			}
// 			if err := beaconBestState.restoreShardCommittee(); (err != nil) != tt.wantErr {
// 				t.Errorf("BeaconBestState.restoreShardCommittee() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestBeaconBestState_restoreBeaconCommittee(t *testing.T) {
// 	type fields struct {
// 		BeaconCommittee []incognitokey.CommitteePublicKey
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 		{},
// 		{},
// 		{},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			beaconBestState := &BeaconBestState{
// 				BeaconCommittee: tt.fields.BeaconCommittee,
// 			}
// 			if err := beaconBestState.restoreBeaconCommittee(); (err != nil) != tt.wantErr {
// 				t.Errorf("BeaconBestState.restoreBeaconCommittee() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestBeaconBestState_storeBeaconPreCommitteeHash(t *testing.T) {
	// type fields struct {
	// 	BestBlockHash                          common.Hash
	// 	PreviousBestBlockHash                  common.Hash
	// 	BestBlock                              BeaconBlock
	// 	BestShardHash                          map[byte]common.Hash
	// 	BestShardHeight                        map[byte]uint64
	// 	Epoch                                  uint64
	// 	BeaconHeight                           uint64
	// 	BeaconProposerIndex                    int
	// 	BeaconCommittee                        []incognitokey.CommitteePublicKey
	// 	BeaconPendingValidator                 []incognitokey.CommitteePublicKey
	// 	CandidateShardWaitingForCurrentRandom  []incognitokey.CommitteePublicKey
	// 	CandidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePublicKey
	// 	CandidateShardWaitingForNextRandom     []incognitokey.CommitteePublicKey
	// 	CandidateBeaconWaitingForNextRandom    []incognitokey.CommitteePublicKey
	// 	ShardCommittee                         map[byte][]incognitokey.CommitteePublicKey
	// 	ShardPendingValidator                  map[byte][]incognitokey.CommitteePublicKey
	// 	AutoStaking                            map[string]bool
	// 	CurrentRandomNumber                    int64
	// 	CurrentRandomTimeStamp                 int64
	// 	IsGetRandomNumber                      bool
	// 	Params                                 map[string]string
	// 	MaxBeaconCommitteeSize                 int
	// 	MinBeaconCommitteeSize                 int
	// 	MaxShardCommitteeSize                  int
	// 	MinShardCommitteeSize                  int
	// 	ActiveShards                           int
	// 	ConsensusAlgorithm                     string
	// 	ShardConsensusAlgorithm                map[byte]string
	// 	RewardReceiver                         map[string]string
	// 	LastCrossShardState                    map[byte]map[byte]uint64
	// 	ShardHandle                            map[byte]bool
	// 	NumOfBlocksByProducers                 map[string]uint64
	// 	BlockInterval                          time.Duration
	// 	BlockMaxCreateTime                     time.Duration
	// 	consensusStateDB                       *statedb.StateDB
	// 	ConsensusStateDBRootHash               common.Hash
	// 	rewardStateDB                          *statedb.StateDB
	// 	RewardStateDBRootHash                  common.Hash
	// 	featureStateDB                         *statedb.StateDB
	// 	FeatureStateDBRootHash                 common.Hash
	// 	slashStateDB                           *statedb.StateDB
	// 	SlashStateDBRootHash                   common.Hash
	// 	BeaconPreCommitteeHash                 common.Hash
	// 	ShardPreCommitteeHash                  common.Hash
	// }
	// type args struct {
	// 	db incdb.KeyValueWriter
	// }
	// tests := []struct {
	// 	name    string
	// 	fields  fields
	// 	args    args
	// 	wantErr bool
	// }{
	// 	// TODO: Add test cases.
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		beaconBestState := &BeaconBestState{
	// 			BestBlockHash:                          tt.fields.BestBlockHash,
	// 			PreviousBestBlockHash:                  tt.fields.PreviousBestBlockHash,
	// 			BestBlock:                              tt.fields.BestBlock,
	// 			BestShardHash:                          tt.fields.BestShardHash,
	// 			BestShardHeight:                        tt.fields.BestShardHeight,
	// 			Epoch:                                  tt.fields.Epoch,
	// 			BeaconHeight:                           tt.fields.BeaconHeight,
	// 			BeaconProposerIndex:                    tt.fields.BeaconProposerIndex,
	// 			BeaconCommittee:                        tt.fields.BeaconCommittee,
	// 			BeaconPendingValidator:                 tt.fields.BeaconPendingValidator,
	// 			CandidateShardWaitingForCurrentRandom:  tt.fields.CandidateShardWaitingForCurrentRandom,
	// 			CandidateBeaconWaitingForCurrentRandom: tt.fields.CandidateBeaconWaitingForCurrentRandom,
	// 			CandidateShardWaitingForNextRandom:     tt.fields.CandidateShardWaitingForNextRandom,
	// 			CandidateBeaconWaitingForNextRandom:    tt.fields.CandidateBeaconWaitingForNextRandom,
	// 			ShardCommittee:                         tt.fields.ShardCommittee,
	// 			ShardPendingValidator:                  tt.fields.ShardPendingValidator,
	// 			AutoStaking:                            tt.fields.AutoStaking,
	// 			CurrentRandomNumber:                    tt.fields.CurrentRandomNumber,
	// 			CurrentRandomTimeStamp:                 tt.fields.CurrentRandomTimeStamp,
	// 			IsGetRandomNumber:                      tt.fields.IsGetRandomNumber,
	// 			Params:                                 tt.fields.Params,
	// 			MaxBeaconCommitteeSize:                 tt.fields.MaxBeaconCommitteeSize,
	// 			MinBeaconCommitteeSize:                 tt.fields.MinBeaconCommitteeSize,
	// 			MaxShardCommitteeSize:                  tt.fields.MaxShardCommitteeSize,
	// 			MinShardCommitteeSize:                  tt.fields.MinShardCommitteeSize,
	// 			ActiveShards:                           tt.fields.ActiveShards,
	// 			ConsensusAlgorithm:                     tt.fields.ConsensusAlgorithm,
	// 			ShardConsensusAlgorithm:                tt.fields.ShardConsensusAlgorithm,
	// 			RewardReceiver:                         tt.fields.RewardReceiver,
	// 			LastCrossShardState:                    tt.fields.LastCrossShardState,
	// 			ShardHandle:                            tt.fields.ShardHandle,
	// 			NumOfBlocksByProducers:                 tt.fields.NumOfBlocksByProducers,
	// 			BlockInterval:                          tt.fields.BlockInterval,
	// 			BlockMaxCreateTime:                     tt.fields.BlockMaxCreateTime,
	// 			consensusStateDB:                       tt.fields.consensusStateDB,
	// 			ConsensusStateDBRootHash:               tt.fields.ConsensusStateDBRootHash,
	// 			rewardStateDB:                          tt.fields.rewardStateDB,
	// 			RewardStateDBRootHash:                  tt.fields.RewardStateDBRootHash,
	// 			featureStateDB:                         tt.fields.featureStateDB,
	// 			FeatureStateDBRootHash:                 tt.fields.FeatureStateDBRootHash,
	// 			slashStateDB:                           tt.fields.slashStateDB,
	// 			SlashStateDBRootHash:                   tt.fields.SlashStateDBRootHash,
	// 			BeaconPreCommitteeHash:                 tt.fields.BeaconPreCommitteeHash,
	// 			ShardPreCommitteeHash:                  tt.fields.ShardPreCommitteeHash,
	// 		}
	// 		if err := beaconBestState.storeBeaconPreCommitteeHash(tt.args.db); (err != nil) != tt.wantErr {
	// 			t.Errorf("BeaconBestState.storeBeaconPreCommitteeHash() error = %v, wantErr %v", err, tt.wantErr)
	// 		}
	// 	})
	// }
}

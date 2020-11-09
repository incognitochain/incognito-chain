package blockchain

import (
	"reflect"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func TestBeaconBestState_filterCommitteeInstructions(t *testing.T) {
	type fields struct {
		BestBlockHash            common.Hash
		PreviousBestBlockHash    common.Hash
		BestBlock                types.BeaconBlock
		BestShardHash            map[byte]common.Hash
		BestShardHeight          map[byte]uint64
		Epoch                    uint64
		BeaconHeight             uint64
		BeaconProposerIndex      int
		CurrentRandomNumber      int64
		CurrentRandomTimeStamp   int64
		IsGetRandomNumber        bool
		Params                   map[string]string
		MaxBeaconCommitteeSize   int
		MinBeaconCommitteeSize   int
		MaxShardCommitteeSize    int
		MinShardCommitteeSize    int
		ActiveShards             int
		ConsensusAlgorithm       string
		ShardConsensusAlgorithm  map[byte]string
		beaconCommitteeEngine    committeestate.BeaconCommitteeEngine
		LastCrossShardState      map[byte]map[byte]uint64
		ShardHandle              map[byte]bool
		NumOfBlocksByProducers   map[string]uint64
		BlockInterval            time.Duration
		BlockMaxCreateTime       time.Duration
		consensusStateDB         *statedb.StateDB
		ConsensusStateDBRootHash common.Hash
		rewardStateDB            *statedb.StateDB
		RewardStateDBRootHash    common.Hash
		featureStateDB           *statedb.StateDB
		FeatureStateDBRootHash   common.Hash
		slashStateDB             *statedb.StateDB
		SlashStateDBRootHash     common.Hash
	}
	type args struct {
		instructions [][]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{},
		{},
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beaconBestState := &BeaconBestState{
				BestBlockHash:            tt.fields.BestBlockHash,
				PreviousBestBlockHash:    tt.fields.PreviousBestBlockHash,
				BestBlock:                tt.fields.BestBlock,
				BestShardHash:            tt.fields.BestShardHash,
				BestShardHeight:          tt.fields.BestShardHeight,
				Epoch:                    tt.fields.Epoch,
				BeaconHeight:             tt.fields.BeaconHeight,
				BeaconProposerIndex:      tt.fields.BeaconProposerIndex,
				CurrentRandomNumber:      tt.fields.CurrentRandomNumber,
				CurrentRandomTimeStamp:   tt.fields.CurrentRandomTimeStamp,
				IsGetRandomNumber:        tt.fields.IsGetRandomNumber,
				Params:                   tt.fields.Params,
				MaxBeaconCommitteeSize:   tt.fields.MaxBeaconCommitteeSize,
				MinBeaconCommitteeSize:   tt.fields.MinBeaconCommitteeSize,
				MaxShardCommitteeSize:    tt.fields.MaxShardCommitteeSize,
				MinShardCommitteeSize:    tt.fields.MinShardCommitteeSize,
				ActiveShards:             tt.fields.ActiveShards,
				ConsensusAlgorithm:       tt.fields.ConsensusAlgorithm,
				ShardConsensusAlgorithm:  tt.fields.ShardConsensusAlgorithm,
				beaconCommitteeEngine:    tt.fields.beaconCommitteeEngine,
				LastCrossShardState:      tt.fields.LastCrossShardState,
				ShardHandle:              tt.fields.ShardHandle,
				NumOfBlocksByProducers:   tt.fields.NumOfBlocksByProducers,
				BlockInterval:            tt.fields.BlockInterval,
				BlockMaxCreateTime:       tt.fields.BlockMaxCreateTime,
				consensusStateDB:         tt.fields.consensusStateDB,
				ConsensusStateDBRootHash: tt.fields.ConsensusStateDBRootHash,
				rewardStateDB:            tt.fields.rewardStateDB,
				RewardStateDBRootHash:    tt.fields.RewardStateDBRootHash,
				featureStateDB:           tt.fields.featureStateDB,
				FeatureStateDBRootHash:   tt.fields.FeatureStateDBRootHash,
				slashStateDB:             tt.fields.slashStateDB,
				SlashStateDBRootHash:     tt.fields.SlashStateDBRootHash,
			}
			if got := beaconBestState.filterCommitteeInstructions(tt.args.instructions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconBestState.filterCommitteeInstructions() = %v, want %v", got, tt.want)
			}
		})
	}
}

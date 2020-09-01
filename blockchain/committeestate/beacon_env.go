package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type BeaconCommitteeStateEnvironment struct {
	BeaconHeight                    uint64
	Epoch                           uint64
	BeaconHash                      common.Hash
	ParamEpoch                      uint64
	BeaconInstructions              [][]string
	EpochBreakPointSwapNewKey       []uint64
	RandomNumber                    int64
	IsFoundRandomNumber             bool
	IsBeaconRandomTime              bool
	AssignOffset                    int
	DefaultOffset                   int
	SwapOffset                      int
	ActiveShards                    int
	MinShardCommitteeSize           int
	MinBeaconCommitteeSize          int
	MaxBeaconCommitteeSize          int
	MaxCommitteeSize                int
	ConsensusStateDB                *statedb.StateDB
	IsReplace                       bool
	NumberOfFixedBlockValidator     uint64
	allCandidateSubstituteCommittee []string
}

type BeaconCommitteeStateHash struct {
	BeaconCommitteeAndValidatorHash common.Hash
	BeaconCandidateHash             common.Hash
	ShardCandidateHash              common.Hash
	ShardCommitteeAndValidatorHash  common.Hash
	AutoStakeHash                   common.Hash
}

package delegation

import (
	"github.com/incognitochain/incognito-chain/config"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/pkg/errors"
)

func NewTestError(tag string, err error) error {
	return errors.Wrap(err, tag)
}

const (
	TestShardStakingWithDelegationError   = "TestShardStakingWithDelegation Error"
	TestDelegationAfterStakeError         = "TestDelegationAfterStake Error"
	TestShardStakingWithReDelegationError = "TestShardStakingWithReDelegation Error"
)

func InitDelegationTest() *testsuite.NodeEngine {
	cfg := testsuite.Config{
		DataDir: "./data/",
		Network: testsuite.ID_TESTNET2,
		ResetDB: true,
	}

	node := testsuite.InitChainParam(cfg, func() {
		config.Param().ActiveShards = 2
		config.Param().BCHeightBreakPointNewZKP = 1
		config.Param().BCHeightBreakPointPrivacyV2 = 2
		config.Param().BeaconHeightBreakPointBurnAddr = 1
		config.Param().ConsensusParam.EnableSlashingHeightV2 = 1
		config.Param().ConsensusParam.StakingFlowV2Height = 1
		config.Param().ConsensusParam.AssignRuleV3Height = 1
		config.Param().ConsensusParam.StakingFlowV3Height = 1
		config.Param().ConsensusParam.StakingFlowV4Height = 1
		config.Param().CommitteeSize.MaxShardCommitteeSize = 8
		config.Param().MaxReward = 95000000000000000
		config.Param().BasicReward = 1386666000
		config.Param().CommitteeSize.MinShardCommitteeSize = 4
		config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 4
		config.Param().ConsensusParam.ConsensusV2Epoch = 1
		config.Param().EpochParam.NumberOfBlockInEpoch = 20
		config.Param().EpochParam.RandomTime = 10
		config.Param().ConsensusParam.EpochBreakPointSwapNewKey = []uint64{1e9}
		config.Config().LimitFee = 0
		config.Param().PDexParams.Pdexv3BreakPointHeight = 1e9
		config.Param().TxPoolVersion = 0
	}, func(node *testsuite.NodeEngine) {})

	return node
}

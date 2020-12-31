package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

type beaconCommitteeEngineSlashingBase struct {
	beaconCommitteeEngineBase
}

func NewBeaconCommitteeEngineSlashingBaseWithValue(
	beaconHeight uint64,
	beaconHash common.Hash,
	finalState *beaconCommitteeStateBase) *beaconCommitteeEngineSlashingBase {
	Logger.log.Infof("Init Beacon Committee Engine V2, %+v", beaconHeight)
	return &beaconCommitteeEngineSlashingBase{
		beaconCommitteeEngineBase: beaconCommitteeEngineBase{
			beaconHeight:     beaconHeight,
			beaconHash:       beaconHash,
			finalState:       finalState,
			uncommittedState: NewBeaconCommitteeStateBase(),
		},
	}
}

func (engine *beaconCommitteeEngineSlashingBase) InitCommitteeState(env *BeaconCommitteeStateEnvironment) {
	engine.beaconCommitteeEngineBase.InitCommitteeState(env)
	////Declare business rules here
	////Declare swaprule interface
	engine.finalState.SetSwapRule(SwapRuleByEnv(env))
}

//Clone :
func (engine *beaconCommitteeEngineSlashingBase) Clone() BeaconCommitteeEngine {
	res := &beaconCommitteeEngineSlashingBase{
		beaconCommitteeEngineBase: *engine.beaconCommitteeEngineBase.Clone().(*beaconCommitteeEngineBase),
	}
	return res
}

//SplitReward ...
func (engine *beaconCommitteeEngineSlashingBase) SplitReward(
	env *BeaconCommitteeStateEnvironment) (
	map[common.Hash]uint64, map[common.Hash]uint64,
	map[common.Hash]uint64, map[common.Hash]uint64, error,
) {
	devPercent := uint64(env.DAOPercent)
	allCoinTotalReward := env.TotalReward
	rewardForBeacon := map[common.Hash]uint64{}
	rewardForShard := map[common.Hash]uint64{}
	rewardForIncDAO := map[common.Hash]uint64{}
	rewardForCustodian := map[common.Hash]uint64{}
	lenBeaconCommittees := uint64(len(engine.GetBeaconCommittee()))
	lenShardCommittees := uint64(len(engine.GetShardCommittee()[env.ShardID]))

	if len(allCoinTotalReward) == 0 {
		Logger.log.Info("Beacon Height %+v, 😭 found NO reward", env.BeaconHeight)
		return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
	}

	for key, totalReward := range allCoinTotalReward {
		totalRewardForDAOAndCustodians := devPercent * totalReward / 100
		totalRewardForShardAndBeaconValidators := totalReward - totalRewardForDAOAndCustodians
		shardWeight := float64(lenShardCommittees)
		beaconWeight := 2 * float64(lenBeaconCommittees) / float64(env.ActiveShards)
		totalValidatorWeight := shardWeight + beaconWeight

		rewardForShard[key] = uint64(shardWeight * float64(totalRewardForShardAndBeaconValidators) / totalValidatorWeight)
		Logger.log.Infof("[test-salary] totalRewardForDAOAndCustodians tokenID %v - %v\n",
			key.String(), totalRewardForDAOAndCustodians)

		if env.IsSplitRewardForCustodian {
			rewardForCustodian[key] += env.PercentCustodianReward * totalRewardForDAOAndCustodians / 100
			rewardForIncDAO[key] += totalRewardForDAOAndCustodians - rewardForCustodian[key]
		} else {
			rewardForIncDAO[key] += totalRewardForDAOAndCustodians
		}
		rewardForBeacon[key] += totalReward - (rewardForShard[key] + totalRewardForDAOAndCustodians)
	}

	return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
}

func (engine *beaconCommitteeEngineSlashingBase) GenerateAllSwapShardInstructions(
	env *BeaconCommitteeStateEnvironment) (
	[]*instruction.SwapShardInstruction, error) {
	swapShardInstructions := []*instruction.SwapShardInstruction{}
	for i := 0; i < env.ActiveShards; i++ {
		shardID := byte(i)
		committees := engine.finalState.ShardCommittee()[shardID]
		substitutes := engine.finalState.ShardSubstitute()[shardID]
		tempCommittees, _ := incognitokey.CommitteeKeyListToString(committees)
		tempSubstitutes, _ := incognitokey.CommitteeKeyListToString(substitutes)

		swapShardInstruction, _, _, _, _ := engine.finalState.SwapRule().GenInstructions(
			shardID,
			tempCommittees,
			tempSubstitutes,
			env.MinShardCommitteeSize,
			env.MaxShardCommitteeSize,
			instruction.SWAP_BY_END_EPOCH,
			env.NumberOfFixedShardBlockValidator,
			env.MissingSignaturePenalty,
		)

		if !swapShardInstruction.IsEmpty() {
			swapShardInstructions = append(swapShardInstructions, swapShardInstruction)
		} else {
			Logger.log.Infof("Generate empty instructions beacon hash: %s & height: %v \n", engine.beaconHash, engine.beaconHash)
		}
	}
	return swapShardInstructions, nil
}

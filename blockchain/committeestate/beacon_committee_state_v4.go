package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"reflect"
	"runtime"
	"sort"
)

type BlockChain interface {
	GetTransactionByHash(txHash common.Hash) (byte, common.Hash, uint64, int, metadata.Transaction, error)
}

type StakerInfo struct {
	cpkStr        incognitokey.CommitteePublicKey
	stakingAmount uint64
	unstake       bool
}
type LockingInfo struct {
	cpkStr        incognitokey.CommitteePublicKey
	lockingEpoch  uint64
	lockingReason int
}

func NewBeaconCommitteeStateV4() *BeaconCommitteeStateV4 {
	return &BeaconCommitteeStateV4{
		beaconCommittee: make(map[string]StakerInfo),
		beaconPending:   make(map[string]StakerInfo),
		beaconWaiting:   make(map[string]StakerInfo),
		beaconLocking:   make(map[string]LockingInfo),
	}
}

type BeaconCommitteeStateV4 struct {
	*BeaconCommitteeStateV3

	//beacon flow
	beaconCommittee map[string]StakerInfo
	beaconPending   map[string]StakerInfo
	beaconWaiting   map[string]StakerInfo
	beaconLocking   map[string]LockingInfo
	stateDB         *statedb.StateDB
	bc              BlockChain
}

func GetKeyStructListFromMapStaker(list map[string]StakerInfo) []incognitokey.CommitteePublicKey {
	keys := []string{}
	for k, _ := range list {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})

	res := make([]incognitokey.CommitteePublicKey, len(keys))
	for i, v := range keys {
		res[i] = list[v].cpkStr
	}
	return res
}

func GetKeyStructListFromMapLocking(list map[string]LockingInfo) []incognitokey.CommitteePublicKey {
	keys := []string{}
	for k, _ := range list {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})

	res := make([]incognitokey.CommitteePublicKey, len(keys))
	for i, v := range keys {
		res[i] = list[v].cpkStr
	}
	return res
}

func (b *BeaconCommitteeStateV4) UpgradeFromV3(stateV3 *BeaconCommitteeStateV3, stateDB *statedb.StateDB) error {
	b.BeaconCommitteeStateV3 = stateV3.Clone(stateDB).(*BeaconCommitteeStateV3)
	for _, cpk := range stateV3.GetBeaconCommittee() {
		cpkStr, _ := cpk.ToBase58()
		info, exists, _ := statedb.GetStakerInfo(stateDB, cpkStr)
		if !exists {
			return fmt.Errorf("Cannot find cpk %v", cpk)
		}
		stakingTx := map[common.Hash]uint64{}
		stakingTx[info.TxStakingID()] = 0
		beaconInfo := statedb.NewBeaconStakerInfoWithValue(info.RewardReceiver(), 1, stakingTx)
		err := statedb.StoreBeaconStakerInfo(stateDB, cpk, beaconInfo)
		if err != nil {

			return err
		}
		//b.BeaconCommitteeStateV3.deleteStakerInfo(cpk, cpkStr, &CommitteeChange{})
		//err = statedb.DeleteStakerInfo(stateDB, []incognitokey.CommitteePublicKey{cpk})
		//if err != nil {
		//
		//	return err
		//}
	}
	err := b.RestoreBeaconCommitteeFromDB(stateDB)
	if err != nil {
		panic(err)
	}

	return nil
}

func (s *BeaconCommitteeStateV4) RemoveBeaconCommittee(cpkList []incognitokey.CommitteePublicKey) error {
	err := statedb.DeleteBeaconCommittee(s.stateDB, cpkList)
	if err != nil {
		return err
	}
	for _, cpk := range cpkList {
		cpkstring, err := cpk.ToBase58()
		if err != nil {
			return err
		}
		delete(s.beaconCommittee, cpkstring)
	}
	return nil
}

func (s *BeaconCommitteeStateV4) RemoveBeaconSubstitute(cpkList []incognitokey.CommitteePublicKey) error {
	err := statedb.DeleteBeaconSubstituteValidator(s.stateDB, cpkList)
	if err != nil {
		return err
	}
	for _, cpk := range cpkList {
		cpkstring, err := cpk.ToBase58()
		if err != nil {
			return err
		}
		delete(s.beaconPending, cpkstring)
	}
	return nil
}

func (s *BeaconCommitteeStateV4) RemoveBeaconWaiting(cpkList []incognitokey.CommitteePublicKey) error {
	err := statedb.DeleteBeaconWaiting(s.stateDB, cpkList)
	if err != nil {
		return err
	}
	for _, cpk := range cpkList {
		cpkstring, err := cpk.ToBase58()
		if err != nil {
			return err
		}
		delete(s.beaconWaiting, cpkstring)
	}
	return nil
}

func (s *BeaconCommitteeStateV4) RemoveBeaconLocking(cpkList []incognitokey.CommitteePublicKey) error {
	err := statedb.DeleteBeaconLocking(s.stateDB, cpkList)
	if err != nil {
		return err
	}
	for _, cpk := range cpkList {
		cpkstring, err := cpk.ToBase58()
		if err != nil {
			return err
		}
		delete(s.beaconLocking, cpkstring)
	}
	err = statedb.DeleteBeaconStakerInfo(s.stateDB, cpkList)
	if err != nil {
		return err
	}
	return nil
}

func (b *BeaconCommitteeStateV4) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	return GetKeyStructListFromMapStaker(b.beaconPending)
}
func (b *BeaconCommitteeStateV4) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	return GetKeyStructListFromMapStaker(b.beaconCommittee)
}
func (b *BeaconCommitteeStateV4) GetBeaconWaiting() []incognitokey.CommitteePublicKey {
	return GetKeyStructListFromMapStaker(b.beaconWaiting)
}
func (b *BeaconCommitteeStateV4) GetBeaconLocking() []incognitokey.CommitteePublicKey {
	return GetKeyStructListFromMapLocking(b.beaconLocking)
}
func NewBeaconCommitteeStateV4WithValue(
	beaconCommittee []string,
	shardCommittee map[byte][]string,
	shardSubstitute map[byte][]string,
	shardCommonPool []string,
	numberOfAssignedCandidates int,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	syncPool map[byte][]string,
	swapRule SwapRuleProcessor,
	assignRule AssignRuleProcessor,
) *BeaconCommitteeStateV4 {
	stateV3 := &BeaconCommitteeStateV3{
		beaconCommitteeStateSlashingBase: *newBeaconCommitteeStateSlashingBaseWithValue(
			nil, shardCommittee, shardSubstitute, autoStake, rewardReceiver, stakingTx,
			shardCommonPool, numberOfAssignedCandidates, swapRule, assignRule,
		),
		syncPool: syncPool,
	}
	stateV4 := NewBeaconCommitteeStateV4()
	stateV4.BeaconCommitteeStateV3 = stateV3
	return stateV4
}

func (b *BeaconCommitteeStateV4) RestoreBeaconCommitteeFromDB(stateDB *statedb.StateDB) error {
	commitee := statedb.GetBeaconCommittee(stateDB)

	for _, cpk := range commitee {
		cpkStr, err := cpk.ToBase58()
		if err != nil {
			return err
		}
		info, exist, _ := statedb.GetBeaconStakerInfo(stateDB, cpkStr)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v", cpkStr)
		}
		b.beaconCommittee[cpkStr] = StakerInfo{cpkStr: cpk, unstake: info.Unstaking(), stakingAmount: info.TotalStakingAmount()}
	}

	pending := statedb.GetBeaconSubstituteValidator(stateDB)
	for _, cpk := range pending {
		cpkStr, err := cpk.ToBase58()
		if err != nil {
			return err
		}
		info, exist, _ := statedb.GetBeaconStakerInfo(stateDB, cpkStr)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v", cpkStr)
		}
		b.beaconPending[cpkStr] = StakerInfo{cpkStr: cpk, unstake: info.Unstaking(), stakingAmount: info.TotalStakingAmount()}
	}

	waiting := statedb.GetBeaconWaiting(stateDB)
	for _, cpk := range waiting {
		cpkStr, err := cpk.ToBase58()
		if err != nil {
			return err
		}
		info, exist, _ := statedb.GetBeaconStakerInfo(stateDB, cpkStr)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v", cpkStr)
		}
		b.beaconWaiting[cpkStr] = StakerInfo{cpkStr: cpk, unstake: info.Unstaking(), stakingAmount: info.TotalStakingAmount()}
	}

	locking := statedb.GetBeaconLocking(stateDB)
	for _, cpk := range locking {
		cpkStr, err := cpk.ToBase58()
		if err != nil {
			return err
		}
		info, exist, _ := statedb.GetBeaconStakerInfo(stateDB, cpkStr)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v", cpkStr)
		}
		b.beaconLocking[cpkStr] = LockingInfo{cpkStr: cpk, lockingEpoch: info.LockingEpoch(), lockingReason: info.LockingReason()}
	}
	return nil
}

func (b *BeaconCommitteeStateV4) Version() int {
	return STAKING_FLOW_V4
}

func (b *BeaconCommitteeStateV4) Clone(cloneState *statedb.StateDB) BeaconCommitteeState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	newState := NewBeaconCommitteeStateV4()
	newState.stateDB = cloneState
	newState.BeaconCommitteeStateV3 = b.BeaconCommitteeStateV3.clone()
	for k, v := range b.beaconCommittee {
		newState.beaconCommittee[k] = v
	}
	for k, v := range b.beaconPending {
		newState.beaconPending[k] = v
	}
	for k, v := range b.beaconWaiting {
		newState.beaconWaiting[k] = v
	}
	for k, v := range b.beaconLocking {
		newState.beaconLocking[k] = v
	}
	return newState
}

func (s *BeaconCommitteeStateV4) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	var stateHash *BeaconCommitteeStateHash
	var instructions [][]string
	var err error

	if _, err := s.ProcessCountShardActiveTime(env); err != nil {
		if err != nil {
			return nil, nil, nil, err
		}
	}

	//Process committee state for shard
	stateHash, changes, instructions, err := s.BeaconCommitteeStateV3.UpdateCommitteeState(env)
	if err != nil {
		return nil, nil, nil, err
	}

	processFuncs := []func(*BeaconCommitteeStateEnvironment) ([][]string, error){
		s.ProcessBeaconUnstakeInstruction,
		s.ProcessBeaconSwapAndSlash,
		s.ProcessBeaconFinishSyncInstruction,
		s.ProcessAssignBeaconPending,
		s.ProcessBeaconStakeInstruction,
		s.ProcessBeaconAddStakingAmountInstruction,
		s.ProcessBeaconUnlocking,
	}

	for _, f := range processFuncs {
		inst, err := f(env)
		if err != nil {
			Logger.log.Error(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "error", err)
		}

		if err != nil {
			return nil, nil, nil, err
		}
		instructions = append(instructions, inst...)
	}

	return stateHash, changes, instructions, nil
}
func firstBlockEpoch(h uint64) bool {
	return h%350 == 1
}

//Process shard active time
func (s *BeaconCommitteeStateV4) ProcessCountShardActiveTime(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	if !firstBlockEpoch(env.BeaconHeight) {
		return nil, nil
	}
	for cpkStr, stakerInfo := range s.beaconWaiting {
		if sig, ok := env.MissingSignature[cpkStr]; ok {
			staker, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpkStr)
			if !exist {
				return nil, fmt.Errorf("Cannot find cpk %v", cpkStr)
			}
			if (sig.Missing*100)/sig.ActualTotal < 80 {
				staker.ResetShardActiveTime()
			} else {
				staker.IncreaseShardActiveTime()
			}
			statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStr, staker)
		}
	}
	return nil, nil
}

//Process slash, unstake and swap
func (s *BeaconCommitteeStateV4) ProcessBeaconSwapAndSlash(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	if !firstBlockEpoch(env.BeaconHeight) {
		return nil, nil
	}

	var slashCpk, unstakeCpk map[string]bool
	var newBeaconCommittee, newBeaconPending map[string]incognitokey.CommitteePublicKey
	var err error

	//check version to swap here
	slashCpk, unstakeCpk, newBeaconCommittee, newBeaconPending, err = s.beacon_swap_v1(env)
	if err != nil {
		return nil, err
	}

	//update slash
	for cpk, _ := range slashCpk {
		info, exists, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exists {
			Logger.log.Errorf("Cannot find %v in beacon staker", cpk)
			continue
		}
		info.SetLocking(env.Epoch, statedb.LOCKING_BY_SLASH)
		var key incognitokey.CommitteePublicKey
		key.FromString(cpk)
		statedb.StoreBeaconStakerInfo(s.stateDB, key, info)
		s.beaconLocking[cpk] = LockingInfo{key, env.Epoch, statedb.LOCKING_BY_SLASH}
		statedb.StoreBeaconLocking(s.stateDB, []incognitokey.CommitteePublicKey{key})
	}

	//update unstake
	for cpk, _ := range unstakeCpk {
		info, exists, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exists {
			Logger.log.Errorf("Cannot find %v in beacon staker", cpk)
			continue
		}
		info.SetLocking(env.Epoch, statedb.LOCKING_BY_UNSTAKE)
		var key incognitokey.CommitteePublicKey
		key.FromString(cpk)
		statedb.StoreBeaconStakerInfo(s.stateDB, key, info)
		s.beaconLocking[cpk] = LockingInfo{key, env.Epoch, statedb.LOCKING_BY_SLASH}
		statedb.StoreBeaconLocking(s.stateDB, []incognitokey.CommitteePublicKey{key})
	}

	//update new beacon committee/pending
	//update statedb
	for k, _ := range s.beaconPending {
		if _, ok := newBeaconPending[k]; !ok {
			statedb.DeleteBeaconSubstituteValidator(s.stateDB, []incognitokey.CommitteePublicKey{s.beaconPending[k].cpkStr})
			delete(s.beaconPending, k)
		}
	}
	for k, _ := range s.beaconCommittee {
		if _, ok := newBeaconCommittee[k]; !ok {
			statedb.DeleteBeaconCommittee(s.stateDB, []incognitokey.CommitteePublicKey{s.beaconPending[k].cpkStr})
			delete(s.beaconCommittee, k)
		}
	}
	for k, cpk := range newBeaconPending {
		if _, ok := s.beaconPending[k]; !ok {
			statedb.StoreBeaconSubstituteValidator(s.stateDB, []incognitokey.CommitteePublicKey{cpk})
		}
	}
	for k, cpk := range newBeaconCommittee {
		if _, ok := s.beaconCommittee[k]; !ok {
			statedb.StoreBeaconCommittee(s.stateDB, []incognitokey.CommitteePublicKey{cpk})
		}
	}
	//update memstate
	beaconPending := map[string]StakerInfo{}
	for cpk, _ := range newBeaconPending {
		if _, ok := s.beaconPending[cpk]; !ok {
			if _, ok := s.beaconCommittee[cpk]; !ok {
				return nil, fmt.Errorf("Cannot find cpl %v in pending and committee list", cpk)
			}
			beaconPending[cpk] = s.beaconCommittee[cpk]
		} else {
			beaconPending[cpk] = s.beaconPending[cpk]
		}
	}

	beaconCommittee := map[string]StakerInfo{}
	for cpk, _ := range newBeaconPending {
		if _, ok := s.beaconPending[cpk]; !ok {
			if _, ok := s.beaconCommittee[cpk]; !ok {
				return nil, fmt.Errorf("Cannot find cpk %v in pending and committee list", cpk)
			}
			beaconCommittee[cpk] = s.beaconPending[cpk]
		} else {
			beaconCommittee[cpk] = s.beaconCommittee[cpk]
		}
	}
	s.beaconPending = beaconPending
	s.beaconCommittee = beaconCommittee

	return nil, nil
}

func (s *BeaconCommitteeStateV4) ProcessBeaconFinishSyncInstruction(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	for _, inst := range env.BeaconInstructions {
		if inst[0] == instruction.FINISH_SYNC_ACTION && inst[1] == "-1" {
			finishSyncInst, err := instruction.ImportFinishSyncInstructionFromString(inst)
			if err != nil {
				return nil, err
			}
			for _, cpk := range finishSyncInst.PublicKeys {
				info, exists, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
				if !exists {
					Logger.log.Errorf("Cannot find %v in beacon staker", cpk)
					continue
				}
				info.SetFinishSync()
				var key incognitokey.CommitteePublicKey
				key.FromString(cpk)
				statedb.StoreBeaconStakerInfo(s.stateDB, key, info)
			}
		}
	}
	return nil, nil
}

//Process assign beacon pending (sync, sync valid time)
func (s *BeaconCommitteeStateV4) ProcessAssignBeaconPending(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	//if finish sync & enough valid time & shard staker is unstaked -> update role to pending
	for cpk, stakerInfo := range s.beaconWaiting {
		staker, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return nil, fmt.Errorf("Cannot find stakerInfo %v", cpk)
		}
		_, shardExist, _ := statedb.GetStakerInfo(s.stateDB, cpk)
		if staker.FinishSync() && staker.ShardActiveTime() > 50 && !shardExist {
			delete(s.beaconWaiting, cpk)
			statedb.DeleteBeaconWaiting(s.stateDB, []incognitokey.CommitteePublicKey{stakerInfo.cpkStr})
			s.beaconPending[cpk] = stakerInfo
			statedb.StoreBeaconSubstituteValidator(s.stateDB, []incognitokey.CommitteePublicKey{stakerInfo.cpkStr})
		}
	}
	return nil, nil
}

//Process stake instruction
//-> update waiting
//-> store beacon staker info
func (s *BeaconCommitteeStateV4) ProcessBeaconStakeInstruction(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	returnStakingList := [][]string{}
	return_cpk := []string{}
	return_amount := []uint64{}
	return_reason := []int{}

	for _, inst := range env.BeaconInstructions {
		if inst[0] == instruction.STAKE_ACTION && inst[2] == "-1" {
			beaconStakeInst := instruction.ImportInitStakeInstructionFromString(inst)
			for i, _ := range beaconStakeInst.TxStakeHashes {
				_, _, _, _, stakingTx, err := s.bc.GetTransactionByHash(beaconStakeInst.TxStakeHashes[i])
				if err != nil {
					return nil, fmt.Errorf("Cannot find staking tx %v", beaconStakeInst.TxStakeHashes[i].String())
				}
				stakingMetadata := stakingTx.GetMetadata().(*metadata.StakingMetadata)

				//return staking if already exist
				_, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, beaconStakeInst.PublicKeys[i])
				if exist {
					return_cpk = append(return_cpk, beaconStakeInst.PublicKeys[i])
					return_amount = append(return_amount, stakingMetadata.StakingAmount)
					return_reason = append(return_reason, statedb.RETURN_BY_DUPLICATE_STAKE)
					continue
				}

				var key incognitokey.CommitteePublicKey
				key.FromString(beaconStakeInst.PublicKeys[i])
				stakingInfo := map[common.Hash]uint64{}
				stakingInfo[*stakingMetadata.Hash()] = stakingMetadata.StakingAmount
				info := statedb.NewBeaconStakerInfoWithValue(beaconStakeInst.RewardReceiverStructs[i], env.BeaconHeight, stakingInfo)
				statedb.StoreBeaconStakerInfo(s.stateDB, key, info)
				s.beaconWaiting[beaconStakeInst.PublicKeys[i]] = StakerInfo{key, stakingMetadata.StakingAmount, false}
				statedb.StoreBeaconWaiting(s.stateDB, []incognitokey.CommitteePublicKey{key})
			}
		}
	}
	if len(return_cpk) == 0 {
		return nil, nil
	}
	returnStakingList = append(returnStakingList, instruction.NewReturnBeaconStakeInsWithValue(return_cpk, return_reason, return_amount).ToString())

	return returnStakingList, nil
}

//Process add stake amount
func (s *BeaconCommitteeStateV4) ProcessBeaconAddStakingAmountInstruction(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	returnStakingInstList := [][]string{}
	return_cpk := []string{}
	return_reason := []int{}
	return_amount := []uint64{}
	for _, inst := range env.BeaconInstructions {
		if inst[0] == instruction.ADD_STAKING_ACTION {
			addStakeInst := instruction.ImportAddStakingInstructionFromString(inst)
			for i, cpk := range addStakeInst.CommitteePublicKeys {
				info, exists, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
				if !exists {
					Logger.log.Errorf("Cannot find %v in beacon staker", cpk)
					return_cpk = append(return_cpk, cpk)
					return_reason = append(return_reason, statedb.RETURN_BY_ADDSTAKE_FAIL)
					return_amount = append(return_amount, addStakeInst.StakingAmount[i])
					continue
				}
				if _, ok := s.beaconLocking[cpk]; ok {
					Logger.log.Errorf("Add staking to locking beacon staker", cpk)
					return_cpk = append(return_cpk, cpk)
					return_reason = append(return_reason, statedb.RETURN_BY_ADDSTAKE_FAIL)
					return_amount = append(return_amount, addStakeInst.StakingAmount[i])
					continue
				}

				stakingTxHash, err := common.Hash{}.NewHashFromStr(addStakeInst.StakingTxIDs[i])
				if err != nil {
					Logger.log.Errorf("Convert staking tx hash fail %v", addStakeInst.StakingTxIDs[i])
					continue
				}
				info.AddStaking(*stakingTxHash, addStakeInst.StakingAmount[i])
				var key incognitokey.CommitteePublicKey
				key.FromString(cpk)
				statedb.StoreBeaconStakerInfo(s.stateDB, key, info)
				if staker, ok := s.beaconCommittee[cpk]; ok {
					staker.stakingAmount = info.TotalStakingAmount()
					s.beaconCommittee[cpk] = staker
				}
				if staker, ok := s.beaconWaiting[cpk]; ok {
					staker.stakingAmount = info.TotalStakingAmount()
					s.beaconWaiting[cpk] = staker
				}
				if staker, ok := s.beaconPending[cpk]; ok {
					staker.stakingAmount = info.TotalStakingAmount()
					s.beaconPending[cpk] = staker
				}

			}
		}
	}
	if len(return_cpk) == 0 {
		return nil, nil
	}
	returnStakingInstList = append(returnStakingInstList, instruction.NewReturnBeaconStakeInsWithValue(return_cpk, return_reason, return_amount).ToString())
	return returnStakingInstList, nil
}

//Process unstaking instruction -> set unstake
func (s *BeaconCommitteeStateV4) ProcessBeaconUnstakeInstruction(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	for _, inst := range env.BeaconInstructions {
		if inst[0] == instruction.UNSTAKE_ACTION {
			unstakeInst := instruction.NewUnstakeInstructionWithValue(inst)
			for _, cpk := range unstakeInst.CommitteePublicKeys {
				info, exists, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
				if !exists {
					Logger.log.Errorf("Cannot find %v in beacon staker", cpk)
					continue
				}
				info.SetUnstaking()
				var key incognitokey.CommitteePublicKey
				key.FromString(cpk)
				statedb.StoreBeaconStakerInfo(s.stateDB, key, info)
			}
		}
	}
	return nil, nil
}

//Process return staking amount (unlocking)
func (s *BeaconCommitteeStateV4) ProcessBeaconUnlocking(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	if !firstBlockEpoch(env.BeaconHeight) {
		return nil, nil
	}
	returnStakingInstList := [][]string{}
	return_cpk := []string{}
	return_reason := []int{}
	return_amount := []uint64{}
	for cpk, lockingInfo := range s.beaconLocking {
		info, exists, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exists {
			Logger.log.Errorf("Cannot find %v in beacon staker", cpk)
			continue
		}
		if env.Epoch >= info.LockingEpoch()+150 {
			return_cpk = append(return_cpk, cpk)
			switch lockingInfo.lockingReason {
			case statedb.LOCKING_BY_SLASH:
				return_reason = append(return_reason, statedb.RETURN_BY_SLASH)
			case statedb.LOCKING_BY_UNSTAKE:
				return_reason = append(return_reason, statedb.RETURN_BY_UNSTAKE)
			}
			return_amount = append(return_amount, info.TotalStakingAmount())
			statedb.DeleteBeaconLocking(s.stateDB, []incognitokey.CommitteePublicKey{lockingInfo.cpkStr})
			statedb.DeleteBeaconStakerInfo(s.stateDB, []incognitokey.CommitteePublicKey{lockingInfo.cpkStr})
		}
	}
	if len(return_cpk) == 0 {
		return nil, nil
	}
	returnStakingInstList = append(returnStakingInstList, instruction.NewReturnBeaconStakeInsWithValue(return_cpk, return_reason, return_amount).ToString())
	return returnStakingInstList, nil
}

func (s *BeaconCommitteeStateV4) beacon_swap_v1(env *BeaconCommitteeStateEnvironment) (
	map[string]bool, map[string]bool,
	map[string]incognitokey.CommitteePublicKey, map[string]incognitokey.CommitteePublicKey,
	error) {

	//slash
	slashCpk := map[string]bool{}
	for cpk, _ := range s.beaconCommittee {
		if sig, ok := env.MissingSignature[cpk]; ok {
			if (sig.Missing*100)/sig.ActualTotal < 50 {
				slashCpk[cpk] = true
			}
		}
	}

	//unstake
	unstakeCpk := map[string]bool{}
	for cpk, stakerInfo := range s.beaconCommittee {
		if stakerInfo.unstake {
			unstakeCpk[cpk] = true
		}
	}

	//swap pending <-> committee
	type CandidateInfo struct {
		cpk    incognitokey.CommitteePublicKey
		cpkStr string
		score  uint64
	}
	candidateList := []CandidateInfo{}
	for cpk, stakerInfo := range s.beaconPending {
		if !slashCpk[cpk] && !unstakeCpk[cpk] {
			score := uint64(env.MissingSignature[cpk].Missing*100/env.MissingSignature[cpk].ActualTotal) * stakerInfo.stakingAmount
			candidateList = append(candidateList, CandidateInfo{stakerInfo.cpkStr, cpk, score})
		}
	}
	for cpk, stakerInfo := range s.beaconCommittee {
		if !slashCpk[cpk] && !unstakeCpk[cpk] {
			score := uint64(env.MissingSignature[cpk].Missing*100/env.MissingSignature[cpk].ActualTotal) * stakerInfo.stakingAmount
			candidateList = append(candidateList, CandidateInfo{stakerInfo.cpkStr, cpk, score})
		}
	}
	//sort candidate list
	sort.Slice(candidateList, func(i, j int) bool {
		return candidateList[i].score > candidateList[j].score
	})
	newBeaconCommittee := map[string]incognitokey.CommitteePublicKey{}
	newBeaconPending := map[string]incognitokey.CommitteePublicKey{}
	for i := 0; i < env.MaxBeaconCommitteeSize; i++ {
		newBeaconCommittee[candidateList[i].cpkStr] = candidateList[i].cpk
	}
	for i := env.MaxBeaconCommitteeSize; i < len(candidateList); i++ {
		newBeaconPending[candidateList[i].cpkStr] = candidateList[i].cpk
	}
	return slashCpk, unstakeCpk, newBeaconCommittee, newBeaconPending, nil
}

func (b BeaconCommitteeStateV4) GetAllStaker() (map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey) {
	sC := make(map[byte][]incognitokey.CommitteePublicKey)
	sPV := make(map[byte][]incognitokey.CommitteePublicKey)
	sSP := make(map[byte][]incognitokey.CommitteePublicKey)
	for shardID, committee := range b.GetShardCommittee() {
		sC[shardID] = append([]incognitokey.CommitteePublicKey{}, committee...)
	}
	for shardID, Substitute := range b.GetShardSubstitute() {
		sPV[shardID] = append([]incognitokey.CommitteePublicKey{}, Substitute...)
	}
	for shardID, syncValidator := range b.GetSyncingValidators() {
		sSP[shardID] = append([]incognitokey.CommitteePublicKey{}, syncValidator...)
	}
	bC := b.GetBeaconCommittee()
	bPV := b.GetBeaconSubstitute()
	bW := b.GetBeaconWaiting()
	bL := b.GetBeaconLocking()
	cSWFCR := b.GetCandidateShardWaitingForCurrentRandom()
	cSWFNR := b.GetCandidateShardWaitingForNextRandom()
	return sC, sPV, sSP, bC, bPV, bW, bL, cSWFCR, cSWFNR
}

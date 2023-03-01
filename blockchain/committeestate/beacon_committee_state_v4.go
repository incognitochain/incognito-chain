package committeestate

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sort"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/key"
)

const INIT_SHARE_PRICE = 1750 * 1e9

// must trigger version for new config
type BeaconCommitteeStateV4Config struct {
	MAX_SCORE             uint64
	MIN_SCORE             uint64
	DEFAULT_PERFORMING    uint64
	INCREASE_PERFORMING   uint64
	DECREASE_PERFORMING   uint64
	MIN_ACTIVE_SHARD      int
	MIN_WAITING_PERIOD    int64
	MIN_PERFORMANCE       uint64
	LOCKING_PERIOD        uint64
	LOCKING_FACTOR        float64
	BEACON_COMMITTEE_SIZE int
}

func NewBeaconCommitteeStateV4Config(version int) BeaconCommitteeStateV4Config {
	switch config.Config().Network() {
	case "mainnet":
		return BeaconCommitteeStateV4Config{
			MAX_SCORE:             1000,
			MIN_SCORE:             100,
			DEFAULT_PERFORMING:    500,
			INCREASE_PERFORMING:   1015,
			DECREASE_PERFORMING:   965,
			MIN_ACTIVE_SHARD:      10,
			MIN_WAITING_PERIOD:    60 * 60 * 24 * 3, //seconds : 3 days
			MIN_PERFORMANCE:       100,
			LOCKING_PERIOD:        160,
			LOCKING_FACTOR:        5,
			BEACON_COMMITTEE_SIZE: 32,
		}
	case "local":
		return BeaconCommitteeStateV4Config{
			MAX_SCORE:             1000,
			MIN_SCORE:             100,
			DEFAULT_PERFORMING:    500,
			INCREASE_PERFORMING:   1015,
			DECREASE_PERFORMING:   965,
			MIN_ACTIVE_SHARD:      2,
			MIN_WAITING_PERIOD:    60, //seconds
			MIN_PERFORMANCE:       370,
			LOCKING_PERIOD:        2,
			LOCKING_FACTOR:        1,
			BEACON_COMMITTEE_SIZE: 6,
		}
	default:
		return BeaconCommitteeStateV4Config{
			MAX_SCORE:             1000,
			MIN_SCORE:             100,
			DEFAULT_PERFORMING:    500,
			INCREASE_PERFORMING:   1015,
			DECREASE_PERFORMING:   965,
			MIN_ACTIVE_SHARD:      4,
			MIN_WAITING_PERIOD:    60 * 60, //seconds
			MIN_PERFORMANCE:       100,
			LOCKING_PERIOD:        4,
			LOCKING_FACTOR:        2,
			BEACON_COMMITTEE_SIZE: 8,
		}
	}
}

const (
	COMMITTEE_POOL = iota
	PENDING_POOL
	WAITING_POOL
	LOCKING_POOL
)

type StakerInfo struct {
	cpkStruct       incognitokey.CommitteePublicKey
	CPK             string
	StakingAmount   uint64
	Unstake         bool
	Performance     uint64
	EpochScore      uint64 // -> sorted list
	FixedNode       bool
	enterTime       int64
	FinishSync      bool
	ShardActiveTime int
	//delegation
	stakeID         string
	TotalDelegators uint64 `json:"TotalDelegators,omitempty"`
}

type LockingInfo struct {
	cpkStr        incognitokey.CommitteePublicKey
	LockingEpoch  uint64
	LockingReason int
}

type StateDataDetail struct {
	Committee []StakerInfoDetail
	Pending   []StakerInfoDetail
	Waiting   []StakerInfoDetail
	Locking   []LockingInfoDetail
}

type StakerInfoDetail struct {
	StakeTime     int64
	PoolEnterTime int64
	StakerInfo
}

type LockingInfoDetail struct {
	CPK           string
	StakeTime     int64
	PoolEnterTime int64
	LockingEpoch  uint64
	LockingReason int
	ReleaseEpoch  uint64
	ReleaseAmount uint64
}

type BeaconCommitteeStateV4 struct {
	config BeaconCommitteeStateV4Config

	*BeaconCommitteeStateV3

	//beacon flow
	beaconCommittee map[string]*StakerInfo
	beaconPending   map[string]*StakerInfo
	beaconWaiting   map[string]*StakerInfo
	beaconLocking   map[string]*LockingInfo
	stateDB         *statedb.StateDB

	enableDelegate bool
}

func NewBeaconCommitteeStateV4WithValue(
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
	stateV4.config = NewBeaconCommitteeStateV4Config(1)
	return stateV4
}

func NewBeaconCommitteeStateV4() *BeaconCommitteeStateV4 {
	return &BeaconCommitteeStateV4{
		beaconCommittee: make(map[string]*StakerInfo),
		beaconPending:   make(map[string]*StakerInfo),
		beaconWaiting:   make(map[string]*StakerInfo),
		beaconLocking:   make(map[string]*LockingInfo),
		config:          NewBeaconCommitteeStateV4Config(1),
	}
}

func (s BeaconCommitteeStateV4) GetConfig() BeaconCommitteeStateV4Config {
	return s.config
}

func (s *BeaconCommitteeStateV4) getBeaconStakerInfo(cpk string) *StakerInfo {
	if stakerInfo, ok := s.beaconCommittee[cpk]; ok {
		return stakerInfo
	}
	if stakerInfo, ok := s.beaconPending[cpk]; ok {
		return stakerInfo
	}
	if stakerInfo, ok := s.beaconWaiting[cpk]; ok {
		return stakerInfo
	}
	return nil
}

func (s *BeaconCommitteeStateV4) setShardActiveTime(cpk string, t int) error {
	if stakerInfo := s.getBeaconStakerInfo(cpk); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", cpk)
		}
		info.SetShardActiveTime(t)
		stakerInfo.ShardActiveTime = t
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStruct, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", cpk)
}

func (s *BeaconCommitteeStateV4) setUnstake(cpk string) error {
	if stakerInfo := s.getBeaconStakerInfo(cpk); stakerInfo != nil {
		stakerInfo.Unstake = true
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", cpk)
		}
		info.SetUnstaking()
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStruct, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", cpk)
}

func (s *BeaconCommitteeStateV4) setEnterTime(cpk string, t int64) error {
	if stakerInfo := s.getBeaconStakerInfo(cpk); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", cpk)
		}
		info.SetEnterTime(t)
		stakerInfo.enterTime = t
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStruct, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", cpk)
}

func (s *BeaconCommitteeStateV4) setFinishSync(cpk string, b bool) error {
	if stakerInfo := s.getBeaconStakerInfo(cpk); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", cpk)
		}
		info.SetFinishSync(b)
		stakerInfo.FinishSync = b
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStruct, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", cpk)
}

func (s *BeaconCommitteeStateV4) addStakingTx(cpk string, tx common.Hash, amount, height uint64) error {
	if stakerInfo := s.getBeaconStakerInfo(cpk); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", cpk)
		}
		info.AddStaking(tx, height, amount)
		stakerInfo.StakingAmount = info.TotalStakingAmount()
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStruct, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", cpk)
}

func (s *BeaconCommitteeStateV4) setLocking(cpk string, epoch, unlockEpoch uint64, reason int) error {
	if stakerInfo := s.getBeaconStakerInfo(cpk); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", cpk)
		}
		info.SetLocking(epoch, unlockEpoch, reason)
		s.beaconLocking[cpk] = &LockingInfo{stakerInfo.cpkStruct, epoch, reason}
		if err := statedb.StoreBeaconLocking(s.stateDB, []incognitokey.CommitteePublicKey{stakerInfo.cpkStruct}); err != nil {
			return err
		}
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStruct, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", cpk)
}

func (s *BeaconCommitteeStateV4) removeFromPool(pool int, cpk string) error {
	cpkStruct := incognitokey.CommitteePublicKey{}
	if err := cpkStruct.FromString(cpk); err != nil {
		return err
	}
	switch pool {
	case COMMITTEE_POOL:
		delete(s.beaconCommittee, cpk)
		if err := statedb.DeleteBeaconCommittee(s.stateDB, []incognitokey.CommitteePublicKey{cpkStruct}); err != nil {
			return err
		}
	case PENDING_POOL:
		delete(s.beaconPending, cpk)
		if err := statedb.DeleteBeaconSubstituteValidator(s.stateDB, []incognitokey.CommitteePublicKey{cpkStruct}); err != nil {
			return err
		}
	case WAITING_POOL:
		delete(s.beaconWaiting, cpk)
		if err := statedb.DeleteBeaconWaiting(s.stateDB, []incognitokey.CommitteePublicKey{cpkStruct}); err != nil {
			return err
		}
	case LOCKING_POOL:
		delete(s.beaconLocking, cpk)
		if err := statedb.DeleteBeaconLocking(s.stateDB, []incognitokey.CommitteePublicKey{cpkStruct}); err != nil {
			return err
		}
	default:
		panic("must not be here")
	}
	return nil
}

func (s *BeaconCommitteeStateV4) addToPool(pool int, cpk string, stakerInfo *StakerInfo) error {
	t := time.Now().UnixNano()
	stakerInfo.enterTime = t
	s.setEnterTime(cpk, t) //save to beacon staker info
	switch pool {
	case COMMITTEE_POOL:
		s.beaconCommittee[cpk] = stakerInfo
		if err := statedb.StoreBeaconCommittee(s.stateDB, []incognitokey.CommitteePublicKey{stakerInfo.cpkStruct}); err != nil {
			return err
		}
	case PENDING_POOL:
		s.beaconPending[cpk] = stakerInfo
		if err := statedb.StoreBeaconSubstituteValidator(s.stateDB, []incognitokey.CommitteePublicKey{stakerInfo.cpkStruct}); err != nil {
			return err
		}
	case WAITING_POOL:
		s.beaconWaiting[cpk] = stakerInfo
		if err := statedb.StoreBeaconWaiting(s.stateDB, []incognitokey.CommitteePublicKey{stakerInfo.cpkStruct}); err != nil {
			return err
		}
	default:
		panic("must not be here")
	}
	return nil
}

func (s *BeaconCommitteeStateV4) UpgradeFromV3(stateV3 *BeaconCommitteeStateV3, stateDB *statedb.StateDB, minBeaconCommitteeSize int) error {
	s.BeaconCommitteeStateV3 = stateV3.Clone(stateDB).(*BeaconCommitteeStateV3)
	s.BeaconCommitteeStateV3.beaconCommittee = []string{}
	scores := map[string]statedb.CurrentEpochCommitteeAndPendingInfo{}
	for _, cpk := range stateV3.GetBeaconCommittee() {
		cpkStr, _ := cpk.ToBase58()

		stakeID := common.HashH([]byte(fmt.Sprintf("%v_%v", cpkStr, 1)))
		scores[cpkStr] = statedb.CurrentEpochCommitteeAndPendingInfo{0, s.config.DEFAULT_PERFORMING, 0, 0, stakeID.String()}
		info, exists, _ := statedb.GetShardStakerInfo(stateDB, cpkStr)
		if !exists {
			return fmt.Errorf("Cannot find cpk %v", cpk)
		}
		stakingTx := map[common.Hash]statedb.StakingTxInfo{}
		stakingTx[info.TxStakingID()] = statedb.StakingTxInfo{0, 1}
		beaconInfo := statedb.NewBeaconStakerInfoWithValue(info.RewardReceiver(), info.RewardReceiver(), 1, 1, time.Now().UnixNano(), stakingTx)
		err := statedb.StoreBeaconStakerInfo(stateDB, cpk, beaconInfo)
		if err != nil {

			return err
		}
	}
	err := statedb.StoreCommitteeData(stateDB, &statedb.CommitteeData{BeginEpochInfo: scores})
	if err != nil {
		panic(err)
	}
	err = s.RestoreBeaconCommitteeFromDB(stateDB, minBeaconCommitteeSize, nil)
	if err != nil {
		panic(err)
	}

	return nil
}

func (s *BeaconCommitteeStateV4) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	return GetKeyStructListFromMapStaker(s.beaconPending)
}

func (s *BeaconCommitteeStateV4) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	return GetKeyStructListFromMapStaker(s.beaconCommittee)
}

func (s *BeaconCommitteeStateV4) GetBeaconWaiting() []incognitokey.CommitteePublicKey {
	return GetKeyStructListFromMapStaker(s.beaconWaiting)
}

// result is not consistent
func (s *BeaconCommitteeStateV4) GetUnsyncBeaconValidator() []incognitokey.CommitteePublicKey {
	res := []incognitokey.CommitteePublicKey{}
	for _, v := range s.beaconWaiting {
		if !v.FinishSync {
			res = append(res, v.cpkStruct)
		}
	}
	for _, v := range s.beaconPending {
		if !v.FinishSync {
			res = append(res, v.cpkStruct)
		}
	}
	return res
}

func (s *BeaconCommitteeStateV4) GetBeaconLocking() []incognitokey.CommitteePublicKey {
	return GetKeyStructListFromMapLocking(s.beaconLocking)
}

func (s *BeaconCommitteeStateV4) GetNonSlashingRewardReceiver(staker []incognitokey.CommitteePublicKey) ([]key.PaymentAddress, error) {
	res := []key.PaymentAddress{}
	for _, k := range staker {
		kString, err := k.ToBase58()
		if err != nil {
			return []key.PaymentAddress{}, fmt.Errorf("Base58 error")
		}
		info, exists, _ := statedb.GetBeaconStakerInfo(s.stateDB, kString)
		if !exists || info == nil {
			return []key.PaymentAddress{}, fmt.Errorf(kString + "not found!")
		}
		if info.LockingReason() != statedb.BY_SLASH {
			res = append(res, info.RewardReceiver())
		}
	}
	return res, nil
}

func (s BeaconCommitteeStateV4) DebugBeaconCommitteeState() *StateDataDetail {

	data := &StateDataDetail{
		Committee: []StakerInfoDetail{},
		Pending:   []StakerInfoDetail{},
		Waiting:   []StakerInfoDetail{},
		Locking:   []LockingInfoDetail{}}

	getLockingDetail := func(cpk string) LockingInfoDetail {
		stakerInfo, has, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !has {
			return LockingInfoDetail{}
		}
		detail := LockingInfoDetail{
			cpk,
			stakerInfo.BeaconConfirmTime(),
			stakerInfo.GetEnterTime(),
			(*s.beaconLocking[cpk]).LockingEpoch,
			(*s.beaconLocking[cpk]).LockingReason,
			stakerInfo.UnlockingEpoch(),
			stakerInfo.TotalStakingAmount(),
		}
		return detail
	}

	for _, v := range s.GetBeaconCommittee() {
		cpk, _ := v.ToBase58()
		stakerInfoDB, _, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		data.Committee = append(data.Committee, StakerInfoDetail{stakerInfoDB.BeaconConfirmTime(), stakerInfoDB.GetEnterTime(), *s.getBeaconStakerInfo(cpk)})
	}

	for _, v := range s.GetBeaconSubstitute() {
		cpk, _ := v.ToBase58()
		stakerInfoDB, _, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		data.Pending = append(data.Pending, StakerInfoDetail{stakerInfoDB.BeaconConfirmTime(), stakerInfoDB.GetEnterTime(), *s.getBeaconStakerInfo(cpk)})
	}
	for _, v := range s.GetBeaconWaiting() {
		cpk, _ := v.ToBase58()
		stakerInfoDB, _, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		data.Waiting = append(data.Waiting, StakerInfoDetail{stakerInfoDB.BeaconConfirmTime(), stakerInfoDB.GetEnterTime(), *s.getBeaconStakerInfo(cpk)})
	}
	for _, v := range s.GetBeaconLocking() {
		cpk, _ := v.ToBase58()
		detail := getLockingDetail(cpk)
		data.Locking = append(data.Locking, detail)
	}

	return data
}

func (s *BeaconCommitteeStateV4) RestoreBeaconCommitteeFromDB(stateDB *statedb.StateDB, minBeaconCommitteeSize int, allBeaconBlock []types.BeaconBlock) error {
	s.stateDB = stateDB
	s.config = NewBeaconCommitteeStateV4Config(1)
	commitee := statedb.GetBeaconCommittee(stateDB)
	commiteeData := statedb.GetCommitteeData(stateDB)
	//
	for index, cpk := range commitee {
		cpkStr, err := cpk.ToBase58()
		if err != nil {
			return err
		}
		info, exist, _ := statedb.GetBeaconStakerInfo(stateDB, cpkStr)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v", cpkStr)
		}
		stakeID, err := s.GetBeaconCandidateUID(cpkStr)
		if err != nil {
			return err
		}
		s.beaconCommittee[cpkStr] = &StakerInfo{cpkStruct: cpk, stakeID: stakeID, CPK: cpkStr, Unstake: info.Unstaking(), StakingAmount: info.TotalStakingAmount(), FinishSync: info.FinishSync(), ShardActiveTime: info.ShardActiveTime(), TotalDelegators: info.TotalDelegators()}
		s.beaconCommittee[cpkStr].EpochScore = commiteeData.BeginEpochInfo[cpkStr].Score
		s.beaconCommittee[cpkStr].Performance = commiteeData.BeginEpochInfo[cpkStr].Performance
		s.beaconCommittee[cpkStr].enterTime = info.GetEnterTime()
		if index < minBeaconCommitteeSize {
			s.beaconCommittee[cpkStr].FixedNode = true
		}
		//if not init share price, init it
		sharePrice, _, _ := statedb.GetBeaconSharePrice(stateDB, stakeID)
		if sharePrice == nil {
			statedb.StoreBeaconSharePrice(stateDB, stakeID, INIT_SHARE_PRICE)
		}
	}

	for _, blk := range allBeaconBlock {
		if lastBlockEpoch(allBeaconBlock[len(allBeaconBlock)-1].GetBeaconHeight()) { //no need to restore performance if the last block is end of epoch
			break
		}
		err := s.updateBeaconPerformance(blk.Header.PreviousValidationData)
		if err != nil {
			panic(err)
			return err
		}
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
		stakeID, err := s.GetBeaconCandidateUID(cpkStr)
		if err != nil {
			return err
		}
		s.beaconPending[cpkStr] = &StakerInfo{cpkStruct: cpk, stakeID: stakeID, CPK: cpkStr, Unstake: info.Unstaking(),
			EpochScore:      commiteeData.BeginEpochInfo[cpkStr].Score,
			StakingAmount:   info.TotalStakingAmount(),
			Performance:     s.config.DEFAULT_PERFORMING,
			enterTime:       info.GetEnterTime(),
			FinishSync:      info.FinishSync(),
			ShardActiveTime: info.ShardActiveTime(),
			TotalDelegators: info.TotalDelegators(),
		}
		//if not init share price, init it
		sharePrice, _, _ := statedb.GetBeaconSharePrice(stateDB, stakeID)
		if sharePrice == nil {
			statedb.StoreBeaconSharePrice(stateDB, stakeID, INIT_SHARE_PRICE)
		}
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
		stakeID, err := s.GetBeaconCandidateUID(cpkStr)
		if err != nil {
			return err
		}
		s.beaconWaiting[cpkStr] = &StakerInfo{cpkStruct: cpk, stakeID: stakeID, CPK: cpkStr, Unstake: info.Unstaking(),
			StakingAmount: info.TotalStakingAmount(), Performance: s.config.DEFAULT_PERFORMING,
			enterTime:       info.GetEnterTime(),
			FinishSync:      info.FinishSync(),
			ShardActiveTime: info.ShardActiveTime()}
		//if not init share price, init it
		sharePrice, _, _ := statedb.GetBeaconSharePrice(stateDB, stakeID)
		if sharePrice == nil {
			statedb.StoreBeaconSharePrice(stateDB, stakeID, INIT_SHARE_PRICE)
		}
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
		s.beaconLocking[cpkStr] = &LockingInfo{cpkStr: cpk, LockingEpoch: info.LockingEpoch(), LockingReason: info.LockingReason()}
	}

	return nil
}

type ProcessContext struct {
	*BeaconCommitteeStateEnvironment
	RemovedStaker []string
}

func (s *BeaconCommitteeStateV4) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	var stateHash *BeaconCommitteeStateHash
	var instructions [][]string
	var err error
	ctx := ProcessContext{BeaconCommitteeStateEnvironment: env}

	if _, err := s.ProcessCountShardActiveTime(ctx); err != nil { //Review Count shard active times after process committee state for shard
		if err != nil {
			return nil, nil, nil, err
		}
	}

	//Process committee state for shard
	stateHash, changes, instructions, err := s.BeaconCommitteeStateV3.UpdateCommitteeState(env)
	if err != nil {
		return nil, nil, nil, err
	}
	ctx.RemovedStaker = changes.RemovedStaker

	processFuncs := []func(ProcessContext) ([][]string, error){
		s.ProcessWithdrawDelegationReward,
		s.ProcessUpdateBeaconPerformance,
		s.ProcessBeaconUnstakeInstruction,
		s.ProcessDelegateRewardForReturnValidator,
		s.ProcessBeaconRedelegateInstruction,
		s.ProcessBeaconSwapAndSlash,
		s.ProcessBeaconFinishSyncInstruction,
		s.ProcessBeaconWaitingCondition,
		s.ProcessBeaconStakeInstruction,
		s.ProcessBeaconAddStakingAmountInstruction,
		s.ProcessBeaconUnlocking,
		s.ProcessBeaconSharePrice,
	}

	for _, f := range processFuncs {
		inst, err := f(ctx)
		if err != nil {
			Logger.log.Error(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "error", err)
		}

		if err != nil {
			return nil, nil, nil, err
		}
		instructions = append(instructions, inst...)
	}

	//udpate beacon state

	stateHash.BeaconCommitteeAndValidatorHash = s.commiteeAndPendingHash()
	stateHash.BeaconCandidateHash = s.waitingAndSlashingHash()

	return stateHash, changes, instructions, nil
}
func (s *BeaconCommitteeStateV4) commiteeAndPendingHash() common.Hash {

	type Data struct {
		Config          BeaconCommitteeStateV4Config
		BeaconCommittee []*StakerInfo
		BeaconPending   []*StakerInfo
	}
	beaconCommittee := s.GetBeaconCommittee()
	beaconPending := s.GetBeaconSubstitute()
	data := Data{
		Config:          s.config,
		BeaconCommittee: []*StakerInfo{},
		BeaconPending:   []*StakerInfo{},
	}
	for _, v := range beaconCommittee {
		vStr, _ := v.ToBase58()
		stakerInfo := s.getBeaconStakerInfo(vStr)
		if stakerInfo == nil {
			panic(fmt.Sprintf("Can not get staker info of this pk %v", vStr))
		}
		data.BeaconCommittee = append(data.BeaconCommittee, stakerInfo)
	}
	for _, v := range beaconPending {
		vStr, _ := v.ToBase58()
		stakerInfo := s.getBeaconStakerInfo(vStr)
		if stakerInfo == nil {
			panic(fmt.Sprintf("Can not get staker info of this pk %v", vStr))
		}
		data.BeaconPending = append(data.BeaconPending, stakerInfo)
	}
	b, _ := json.Marshal(data)
	hash := common.HashH(b)
	return hash
}

func (s *BeaconCommitteeStateV4) waitingAndSlashingHash() common.Hash {
	type Data struct {
		Config        BeaconCommitteeStateV4Config
		BeaconWaiting []*StakerInfo
		BeaconLocking []*LockingInfo
	}
	beaconWaiting := s.GetBeaconWaiting()
	beaconLocking := s.GetBeaconLocking()
	data := Data{
		Config:        s.config,
		BeaconWaiting: []*StakerInfo{},
		BeaconLocking: []*LockingInfo{},
	}
	for _, v := range beaconWaiting {
		vStr, _ := v.ToBase58()
		stakerInfo := s.getBeaconStakerInfo(vStr)
		if stakerInfo == nil {
			panic(fmt.Sprintf("Can not get staker info of this pk %v", vStr))
		}
		data.BeaconWaiting = append(data.BeaconWaiting, stakerInfo)
	}
	for _, v := range beaconLocking {
		vStr, _ := v.ToBase58()
		stakerLockingInfo, ok := s.beaconLocking[vStr]
		if !ok {
			panic(fmt.Sprintf("Can not get staker locking info of this pk %v", vStr))
		}
		data.BeaconLocking = append(data.BeaconLocking, stakerLockingInfo)
	}
	b, _ := json.Marshal(data)
	hash := common.HashH(b)
	return hash
}

func (s *BeaconCommitteeStateV4) updateBeaconPerformance(previousData string) error {
	if previousData != "" {
		prevValidationData, err := consensustypes.DecodeValidationData(previousData)
		if err != nil {
			return fmt.Errorf("Cannot decode previous validation data")
		}
		//log.Println("ProcessUpdateBeaconPerformance ", prevValidationData.ValidatiorsIdx)
		beaconCommittee := s.GetBeaconCommittee()
		for index, cpk := range beaconCommittee {
			if common.IndexOfInt(index, prevValidationData.ValidatiorsIdx) == -1 {
				cpkStr, _ := cpk.ToBase58()
				stakerInfo := s.beaconCommittee[cpkStr]
				stakerInfo.Performance = (stakerInfo.Performance * s.config.DECREASE_PERFORMING) / s.config.MAX_SCORE
				if stakerInfo.Performance < s.config.MIN_SCORE {
					stakerInfo.Performance = s.config.MIN_SCORE
				}
			} else {
				cpkStr, _ := cpk.ToBase58()
				stakerInfo := s.beaconCommittee[cpkStr]
				stakerInfo.Performance = (stakerInfo.Performance * s.config.INCREASE_PERFORMING) / s.config.MAX_SCORE
				if stakerInfo.Performance > s.config.MAX_SCORE {
					stakerInfo.Performance = s.config.MAX_SCORE
				}
			}
		}
	}
	return nil
}

func (s *BeaconCommitteeStateV4) ProcessUpdateBeaconPerformance(env ProcessContext) ([][]string, error) {
	if firstBlockEpoch(env.BeaconHeight) {
		return nil, nil
	}
	return nil, s.updateBeaconPerformance(env.BeaconHeader.PreviousValidationData)
}

// Process shard active time
func (s *BeaconCommitteeStateV4) ProcessCountShardActiveTime(env ProcessContext) ([][]string, error) {
	if !firstBlockEpoch(env.BeaconHeight) { //Review Using BeaconCommitteeStateEnvironment?
		return nil, nil
	}

	for cpkStr, stakerInfo := range s.beaconWaiting {
		staker, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpkStr)
		if !exist {
			return nil, fmt.Errorf("Cannot find cpk %v", cpkStr)
		}

		if sig, ok := env.MissingSignature[cpkStr]; ok && sig.ActualTotal != 0 {
			//update shard active time
			activeTimes := staker.ShardActiveTime()
			if (sig.Missing*100)/sig.ActualTotal > 20 {
				activeTimes -= 2
			} else {
				activeTimes++
			}
			if activeTimes < 0 {
				activeTimes = 0
			}
			//if this pubkey is slashed in this block
			if _, ok := env.MissingSignaturePenalty[cpkStr]; ok {
				activeTimes = 0
			}
			s.setShardActiveTime(cpkStr, activeTimes)

			shardStakerInfo, exists, _ := statedb.GetShardStakerInfo(s.stateDB, cpkStr)
			if exists && staker.ShardActiveTime() >= s.config.MIN_ACTIVE_SHARD {
				shardStakerInfo.SetAutoStaking(false)
				err := statedb.StoreStakerInfoV2(s.stateDB, stakerInfo.cpkStruct, shardStakerInfo)
				if err != nil {
					return nil, err
				}
			}
		}

	}
	return nil, nil
}

// Process slash, unstake and swap
func (s *BeaconCommitteeStateV4) ProcessBeaconSwapAndSlash(env ProcessContext) ([][]string, error) {
	if !lastBlockEpoch(env.BeaconHeight) {
		return nil, nil
	}

	slashCpk := make(map[string]uint64)
	unstakeCpk := make(map[string]uint64)
	var err error

	//snapshot performance, delegator list of current committee and pending
	//performance: using current performance score; delegator: using the delegators register at beginning epoch
	committeeData := statedb.GetCommitteeData(s.stateDB)
	lastEpochCommitteeData := map[string]statedb.LastEpochCommitteeAndPendingInfo{}
	for cpkStr, _ := range s.beaconCommittee {
		stakeHash, err := s.GetBeaconCandidateUID(cpkStr)
		if err != nil {
			return nil, err
		}
		lastEpochCommitteeData[cpkStr] = statedb.LastEpochCommitteeAndPendingInfo{s.beaconCommittee[cpkStr].Performance, committeeData.BeginEpochInfo[cpkStr].StakeAmount, committeeData.BeginEpochInfo[cpkStr].Delegators, stakeHash}
	}
	for cpkStr, _ := range s.beaconPending {
		stakeHash, err := s.GetBeaconCandidateUID(cpkStr)
		if err != nil {
			return nil, err
		}
		lastEpochCommitteeData[cpkStr] = statedb.LastEpochCommitteeAndPendingInfo{s.beaconPending[cpkStr].Performance, committeeData.BeginEpochInfo[cpkStr].StakeAmount, committeeData.BeginEpochInfo[cpkStr].Delegators, stakeHash}
	}

	//slash
	for cpk, stakerInfo := range s.beaconCommittee {
		if stakerInfo.Performance <= s.config.MIN_PERFORMANCE && !stakerInfo.FixedNode {
			slashCpk[cpk] = env.Epoch + s.getTotalLockingEpoch(stakerInfo.Performance)
		}
	}
	for cpk, unlockEpoch := range slashCpk {
		if err = s.setLocking(cpk, env.Epoch, unlockEpoch, statedb.BY_SLASH); err != nil {
			return nil, err
		}
		if err = s.removeFromPool(COMMITTEE_POOL, cpk); err != nil {
			return nil, err
		}
	}

	//unstake
	for cpk, stakerInfo := range s.beaconCommittee {
		if stakerInfo.Unstake && !stakerInfo.FixedNode {
			unstakeCpk[cpk] = env.Epoch + s.getTotalLockingEpoch(stakerInfo.Performance)
		}
	}

	for cpk, stakerInfo := range s.beaconPending {
		if stakerInfo.Unstake && !stakerInfo.FixedNode {
			unstakeCpk[cpk] = env.Epoch + s.getTotalLockingEpoch(stakerInfo.Performance)
		}
	}
	for cpk, stakerInfo := range s.beaconWaiting {
		if stakerInfo.Unstake && !stakerInfo.FixedNode {
			unstakeCpk[cpk] = env.Epoch + s.getTotalLockingEpoch(stakerInfo.Performance)
		}
	}
	for cpk, unlockEpoch := range unstakeCpk {
		if err = s.setLocking(cpk, env.Epoch, unlockEpoch, statedb.BY_UNSTAKE); err != nil {
			return nil, err
		}
		if err = s.removeFromPool(COMMITTEE_POOL, cpk); err != nil {
			return nil, err
		}
		if err = s.removeFromPool(PENDING_POOL, cpk); err != nil {
			return nil, err
		}
		if err = s.removeFromPool(WAITING_POOL, cpk); err != nil {
			return nil, err
		}
	}

	//check version to swap here
	newBeaconCommittee, newBeaconPending := s.beacon_swap_v1(env)
	//update new beacon committee/pending
	//update statedb/memdb

	for _, cpk := range newBeaconPending {
		k, _ := cpk.ToBase58()
		if stakerInfo, ok := s.beaconPending[k]; !ok {
			stakerInfo = s.getBeaconStakerInfo(k)
			stakerInfo.EpochScore = s.config.DEFAULT_PERFORMING * stakerInfo.getScore()
			stakerInfo.Performance = s.config.DEFAULT_PERFORMING
			log.Println("disable finish sync", false, k)
			s.setFinishSync(k, false)

			if err = s.addToPool(PENDING_POOL, k, stakerInfo); err != nil {
				return nil, err
			}
			if err = s.removeFromPool(COMMITTEE_POOL, k); err != nil {
				return nil, err
			}
			if err = s.removeFromPool(WAITING_POOL, k); err != nil {
				return nil, err
			}
		} else {
			stakerInfo.EpochScore = s.config.DEFAULT_PERFORMING * stakerInfo.getScore()
			stakerInfo.Performance = s.config.DEFAULT_PERFORMING
		}
	}

	for _, cpk := range newBeaconCommittee {
		k, _ := cpk.ToBase58()
		if stakerInfo, ok := s.beaconCommittee[k]; !ok { //new committee
			stakerInfo = s.getBeaconStakerInfo(k)
			stakerInfo.EpochScore = s.config.DEFAULT_PERFORMING * stakerInfo.getScore()
			stakerInfo.Performance = s.config.DEFAULT_PERFORMING
			if err = s.addToPool(COMMITTEE_POOL, k, stakerInfo); err != nil {
				return nil, err
			}
			if err = s.removeFromPool(PENDING_POOL, k); err != nil {
				return nil, err
			}
		} else { // old committee
			stakerInfo.EpochScore = stakerInfo.Performance * stakerInfo.getScore()
		}
	}

	//store committee data (epoch score)
	beaconCommitteeList := s.GetBeaconCommittee()
	beaconPendingList := s.GetBeaconSubstitute()

	currentEpochCommitteeData := map[string]statedb.CurrentEpochCommitteeAndPendingInfo{}
	for _, cpk := range beaconCommitteeList {
		cpkStr, _ := cpk.ToBase58()
		if s.beaconCommittee[cpkStr].stakeID == "" {
			panic(1)
		}
		currentEpochCommitteeData[cpkStr] = statedb.CurrentEpochCommitteeAndPendingInfo{s.beaconCommittee[cpkStr].EpochScore, s.beaconCommittee[cpkStr].Performance, s.beaconCommittee[cpkStr].StakingAmount, s.beaconCommittee[cpkStr].TotalDelegators, s.beaconCommittee[cpkStr].stakeID}
	}
	for _, cpk := range beaconPendingList {
		cpkStr, _ := cpk.ToBase58()
		if s.beaconPending[cpkStr].stakeID == "" {
			panic(1)
		}
		currentEpochCommitteeData[cpkStr] = statedb.CurrentEpochCommitteeAndPendingInfo{s.beaconPending[cpkStr].EpochScore, s.beaconPending[cpkStr].Performance, s.beaconPending[cpkStr].StakingAmount, s.beaconPending[cpkStr].TotalDelegators, s.beaconPending[cpkStr].stakeID}
	}
	err = statedb.StoreCommitteeData(s.stateDB, &statedb.CommitteeData{BeginEpochInfo: currentEpochCommitteeData, LastEpoch: lastEpochCommitteeData})
	if err != nil {
		Logger.log.Errorf("Cannot store committee data %+v", err)
		return nil, err
	}
	return nil, nil
}

func (s *BeaconCommitteeStateV4) ProcessBeaconFinishSyncInstruction(env ProcessContext) ([][]string, error) {
	for _, inst := range env.BeaconInstructions {
		if inst[0] == instruction.FINISH_SYNC_ACTION && inst[1] == "-1" {
			finishSyncInst, err := instruction.ImportFinishSyncInstructionFromString(inst)
			if err != nil {
				return nil, err
			}
			for _, cpk := range finishSyncInst.PublicKeys {
				log.Println("set finish sync", cpk)
				if err = s.setFinishSync(cpk, true); err != nil {
					return nil, err
				}
			}
		}
	}
	return nil, nil
}

// Process assign beacon pending (sync, sync valid time)
func (s *BeaconCommitteeStateV4) ProcessBeaconWaitingCondition(env ProcessContext) ([][]string, error) {

	for _, k := range s.GetBeaconWaiting() {
		cpk, _ := k.ToBase58()
		stakerInfo := s.getBeaconStakerInfo(cpk)
		//Check 1: waiting -> unstake
		//if this staker not have valid active time, and not stake shard any more -> unstake beacon
		staker, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return nil, fmt.Errorf("Cannot find stakerInfo %v", cpk)
		}

		_, shardExist, _ := statedb.GetShardStakerInfo(s.stateDB, cpk)
		if !shardExist && staker.ShardActiveTime() < s.config.MIN_ACTIVE_SHARD {
			s.setUnstake(cpk)
			continue
		}

		//Check 2: waiting -> pending
		//if finish sync & enough valid time & shard staker is unstaked -> update role to pending
		log.Printf("ProcessBeaconWaitingCondition %v %v %v %v %v %+v", staker.FinishSync(), staker.ShardActiveTime(), staker.BeaconConfirmTime(), s.config.MIN_ACTIVE_SHARD, shardExist, staker.ToString())
		if env.BeaconHeader.Timestamp < staker.BeaconConfirmTime()+s.config.MIN_WAITING_PERIOD || staker.ShardActiveTime() < s.config.MIN_ACTIVE_SHARD {
			continue
		}

		if staker.FinishSync() && !shardExist {
			if err := s.addToPool(PENDING_POOL, cpk, stakerInfo); err != nil {
				return nil, err
			}
			if err := s.removeFromPool(WAITING_POOL, cpk); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

// Process stake instruction
// -> update waiting
// -> store beacon staker info
func (s *BeaconCommitteeStateV4) ProcessBeaconStakeInstruction(env ProcessContext) ([][]string, error) {
	returnStakingList := [][]string{}
	return_cpk := []string{}
	return_amount := []uint64{}
	return_reason := []int{}

	for _, inst := range env.BeaconInstructions {
		if inst[0] == instruction.BEACON_STAKE_ACTION {
			beaconStakeInst := instruction.ImportBeaconStakeInstructionFromString(inst)
			for i, txHash := range beaconStakeInst.TxStakeHashes {
				//return staking if already exist
				_, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, beaconStakeInst.PublicKeys[i])
				if exist {
					return_cpk = append(return_cpk, beaconStakeInst.PublicKeys[i])
					return_amount = append(return_amount, beaconStakeInst.StakingAmount[i])
					return_reason = append(return_reason, statedb.BY_DUPLICATE_STAKE)
					continue
				}
				var key incognitokey.CommitteePublicKey
				if err := key.FromString(beaconStakeInst.PublicKeys[i]); err != nil {
					return nil, err
				}
				stakingInfo := map[common.Hash]statedb.StakingTxInfo{}
				stakingInfo[txHash] = statedb.StakingTxInfo{beaconStakeInst.StakingAmount[i], env.BeaconHeight}
				info := statedb.NewBeaconStakerInfoWithValue(beaconStakeInst.FunderAddressStructs[i], beaconStakeInst.RewardReceiverStructs[i],
					env.BeaconHeight, env.BeaconHeader.Timestamp, time.Now().UnixNano(), stakingInfo)
				if err := statedb.StoreBeaconStakerInfo(s.stateDB, key, info); err != nil {
					return nil, err
				}
				stakeID := common.HashH([]byte(fmt.Sprintf("%v-%v", beaconStakeInst.PublicKeys[i], env.BeaconHeader.Timestamp)))
				newStakerInfo := &StakerInfo{key, beaconStakeInst.PublicKeys[i], beaconStakeInst.StakingAmount[i],
					false, 500, 0, false, time.Now().UnixNano(),
					false, 0, stakeID.String(), 0}
				statedb.StoreBeaconSharePrice(s.stateDB, stakeID.String(), INIT_SHARE_PRICE)
				if err := s.addToPool(WAITING_POOL, beaconStakeInst.PublicKeys[i], newStakerInfo); err != nil {
					return nil, err
				}
			}
		}
	}

	if len(return_cpk) == 0 {
		return nil, nil
	}
	returnStakingList = append(returnStakingList, instruction.NewReturnBeaconStakeInsWithValue(return_cpk, return_reason, return_amount).ToString())

	return returnStakingList, nil
}

// Process add stake amount
func (s *BeaconCommitteeStateV4) ProcessBeaconAddStakingAmountInstruction(env ProcessContext) ([][]string, error) {
	for _, inst := range env.BeaconInstructions {
		if inst[0] == instruction.ADD_STAKING_ACTION {
			addStakeInst := instruction.ImportAddStakingInstructionFromString(inst)
			for i, cpk := range addStakeInst.CommitteePublicKeys {
				stakingTxHash, err := common.Hash{}.NewHashFromStr(addStakeInst.StakingTxIDs[i])
				if err != nil {
					return nil, fmt.Errorf("Cannot convert staing tx hash, %v", addStakeInst.StakingTxIDs[i])
				}
				err = s.addStakingTx(cpk, *stakingTxHash, addStakeInst.StakingAmount[i], env.BeaconHeight)
				if err != nil {
					err = fmt.Errorf("Add Staking tx error, %v", err.Error())
					Logger.log.Error(err)
				}
			}
		}
	}
	return nil, nil
}

// unstaking instruction -> set unstake

func (s *BeaconCommitteeStateV4) ProcessBeaconUnstakeInstruction(env ProcessContext) ([][]string, error) {
	for _, inst := range env.BeaconInstructions {

		unstakeCPKs := []string{}
		if inst[0] == instruction.STOP_AUTO_STAKE_ACTION {

			unstakeInst, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				return nil, err
			}
			unstakeCPKs = unstakeInst.CommitteePublicKeys
		}
		if inst[0] == instruction.UNSTAKE_ACTION {

			unstakeInst, err := instruction.ValidateAndImportUnstakeInstructionFromString(inst)
			if err != nil {
				return nil, err
			}
			unstakeCPKs = unstakeInst.CommitteePublicKeys
		}

		for _, cpk := range unstakeCPKs {
			if stakerInfo := s.getBeaconStakerInfo(cpk); stakerInfo != nil {
				if err := s.setUnstake(cpk); err != nil {
					return nil, err
				}
			}
		}
	}
	return nil, nil
}

// Process return staking amount (unlocking)
func (s *BeaconCommitteeStateV4) ProcessBeaconUnlocking(env ProcessContext) ([][]string, error) {
	if !lastBlockEpoch(env.BeaconHeight) {
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
		// if env.Epoch >= info.LockingEpoch()+s.config.LOCKING_PERIOD {
		if env.Epoch >= info.UnlockingEpoch() {
			return_cpk = append(return_cpk, cpk)
			switch lockingInfo.LockingReason {
			case statedb.BY_SLASH:
				return_reason = append(return_reason, statedb.BY_SLASH)
			case statedb.BY_UNSTAKE:
				return_reason = append(return_reason, statedb.BY_UNSTAKE)
			}
			return_amount = append(return_amount, info.TotalStakingAmount())
			if err := s.removeFromPool(LOCKING_POOL, cpk); err != nil {
				return nil, err
			}
			if err := statedb.DeleteBeaconStakerInfo(s.stateDB, []incognitokey.CommitteePublicKey{lockingInfo.cpkStr}); err != nil {
				return nil, err
			}
		}
	}
	if len(return_cpk) == 0 {
		return nil, nil
	}
	returnStakingInstList = append(returnStakingInstList, instruction.NewReturnBeaconStakeInsWithValue(return_cpk, return_reason, return_amount).ToString())
	return returnStakingInstList, nil
}

func (s BeaconCommitteeStateV4) GetAllStaker() (map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey) {
	sC := make(map[byte][]incognitokey.CommitteePublicKey)
	sPV := make(map[byte][]incognitokey.CommitteePublicKey)
	sSP := make(map[byte][]incognitokey.CommitteePublicKey)
	for shardID, committee := range s.GetShardCommittee() {
		sC[shardID] = append([]incognitokey.CommitteePublicKey{}, committee...)
	}
	for shardID, Substitute := range s.GetShardSubstitute() {
		sPV[shardID] = append([]incognitokey.CommitteePublicKey{}, Substitute...)
	}
	for shardID, syncValidator := range s.GetSyncingValidators() {
		sSP[shardID] = append([]incognitokey.CommitteePublicKey{}, syncValidator...)
	}
	bC := s.GetBeaconCommittee()
	bPV := s.GetBeaconSubstitute()
	bW := s.GetBeaconWaiting()
	bL := s.GetBeaconLocking()
	cSWFCR := s.GetCandidateShardWaitingForCurrentRandom()
	cSWFNR := s.GetCandidateShardWaitingForNextRandom()
	return sC, sPV, sSP, bC, bPV, bW, bL, cSWFCR, cSWFNR
}

func (s *BeaconCommitteeStateV4) Version() int {
	return STAKING_FLOW_V4
}

// Review Should we use get/set for newState.stateDB?
func (s *BeaconCommitteeStateV4) Clone(cloneState *statedb.StateDB) BeaconCommitteeState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	newState := NewBeaconCommitteeStateV4()
	newState.stateDB = cloneState
	newState.BeaconCommitteeStateV3 = s.BeaconCommitteeStateV3.clone()
	for k, v := range s.beaconCommittee {
		infoClone := *v
		newState.beaconCommittee[k] = &infoClone
	}
	for k, v := range s.beaconPending {
		infoClone := *v
		newState.beaconPending[k] = &infoClone
	}
	for k, v := range s.beaconWaiting {
		infoClone := *v
		newState.beaconWaiting[k] = &infoClone
	}
	for k, v := range s.beaconLocking {
		infoClone := *v
		newState.beaconLocking[k] = &infoClone
	}
	newState.config = s.config
	return newState
}

func GetKeyStructListFromMapStaker(list map[string]*StakerInfo) []incognitokey.CommitteePublicKey {
	fixNode := []*StakerInfo{}
	keys := []*StakerInfo{}
	for _, staker := range list {
		if staker.FixedNode {
			fixNode = append(fixNode, staker)
		} else {
			keys = append(keys, staker)
		}
	}
	sort.Slice(fixNode, func(i, j int) bool {
		return fixNode[i].enterTime < fixNode[j].enterTime
	})
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].EpochScore == keys[j].EpochScore {
			return keys[i].enterTime < keys[j].enterTime
		}
		return keys[i].EpochScore > keys[j].EpochScore
	})
	res := make([]incognitokey.CommitteePublicKey, len(keys)+len(fixNode))
	for i, v := range fixNode {
		res[i] = v.cpkStruct
	}
	for i, v := range keys {
		res[len(fixNode)+i] = v.cpkStruct
	}
	return res
}

func GetKeyStructListFromMapLocking(list map[string]*LockingInfo) []incognitokey.CommitteePublicKey {
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

func firstBlockEpoch(h uint64) bool {
	return h%config.Param().EpochParam.NumberOfBlockInEpoch == 1
}
func lastBlockEpoch(h uint64) bool {
	return h%config.Param().EpochParam.NumberOfBlockInEpoch == 0
}

func (s BeaconCommitteeStateV4) IsFinishSync(key string) bool {
	info, exists, _ := statedb.GetBeaconStakerInfo(s.stateDB, key)
	if exists && info.FinishSync() {
		return true
	}
	return false
}

func (s *BeaconCommitteeStateV4) GetAllCandidateSubstituteCommittee() []string {
	stateV3Res := s.BeaconCommitteeStateV3.getAllCandidateSubstituteCommittee()
	for cpk, _ := range s.beaconCommittee {
		stateV3Res = append(stateV3Res, cpk)
	}
	for cpk, _ := range s.beaconWaiting {
		stateV3Res = append(stateV3Res, cpk)
	}
	for cpk, _ := range s.beaconPending {
		stateV3Res = append(stateV3Res, cpk)
	}
	for cpk, _ := range s.beaconLocking {
		stateV3Res = append(stateV3Res, cpk)
	}
	return stateV3Res
}
func (s BeaconCommitteeStateV4) getTotalLockingEpoch(perf uint64) uint64 {
	return s.config.LOCKING_PERIOD + uint64(s.config.LOCKING_FACTOR*float64(s.config.MAX_SCORE-perf)/float64(s.config.MIN_SCORE))
}

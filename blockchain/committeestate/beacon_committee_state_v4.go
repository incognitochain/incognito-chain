package committeestate

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"log"
	"reflect"
	"runtime"
	"sort"
	"time"
)

type BeaconCommitteeStateV4Config struct {
	DEFAULT_PERFORMING  uint64
	INCREASE_PERFORMING uint64
	DECREASE_PERFORMING uint64
	MIN_ACTIVE_SHARD    int
	MIN_WAITING_PERIOD  int64
	MIN_PERFORMANCE     uint64
	LOCKING_PERIOD      uint64
}

func NewBeaconCommitteeStateV4Config(version int) BeaconCommitteeStateV4Config {
	return BeaconCommitteeStateV4Config{
		DEFAULT_PERFORMING:  500,
		INCREASE_PERFORMING: 1015,
		DECREASE_PERFORMING: 965,
		MIN_ACTIVE_SHARD:    2,
		MIN_WAITING_PERIOD:  60, //seconds
		MIN_PERFORMANCE:     370,
		LOCKING_PERIOD:      3,
	}
}

const (
	COMMITEE_POOL = iota
	PENDING_POOL
	WAITING_POOL
	LOCKING_POOL
)

type StakerInfo struct {
	cpkStr        incognitokey.CommitteePublicKey
	StakingAmount uint64
	Unstake       bool
	Performance   uint64
	EpochScore    uint64 // -> sorted list
	FixedNode     bool
	EnterTime     time.Time
}

type StakerInfoDetail struct {
	CPK             string
	StakingAmount   uint64
	Unstake         bool
	Performance     uint64
	EpochScore      uint64 // -> sorted list
	FixedNode       bool
	FinishSync      bool
	ShardActiveTime int
}

type LockingInfo struct {
	cpkStr        incognitokey.CommitteePublicKey
	LockingEpoch  uint64
	LockingReason int
}

type LockingInfoDetail struct {
	CPK           string
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
	}
}

func (s *BeaconCommitteeStateV4) getStakerInfo(cpk string) *StakerInfo {
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
	if stakerInfo := s.getStakerInfo(cpk); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", cpk)
		}
		info.SetShardActiveTime(t)
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStr, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", cpk)
}

func (s *BeaconCommitteeStateV4) setUnstake(cpk string) error {
	if stakerInfo := s.getStakerInfo(cpk); stakerInfo != nil {
		stakerInfo.Unstake = true
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", cpk)
		}
		info.SetUnstaking()
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStr, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", cpk)
}

func (s *BeaconCommitteeStateV4) setFinishSync(cpk string) error {
	if stakerInfo := s.getStakerInfo(cpk); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", cpk)
		}
		info.SetFinishSync()
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStr, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", cpk)
}

func (s *BeaconCommitteeStateV4) addStakingTx(cpk string, tx common.Hash, amount, height uint64) error {
	if stakerInfo := s.getStakerInfo(cpk); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", cpk)
		}
		info.AddStaking(tx, height, amount)
		stakerInfo.StakingAmount = info.TotalStakingAmount()
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStr, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", cpk)
}

func (s *BeaconCommitteeStateV4) setLocking(cpk string, epoch uint64, reason int) error {
	if stakerInfo := s.getStakerInfo(cpk); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", cpk)
		}
		info.SetLocking(epoch, reason)
		s.beaconLocking[cpk] = &LockingInfo{stakerInfo.cpkStr, epoch, reason}
		if err := statedb.StoreBeaconLocking(s.stateDB, []incognitokey.CommitteePublicKey{stakerInfo.cpkStr}); err != nil {
			return err
		}
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStr, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", cpk)
}

func (s *BeaconCommitteeStateV4) removeFromPool(pool int, cpk string) error {
	cpkStruct := incognitokey.CommitteePublicKey{}
	if err := cpkStruct.FromString(cpk); err != nil {
		return err
	}
	switch pool {
	case COMMITEE_POOL:
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
	switch pool {
	case COMMITEE_POOL:
		s.beaconCommittee[cpk] = stakerInfo
		if err := statedb.StoreBeaconCommittee(s.stateDB, []incognitokey.CommitteePublicKey{stakerInfo.cpkStr}); err != nil {
			return err
		}
	case PENDING_POOL:
		s.beaconPending[cpk] = stakerInfo
		if err := statedb.StoreBeaconSubstituteValidator(s.stateDB, []incognitokey.CommitteePublicKey{stakerInfo.cpkStr}); err != nil {
			return err
		}
	case WAITING_POOL:
		s.beaconWaiting[cpk] = stakerInfo
		if err := statedb.StoreBeaconWaiting(s.stateDB, []incognitokey.CommitteePublicKey{stakerInfo.cpkStr}); err != nil {
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
	scores := map[string]uint64{}
	for _, cpk := range stateV3.GetBeaconCommittee() {
		cpkStr, _ := cpk.ToBase58()
		scores[cpkStr] = s.config.DEFAULT_PERFORMING
		info, exists, _ := statedb.GetStakerInfo(stateDB, cpkStr)
		if !exists {
			return fmt.Errorf("Cannot find cpk %v", cpk)
		}
		stakingTx := map[common.Hash]statedb.StakingTxInfo{}
		stakingTx[info.TxStakingID()] = statedb.StakingTxInfo{0, 1}
		beaconInfo := statedb.NewBeaconStakerInfoWithValue(info.RewardReceiver(), info.RewardReceiver(), 1, 1, stakingTx)
		err := statedb.StoreBeaconStakerInfo(stateDB, cpk, beaconInfo)
		if err != nil {

			return err
		}
	}
	err := statedb.StoreCommitteeData(stateDB, &statedb.CommitteeData{CommitteeScore: scores})
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

func (s *BeaconCommitteeStateV4) GetBeaconLocking() []incognitokey.CommitteePublicKey {
	return GetKeyStructListFromMapLocking(s.beaconLocking)
}

func (s BeaconCommitteeStateV4) DebugBeaconCommitteeState() string {
	type StateData struct {
		Committee []StakerInfoDetail
		Pending   []StakerInfoDetail
		Waiting   []StakerInfoDetail
		Locking   []LockingInfoDetail
	}
	data := &StateData{
		Committee: []StakerInfoDetail{},
		Pending:   []StakerInfoDetail{},
		Waiting:   []StakerInfoDetail{},
		Locking:   []LockingInfoDetail{}}

	getStakerDetail := func(cpk string) StakerInfoDetail {
		stakerStateInfo, has, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !has || stakerStateInfo == nil {
			return StakerInfoDetail{}
		}

		stakerInfo := s.getStakerInfo(cpk)
		if stakerInfo == nil {
			panic(1)
		}
		detail := StakerInfoDetail{
			cpk,
			stakerInfo.StakingAmount,
			stakerInfo.Unstake,
			stakerInfo.Performance,
			stakerInfo.EpochScore,
			stakerInfo.FixedNode,
			stakerStateInfo.FinishSync(),
			stakerStateInfo.ShardActiveTime(),
		}
		return detail
	}

	getLockingDetail := func(cpk string) LockingInfoDetail {
		stakerInfo, has, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
		if !has {
			return LockingInfoDetail{}
		}
		detail := LockingInfoDetail{
			cpk,
			(*s.beaconLocking[cpk]).LockingEpoch,
			(*s.beaconLocking[cpk]).LockingReason,
			(*s.beaconLocking[cpk]).LockingEpoch + s.config.LOCKING_PERIOD,
			stakerInfo.TotalStakingAmount(),
		}
		return detail
	}
	for _, v := range s.GetBeaconCommittee() {
		cpk, _ := v.ToBase58()
		detail := getStakerDetail(cpk)
		data.Committee = append(data.Committee, detail)
	}

	for _, v := range s.GetBeaconSubstitute() {
		cpk, _ := v.ToBase58()
		detail := getStakerDetail(cpk)
		data.Pending = append(data.Pending, detail)
	}
	for _, v := range s.GetBeaconWaiting() {
		cpk, _ := v.ToBase58()
		detail := getStakerDetail(cpk)
		data.Waiting = append(data.Waiting, detail)
	}
	for _, v := range s.GetBeaconLocking() {
		cpk, _ := v.ToBase58()
		detail := getLockingDetail(cpk)
		data.Locking = append(data.Locking, detail)
	}
	b, _ := json.Marshal(data)
	return string(b)
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
		s.beaconCommittee[cpkStr] = &StakerInfo{cpkStr: cpk, Unstake: info.Unstaking(), StakingAmount: info.TotalStakingAmount()}
		s.beaconCommittee[cpkStr].EpochScore = commiteeData.CommitteeScore[cpkStr]
		s.beaconCommittee[cpkStr].Performance = s.config.DEFAULT_PERFORMING
		s.beaconCommittee[cpkStr].EnterTime = time.Now() //the list is already sort, this use execution time to know which committee go first
		if index < minBeaconCommitteeSize {
			s.beaconCommittee[cpkStr].FixedNode = true
		}
	}
	for _, blk := range allBeaconBlock {
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
		s.beaconPending[cpkStr] = &StakerInfo{cpkStr: cpk, Unstake: info.Unstaking(),
			EpochScore:    commiteeData.CommitteeScore[cpkStr],
			StakingAmount: info.TotalStakingAmount(),
			Performance:   s.config.DEFAULT_PERFORMING,
			EnterTime:     time.Now()} //the list is already sort, this use execution time to know which committee go first
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
		s.beaconWaiting[cpkStr] = &StakerInfo{cpkStr: cpk, Unstake: info.Unstaking(),
			StakingAmount: info.TotalStakingAmount(), Performance: s.config.DEFAULT_PERFORMING,
			EnterTime: time.Now()} //the list is already sort, this use execution time to know which committee go first
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
		s.ProcessUpdateBeaconPerformance,
		s.ProcessBeaconUnstake,
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

	//udpate beacon state
	beaconValidators := append(s.GetBeaconCommittee(), s.GetBeaconSubstitute()...)
	beaconValidators = append(beaconValidators, s.GetBeaconLocking()...)

	beaconValidatorList, _ := incognitokey.CommitteeKeyListToString(beaconValidators)

	stateHash.BeaconCommitteeAndValidatorHash, _ = common.GenerateHashFromStringArray(beaconValidatorList)

	beaconCandidateList, _ := incognitokey.CommitteeKeyListToString(s.GetBeaconWaiting())
	stateHash.BeaconCandidateHash, _ = common.GenerateHashFromStringArray(beaconCandidateList)

	return stateHash, changes, instructions, nil
}

func (s *BeaconCommitteeStateV4) updateBeaconPerformance(previousData string) error {
	if previousData != "" {
		prevValidationData, err := consensustypes.DecodeValidationData(previousData)
		if err != nil {
			return fmt.Errorf("Cannot decode previous validation data")
		}
		log.Println("ProcessUpdateBeaconPerformance ", prevValidationData.ValidatiorsIdx)
		beaconCommittee := s.GetBeaconCommittee()
		for index, cpk := range beaconCommittee {
			if common.IndexOfInt(index, prevValidationData.ValidatiorsIdx) == -1 {
				cpkStr, _ := cpk.ToBase58()
				stakerInfo := s.beaconCommittee[cpkStr]
				stakerInfo.Performance = (stakerInfo.Performance * s.config.DECREASE_PERFORMING) / 1000
				if stakerInfo.Performance < 100 {
					stakerInfo.Performance = 100
				}
			} else {
				cpkStr, _ := cpk.ToBase58()
				stakerInfo := s.beaconCommittee[cpkStr]
				stakerInfo.Performance = (stakerInfo.Performance * s.config.INCREASE_PERFORMING) / 1000
				if stakerInfo.Performance > 1000 {
					stakerInfo.Performance = 1000
				}
			}
		}
	}
	return nil
}

func (s *BeaconCommitteeStateV4) ProcessUpdateBeaconPerformance(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	//reset if is first block epoch
	if firstBlockEpoch(env.BeaconHeight) {
		for _, stakerInfo := range s.beaconCommittee {
			stakerInfo.Performance = s.config.DEFAULT_PERFORMING
		}
		return nil, nil
	}

	return nil, s.updateBeaconPerformance(env.BeaconHeader.PreviousValidationData)
}

//Process shard active time
func (s *BeaconCommitteeStateV4) ProcessCountShardActiveTime(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	for cpkStr, _ := range s.beaconWaiting {
		if _, ok := env.MissingSignature[cpkStr]; ok {
			log.Println(env.BeaconHeight, cpkStr, env.MissingSignature[cpkStr])
		}
	}
	if !firstBlockEpoch(env.BeaconHeight) {
		return nil, nil
	}

	for cpkStr, stakerInfo := range s.beaconWaiting {
		staker, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpkStr)
		if !exist {
			return nil, fmt.Errorf("Cannot find cpk %v", cpkStr)
		}

		if sig, ok := env.MissingSignature[cpkStr]; ok && sig.ActualTotal != 0 {
			if sig.ActualTotal == 0 {
				log.Println(cpkStr)
				log.Println(env.MissingSignature[cpkStr])
				log.Println(env.BeaconHeight)
				panic(1)
			}
			//update shard active time
			if (sig.Missing*100)/sig.ActualTotal > 20 {
				s.setShardActiveTime(cpkStr, 0)
			} else {
				s.setShardActiveTime(cpkStr, staker.ShardActiveTime()+1)
			}
			//if this pubkey is slashed in this block
			if _, ok := env.MissingSignaturePenalty[cpkStr]; ok {
				s.setUnstake(cpkStr)
			}

			shardStakerInfo, exists, _ := statedb.GetStakerInfo(s.stateDB, cpkStr)
			if exists && staker.ShardActiveTime() >= s.config.MIN_ACTIVE_SHARD {
				shardStakerInfo.SetAutoStaking(false)
				err := statedb.StoreStakerInfoV2(s.stateDB, stakerInfo.cpkStr, shardStakerInfo)
				if err != nil {
					return nil, err
				}
			}
		}

		//if this staker not have valid active time, and not stake shard any more -> unstake beacon
		_, exists, _ := statedb.GetStakerInfo(s.stateDB, cpkStr)
		if !exists && staker.ShardActiveTime() < s.config.MIN_ACTIVE_SHARD {
			s.setUnstake(cpkStr)
		}
	}
	return nil, nil
}

//Process slash, unstake and swap
func (s *BeaconCommitteeStateV4) ProcessBeaconSwapAndSlash(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	if !lastBlockEpoch(env.BeaconHeight) {
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
		if err = s.setLocking(cpk, env.Epoch, statedb.LOCKING_BY_SLASH); err != nil {
			return nil, err
		}
		if err = s.removeFromPool(COMMITEE_POOL, cpk); err != nil {
			return nil, err
		}
		if err = s.removeFromPool(PENDING_POOL, cpk); err != nil {
			return nil, err
		}
		if err = s.removeFromPool(WAITING_POOL, cpk); err != nil {
			return nil, err
		}
	}

	//update unstake
	for cpk, _ := range unstakeCpk {
		if err = s.setLocking(cpk, env.Epoch, statedb.LOCKING_BY_UNSTAKE); err != nil {
			return nil, err
		}
		if err = s.removeFromPool(COMMITEE_POOL, cpk); err != nil {
			return nil, err
		}
		if err = s.removeFromPool(PENDING_POOL, cpk); err != nil {
			return nil, err
		}
		if err = s.removeFromPool(WAITING_POOL, cpk); err != nil {
			return nil, err
		}
	}

	//update new beacon committee/pending
	//update statedb/memdb
	for cpk, staker := range s.beaconCommittee {
		log.Println("before commmittee", cpk, staker.Performance, staker.EpochScore)
	}
	for cpk, staker := range s.beaconPending {
		log.Println("before pending", cpk, staker.Performance, staker.EpochScore)
	}

	for k, _ := range newBeaconPending {
		if stakerInfo, ok := s.beaconPending[k]; !ok {
			stakerInfo = s.getStakerInfo(k)
			stakerInfo.EpochScore = s.config.DEFAULT_PERFORMING * stakerInfo.StakingAmount
			stakerInfo.Performance = s.config.DEFAULT_PERFORMING
			if err = s.addToPool(PENDING_POOL, k, stakerInfo); err != nil {
				return nil, err
			}
			if err = s.removeFromPool(COMMITEE_POOL, k); err != nil {
				return nil, err
			}
			if err = s.removeFromPool(WAITING_POOL, k); err != nil {
				return nil, err
			}
		} else {
			stakerInfo.EpochScore = s.config.DEFAULT_PERFORMING * stakerInfo.StakingAmount
			stakerInfo.Performance = s.config.DEFAULT_PERFORMING
		}
	}

	for k, _ := range newBeaconCommittee {
		if stakerInfo, ok := s.beaconCommittee[k]; !ok { //new committee
			stakerInfo = s.getStakerInfo(k)
			stakerInfo.EpochScore = s.config.DEFAULT_PERFORMING * stakerInfo.StakingAmount
			stakerInfo.Performance = s.config.DEFAULT_PERFORMING
			if err = s.addToPool(COMMITEE_POOL, k, stakerInfo); err != nil {
				return nil, err
			}
			if err = s.removeFromPool(PENDING_POOL, k); err != nil {
				return nil, err
			}
		} else { // old committee
			stakerInfo.EpochScore = stakerInfo.Performance * stakerInfo.StakingAmount
			stakerInfo.Performance = s.config.DEFAULT_PERFORMING
		}
	}

	//store committee data (epoch score)
	beaconCommitteeList := s.GetBeaconCommittee()
	beaconPendingList := s.GetBeaconSubstitute()

	score := map[string]uint64{}
	for _, cpk := range beaconCommitteeList {
		cpkStr, _ := cpk.ToBase58()
		score[cpkStr] = s.beaconCommittee[cpkStr].EpochScore
	}
	for _, cpk := range beaconPendingList {
		cpkStr, _ := cpk.ToBase58()
		score[cpkStr] = s.beaconPending[cpkStr].EpochScore
	}
	err = statedb.StoreCommitteeData(s.stateDB, &statedb.CommitteeData{CommitteeScore: score})
	if err != nil {
		Logger.log.Errorf("Cannot store committee data %+v", err)
		return nil, err
	}
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
				if err = s.setFinishSync(cpk); err != nil {
					return nil, err
				}
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
		log.Printf("ProcessAssignBeaconPending %v %v %v %v %v %+v", staker.FinishSync(), staker.ShardActiveTime(), staker.BeaconConfirmTime(), s.config.MIN_ACTIVE_SHARD, shardExist, staker.ToString())

		if env.BeaconHeader.Timestamp < staker.BeaconConfirmTime()+s.config.MIN_WAITING_PERIOD && staker.ShardActiveTime() < s.config.MIN_ACTIVE_SHARD {
			continue
		}

		if staker.FinishSync() && !shardExist {
			if err := s.removeFromPool(WAITING_POOL, cpk); err != nil {
				return nil, err
			}
			if err := s.addToPool(PENDING_POOL, cpk, stakerInfo); err != nil {
				return nil, err
			}
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
		if inst[0] == instruction.BEACON_STAKE_ACTION {
			beaconStakeInst := instruction.ImportBeaconStakeInstructionFromString(inst)
			for i, txHash := range beaconStakeInst.TxStakeHashes {
				//return staking if already exist
				_, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, beaconStakeInst.PublicKeys[i])
				if exist {
					return_cpk = append(return_cpk, beaconStakeInst.PublicKeys[i])
					return_amount = append(return_amount, beaconStakeInst.StakingAmount[i])
					return_reason = append(return_reason, statedb.RETURN_BY_DUPLICATE_STAKE)
					continue
				}
				var key incognitokey.CommitteePublicKey
				if err := key.FromString(beaconStakeInst.PublicKeys[i]); err != nil {
					return nil, err
				}
				stakingInfo := map[common.Hash]statedb.StakingTxInfo{}
				stakingInfo[txHash] = statedb.StakingTxInfo{beaconStakeInst.StakingAmount[i], env.BeaconHeight}
				info := statedb.NewBeaconStakerInfoWithValue(beaconStakeInst.FunderAddressStructs[i], beaconStakeInst.RewardReceiverStructs[i], env.BeaconHeight, env.BeaconHeader.Timestamp, stakingInfo)
				if err := statedb.StoreBeaconStakerInfo(s.stateDB, key, info); err != nil {
					return nil, err
				}

				newStakerInfo := &StakerInfo{key, beaconStakeInst.StakingAmount[i], false, 500, 0, false, time.Now()}
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
				stakingTxHash, err := common.Hash{}.NewHashFromStr(addStakeInst.StakingTxIDs[i])
				if err != nil {
					return nil, fmt.Errorf("Cannot convert staing tx hash, %v", addStakeInst.StakingTxIDs[i])
				}
				err = s.addStakingTx(cpk, *stakingTxHash, addStakeInst.StakingAmount[i], env.BeaconHeight)
				if err != nil {
					Logger.log.Error(err)
					return_cpk = append(return_cpk, cpk)
					return_reason = append(return_reason, statedb.RETURN_BY_ADDSTAKE_FAIL)
					return_amount = append(return_amount, addStakeInst.StakingAmount[i])
					continue
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

//unstaking instruction -> set unstake
func (s *BeaconCommitteeStateV4) ProcessBeaconUnstake(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
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
			if err := s.setUnstake(cpk); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

//Process return staking amount (unlocking)
func (s *BeaconCommitteeStateV4) ProcessBeaconUnlocking(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
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
		if env.Epoch >= info.LockingEpoch()+s.config.LOCKING_PERIOD {
			return_cpk = append(return_cpk, cpk)
			switch lockingInfo.LockingReason {
			case statedb.LOCKING_BY_SLASH:
				return_reason = append(return_reason, statedb.RETURN_BY_SLASH)
			case statedb.LOCKING_BY_UNSTAKE:
				return_reason = append(return_reason, statedb.RETURN_BY_UNSTAKE)
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

func (s *BeaconCommitteeStateV4) beacon_swap_v1(env *BeaconCommitteeStateEnvironment) (
	map[string]bool, map[string]bool,
	map[string]incognitokey.CommitteePublicKey, map[string]incognitokey.CommitteePublicKey,
	error) {

	//slash
	slashCpk := map[string]bool{}
	for cpk, stakerInfo := range s.beaconCommittee {
		if stakerInfo.Performance < s.config.MIN_PERFORMANCE && !stakerInfo.FixedNode {
			slashCpk[cpk] = true
		}
	}

	//unstake
	unstakeCpk := map[string]bool{}
	for cpk, stakerInfo := range s.beaconCommittee {
		if stakerInfo.Unstake && !stakerInfo.FixedNode {
			unstakeCpk[cpk] = true
		}
	}
	for cpk, stakerInfo := range s.beaconPending {
		if stakerInfo.Unstake && !stakerInfo.FixedNode {
			unstakeCpk[cpk] = true
		}
	}
	for cpk, stakerInfo := range s.beaconWaiting {
		if stakerInfo.Unstake && !stakerInfo.FixedNode {
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
			score := s.config.DEFAULT_PERFORMING * stakerInfo.StakingAmount
			candidateList = append(candidateList, CandidateInfo{stakerInfo.cpkStr, cpk, score})
		}
	}

	fixNode := []CandidateInfo{}
	for cpk, stakerInfo := range s.beaconCommittee {
		if !slashCpk[cpk] && !unstakeCpk[cpk] {
			score := stakerInfo.Performance * stakerInfo.StakingAmount
			if !stakerInfo.FixedNode {
				candidateList = append(candidateList, CandidateInfo{stakerInfo.cpkStr, cpk, score})
			} else {
				fixNode = append(fixNode, CandidateInfo{stakerInfo.cpkStr, cpk, score})
			}
		}
	}

	//sort candidate list
	sort.Slice(candidateList, func(i, j int) bool {
		return candidateList[i].score > candidateList[j].score
	})
	newBeaconCommittee := map[string]incognitokey.CommitteePublicKey{}
	newBeaconPending := map[string]incognitokey.CommitteePublicKey{}
	//fixed node
	for i, _ := range fixNode {
		newBeaconCommittee[fixNode[i].cpkStr] = fixNode[i].cpk
	}

	//other candidate

	for i := 0; i < env.MaxBeaconCommitteeSize-len(fixNode) && i < len(candidateList); i++ {
		log.Println("candidateList", i, len(candidateList))
		newBeaconCommittee[candidateList[i].cpkStr] = candidateList[i].cpk
	}
	for i := env.MaxBeaconCommitteeSize - len(fixNode); i < len(candidateList); i++ {
		newBeaconPending[candidateList[i].cpkStr] = candidateList[i].cpk
	}
	log.Println("fixNode", len(fixNode))
	log.Println("newBeaconCommittee", len(newBeaconCommittee))
	log.Println("newBeaconPending", len(newBeaconPending))
	return slashCpk, unstakeCpk, newBeaconCommittee, newBeaconPending, nil
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
		return fixNode[i].EnterTime.UnixNano() < fixNode[j].EnterTime.UnixNano()
	})
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].EpochScore == keys[j].EpochScore {
			return keys[i].EnterTime.UnixNano() < keys[j].EnterTime.UnixNano()
		}
		return keys[i].EpochScore > keys[j].EpochScore
	})
	res := make([]incognitokey.CommitteePublicKey, len(keys)+len(fixNode))
	for i, v := range fixNode {
		res[i] = v.cpkStr
	}
	for i, v := range keys {
		res[len(fixNode)+i] = v.cpkStr
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

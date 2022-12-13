package committeestate

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type AssignEnvironment struct {
	ConsensusStateDB   *statedb.StateDB
	delegateState      map[string]*BeaconDelegatorInfo
	shardCommittee     map[byte][]string
	shardSubstitute    map[byte][]string
	shardNewCandidates []string
}

type AssignRuleProcessor interface {
	Process(candidates []string, numberOfValidators []int, randomNumber int64) map[byte][]string
	ProcessBeacon(candidates []string, env *AssignEnvironment) ([]string, []string, []string)
	Version() int
}

// VersionByBeaconHeight get version of committee engine by beaconHeight and config of blockchain
func GetAssignRuleVersion(beaconHeight, assignRuleV2, assignRuleV3 uint64) AssignRuleProcessor {
	if beaconHeight < assignRuleV2 && beaconHeight < assignRuleV3 {
		Logger.log.Infof("Beacon Height %+v, using Assign Rule V1 (Nil Assign Rule)", beaconHeight)
		return NewNilAssignRule()
	}
	if beaconHeight >= assignRuleV3 {
		Logger.log.Infof("Beacon Height %+v, using Assign Rule V3", beaconHeight)
		return NewAssignRuleV3()
	}

	Logger.log.Infof("Beacon Height %+v, using Assign Rule V2", beaconHeight)

	if beaconHeight >= assignRuleV2 {
		return NewAssignRuleV2()
	}

	return NewAssignRuleV2()
}

type NilAssignRule struct {
}

func NewNilAssignRule() *NilAssignRule {
	return &NilAssignRule{}
}

func (a NilAssignRule) Process(candidates []string, numberOfValidators []int, randomNumber int64) map[byte][]string {
	panic("implement me")
}

func (a NilAssignRule) Version() int {
	panic("implement me")
}

func (a NilAssignRule) ProcessBeacon(candidates []string, env *AssignEnvironment) ([]string, []string, []string) {
	panic("implement me")
}

type AssignRuleV2 struct {
}

func NewAssignRuleV2() *AssignRuleV2 {
	return &AssignRuleV2{}
}

func (AssignRuleV2) Version() int {
	return ASSIGN_RULE_V2
}

// Process assign unassignedCommonPool into shard pool with random number
func (AssignRuleV2) Process(candidates []string, numberOfValidators []int, rand int64) map[byte][]string {
	total := 0
	for _, v := range numberOfValidators {
		total += v
	}
	n := byte(len(numberOfValidators))
	sortedShardIDs := sortShardIDByIncreaseOrder(numberOfValidators)
	m := getShardIDPositionFromArray(sortedShardIDs)
	assignedCandidates := make(map[byte][]string)
	candidateRandomShardID := make(map[string]byte)
	for _, candidate := range candidates {
		randomPosition := calculateCandidatePosition(candidate, rand, total)
		shardID := 0
		tempPosition := numberOfValidators[shardID]
		for randomPosition > tempPosition {
			shardID++
			tempPosition += numberOfValidators[shardID]
		}
		candidateRandomShardID[candidate] = byte(shardID)
	}
	for _, candidate := range candidates {
		randomShardID := candidateRandomShardID[candidate]
		assignShardID := sortedShardIDs[n-1-m[randomShardID]]
		assignedCandidates[byte(assignShardID)] = append(assignedCandidates[byte(assignShardID)], candidate)
	}
	return assignedCandidates
}

func (a AssignRuleV2) ProcessBeacon(candidates []string, env *AssignEnvironment) ([]string, []string, []string) {
	panic("implement me")
}

// calculateCandidatePosition calculate reverse shardID for candidate
// randomPosition = sum(hash(candidate+rand)) % total, if randomPosition == 0 then randomPosition = 1
// randomPosition in range (1, total)
func calculateCandidatePosition(candidate string, randomNumber int64, total int) (pos int) {
	rand.Seed(randomNumber)
	seed := candidate + fmt.Sprintf("%v", randomNumber)
	hash := common.HashB([]byte(seed))
	data := 0
	for _, v := range hash {
		data += int(v)
	}
	pos = data % total
	if pos == 0 {
		pos = 1
	}
	return pos
}

// sortShardIDByIncreaseOrder take an array and sort array, return sorted index of array
func sortShardIDByIncreaseOrder(arr []int) []byte {
	sortedIndex := []byte{}
	tempArr := []struct {
		shardID byte
		value   int
	}{}
	for i, v := range arr {
		tempArr = append(tempArr, struct {
			shardID byte
			value   int
		}{byte(i), v})
	}
	sort.Slice(tempArr, func(i, j int) bool {
		return tempArr[i].value < tempArr[j].value
	})
	for _, v := range tempArr {
		sortedIndex = append(sortedIndex, v.shardID)
	}
	return sortedIndex
}

func getShardIDPositionFromArray(arr []byte) map[byte]byte {
	m := make(map[byte]byte)
	for i, v := range arr {
		m[v] = byte(i)
	}
	return m
}

type FilterBeaconRule func(candidates []string, env *AssignEnvironment) (accepted, rejected []int)

type AssignRuleV3 struct {
	filterBeaconRules []FilterBeaconRule
}

func NewAssignRuleV3() *AssignRuleV3 {
	return &AssignRuleV3{
		filterBeaconRules: []FilterBeaconRule{hasEnoughActiveTimes, notInShardCycle, hasEnoughDelegator},
	}
}

func (AssignRuleV3) Version() int {
	return ASSIGN_RULE_V3
}

func (AssignRuleV3) Process(candidates []string, numberOfValidators []int, randomNumber int64) map[byte][]string {

	sum := 0
	for _, v := range numberOfValidators {
		sum += v
	}

	totalShard := len(numberOfValidators)
	tempMean := float64(sum) / float64(totalShard)
	mean := int(tempMean)
	if tempMean > float64(mean) {
		mean += 1
	}

	lowerSet := getOrderedLowerSet(mean, numberOfValidators)

	diff := []int{}
	totalDiff := 0
	for _, shardID := range lowerSet {
		shardDiff := mean - numberOfValidators[shardID]

		// special case: mean == numberOfValidators[shardID] ||
		// shard committee size is equal among all shard ||
		// len(numberOfValidators) == 1
		if shardDiff == 0 {
			shardDiff = 1
		}

		diff = append(diff, shardDiff)
		totalDiff += shardDiff
	}

	assignedCandidates := make(map[byte][]string)
	rand.Seed(randomNumber)
	for _, candidate := range candidates {
		randomPosition := calculateCandidatePositionV2(totalDiff)
		position := 0
		tempPosition := diff[position]
		for randomPosition >= tempPosition && position < len(diff)-1 {
			position++
			tempPosition += diff[position]
		}
		shardID := lowerSet[position]
		assignedCandidates[byte(shardID)] = append(assignedCandidates[byte(shardID)], candidate)
	}

	return assignedCandidates
}

func (a AssignRuleV3) ProcessBeacon(candidates []string, env *AssignEnvironment) (
	toPending []string,
	stayWaiting []string,
	shouldRemove []string,
) {
	var (
		accepted []int
		rejected []int
	)
	processList := candidates
	notInShard, stillInShard := notInShardCycle(processList, env)
	Logger.log.Infof("Filter beacon candidate by rule notInShardCycle, accepted idx %+v, rejected idx %+v", notInShard, stillInShard)
	for _, idx := range stillInShard {
		stayWaiting = append(stayWaiting, processList[idx])
	}
	processListTmp := []string{}
	for _, idx := range notInShard {
		processListTmp = append(processListTmp, processList[idx])
	}
	processList = processListTmp
	enoughTimes, notEnoughTimes := hasEnoughActiveTimes(processList, env)
	Logger.log.Infof("Filter beacon candidate by rule hasEnoughActiveTimes, accepted idx %+v, rejected idx %+v", enoughTimes, notEnoughTimes)
	for _, idx := range notEnoughTimes {
		shouldRemove = append(shouldRemove, processList[idx])
	}

	processListTmp = []string{}
	for _, idx := range enoughTimes {
		processListTmp = append(processListTmp, processList[idx])
	}
	processList = processListTmp
	accepted, rejected = hasEnoughDelegator(processList, env)

	Logger.log.Infof("Filter beacon candidate by rule hasEnoughActiveTimes, accepted idx %+v, rejected idx %+v", accepted, rejected)
	for _, idx := range rejected {
		stayWaiting = append(stayWaiting, processListTmp[idx])
	}
	for _, idx := range accepted {
		toPending = append(toPending, processListTmp[idx])
	}
	Logger.log.Infof("Filter beacon candidate done, to pending %+v, stay waiting %+v, should remove %+v", common.ShortPKList(toPending), common.ShortPKList(stayWaiting), common.ShortPKList(shouldRemove))
	return toPending, stayWaiting, shouldRemove
}

func notInShardCycle(candidates []string, env *AssignEnvironment) (
	accepted []int,
	rejected []int,
) {
	allShardStakerM := map[string]interface{}{}
	for _, pk := range env.shardNewCandidates {
		allShardStakerM[pk] = nil
	}
	for _, publicKeys := range env.shardCommittee {
		for _, pk := range publicKeys {
			allShardStakerM[pk] = nil
		}
	}
	for _, publicKeys := range env.shardSubstitute {
		for _, pk := range publicKeys {
			allShardStakerM[pk] = nil
		}
	}
	for id, candidate := range candidates {
		if _, has := allShardStakerM[candidate]; has {
			rejected = append(rejected, id)
		} else {
			accepted = append(accepted, id)
		}
	}
	return accepted, rejected
}

func hasEnoughActiveTimes(candidates []string, env *AssignEnvironment) (
	accepted []int,
	rejected []int,
) {
	for id, candidate := range candidates {
		if stakerInfo, has, err := statedb.GetBeaconStakerInfo(env.ConsensusStateDB, candidate); (err != nil) || (!has) || (stakerInfo.ActiveTimesInCommittee() < uint(config.Param().ConsensusParam.RequiredActiveTimes)) {
			rejected = append(rejected, id)
		} else {
			accepted = append(accepted, id)
		}
	}
	return accepted, rejected
}

func hasEnoughDelegator(candidates []string, env *AssignEnvironment) (
	accepted []int,
	rejected []int,
) {
	for id, candidate := range candidates {
		if dState, has := env.delegateState[candidate]; has {
			//TODO remove hardcode here
			if (dState != nil) && (dState.CurrentDelegators >= 0) {
				accepted = append(accepted, id)
				continue
			}
		}
		rejected = append(rejected, id)
	}
	return accepted, rejected
}

func getOrderedLowerSet(mean int, numberOfValidators []int) []int {

	lowerSet := []int{}
	totalShard := len(numberOfValidators)
	sortedShardIDs := sortShardIDByIncreaseOrder(numberOfValidators)

	halfOfShard := totalShard / 2
	if halfOfShard == 0 {
		halfOfShard = 1
	}

	for _, shardID := range sortedShardIDs {
		if numberOfValidators[shardID] < mean && len(lowerSet) < halfOfShard {
			lowerSet = append(lowerSet, int(shardID))
		}
	}

	//special case: mean == 0 || shard committee size is equal among all shard || len(numberOfValidators) == 1
	if len(lowerSet) == 0 {
		for i, _ := range numberOfValidators {
			if i == halfOfShard {
				break
			}
			lowerSet = append(lowerSet, i)
		}
	}

	return lowerSet
}

// calculateCandidatePositionV2 random a position in total
func calculateCandidatePositionV2(total int) (pos int) {
	return rand.Intn(total)
}

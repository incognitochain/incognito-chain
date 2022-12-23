package committeestate

import (
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/instruction"
)

// swapRuleV3 ...
type swapRuleV3 struct {
}

func NewSwapRuleV3() *swapRuleV3 {
	return &swapRuleV3{}
}

func (s *swapRuleV3) ProcessBeacon(
	committees, substitutes []string,
	minCommitteeSize, maxCommitteeSize, numberOfFixedValidators int,
	reputation map[string]uint64,
	performance map[string]uint64,
) (
	newCommittees []string,
	newSubstitutes []string,
	swapOutList []string,
	slashedList []string,
) {
	committeeClone := append([]string{}, committees...)
	substitutesClone := append([]string{}, substitutes...)
	maxSwapout := len(committees) - minCommitteeSize
	if len(substitutes) > (minCommitteeSize - numberOfFixedValidators) {
		maxSwapout = len(committeeClone) - numberOfFixedValidators
	} else {
		maxSwapout += len(substitutesClone)
	}
	totalVotePower := uint64(0)
	for _, pk := range committeeClone {
		votePower, ok := reputation[pk]
		if ok {
			totalVotePower += votePower
		}
	}
	slashedVotePower := uint64(0)
	Logger.log.Infof("Process swap beacon 1, env: min %v max %v n.o. fixed nodes %v max swapout %v, subs %v, cmt %v, total vote power %v", minCommitteeSize, maxCommitteeSize, numberOfFixedValidators, maxSwapout, common.ShortPKList(substitutes), common.ShortPKList(committees), totalVotePower)
	committeeClone, slashedList, slashedVotePower = s.slashingBeaconSwapOut(
		committeeClone,
		maxSwapout,
		numberOfFixedValidators,
		reputation,
		performance,
		totalVotePower,
	)
	totalVotePower -= slashedVotePower
	maxSwapout -= len(slashedList)
	Logger.log.Infof("Process swap beacon 2, env: min %v max %v n.o. fixed nodes %v max swapout %v, slashed %v, slashed votepower %v total vote power %v", minCommitteeSize, maxCommitteeSize, numberOfFixedValidators, maxSwapout, common.ShortPKList(slashedList), slashedVotePower, totalVotePower)
	committeeClone, swapOutList, substitutesClone = s.beaconSwapOut(
		committeeClone,
		substitutesClone,
		numberOfFixedValidators,
		maxCommitteeSize,
		maxSwapout,
		reputation,
		totalVotePower,
	)
	Logger.log.Infof("Process swap beacon 3, env: min %v max %v n.o. fixed nodes %v max swapout %v, new cmt %v, new subs %v swapout %v", minCommitteeSize, maxCommitteeSize, numberOfFixedValidators, maxSwapout, common.ShortPKList(committeeClone), common.ShortPKList(substitutesClone), common.ShortPKList(swapOutList))
	committeeM := map[string]interface{}{}
	substituteM := map[string]interface{}{}
	newCommittees = []string{}
	newSubstitutes = []string{}
	for _, pk := range committeeClone {
		committeeM[pk] = nil
	}
	for _, pk := range substitutesClone {
		substituteM[pk] = nil
	}
	for _, pk := range committees {
		if _, ok := committeeM[pk]; ok {
			newCommittees = append(newCommittees, pk)
		}
		if _, ok := substituteM[pk]; ok {
			newSubstitutes = append(newSubstitutes, pk)
		}
	}
	for _, pk := range substitutes {
		if _, ok := committeeM[pk]; ok {
			newCommittees = append(newCommittees, pk)
		}
		if _, ok := substituteM[pk]; ok {
			newSubstitutes = append(newSubstitutes, pk)
		}
	}
	Logger.log.Infof("Process swap beacon done, env: min %v max %v n.o. fixed nodes %v max swapout %v, new cmt %v, new subs %v swapout %v", minCommitteeSize, maxCommitteeSize, numberOfFixedValidators, maxSwapout, common.ShortPKList(newCommittees), common.ShortPKList(newSubstitutes), common.ShortPKList(swapOutList))
	return newCommittees, newSubstitutes, swapOutList, slashedList
}

// GenInstructions generate instructions for swap rule v3
func (s *swapRuleV3) Process(
	shardID byte,
	committees, substitutes []string,
	minCommitteeSize, maxCommitteeSize, typeIns, numberOfFixedValidators int,
	penalty map[string]signaturecounter.Penalty,
) (*instruction.SwapShardInstruction, []string, []string, []string, []string) {

	// @NOTICE: hack code to reduce code complexity
	// All running network need to maintain numberOfFixedValidators equal to minCommitteeSize
	// if numberOfFixedValidators = 0, code execution may go wrong
	minNumberOfValidators := numberOfFixedValidators
	if minNumberOfValidators < minCommitteeSize {
		minNumberOfValidators = minCommitteeSize
	}
	//get slashed nodes
	newCommittees, slashingCommittees :=
		s.slashingSwapOut(committees, penalty, minNumberOfValidators)
	lenSlashedCommittees := len(slashingCommittees)
	//get normal swap out nodes
	newCommittees, normalSwapOutCommittees :=
		s.normalSwapOut(newCommittees, substitutes, len(committees), lenSlashedCommittees, minNumberOfValidators, maxCommitteeSize)
	swappedOutCommittees := append(slashingCommittees, normalSwapOutCommittees...)

	newCommittees, newSubstitutes, swapInCommittees :=
		s.swapInAfterSwapOut(newCommittees, substitutes, maxCommitteeSize, len(slashingCommittees), len(committees))

	if len(swapInCommittees) == 0 && len(swappedOutCommittees) == 0 {
		return instruction.NewSwapShardInstructionWithShardID(int(shardID)), newCommittees, newSubstitutes, slashingCommittees, normalSwapOutCommittees
	}

	swapShardInstruction := instruction.NewSwapShardInstructionWithValue(
		swapInCommittees,
		swappedOutCommittees,
		int(shardID),
		typeIns,
	)

	return swapShardInstruction, newCommittees, newSubstitutes, slashingCommittees, normalSwapOutCommittees
}

func (s *swapRuleV3) CalculateAssignOffset(lenShardSubstitute, lenCommittees, numberOfFixedValidators, minCommitteeSize int) int {
	assignOffset := lenCommittees / MAX_ASSIGN_PERCENT_V3
	if assignOffset == 0 && lenCommittees < MAX_ASSIGN_PERCENT_V3 {
		assignOffset = 1
	}
	return assignOffset
}

func (s *swapRuleV3) swapInAfterSwapOut(lenCommitteesAfterSwapOut []string, substitutes []string, maxCommitteeSize int, numberOfSlashingValidators int, lenCommitteesBeforeSwapOut int) ([]string, []string, []string) {
	swapInOffset := s.getSwapInOffset(len(lenCommitteesAfterSwapOut), len(substitutes), maxCommitteeSize, numberOfSlashingValidators, lenCommitteesBeforeSwapOut)

	newCommittees := common.DeepCopyString(lenCommitteesAfterSwapOut)
	swapInCommittees := common.DeepCopyString(substitutes[:swapInOffset])
	newSubstitutes := common.DeepCopyString(substitutes[swapInOffset:])
	newCommittees = append(newCommittees, swapInCommittees...)

	return newCommittees, newSubstitutes, swapInCommittees
}

// getSwapInOffset calculate based on lenCommitteesAfterSwapOut
// swap_in = min(lenSubstitute, lenCommitteesAfterSwapOut/8) but no more than maxCommitteeSize
// Special case: when committees size reach max and no slashing, swap in equal to normal swap out
func (s *swapRuleV3) getSwapInOffset(lenCommitteesAfterSwapOut, lenSubstitutes, maxCommitteeSize, numberOfSlashingValidators, lenCommitteesBeforeSwapOut int) int {
	var offset int
	// special case: no slashing node && committee reach max committee size => normal swap out == swap in
	if numberOfSlashingValidators == 0 && lenCommitteesBeforeSwapOut == maxCommitteeSize {
		offset = lenCommitteesBeforeSwapOut / MAX_SWAP_IN_PERCENT_V3
	} else {
		// normal case
		offset = lenCommitteesAfterSwapOut / MAX_SWAP_IN_PERCENT_V3
	}

	// if committee size after swap out below than maxSwapInPercent => no swap in
	// try to swap in at least one
	if offset == 0 && lenCommitteesAfterSwapOut < MAX_SWAP_IN_PERCENT_V3 {
		offset = 1
	}

	if lenSubstitutes < offset {
		offset = lenSubstitutes
	}

	if lenCommitteesAfterSwapOut+offset > maxCommitteeSize {
		offset = maxCommitteeSize - lenCommitteesAfterSwapOut
	}

	return offset
}

func (s *swapRuleV3) normalSwapOut(committeesAfterSlashing, substitutes []string, lenCommitteesBeforeSlash, lenSlashedCommittees, minNumberOfValidators, maxCommitteeSize int) ([]string, []string) {

	resNormalSwapOut := []string{}
	tempCommittees := make([]string, len(committeesAfterSlashing))
	copy(tempCommittees, committeesAfterSlashing)

	normalSwapOutOffset := s.getNormalSwapOutOffset(lenCommitteesBeforeSlash, len(substitutes), lenSlashedCommittees, minNumberOfValidators, maxCommitteeSize)

	resCommittees := append(tempCommittees[:minNumberOfValidators], tempCommittees[(minNumberOfValidators+normalSwapOutOffset):]...)
	resNormalSwapOut = committeesAfterSlashing[minNumberOfValidators : minNumberOfValidators+normalSwapOutOffset]

	return resCommittees, resNormalSwapOut
}

// getNormalSwapOutOffset calculate normal swapout offset
func (s *swapRuleV3) getNormalSwapOutOffset(lenCommitteesBeforeSlash, lenSubstitutes, lenSlashedCommittees, numberOfFixedValidators, maxCommitteeSize int) int {
	if lenSubstitutes == 0 {
		return 0
	}

	if maxCommitteeSize != lenCommitteesBeforeSlash {
		return 0
	}

	maxSlashingOffset := s.getMaxSlashingOffset(lenCommitteesBeforeSlash, numberOfFixedValidators)
	if lenSlashedCommittees == maxSlashingOffset {
		return 0
	}

	maxNormalSwapOutOffset := lenCommitteesBeforeSlash / MAX_SWAP_OUT_PERCENT_V3
	if maxNormalSwapOutOffset > lenCommitteesBeforeSlash-numberOfFixedValidators {
		maxNormalSwapOutOffset = lenCommitteesBeforeSlash - numberOfFixedValidators
	}

	if lenSlashedCommittees >= maxNormalSwapOutOffset {
		return 0
	}

	offset := maxNormalSwapOutOffset - lenSlashedCommittees

	if offset > lenSubstitutes {
		offset = lenSubstitutes
	}

	lenCommitteesAfterSlash := lenCommitteesBeforeSlash - lenSlashedCommittees
	if offset > lenCommitteesAfterSlash-numberOfFixedValidators {
		offset = lenCommitteesAfterSlash - numberOfFixedValidators
	}

	return offset
}

// slashingSwapOut only consider all penalties type as one type
func (s *swapRuleV3) slashingSwapOut(committees []string, penalty map[string]signaturecounter.Penalty, minNumberOfValidators int) ([]string, []string) {
	fixedCommittees := common.DeepCopyString(committees[:minNumberOfValidators])
	flexCommittees := common.DeepCopyString(committees[minNumberOfValidators:])
	flexAfterSlashingCommittees := []string{}
	slashingCommittees := []string{}

	maxSlashingOffset := s.getMaxSlashingOffset(len(committees), minNumberOfValidators)

	for _, flexCommittee := range flexCommittees {
		if _, ok := penalty[flexCommittee]; ok && maxSlashingOffset > 0 {
			slashingCommittees = append(slashingCommittees, flexCommittee)
			maxSlashingOffset--
		} else {
			flexAfterSlashingCommittees = append(flexAfterSlashingCommittees, flexCommittee)
		}
	}

	newCommittees := append(fixedCommittees, flexAfterSlashingCommittees...)
	return newCommittees, slashingCommittees
}

func (s *swapRuleV3) slashingBeaconSwapOut(
	committees []string,
	maxSlashedValidator, numberOfFixedValidators int,
	reputation map[string]uint64,
	performance map[string]uint64,
	totalVotePower uint64,
) (
	newCommittees []string,
	slashedList []string,
	slashedVotePower uint64,
) {
	sortedCommittee := append([]string{}, committees[numberOfFixedValidators:]...)
	sortedCommittee, err := sortStakersByReputation(sortedCommittee, true, reputation, performance)
	if err != nil {
		panic(err)
	}
	slashed := map[string]interface{}{}
	slashedVotePower = uint64(0)
	for idx, pk := range sortedCommittee {
		pi := uint64(0)
		if _, ok := performance[sortedCommittee[idx]]; ok {
			pi = performance[sortedCommittee[idx]]
		}
		votePower := uint64(0)
		if _, ok := reputation[sortedCommittee[idx]]; ok {
			votePower = reputation[sortedCommittee[idx]]
		}
		//TODO: remove hardcode constant
		if pi < config.Param().ConsensusParam.LowestPerformance {
			if votePower+uint64(slashedVotePower) < totalVotePower/3 {
				if len(slashedList) == maxSlashedValidator {
					break
				}
				slashedVotePower += votePower
				slashed[pk] = nil
			}
		} else {
			break
		}
	}
	notSlashed := []string{}
	for _, pk := range committees[numberOfFixedValidators:] {
		if _, ok := slashed[pk]; ok {
			slashedList = append(slashedList, pk)
		} else {
			notSlashed = append(notSlashed, pk)
		}
	}
	newCommittees = append(committees[:numberOfFixedValidators], notSlashed...)
	return newCommittees, slashedList, slashedVotePower
}

func (s *swapRuleV3) beaconSwapOut(
	committee, substitutes []string,
	numberOfFixedValidators, maxCommitteeSize, maxSwapOut int,
	reputation map[string]uint64,
	totalVotePower uint64,
) (
	newCommittee []string,
	swapOut []string,
	newSubtitute []string,
) {
	var err error
	sortedCommittee := append([]string{}, committee[numberOfFixedValidators:]...)
	sortedCommittee, err = sortStakersByReputation(sortedCommittee, true, reputation, reputation)
	if err != nil {
		panic(err)
	}
	substitutes, err = sortStakersByReputation(substitutes, false, reputation, reputation)
	if err != nil {
		panic(err)
	}
	for _, pk := range sortedCommittee {
		Logger.log.Infof("swapbeacon process committee: key %v, rep %v", pk[len(pk)-5:], reputation[pk])
	}
	for _, pk := range substitutes {
		Logger.log.Infof("swapbeacon process pending: key %v, rep %v", pk[len(pk)-5:], reputation[pk])
	}
	swappedInIndex := -1
	for idx, key := range sortedCommittee {
		if idx+1 > maxSwapOut {
			break
		}
		if (swappedInIndex+1 < len(substitutes)) && (reputation[key] < reputation[substitutes[swappedInIndex+1]]) {
			swapOut = append(swapOut, key)
			swappedInIndex++
		} else {
			break
		}
	}
	newCommittee = append([]string{}, committee[:numberOfFixedValidators]...)
	newCommittee = append(newCommittee, sortedCommittee[len(swapOut):]...)
	newCommittee = append(newCommittee, substitutes[:swappedInIndex+1]...)
	substitutes = substitutes[swappedInIndex+1:]
	if (len(substitutes) > 0) && (len(newCommittee) < maxCommitteeSize) {
		if len(substitutes) > maxCommitteeSize-len(newCommittee) {
			swappedInIndex = maxCommitteeSize - len(newCommittee) - 1
		} else {
			swappedInIndex = len(substitutes) - 1
		}
		newCommittee = append(newCommittee, substitutes[:swappedInIndex+1]...)
		substitutes = substitutes[swappedInIndex+1:]
	}
	newSubtitute = substitutes
	return
}

// getMaxSlashingOffset calculate maximum slashing offset, fixed nodes must be spare
// max_slashing_offset = 1/3 committee length
func (s *swapRuleV3) getMaxSlashingOffset(lenCommittees, numberOfFixedValidators int) int {
	if lenCommittees == numberOfFixedValidators {
		return 0
	}

	offset := lenCommittees / MAX_SLASH_PERCENT_V3
	if offset > lenCommittees-numberOfFixedValidators {
		offset = lenCommittees - numberOfFixedValidators
	}

	return offset
}

// calculateNewSubstitutePosition calculate reverse shardID for candidate
func calculateNewSubstitutePosition(candidate string, rand int64, total int) (pos int) {

	seed := candidate + fmt.Sprintf("%v", rand)
	hash := common.HashB([]byte(seed))

	// Using big.Int to convert a random Hash value to an integer
	temp := new(big.Int)
	temp.SetBytes(hash[:])
	data := int(temp.Int64())
	if data < 0 {
		data *= -1
	}

	pos = data % total

	return pos
}

func (s *swapRuleV3) Version() int {
	return swapRuleDCSVersion
}

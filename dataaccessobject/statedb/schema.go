package statedb

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"strconv"
)

var (
	serialNumberPrefix      = []byte("serial-number-")
	allShardCommitteePrefix = []byte("all-shard-committee-")
	nextCandidatePrefix     = []byte("next-cand-")
	currentCandidatePrefix  = []byte("cur-cand-")
	substitutePrefix        = []byte("shard-sub-")
	committeePrefix         = []byte("shard-com-")
	rewardReceiverPrefix    = []byte("reward-receiver-")
)

func GetCommitteePrefixWithRole(role int, shardID int) []byte {
	switch role {
	case NextEpochCandidate:
		temp := []byte(string(nextCandidatePrefix))
		h := common.HashH(temp)
		return h[:][:prefixHashKeyLength]
	case CurrentEpochCandidate:
		temp := []byte(string(currentCandidatePrefix))
		h := common.HashH(temp)
		return h[:][:prefixHashKeyLength]
	case SubstituteValidator:
		temp := []byte(string(substitutePrefix) + strconv.Itoa(shardID))
		h := common.HashH(temp)
		return h[:][:prefixHashKeyLength]
	case CurrentValidator:
		temp := []byte(string(committeePrefix) + strconv.Itoa(shardID))
		h := common.HashH(temp)
		return h[:][:prefixHashKeyLength]
	default:
		panic("no role exist")
	}
}
func GetRewardReceiverPrefix() []byte {
	temp := []byte(rewardReceiverPrefix)
	h := common.HashH(temp)
	return h[:][:prefixHashKeyLength]
}
func GetSerialNumberPrefix() []byte {
	h := common.HashH(serialNumberPrefix)
	return h[:][:prefixHashKeyLength]
}

var _ = func() (_ struct{}) {
	m := make(map[string]string)
	prefixs := [][]byte{}
	// Current validator
	for i := -1; i < 256; i++ {
		temp := GetCommitteePrefixWithRole(CurrentValidator, i)
		prefixs = append(prefixs, temp)
		if v, ok := m[string(temp)]; ok {
			panic("shard-com-" + strconv.Itoa(i) + " same prefix " + v)
		}
		m[string(temp)] = "shard-com-" + strconv.Itoa(i)
	}
	// Substitute validator
	for i := -1; i < 256; i++ {
		temp := GetCommitteePrefixWithRole(SubstituteValidator, i)
		prefixs = append(prefixs, temp)
		if v, ok := m[string(temp)]; ok {
			panic("shard-sub-" + strconv.Itoa(i) + " same prefix " + v)
		}
		m[string(temp)] = "shard-sub-" + strconv.Itoa(i)
	}
	// Current Candidate
	tempCurrentCandidate := GetCommitteePrefixWithRole(CurrentEpochCandidate, -2)
	prefixs = append(prefixs, tempCurrentCandidate)
	if v, ok := m[string(tempCurrentCandidate)]; ok {
		panic("cur-cand-" + " same prefix " + v)
	}
	m[string(tempCurrentCandidate)] = "cur-cand-"
	// Next candidate
	tempNextCandidate := GetCommitteePrefixWithRole(NextEpochCandidate, -2)
	prefixs = append(prefixs, tempNextCandidate)
	if v, ok := m[string(tempNextCandidate)]; ok {
		panic("next-cand-" + " same prefix " + v)
	}
	m[string(tempNextCandidate)] = "next-cand-"
	// serial number
	tempSerialNumber := GetSerialNumberPrefix()
	prefixs = append(prefixs, tempSerialNumber)
	if v, ok := m[string(tempSerialNumber)]; ok {
		panic("serial-number-" + " same prefix " + v)
	}
	m[string(tempSerialNumber)] = "serial-number-"
	// reward receiver
	tempRewardReceiver := GetRewardReceiverPrefix()
	prefixs = append(prefixs, tempRewardReceiver)
	if v, ok := m[string(tempRewardReceiver)]; ok {
		panic("reward-receiver-" + " same prefix " + v)
	}
	m[string(tempRewardReceiver)] = "reward-receiver-"
	for i, v1 := range prefixs {
		for j, v2 := range prefixs {
			if i == j {
				continue
			}
			if bytes.HasPrefix(v1, v2) || bytes.HasPrefix(v2, v1) {
				panic("(prefix: " + fmt.Sprintf("%+v", v1) + ", value: " + m[string(v1)] + ")" + " is prefix or being prefix of " + " (prefix: " + fmt.Sprintf("%+v", v1) + ", value: " + m[string(v2)] + ")")
			}
		}
	}
	return
}()

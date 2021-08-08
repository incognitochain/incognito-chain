package committeestate

import (
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

func TestUtils(t *testing.T) {
	temp := []string{}
	shardIDs := make(map[string]byte)
	m := make(map[string]string)
	shuffledCandidates := []string{}
	for _, candidate := range candidates {
		seed := strconv.Itoa(int(randomNumber)) + candidate
		hash := common.HashH([]byte(seed)).String()
		temp = append(temp, hash)
		m[hash] = candidate
		shardID := calculateCandidateShardID(candidate, randomNumber, 2)
		shardIDs[candidate] = shardID
	}
	sort.Strings(temp)
	for _, key := range temp {
		shuffledCandidates = append(shuffledCandidates, m[key])
	}
}

func TestShuffleShardCandidate(t *testing.T) {
	shuffledCandiates := shuffleShardCandidate(candidates, randomNumber)
	if !reflect.DeepEqual(shuffledCandiates, expectedShuffledCandidates) {
		t.Fatalf("Expect shuffled unassignedCommonPool to be %+v \n but get %+v", priorityKeys, shuffledCandiates)
	}
}
func TestAssignShardCandidate(t *testing.T) {
	numberOfPendingValidator := make(map[byte]int)
	numberOfPendingValidator[0] = 0
	numberOfPendingValidator[1] = 0
	testnetAssignOffset := 3
	activeShards := 2
	remainCandidates, newAssignCandidates := assignShardCandidate(candidates, numberOfPendingValidator, randomNumber, testnetAssignOffset, activeShards)
	if !reflect.DeepEqual(remainCandidates, expectedRemainCandidates) {
		t.Fatalf("Expected remain candidate to be %+v \n but get %+v", expectedRemainCandidates, remainCandidates)
	}
	for shardID, candidates := range newAssignCandidates {
		if len(candidates) != 3 {
			t.Fatalf("Expect New Assigned Candidate have 3 but get %+v", len(candidates))
		}
		if shardID == 0 {
			if common.IndexOfStr("121VhftSAygpEJZ6i9jGk9a1qL1ZjK3QhzqpDT4uJop7tnEfahvCVRVbKPsuH94uGMqj1a1nLQSAcQypiUP3yc1s3t4jCRae8Kf6VXJRyHngt9X9iT2yJexKgRhnjzTfJYU4VkXV3w3xwiFCp5pGnDdocA8aq1SUCVMwfKAnhxjDHQDDWMMaxSLvjU56ZqaNZtsCQGTaySZLxc2RmCgMUeC5JQkF4t3P2NbTaZzz9JreLtZqSL3DVLacPwXy2enu3QMQYcPbZRVbz2Rxuf8e3hHYwjbAqvDJayCk9iasJnmZKP2gVRQhHUcX7cED4U22TMzi2rE6FoVMebThoDB2Dp21BW21qthS26Rkxe7UbxTesss9Xk1LSQsh8tRV1yxyGJ3DgtY2csBicRmT5PxQ3j9FwdSViKX82M1u9qv7ZcQSnvrJ", candidates) == -1 {
				t.Fatalf("Expect %+v in shard 0 but get %+v", "121VhftSAygpEJZ6i9jGk9a1qL1ZjK3QhzqpDT4uJop7tnEfahvCVRVbKPsuH94uGMqj1a1nLQSAcQypiUP3yc1s3t4jCRae8Kf6VXJRyHngt9X9iT2yJexKgRhnjzTfJYU4VkXV3w3xwiFCp5pGnDdocA8aq1SUCVMwfKAnhxjDHQDDWMMaxSLvjU56ZqaNZtsCQGTaySZLxc2RmCgMUeC5JQkF4t3P2NbTaZzz9JreLtZqSL3DVLacPwXy2enu3QMQYcPbZRVbz2Rxuf8e3hHYwjbAqvDJayCk9iasJnmZKP2gVRQhHUcX7cED4U22TMzi2rE6FoVMebThoDB2Dp21BW21qthS26Rkxe7UbxTesss9Xk1LSQsh8tRV1yxyGJ3DgtY2csBicRmT5PxQ3j9FwdSViKX82M1u9qv7ZcQSnvrJ", candidates)
			}
			if common.IndexOfStr("121VhftSAygpEJZ6i9jGk9dPK5DE41KmqUhp2EFJXyyHNrQomEh7QW7icEh2zymM5953C2HN2kYYCbepkL37Qny6mgyxGiAxxqEEkCHnqK4FsNBDbGwojMmVCgURwmS1oRqpfoFYhKwFUTzjyYduxBNegdgUw9VEYXJe3xSMBib24oBRhjbmK7LZ5qxRqKMsV7Lp49rPi2QGiizqiVKSsaAD2QMHEaC5ZBsGJvtRRMdGarHfe6yLtWAvcPLpsuSkYHfNni5Nh61LT1LwYLqSwDqNVKC3ew6RfRvmFfFTPhBGUcbhRKog6djipwKW9RZ7jvpgSiwifLriiz8h25ziR3Guh7k1cgyYm6TxQHrrLBRUgXKGvLkeAEi9ThrCsuQnbE61re6UmjwCcbPd9j53fzVmm4JJ19kMsyGUL2m7obSsdxvM", candidates) == -1 {
				t.Fatalf("Expect %+v in shard 0 but get %+v", "121VhftSAygpEJZ6i9jGk9dPK5DE41KmqUhp2EFJXyyHNrQomEh7QW7icEh2zymM5953C2HN2kYYCbepkL37Qny6mgyxGiAxxqEEkCHnqK4FsNBDbGwojMmVCgURwmS1oRqpfoFYhKwFUTzjyYduxBNegdgUw9VEYXJe3xSMBib24oBRhjbmK7LZ5qxRqKMsV7Lp49rPi2QGiizqiVKSsaAD2QMHEaC5ZBsGJvtRRMdGarHfe6yLtWAvcPLpsuSkYHfNni5Nh61LT1LwYLqSwDqNVKC3ew6RfRvmFfFTPhBGUcbhRKog6djipwKW9RZ7jvpgSiwifLriiz8h25ziR3Guh7k1cgyYm6TxQHrrLBRUgXKGvLkeAEi9ThrCsuQnbE61re6UmjwCcbPd9j53fzVmm4JJ19kMsyGUL2m7obSsdxvM", candidates)
			}
			if common.IndexOfStr("121VhftSAygpEJZ6i9jGk9a5oQvgecAms7BtyyjGxpySigmDAdo8af26UKNmXYjNUhFVp4NN6RpRJFTGn57w7evPi3HkaF8ToXsCJ7ceaZ5p6hrCHLN6tKm5sjBEo3yusZZayurNMsrLGRhBE2i8Xhxkns8uN8c9WY5kcwSVsyD9f7399fMzRqtMB7TjyE4ad2KWDmteZZWuZB79jYB5wHWcxRyUxYaQ761gT8oMJ6FKr7wdPYAeuJ7Pai71Xi9YdPUDNQ1dZ7Uq1m7wxavKKax14Tuf9onu9oZDTeat7SNK1PDxjvf2uwkEmcHAp7qzp8c7igm8X6VjC6685gdThcdEHPxiwDsu3UxQyXK1fqwSHDHx3Ff7w5xKeDK8zJNghCLBbZ3HowQbT7hAKqu3N5puMKw5cjQJndA4trRw5yuzXxbf", candidates) == -1 {
				t.Fatalf("Expect %+v in shard 0 but get %+v", "121VhftSAygpEJZ6i9jGk9a5oQvgecAms7BtyyjGxpySigmDAdo8af26UKNmXYjNUhFVp4NN6RpRJFTGn57w7evPi3HkaF8ToXsCJ7ceaZ5p6hrCHLN6tKm5sjBEo3yusZZayurNMsrLGRhBE2i8Xhxkns8uN8c9WY5kcwSVsyD9f7399fMzRqtMB7TjyE4ad2KWDmteZZWuZB79jYB5wHWcxRyUxYaQ761gT8oMJ6FKr7wdPYAeuJ7Pai71Xi9YdPUDNQ1dZ7Uq1m7wxavKKax14Tuf9onu9oZDTeat7SNK1PDxjvf2uwkEmcHAp7qzp8c7igm8X6VjC6685gdThcdEHPxiwDsu3UxQyXK1fqwSHDHx3Ff7w5xKeDK8zJNghCLBbZ3HowQbT7hAKqu3N5puMKw5cjQJndA4trRw5yuzXxbf", candidates)
			}
		}
		if shardID == 1 {
			if common.IndexOfStr("121VhftSAygpEJZ6i9jGk9dTsGQVZNEnYLWhjXjomneDCT9XC6zGjeTW8ENTJVPooNnQBRuRdpwe9DoNVdLF4j3iy6Ld2p1eNLK8ek1bNSNrHFVjtsHaQVpcoHBy6nGghA9y7pr6ne6W7MRdAXLPHEDgpAodnRVxFSc2zyUbA48XKQVRtSVUneKjcdhDHP1hY9gC1EYLWtqn4weLzHKA84w77rHFhwwV2sbseHDqsmtQa4k5aiTCcXoUdmMpEuvadtSmKR33wciA3FNr5h125ce9ge1eSFPznvXNsCy7sA1rc4YJxipgwGoDDWSftUfnh5vY37WbwwNLhmMRvxtjHP7WBkSrLrKdamr8TdUu93cQyrykYRxbCben7pK1N75NknsSqLSihBnMTLo8gcFDGaDtQixPiJxMkefZ3qfHHwdvaL27", candidates) == -1 {
				t.Fatalf("Expect %+v in shard 0 but get %+v", "121VhftSAygpEJZ6i9jGk9dTsGQVZNEnYLWhjXjomneDCT9XC6zGjeTW8ENTJVPooNnQBRuRdpwe9DoNVdLF4j3iy6Ld2p1eNLK8ek1bNSNrHFVjtsHaQVpcoHBy6nGghA9y7pr6ne6W7MRdAXLPHEDgpAodnRVxFSc2zyUbA48XKQVRtSVUneKjcdhDHP1hY9gC1EYLWtqn4weLzHKA84w77rHFhwwV2sbseHDqsmtQa4k5aiTCcXoUdmMpEuvadtSmKR33wciA3FNr5h125ce9ge1eSFPznvXNsCy7sA1rc4YJxipgwGoDDWSftUfnh5vY37WbwwNLhmMRvxtjHP7WBkSrLrKdamr8TdUu93cQyrykYRxbCben7pK1N75NknsSqLSihBnMTLo8gcFDGaDtQixPiJxMkefZ3qfHHwdvaL27", candidates)
			}
			if common.IndexOfStr("121VhftSAygpEJZ6i9jGk9dnzbXovyLBfFHs6cuMcjFEAB6FphEtuXPXCCNga16A1LkSHHSgLq9rFGQ496VnxHxDL8Ar6duhJxpEdcr75rkRFBzEvPNTkRYVbVcBUMtWd5PugsB6QjYt5CJQyUtzzbctC15AeX64aHrK1QwHJjFz8hnMz3eh8P8SPqKQ6zXhByqHm7YrcCY5uKZUn7CqM9RTwwJaqUhtqsHKyozBkfw26XhfkwvFh4vvL6McE5Ty1ztgiygUp8tt7haPEjVNGCqnwpPEB76oPVFTiKGePY1XqHy6aAFvuUBdmnNFEXrQPnh5xz5ULpq6PMJjVFvpu5kXcnaJbFBRZJcsJm1y69n56zUFKSg4LiNwqymf84U2SiPKT12cFehJedwfJkBMwnGpDAfaEZZvqcNRD9nVRUG4no9X", candidates) == -1 {
				t.Fatalf("Expect %+v in shard 0 but get %+v", "121VhftSAygpEJZ6i9jGk9dnzbXovyLBfFHs6cuMcjFEAB6FphEtuXPXCCNga16A1LkSHHSgLq9rFGQ496VnxHxDL8Ar6duhJxpEdcr75rkRFBzEvPNTkRYVbVcBUMtWd5PugsB6QjYt5CJQyUtzzbctC15AeX64aHrK1QwHJjFz8hnMz3eh8P8SPqKQ6zXhByqHm7YrcCY5uKZUn7CqM9RTwwJaqUhtqsHKyozBkfw26XhfkwvFh4vvL6McE5Ty1ztgiygUp8tt7haPEjVNGCqnwpPEB76oPVFTiKGePY1XqHy6aAFvuUBdmnNFEXrQPnh5xz5ULpq6PMJjVFvpu5kXcnaJbFBRZJcsJm1y69n56zUFKSg4LiNwqymf84U2SiPKT12cFehJedwfJkBMwnGpDAfaEZZvqcNRD9nVRUG4no9X", candidates)
			}
			if common.IndexOfStr("121VhftSAygpEJZ6i9jGk9a4wsKYVdTBhFsvZ4jwhzaWjFsondoUtyYpqF3hjWmPEUvMJF1Wh9NMhRVGaLYW8JtmS2JcP3Zq1L2AvYjYJaWm5CRqoYPz3DUEeJFSbUDdZPcp5YAL3FizX3J192zR3kWA967p6cZUgVP6e9LbHBdQBFGCQYRYKw8DNRCUzophTmoFeFwjWJF3EuXAWX2gv5ASj2Nem9YytNtjhZHCRS7Vz8pvgLdMJVa7B9fpS6oBV7nS4LWdeZc8NPC9VGKbTa1MCy7YkXfjbHfKpowEArnB9CLaLpWmSiaZruTiRxtZQZkU9z3YCMZW2dW5SHmwMGEseu6WwPPqgLz32tazKzNzHgiAJp561pxfCm7HF4r6VxtTqPKe8gjwLfDZqw3X2ew6rQ8Vo2csnjWrSQJAxYuMeinW", candidates) == -1 {
				t.Fatalf("Expect %+v in shard 0 but get %+v", "121VhftSAygpEJZ6i9jGk9a4wsKYVdTBhFsvZ4jwhzaWjFsondoUtyYpqF3hjWmPEUvMJF1Wh9NMhRVGaLYW8JtmS2JcP3Zq1L2AvYjYJaWm5CRqoYPz3DUEeJFSbUDdZPcp5YAL3FizX3J192zR3kWA967p6cZUgVP6e9LbHBdQBFGCQYRYKw8DNRCUzophTmoFeFwjWJF3EuXAWX2gv5ASj2Nem9YytNtjhZHCRS7Vz8pvgLdMJVa7B9fpS6oBV7nS4LWdeZc8NPC9VGKbTa1MCy7YkXfjbHfKpowEArnB9CLaLpWmSiaZruTiRxtZQZkU9z3YCMZW2dW5SHmwMGEseu6WwPPqgLz32tazKzNzHgiAJp561pxfCm7HF4r6VxtTqPKe8gjwLfDZqw3X2ew6rQ8Vo2csnjWrSQJAxYuMeinW", candidates)
			}
		}
	}
}

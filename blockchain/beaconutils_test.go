package blockchain

import (
	"log"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

var (
	randomNumber = int64(1000)
	candidates   = []string{
		"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",
	}
	assignedCandidates = make(map[string]byte)
	priorityKeys       = []string{
		"002861fd7eede6f3e28394a6ccc0b559e696509151dc75d9e5630356e23da90f",
		"1ecbbbd5076fbbc574637893692de5d38b59de23c44c063da835bd44be943a97",
		"2c86c4a23fb3245ebda2fb6d70c6487804672ab5c5038e70613ebe4562903a85",
		"44b83685be993e0b2a689a3c15b5fd26f2b27a3d40015bbbecfbfe6282efb925",
		"8eda59aa90e707ce8afc3eca7473aabcc5a505c4e3ec9c4248a211f733963c8f",
		"945605ea416e86ca6aa855fbaae0b4335c60cb470e1c94fbb12ce681d06d7c43",
		"a4b23abd534a82e6a5969f1505e7a3dd91d22ff842571d328ac746f073b84ac9",
		"a6b16be39ead5680aa799a213e9bf22dac6a1f16598ba8813a79fb8150dd7961",
		"a9c5558021c0b8277e0291258ad7d42ba0a9e868af59fb0ea205cb4a53442830",
		"b43a5a6e87d98bdf219766c84734df0760416682aa75f9243a9148dfeb2f9e67",
		"e7191ee68362219eabdeff08178adcab6690f94e28dee1e525249433bdea5abf",
	}
	expectedShuffledCandidates = []string{
		"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",
	}
	expectedRemainCandidates = []string{
		"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",
	}
)
var _ = func() (_ struct{}) {
	assignedCandidates["121VhftSAygpEJZ6i9jGk9a5Kfw77TTKCB5FhKUBU1JJKrvogDS3g9JhfQXY4PP9xEzfnTRB43MhcpxmbR7qempyNsRu7k3oY59xphaP843bUnZWk17LiedGSaDUfc7xEB2jNt94rpm1FXF9hjPRijDqeqyyBhFV3uyqhmfCdnH1xxYpJW8XLk45Jhpf5vGZpy2qFn3vVUtmxk2TMdW68AsBvkT2PkFGQDqiWYRMBXStV1Npzxu3CUKamd74ZXA66tSY7rP1QE4vSFDCLX23rJE9tjVZuty74Edbin4ZzgG2PhLqvv7s3gpNzaY6oaaxRSbon7JWujnF3uv8o6DvdraGkWiJa1VhXrJjVDzNs5WVNtETGMPy58uUPkemv8oto6yCJzxiDgjUR9Yjy3mfnY3eNWV89BocWjNkBT4Y9JU8HMD8"] = 1
	assignedCandidates["121VhftSAygpEJZ6i9jGk9dnzbXovyLBfFHs6cuMcjFEAB6FphEtuXPXCCNga16A1LkSHHSgLq9rFGQ496VnxHxDL8Ar6duhJxpEdcr75rkRFBzEvPNTkRYVbVcBUMtWd5PugsB6QjYt5CJQyUtzzbctC15AeX64aHrK1QwHJjFz8hnMz3eh8P8SPqKQ6zXhByqHm7YrcCY5uKZUn7CqM9RTwwJaqUhtqsHKyozBkfw26XhfkwvFh4vvL6McE5Ty1ztgiygUp8tt7haPEjVNGCqnwpPEB76oPVFTiKGePY1XqHy6aAFvuUBdmnNFEXrQPnh5xz5ULpq6PMJjVFvpu5kXcnaJbFBRZJcsJm1y69n56zUFKSg4LiNwqymf84U2SiPKT12cFehJedwfJkBMwnGpDAfaEZZvqcNRD9nVRUG4no9X"] = 1
	assignedCandidates["121VhftSAygpEJZ6i9jGk9ZogHDZ369maVpypAooJmsxv3QK9apuYCZxe5iM61B1W8CgRRNuQuBEwBpjMM1bMBeL1buAa2LpokfP8FCKA2gfiLn9QCLew61SQYeTAgRG9rqouR5zph6ECxDbU7qYTytd5f8QFw2jedB9Y7C9eBCRgg6XHiKgaRK7WFN9iNYB3L4KGYGQCJZUTW7zLNxX2uQrpR3sPsERnsQYSfar2dLhH2d7R8Avywn1aZPH2hDLp1ytjv8kgbUM74fT7Dton9dCVm3AV4zgc39eX9j42kx2U7mXWeG7MKDZUBiAVAfWyxvecjpMt5juzuXvzRmiN5qkZJ8QVFAQ2XLSpszFZtuPJET5pSbfYn1StqvALbGKM5jU4n6oSGuQdFmSehMbkjZwZJ5oNBf1X77ax9hXigt3ejqh"] = 1
	assignedCandidates["121VhftSAygpEJZ6i9jGk9dTsGQVZNEnYLWhjXjomneDCT9XC6zGjeTW8ENTJVPooNnQBRuRdpwe9DoNVdLF4j3iy6Ld2p1eNLK8ek1bNSNrHFVjtsHaQVpcoHBy6nGghA9y7pr6ne6W7MRdAXLPHEDgpAodnRVxFSc2zyUbA48XKQVRtSVUneKjcdhDHP1hY9gC1EYLWtqn4weLzHKA84w77rHFhwwV2sbseHDqsmtQa4k5aiTCcXoUdmMpEuvadtSmKR33wciA3FNr5h125ce9ge1eSFPznvXNsCy7sA1rc4YJxipgwGoDDWSftUfnh5vY37WbwwNLhmMRvxtjHP7WBkSrLrKdamr8TdUu93cQyrykYRxbCben7pK1N75NknsSqLSihBnMTLo8gcFDGaDtQixPiJxMkefZ3qfHHwdvaL27"] = 1
	assignedCandidates["121VhftSAygpEJZ6i9jGk9dXXyF2XfP9G56ZoZ8hAtLMc3i33FWSqa3FJgMkc7pTs5qdaeJHAFfJpwjoGWazHWzknPSh8d319L9xMoz2TsLCefqPeP8Kqf24C8fuY9RCTCvnAmecnXL6SJyiVTP6Vjjhcvvdk8cQHVUSnnXxbuufgRckyw9Mc7VrgpG3qzBfeYCfWkurDmdVnyjh7jPsZQM1sBjRFiNSjQQo7HQgaMi2YQp9WGpE4kJfF265eqXySrT8BycKLnjunED9B1TU6WNs2e9aFB2u82tMzwoRwTHWbAgc9rixwM3UiAAnhiEhx56nkidGRsqo2LR5AQASHEUP6aHtnz1wcwovEPthXXChDhRuDAa5PHvfz57LQpWZANf9HV26J5wYFysJb4bN9vRcxTJdijRzDAt5QbBJs5DJo64N"] = 1
	assignedCandidates["121VhftSAygpEJZ6i9jGk9eKW2e2edvTfmZXDVdt8qTxwzuUVFHUtpf9JGKZN9damEuDoy381bwjwq6g5D4M6Zn2KUh2giSEktUca7nbvoM7L23aq9XXtsmyQKvVBseCVvUNmyERNHRZpzvzNCn6gwnzyMR58uBcibUBwV5441jYR18RxwwyKh2w8S6ogEmdrERAxdMYdxPXwj43Ve5aHnZtT8ZfV6vPPKnPmgyM95Bpw5ep1HmWoZvtF4s5WkqbCoaAYoBd94Bcysx4wzVQbvmU1SCD5hCpF3nDtE9n7G3SnNAerDE3DWiZt5GjpLtSpQyAjbaQekqpBnyKVqbsnPiDh5EynFqzETZLkt9p9hvAK2EuzQXBVZXpFwbwoiiw918LhZ2nw11xy2eQF6hGXG5GzaYYiBHmA4Fyd8rywPGA1QH4"] = 1
	assignedCandidates["121VhftSAygpEJZ6i9jGk9a4wsKYVdTBhFsvZ4jwhzaWjFsondoUtyYpqF3hjWmPEUvMJF1Wh9NMhRVGaLYW8JtmS2JcP3Zq1L2AvYjYJaWm5CRqoYPz3DUEeJFSbUDdZPcp5YAL3FizX3J192zR3kWA967p6cZUgVP6e9LbHBdQBFGCQYRYKw8DNRCUzophTmoFeFwjWJF3EuXAWX2gv5ASj2Nem9YytNtjhZHCRS7Vz8pvgLdMJVa7B9fpS6oBV7nS4LWdeZc8NPC9VGKbTa1MCy7YkXfjbHfKpowEArnB9CLaLpWmSiaZruTiRxtZQZkU9z3YCMZW2dW5SHmwMGEseu6WwPPqgLz32tazKzNzHgiAJp561pxfCm7HF4r6VxtTqPKe8gjwLfDZqw3X2ew6rQ8Vo2csnjWrSQJAxYuMeinW"] = 1
	assignedCandidates["121VhftSAygpEJZ6i9jGk9a1qL1ZjK3QhzqpDT4uJop7tnEfahvCVRVbKPsuH94uGMqj1a1nLQSAcQypiUP3yc1s3t4jCRae8Kf6VXJRyHngt9X9iT2yJexKgRhnjzTfJYU4VkXV3w3xwiFCp5pGnDdocA8aq1SUCVMwfKAnhxjDHQDDWMMaxSLvjU56ZqaNZtsCQGTaySZLxc2RmCgMUeC5JQkF4t3P2NbTaZzz9JreLtZqSL3DVLacPwXy2enu3QMQYcPbZRVbz2Rxuf8e3hHYwjbAqvDJayCk9iasJnmZKP2gVRQhHUcX7cED4U22TMzi2rE6FoVMebThoDB2Dp21BW21qthS26Rkxe7UbxTesss9Xk1LSQsh8tRV1yxyGJ3DgtY2csBicRmT5PxQ3j9FwdSViKX82M1u9qv7ZcQSnvrJ"] = 0
	assignedCandidates["121VhftSAygpEJZ6i9jGk9dPK5DE41KmqUhp2EFJXyyHNrQomEh7QW7icEh2zymM5953C2HN2kYYCbepkL37Qny6mgyxGiAxxqEEkCHnqK4FsNBDbGwojMmVCgURwmS1oRqpfoFYhKwFUTzjyYduxBNegdgUw9VEYXJe3xSMBib24oBRhjbmK7LZ5qxRqKMsV7Lp49rPi2QGiizqiVKSsaAD2QMHEaC5ZBsGJvtRRMdGarHfe6yLtWAvcPLpsuSkYHfNni5Nh61LT1LwYLqSwDqNVKC3ew6RfRvmFfFTPhBGUcbhRKog6djipwKW9RZ7jvpgSiwifLriiz8h25ziR3Guh7k1cgyYm6TxQHrrLBRUgXKGvLkeAEi9ThrCsuQnbE61re6UmjwCcbPd9j53fzVmm4JJ19kMsyGUL2m7obSsdxvM"] = 0
	assignedCandidates["121VhftSAygpEJZ6i9jGk9drax2iDha73FtSVnju8AYxEHLxLqrgcB5ocJPiJ3BBRcRgZ1TmTQxnEsSpSm3wEdaRd98Y7YEHBwrMsQdaPsA66MJeTxy9ZDpyAD82sWfYzHNA7Q8pjpBvCrvxKHQTz6NBRZXspvCtxozStN6mJMJWoMUyMBccZLgRMTN7dDXArcJVPtTVQWqjT15DToLbzY3qdnc1vdZDTq916qNdQ9PbCVwbswdqtdCxEwCoYo9uLS9gdkvJaJdU1wNuYFYvFgiAQFa6mgjNZWiDnLyYBtVX3VyfVGe4K8fRgG9bgj15ZG7UypBoQTjxxJJHDmMy23VHV3qSDr8bjLnhLVYgHmkpuHfhxFX2B9KXXhkc4XMgxxyC83HWaz2XvS1eNuTMVbKUd3tjCBkZQJszBDsKa5R7gJqH"] = 0
	assignedCandidates["121VhftSAygpEJZ6i9jGk9a5oQvgecAms7BtyyjGxpySigmDAdo8af26UKNmXYjNUhFVp4NN6RpRJFTGn57w7evPi3HkaF8ToXsCJ7ceaZ5p6hrCHLN6tKm5sjBEo3yusZZayurNMsrLGRhBE2i8Xhxkns8uN8c9WY5kcwSVsyD9f7399fMzRqtMB7TjyE4ad2KWDmteZZWuZB79jYB5wHWcxRyUxYaQ761gT8oMJ6FKr7wdPYAeuJ7Pai71Xi9YdPUDNQ1dZ7Uq1m7wxavKKax14Tuf9onu9oZDTeat7SNK1PDxjvf2uwkEmcHAp7qzp8c7igm8X6VjC6685gdThcdEHPxiwDsu3UxQyXK1fqwSHDHx3Ff7w5xKeDK8zJNghCLBbZ3HowQbT7hAKqu3N5puMKw5cjQJndA4trRw5yuzXxbf"] = 0
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

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

// TODO: @lam
// TESTCASE
// Add 4 more cases, Result and Expected Result MUST BE PRE-CALCULATED Somewhere ELSE
func TestShuffleShardCandidate(t *testing.T) {
	shuffledCandiates := shuffleShardCandidate(candidates, randomNumber)
	if !reflect.DeepEqual(shuffledCandiates, expectedShuffledCandidates) {
		log.Fatalf("Expect shuffled candidates to be %+v \n but get %+v", priorityKeys, shuffledCandiates)
	}
}

// TODO: @lam
// TESTCASE
// Add 4 more cases, Result and Expected Result MUST BE PRE-CALCULATED Somewhere ELSE
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

// TODO: @lam
// NOTICE: badPendingValidators is always empty
// TESTCASE
// 1. RETURN SAME-RESULT-AS-INPUT,NO-ERROR check offset is zero
// 2. RETURN SAME-RESULT-AS-INPUT,NO-ERROR check offset > maxCommittee -------- @hung ERROR offset > maxCommittee will return error
// 3. Push all pending validator to producer list,NO-ERROR (maxCommittee - len(currentGoodProducers)) >= offset
// 4. Push some pending validator to producer list and swap the remaining offset (maxCommittee - len(currentGoodProducers)) <= offset
// 5. Only swap,NO-ERROR len(goodPendingValidators) < offset, (maxCommittee - len(currentGoodProducers)) == 0 -------- @hung ERROR offset > maxCommittee will return error
// 6. Only swap,NO-ERROR len(goodPendingValidators) > offset, (maxCommittee - len(currentGoodProducers)) == 0
func Test_swap(t *testing.T) {
	type args struct {
		badPendingValidators  []string
		goodPendingValidators []string
		currentGoodProducers  []string
		currentBadProducers   []string
		maxCommittee          int
		offset                int
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		want1   []string
		want2   []string
		want3   []string
		wantErr bool
	}{
		{
			name: "swap case 1",
			args: args{
				goodPendingValidators: []string{"val12"},
				currentGoodProducers:  []string{"val1", "val2", "val3"},
				maxCommittee:          3,
				offset:                0,
			},
			want:    []string{"val12"},
			want1:   []string{"val1", "val2", "val3"},
			want2:   nil,
			want3:   []string{},
			wantErr: false,
		},
		{
			name: "swap case 2",
			args: args{
				goodPendingValidators: []string{"val12"},
				currentGoodProducers:  []string{"val1", "val2", "val3"},
				maxCommittee:          3,
				offset:                4,
			},
			want:    []string{"val12"},
			want1:   []string{"val1", "val2", "val3"},
			want2:   nil,
			want3:   []string{},
			wantErr: true,
		},
		{
			name: "swap case 3",
			args: args{
				goodPendingValidators: []string{"val12", "val22", "val32"},
				currentGoodProducers:  []string{"val1", "val2", "val3"},
				maxCommittee:          6,
				offset:                3,
			},
			want:    []string{},
			want1:   []string{"val1", "val2", "val3", "val12", "val22", "val32"},
			want2:   nil,
			want3:   []string{"val12", "val22", "val32"},
			wantErr: false,
		},
		{
			name: "swap case 4",
			args: args{
				goodPendingValidators: []string{"val12", "val22", "val32", "val42"},
				currentGoodProducers:  []string{"val1", "val2", "val3"},
				maxCommittee:          4,
				offset:                3,
			},
			want:    []string{"val42"},
			want1:   []string{"val3", "val12", "val22", "val32"},
			want2:   []string{"val1", "val2"},
			want3:   []string{"val12", "val22", "val32"},
			wantErr: false,
		},
		{
			name: "swap case 5",
			args: args{
				goodPendingValidators: []string{"val12", "val22", "val32"},
				currentGoodProducers:  []string{"val1", "val2", "val3"},
				maxCommittee:          3,
				offset:                4,
			},
			want:    []string{"val12", "val22", "val32"},
			want1:   []string{"val1", "val2", "val3"},
			want2:   nil,
			want3:   []string{},
			wantErr: true,
		},
		{
			name: "swap case 6",
			args: args{
				goodPendingValidators: []string{"val12", "val22", "val32", "val42"},
				currentGoodProducers:  []string{"val1", "val2", "val3"},
				maxCommittee:          3,
				offset:                3,
			},
			want:    []string{"val42"},
			want1:   []string{"val12", "val22", "val32"},
			want2:   []string{"val1", "val2", "val3"},
			want3:   []string{"val12", "val22", "val32"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3, err := swap(tt.args.badPendingValidators, tt.args.goodPendingValidators, tt.args.currentGoodProducers, tt.args.currentBadProducers, tt.args.maxCommittee, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("swap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("swap() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("swap() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("swap() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("swap() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}

// TODO: @lam
// NOTICE: badPendingValidators is always empty, ignore lines 283-304, code not reached in reality
// TESTCASE
// 1. Swap when currentValidator length reach maxCommittee len(currentGoodProducers) == maxCommittee, NO-ERROR
// 2 Swap when currentValidator length reach maxCommittee and offset > goodPendingValidatorsLen, NO-ERROR
// 3. Swap when currentValidator length reach maxCommittee and offset < goodPendingValidatorsLen, NO-ERROR
func TestSwapValidator(t *testing.T) {
	type args struct {
		pendingValidators  []string
		currentValidators  []string
		maxCommittee       int
		minCommittee       int
		offset             int
		producersBlackList map[string]uint8
		swapOffset         int
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		want1   []string
		want2   []string
		want3   []string
		wantErr bool
	}{
		{
			name: "swap case 1",
			args: args{
				pendingValidators: []string{"val12"},
				currentValidators: []string{"val1", "val2", "val3"},
				maxCommittee:      3,
				minCommittee:      1,
				offset:            1,
				swapOffset:        1,
			},
			want:    []string{},
			want1:   []string{"val2", "val3", "val12"},
			want2:   []string{"val1"},
			want3:   []string{"val12"},
			wantErr: false,
		},
		{
			name: "swap case 2",
			args: args{
				pendingValidators: []string{"val12"},
				currentValidators: []string{"val1", "val2", "val3"},
				maxCommittee:      4,
				minCommittee:      1,
				offset:            3,
				swapOffset:        1,
			},
			want:    []string{},
			want1:   []string{"val1", "val2", "val3", "val12"},
			want2:   []string{},
			want3:   []string{"val12"},
			wantErr: false,
		},
		{
			name: "swap case 3",
			args: args{
				pendingValidators: []string{"val12", "val22"},
				currentValidators: []string{"val1", "val2", "val3"},
				maxCommittee:      3,
				minCommittee:      1,
				offset:            1,
				swapOffset:        1,
			},
			want:    []string{"val22"},
			want1:   []string{"val2", "val3", "val12"},
			want2:   []string{"val1"},
			want3:   []string{"val12"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3, err := SwapValidator(tt.args.pendingValidators, tt.args.currentValidators, tt.args.maxCommittee, tt.args.minCommittee, tt.args.offset, tt.args.producersBlackList, tt.args.swapOffset)
			if (err != nil) != tt.wantErr {
				t.Errorf("SwapValidator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SwapValidator() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("SwapValidator() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("SwapValidator() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("SwapValidator() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}

// TODO: @lam
// TESTCASE
// 1. EMPTY-STRING-ARRAY,ERROR NOT PASS CONDITION len(removedValidators) > len(validators)
// 2. 3 cases remove success
func TestRemoveValidator(t *testing.T) {
	type args struct {
		validators        []string
		removedValidators []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "error case",
			args: args{
				validators:        []string{"val1", "val2", "val3"},
				removedValidators: []string{"val1", "val2", "val3", "val4"},
			},
			want:    []string{"val1", "val2", "val3"},
			wantErr: true,
		},
		{
			name: "happy case 1",
			args: args{
				validators:        []string{"val1", "val2", "val3"},
				removedValidators: []string{"val1"},
			},
			want:    []string{"val2", "val3"},
			wantErr: false,
		},
		{
			name: "happy case 2",
			args: args{
				validators:        []string{"val1", "val2", "val3"},
				removedValidators: []string{"val2"},
			},
			want:    []string{"val1", "val3"},
			wantErr: false,
		},
		{
			name: "happy case 3",
			args: args{
				validators:        []string{"val1", "val2", "val3"},
				removedValidators: []string{"val3"},
			},
			want:    []string{"val1", "val2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RemoveValidator(tt.args.validators, tt.args.removedValidators)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveValidator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveValidator() got = %v, want %v", got, tt.want)
			}
		})
	}
}

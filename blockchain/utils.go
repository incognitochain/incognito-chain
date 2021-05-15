package blockchain

import (
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/instruction"
)

func GetBeaconSwapInstructionKeyListV2(epoch uint64) ([]string, []string) {
	newCommittees := config.Param().GenesisParam.SelectBeaconNodeSerializedPubkeyV2[epoch]
	newRewardReceivers := config.Param().GenesisParam.SelectBeaconNodeSerializedPaymentAddressV2[epoch]

	// TODO - in next replacement of committee validator key -> need read oldCommittees from prev-committee instead of from genesis block
	oldCommittees := config.Param().GenesisParam.PreSelectBeaconNodeSerializedPubkey
	beaconSwapInstructionKeyListV2 := []string{instruction.SWAP_ACTION, strings.Join(newCommittees, ","), strings.Join(oldCommittees, ","), "beacon", "", "", strings.Join(newRewardReceivers, ",")}
	return beaconSwapInstructionKeyListV2, newCommittees
}

func GetShardSwapInstructionKeyListV2(epoch uint64, minCommitteeSize int, activeShard int) (map[byte][]string, map[byte][]string) {
	allShardSwapInstructionKeyListV2 := make(map[byte][]string)
	allShardNewKeyListV2 := make(map[byte][]string)
	selectShardNodeSerializedPubkeyV2 := config.Param().GenesisParam.SelectShardNodeSerializedPubkeyV2[epoch]
	selectShardNodeSerializedPaymentAddressV2 := config.Param().GenesisParam.SelectShardNodeSerializedPaymentAddressV2[epoch]
	preSelectShardNodeSerializedPubkey := config.Param().GenesisParam.PreSelectShardNodeSerializedPubkey
	shardCommitteeSize := minCommitteeSize
	for i := 0; i < activeShard; i++ {
		shardID := byte(i)
		newCommittees := selectShardNodeSerializedPubkeyV2[:shardCommitteeSize]

		// TODO - in next replacement of committee validator key -> need read oldCommittees from prev-committee instead of from genesis block
		oldCommittees := preSelectShardNodeSerializedPubkey[:shardCommitteeSize]
		// TODO

		newRewardReceiver := selectShardNodeSerializedPaymentAddressV2[:shardCommitteeSize]
		shardSwapInstructionKeyListV2 := []string{instruction.SWAP_ACTION, strings.Join(newCommittees, ","), strings.Join(oldCommittees, ","), "shard", strconv.Itoa(i), "", strings.Join(newRewardReceiver, ",")}
		allShardNewKeyListV2[shardID] = newCommittees
		selectShardNodeSerializedPubkeyV2 = selectShardNodeSerializedPubkeyV2[shardCommitteeSize:]
		preSelectShardNodeSerializedPubkey = preSelectShardNodeSerializedPubkey[shardCommitteeSize:]
		selectShardNodeSerializedPaymentAddressV2 = selectShardNodeSerializedPaymentAddressV2[shardCommitteeSize:]
		allShardSwapInstructionKeyListV2[shardID] = shardSwapInstructionKeyListV2
	}
	return allShardSwapInstructionKeyListV2, allShardNewKeyListV2
}

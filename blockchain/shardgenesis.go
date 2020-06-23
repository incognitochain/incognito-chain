package blockchain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/transaction"
)

func CreateShardGenesisBlock(
	version int,
	net uint16,
	genesisBlockTime string,
	icoParams *GenesisParams,
) *ShardBlock {
	body := ShardBody{}
	layout := "2006-01-02T15:04:05.000Z"
	str := genesisBlockTime
	genesisTime, err := time.Parse(layout, str)
	if err != nil {
		fmt.Println(err)
	}
	header := ShardHeader{
		Timestamp:         genesisTime.Unix(),
		Version:           version,
		BeaconHeight:      1,
		Epoch:             1,
		Round:             1,
		Height:            1,
		PreviousBlockHash: common.Hash{},
	}

	for _, tx := range icoParams.InitialIncognito {
		testSalaryTX, err := transaction.NewTransactionFromJsonBytes([]byte(tx))
		if err != nil {
			panic("Something is wrong when NewTransactionFromJsonBytes")
		}
		body.Transactions = append(body.Transactions, testSalaryTX)
	}

	block := &ShardBlock{
		Body:   body,
		Header: header,
	}

	return block
}

func GetShardSwapInstructionKeyListV2(genesisParams *GenesisParams, epoch uint64, minCommitteeSize int, activeShard int) (map[byte][]string, map[byte][]string) {
	allShardSwapInstructionKeyListV2 := make(map[byte][]string)
	allShardNewKeyListV2 := make(map[byte][]string)
	selectShardNodeSerializedPubkeyV2 := genesisParams.SelectShardNodeSerializedPubkeyV2[epoch]
	selectShardNodeSerializedPaymentAddressV2 := genesisParams.SelectShardNodeSerializedPaymentAddressV2[epoch]
	preSelectShardNodeSerializedPubkey := genesisParams.PreSelectShardNodeSerializedPubkey
	shardCommitteeSize := minCommitteeSize
	for i := 0; i < activeShard; i++ {
		shardID := byte(i)
		newCommittees := selectShardNodeSerializedPubkeyV2[:shardCommitteeSize]
		oldCommittees := preSelectShardNodeSerializedPubkey[:shardCommitteeSize]
		newRewardReceiver := selectShardNodeSerializedPaymentAddressV2[:shardCommitteeSize]
		shardSwapInstructionKeyListV2 := []string{SwapAction, strings.Join(newCommittees, ","), strings.Join(oldCommittees, ","), "shard", strconv.Itoa(i), "", strings.Join(newRewardReceiver, ",")}
		allShardNewKeyListV2[shardID] = newCommittees
		selectShardNodeSerializedPubkeyV2 = selectShardNodeSerializedPubkeyV2[shardCommitteeSize:]
		preSelectShardNodeSerializedPubkey = preSelectShardNodeSerializedPubkey[shardCommitteeSize:]
		selectShardNodeSerializedPaymentAddressV2 = selectShardNodeSerializedPaymentAddressV2[shardCommitteeSize:]
		allShardSwapInstructionKeyListV2[shardID] = shardSwapInstructionKeyListV2
	}
	return allShardSwapInstructionKeyListV2, allShardNewKeyListV2
}

package blockchain

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/transaction"
)

func CreateGenesisBeaconBlock() *types.BeaconBlock {
	inst := [][]string{}
	shardAutoStaking := []string{}
	beaconAutoStaking := []string{}
	txStakes := []string{}
	param := config.Param()
	for i := 0; i < len(param.GenesisParam.PreSelectBeaconNodeSerializedPubkey); i++ {
		beaconAutoStaking = append(beaconAutoStaking, "false")
		txStakes = append(txStakes, param.GenesisParam.TxStake)
	}
	for i := 0; i < len(param.GenesisParam.PreSelectShardNodeSerializedPubkey); i++ {
		shardAutoStaking = append(shardAutoStaking, "false")
		txStakes = append(txStakes, param.GenesisParam.TxStake)
	}
	// build validator beacon
	// test generate public key in utility/generateKeys
	beaconAssignInstruction := []string{instruction.STAKE_ACTION}
	beaconAssignInstruction = append(beaconAssignInstruction, strings.Join(param.GenesisParam.PreSelectBeaconNodeSerializedPubkey[:], ","))
	beaconAssignInstruction = append(beaconAssignInstruction, "beacon")
	beaconAssignInstruction = append(beaconAssignInstruction, strings.Join(txStakes[:len(param.GenesisParam.PreSelectBeaconNodeSerializedPubkey)], ","))
	beaconAssignInstruction = append(beaconAssignInstruction, strings.Join(param.GenesisParam.PreSelectBeaconNodeSerializedPaymentAddress[:], ","))
	beaconAssignInstruction = append(beaconAssignInstruction, strings.Join(beaconAutoStaking[:], ","))

	shardAssignInstruction := []string{instruction.STAKE_ACTION}
	shardAssignInstruction = append(shardAssignInstruction, strings.Join(param.GenesisParam.PreSelectShardNodeSerializedPubkey[:], ","))
	shardAssignInstruction = append(shardAssignInstruction, "shard")
	shardAssignInstruction = append(shardAssignInstruction, strings.Join(txStakes[len(param.GenesisParam.PreSelectBeaconNodeSerializedPubkey):], ","))
	shardAssignInstruction = append(shardAssignInstruction, strings.Join(param.GenesisParam.PreSelectShardNodeSerializedPaymentAddress[:], ","))
	shardAssignInstruction = append(shardAssignInstruction, strings.Join(shardAutoStaking[:], ","))

	inst = append(inst, beaconAssignInstruction)
	inst = append(inst, shardAssignInstruction)

	// init network param
	inst = append(inst, []string{instruction.SET_ACTION, "randomnumber", strconv.Itoa(int(0))})

	layout := "2006-01-02T15:04:05.000Z"
	str := param.GenesisParam.BlockTimestamp
	genesisTime, err := time.Parse(layout, str)

	if err != nil {
		Logger.log.Error(err)
	}
	body := types.BeaconBody{ShardState: nil, Instructions: inst}
	header := types.BeaconHeader{
		Timestamp:                       genesisTime.Unix(),
		Version:                         1,
		Epoch:                           1,
		Height:                          1,
		Round:                           1,
		PreviousBlockHash:               common.Hash{},
		BeaconCommitteeAndValidatorRoot: common.Hash{},
		BeaconCandidateRoot:             common.Hash{},
		ShardCandidateRoot:              common.Hash{},
		ShardCommitteeAndValidatorRoot:  common.Hash{},
		ShardStateHash:                  common.Hash{},
		InstructionHash:                 common.Hash{},
	}

	block := &types.BeaconBlock{
		Body:   body,
		Header: header,
	}

	return block
}

func CreateGenesisShardBlock() *types.ShardBlock {
	body := types.ShardBody{}
	layout := "2006-01-02T15:04:05.000Z"
	str := config.Param().GenesisParam.BlockTimestamp
	genesisTime, err := time.Parse(layout, str)
	if err != nil {
		Logger.log.Error(err)
	}
	header := types.ShardHeader{
		Timestamp:         genesisTime.Unix(),
		Version:           1,
		BeaconHeight:      1,
		Epoch:             1,
		Round:             1,
		Height:            1,
		PreviousBlockHash: common.Hash{},
	}

	for _, v := range config.Param().GenesisParam.InitialIncognito {
		tx, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		initSalaryTx := new(transaction.TxVersion1)
		err = initSalaryTx.UnmarshalJSON(tx)
		if err != nil {
			panic(err)
		}
		body.Transactions = append(body.Transactions, initSalaryTx)
	}

	block := &types.ShardBlock{
		Body:   body,
		Header: header,
	}

	return block
}

var genesisBeaconBlock *types.BeaconBlock
var genesisShardBlock *types.ShardBlock

func CreateGenesisBlocks() {
	genesisBeaconBlock = CreateGenesisBeaconBlock()
	genesisShardBlock = CreateGenesisShardBlock()
}

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

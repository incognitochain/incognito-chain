package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"strings"
)

func (blockchain *BlockChain) getUpdatedProducersBlackListV2(slashStateDB *statedb.StateDB, isBeacon bool, shardID int, committee []string, beaconHeight uint64) (map[string]uint8, error) {
	producersBlackList := statedb.GetProducersBlackList(slashStateDB, beaconHeight)
	if isBeacon {
		punishedProducersFinished := []string{}
		for producer, punishedEpoches := range producersBlackList {
			if punishedEpoches == 1 {
				punishedProducersFinished = append(punishedProducersFinished, producer)
			}
		}
		for _, producer := range punishedProducersFinished {
			delete(producersBlackList, producer)
		}
	}
	badProducersWithPunishment := blockchain.buildBadProducersWithPunishment(isBeacon, shardID, committee)
	for producer, punishedEpoches := range badProducersWithPunishment {
		epoches, found := producersBlackList[producer]
		if !found || epoches < punishedEpoches {
			producersBlackList[producer] = punishedEpoches
		}
	}
	return sortMapStringUint8Keys(producersBlackList), nil
}

func (blockchain *BlockChain) processForSlashingV2(slashStateDB *statedb.StateDB, beaconBlock *BeaconBlock) error {
	var err error
	punishedProducersFinished := []string{}
	fliterPunishedProducersFinished := []string{}
	beaconHeight := beaconBlock.GetHeight()
	producersBlackList := statedb.GetProducersBlackList(slashStateDB, beaconHeight-1)
	chainParamEpoch := blockchain.config.ChainParams.Epoch
	newBeaconHeight := beaconBlock.GetHeight()
	if newBeaconHeight%uint64(chainParamEpoch) == 0 { // end of epoch
		for producer := range producersBlackList {
			producersBlackList[producer]--
			if producersBlackList[producer] == 0 {
				punishedProducersFinished = append(punishedProducersFinished, producer)
			}
		}
		for _, producer := range punishedProducersFinished {
			delete(producersBlackList, producer)
		}
	}

	for _, inst := range beaconBlock.GetInstructions() {
		if len(inst) == 0 {
			continue
		}
		if inst[0] != SwapAction {
			continue
		}
		badProducersWithPunishmentBytes := []byte{}
		if len(inst) == 6 && inst[3] == "shard" {
			badProducersWithPunishmentBytes = []byte(inst[5])
		}
		if len(inst) == 5 && inst[3] == "beacon" {
			badProducersWithPunishmentBytes = []byte(inst[4])
		}
		if len(badProducersWithPunishmentBytes) == 0 {
			continue
		}

		var badProducersWithPunishment map[string]uint8
		err = json.Unmarshal(badProducersWithPunishmentBytes, &badProducersWithPunishment)
		if err != nil {
			return err
		}
		for producer, punishedEpoches := range badProducersWithPunishment {
			epoches, found := producersBlackList[producer]
			if !found || epoches < punishedEpoches {
				producersBlackList[producer] = punishedEpoches
			}
		}
	}
	for _, punishedProducerFinished := range punishedProducersFinished {
		flag := false
		for producerBlaskList, _ := range producersBlackList {
			if strings.Compare(producerBlaskList, punishedProducerFinished) == 0 {
				flag = true
				break
			}
		}
		if !flag {
			fliterPunishedProducersFinished = append(fliterPunishedProducersFinished, punishedProducerFinished)
		}
	}
	statedb.RemoveProducerBlackList(slashStateDB, fliterPunishedProducersFinished)
	err = statedb.StoreProducersBlackList(slashStateDB, beaconHeight, producersBlackList)
	return err
}

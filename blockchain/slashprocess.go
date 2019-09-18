package blockchain

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

func (blockchain *BlockChain) buildBadProducersWithPunishment(
	isBeacon bool,
	shardID int,
	committee []string,
) map[string]uint8 {
	slashLevels := blockchain.config.ChainParams.SlashLevels
	numOfBlocksByProducers := map[string]uint64{}
	if isBeacon {
		numOfBlocksByProducers = blockchain.BestState.Beacon.NumOfBlocksByProducers
	} else {
		numOfBlocksByProducers = blockchain.BestState.Shard[byte(shardID)].NumOfBlocksByProducers
	}
	// numBlkPerEpoch := blockchain.config.ChainParams.Epoch
	numBlkPerEpoch := uint64(0)
	for _, numBlk := range numOfBlocksByProducers {
		numBlkPerEpoch += numBlk
	}
	expectedNumBlkByEachProducer := numBlkPerEpoch / uint64(len(committee))
	badProducersWithPunishment := map[string]uint8{}
	for producer, numBlk := range numOfBlocksByProducers {
		missingPercent := uint8((numBlk * 100) / expectedNumBlkByEachProducer)
		var selectedSlLev *SlashLevel
		for _, slLev := range slashLevels {
			if missingPercent >= slLev.MinRange {
				selectedSlLev = &slLev
			}
		}
		badProducersWithPunishment[producer] = selectedSlLev.PunishedEpoches
	}
	return badProducersWithPunishment
}

func (blockchain *BlockChain) getUpdatedProducersBlackList(
	isBeacon bool,
	shardID int,
	committee []string,
) (map[string]uint8, error) {
	db := blockchain.GetDatabase()
	producersBlackList, err := db.GetProducersBlackList()
	if err != nil {
		return nil, err
	}
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
	return producersBlackList, nil
}

func (blockchain *BlockChain) processForSlashing(block *BeaconBlock) error {
	db := blockchain.GetDatabase()
	chainParamEpoch := blockchain.config.ChainParams.Epoch
	newBeaconHeight := block.GetHeight()
	updatedProducersBlackList := map[string]uint8{}
	var err error
	if newBeaconHeight%uint64(chainParamEpoch) == 0 { // process for beacon swap
		beaconBestState := blockchain.BestState.Beacon
		beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconCommittee)
		if err != nil {
			return err
		}
		updatedProducersBlackList, err = blockchain.getUpdatedProducersBlackList(true, -1, beaconCommitteeStr)
		if err != nil {
			return err
		}
	}
	// process for shards swap
	for _, inst := range block.GetInstructions() {
		if len(inst) != 6 { // length  of swap instruction should be 6
			continue
		}
		if inst[0] != SwapAction {
			continue
		}
		var badProducersWithPunishment map[string]uint8
		err = json.Unmarshal([]byte(inst[5]), &badProducersWithPunishment)
		if err != nil {
			return err
		}
		for producer, punishedEpoches := range badProducersWithPunishment {
			epoches, found := updatedProducersBlackList[producer]
			if !found || epoches < punishedEpoches {
				updatedProducersBlackList[producer] = punishedEpoches
			}
		}
	}
	err = db.StoreProducersBlackList(updatedProducersBlackList)
	return err
}

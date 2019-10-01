package blockchain

import (
	"encoding/json"
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
	badProducersWithPunishment := make(map[string]uint8)
	committeeLen := len(committee)
	if committeeLen == 0 {
		return badProducersWithPunishment
	}
	expectedNumBlkByEachProducer := numBlkPerEpoch / uint64(committeeLen)

	if expectedNumBlkByEachProducer == 0 {
		return badProducersWithPunishment
	}
	// for producer, numBlk := range numOfBlocksByProducers {
	for _, producer := range committee {
		numBlk, found := numOfBlocksByProducers[producer]
		if !found {
			numBlk = 0
		}
		if numBlk >= expectedNumBlkByEachProducer {
			continue
		}
		missingPercent := uint8((-(numBlk - expectedNumBlkByEachProducer) * 100) / expectedNumBlkByEachProducer)
		var selectedSlLev *SlashLevel
		for _, slLev := range slashLevels {
			if missingPercent >= slLev.MinRange {
				selectedSlLev = &slLev
			}
		}
		if selectedSlLev != nil {
			badProducersWithPunishment[producer] = selectedSlLev.PunishedEpoches
		}
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
	var err error
	db := blockchain.GetDatabase()
	producersBlackList, err := db.GetProducersBlackList()
	if err != nil {
		return err
	}
	chainParamEpoch := blockchain.config.ChainParams.Epoch
	newBeaconHeight := block.GetHeight()
	if newBeaconHeight%uint64(chainParamEpoch) == 0 { // end of epoch
		punishedProducersFinished := []string{}
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

	for _, inst := range block.GetInstructions() {
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
	err = db.StoreProducersBlackList(producersBlackList)
	return err
}

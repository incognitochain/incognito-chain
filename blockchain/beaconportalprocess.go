package blockchain

import (
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"strconv"
)

func (blockchain *BlockChain) processPortalInstructions(block *BeaconBlock, bd *[]database.BatchData) error {
	beaconHeight := block.Header.Height - 1
	db := blockchain.GetDatabase()
	currentPortalState, err := InitCurrentPortalStateFromDB(db, beaconHeight)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not Portal instruction
		}
		var err error
		switch inst[0] {
		case strconv.Itoa(metadata.PortalCustodianDepositMeta):
			err = blockchain.processPortalCustodianDeposit(beaconHeight, inst, currentPortalState)
		}
		//case //
		//todo:

		if err != nil {
			Logger.log.Error(err)
			return nil
		}
	}
	// store updated currentPortalState to leveldb with new beacon height
	err = storePortalStateToDB(db, beaconHeight+1, currentPortalState)
	if err != nil {
		Logger.log.Error(err)
	}
	return nil
}

func InitCurrentPortalStateFromDB(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (*CurrentPortalState, error) {
	/*waitingPDEContributions, err := getWaitingPortalContributions(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	pdePoolPairs, err := getPortalPoolPair(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	pdeShares, err := getPortalShares(db, beaconHeight)
	if err != nil {
		return nil, err
	}*/

	return &CurrentPortalState{
		CustodianPoolState: waitingPDEContributions,
		PortingRequests:            pdePoolPairs,
		RedeemRequests:               pdeShares,
	}, nil
}



// todo
func (blockchain *BlockChain) processPortalCustodianDeposit(
	beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
		if currentPortalState == nil {
			Logger.log.Errorf("current portal state is nil")
			return nil
		}
		return nil
}
package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/tendermint/tendermint/types"
	"strconv"
)

func (blockchain *BlockChain) processRelayingInstructions(block *BeaconBlock) error {
	beaconHeight := block.Header.Height - 1
	db := blockchain.GetDatabase()

	relayingState, err := blockchain.InitRelayingHeaderChainStateFromDB(db, beaconHeight)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}

	// because relaying instructions in received beacon block were sorted already as desired so dont need to do sorting again over here
	for _, inst := range block.Body.Instructions {
		if len(inst) < 4 {
			continue // Not error, just not relaying instruction
		}
		var err error
		switch inst[0] {
		case strconv.Itoa(metadata.RelayingBNBHeaderMeta):
			err = blockchain.processRelayingBNBHeaderInst(inst, relayingState)
		case strconv.Itoa(metadata.RelayingBTCHeaderMeta):
			err = blockchain.processRelayingBTCHeaderInst(inst, relayingState)
		}
		if err != nil {
			Logger.log.Error(err)
		}
	}

	// store updated relayingState to leveldb with new beacon height
	err = storeRelayingHeaderStateToDB(db, beaconHeight+1, relayingState)
	if err != nil {
		Logger.log.Error(err)
	}
	return nil
}

func (blockchain *BlockChain) processRelayingBTCHeaderInst(
	instruction []string,
	relayingState *RelayingHeaderChainState,
) error {
	Logger.log.Info("[BTC Relaying] - Processing processRelayingBTCHeaderInst...")
	btcHeaderChain := relayingState.BTCHeaderChain
	if btcHeaderChain == nil {
		return errors.New("[processRelayingBTCHeaderInst] BTC Header chain instance should not be nil")
	}

	if len(instruction) != 4 {
		return nil // skip the instruction
	}

	var relayingHeaderContent metadata.RelayingHeaderContent
	err := json.Unmarshal([]byte(instruction[3]), &relayingHeaderContent)
	if err != nil {
		return err
	}

	headerBytes, err := base64.StdEncoding.DecodeString(relayingHeaderContent.Header)
	if err != nil {
		return err
	}
	var msgBlk *wire.MsgBlock
	err = json.Unmarshal(headerBytes, &msgBlk)
	if err != nil {
		return err
	}
	block := btcutil.NewBlock(msgBlk)
	isMainChain, isOrphan, err := btcHeaderChain.ProcessBlockV2(block, btcrelaying.BFNone)
	if err != nil {
		Logger.log.Errorf("ProcessBlock fail with error: %v", err)
		return err
	}
	Logger.log.Infof("ProcessBlock (%s) success with result: isMainChain: %v, isOrphan: %v", block.Hash(), isMainChain, isOrphan)
	return nil
}

func (blockchain *BlockChain) processRelayingBNBHeaderInst(
	instructions []string,
	relayingState *RelayingHeaderChainState,
) error {
	if relayingState == nil {
		Logger.log.Errorf("relaying header state is nil")
		return errors.New("relaying header state is nil")
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}
	db := blockchain.GetDatabase()

	// unmarshal instructions content
	var actionData metadata.RelayingHeaderContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		return err
	}

	var header types.Block
	headerBytes, err := base64.StdEncoding.DecodeString(actionData.Header)
	if err != nil {
		return err
	}
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		return err
	}

	reqStatus := instructions[2]
	if reqStatus == common.RelayingHeaderUnconfirmedAcceptedChainStatus {
		//update relaying state
		relayingState.BNBHeaderChain.UnconfirmedBlocks = append(relayingState.BNBHeaderChain.UnconfirmedBlocks, &header)

	} else if reqStatus == common.RelayingHeaderConfirmedAcceptedChainStatus {
		// check newLatestBNBHeader is genesis header or not
		genesisHeaderHeight, _ := bnb.GetGenesisBNBHeaderBlockHeight(blockchain.config.ChainParams.BNBRelayingHeaderChainID)

		if header.Header.Height == genesisHeaderHeight {
			relayingState.BNBHeaderChain.LatestBlock = &header

			// store new confirmed header into db
			newConfirmedheader := relayingState.BNBHeaderChain.LatestBlock
			// don't need to store Data and Evidence into db
			newConfirmedheader.Data = types.Data{}
			newConfirmedheader.Evidence = types.EvidenceData{}
			newConfirmedheaderBytes, _ := json.Marshal(newConfirmedheader)

			err := rawdbv2.StoreRelayingBNBHeaderChain(db, uint64(newConfirmedheader.Height), newConfirmedheaderBytes)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occured while storing new confirmed header: %+v", err)
				return err
			}
			return nil
		}

		// get new latest header
		blockIDNewLatestHeader := header.Header.LastBlockID
		for _, header := range relayingState.BNBHeaderChain.UnconfirmedBlocks {
			if bytes.Equal(header.Hash().Bytes(), blockIDNewLatestHeader.Hash) {
				relayingState.BNBHeaderChain.LatestBlock = header
				break
			}
		}

		//update relaying state
		relayingState.BNBHeaderChain.UnconfirmedBlocks = []*types.Block{&header}

		// store new confirmed header into db
		newConfirmedheader := relayingState.BNBHeaderChain.LatestBlock
		newConfirmedheader.Data = types.Data{}
		newConfirmedheader.Evidence = types.EvidenceData{}
		newConfirmedheaderBytes, _ := json.Marshal(newConfirmedheader)

		err := rawdbv2.StoreRelayingBNBHeaderChain(db, uint64(newConfirmedheader.Height), newConfirmedheaderBytes)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while storing new confirmed header: %+v", err)
			return err
		}
	}

	return nil
}

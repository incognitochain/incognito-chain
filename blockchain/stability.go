package blockchain

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

// build actions from txs and ins at shard
func buildStabilityActions(
	txs []metadata.Transaction,
	bc *BlockChain,
	shardID byte,
) ([][]string, error) {
	actions := [][]string{}
	for _, tx := range txs {
		meta := tx.GetMetadata()
		if meta != nil {
			actionPairs, err := meta.BuildReqActions(tx, bc, shardID)
			if err != nil {
				continue
			}
			actions = append(actions, actionPairs...)
		}
	}
	return actions, nil
}

// pickPubkeyRootInstruction finds all instructions of type BeaconPubkeyRootMeta returns them to save in bridge block
// These instructions contain merkle root of beacon/bridge committee's pubkey
func pickPubkeyRootInstruction(
	beaconBlocks []*BeaconBlock,
) [][]string {
	beaconType := strconv.Itoa(metadata.BeaconPubkeyRootMeta)
	bridgeType := strconv.Itoa(metadata.BridgePubkeyRootMeta)
	commPubkeyInst := [][]string{}
	for _, block := range beaconBlocks {
		for _, inst := range block.Body.Instructions {
			instType := inst[0]
			if instType != beaconType && instType != bridgeType {
				continue
			}
			fmt.Printf("[db] found root inst: %v, beacon block %d\n", inst, block.Header.Height)
			commPubkeyInst = append(commPubkeyInst, inst)
		}
	}
	return commPubkeyInst
}

// build instructions at beacon chain before syncing to shards
func (blockChain *BlockChain) buildStabilityInstructions(
	shardID byte,
	shardBlockInstructions [][]string,
	beaconBestState *BestStateBeacon,
) ([][]string, error) {
	instructions := [][]string{}
	for _, inst := range shardBlockInstructions {
		if len(inst) == 0 {
			continue
		}
		if inst[0] == StakeAction || inst[0] == SwapAction || inst[0] == RandomAction {
			continue
		}

		newInst := [][]string{}
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			return [][]string{}, err
		}
		switch metaType {
		case metadata.IssuingRequestMeta, metadata.ContractingRequestMeta:
			newInst = [][]string{inst}

		default:
			continue
		}

		if err != nil {
			Logger.log.Error(err)
			continue
		}
		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}
	return instructions, nil
}

func (blockgen *BlkTmplGenerator) buildResponseTxsFromBeaconInstructions(
	beaconBlocks []*BeaconBlock,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) ([]metadata.Transaction, error) {
	resTxs := []metadata.Transaction{}
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			if l[0] == SwapAction {
				//fmt.Println("SA: swap instruction ", l, beaconBlock.Header.Height, blockgen.chain.BestState.Beacon.GetShardCommittee())
				for _, v := range strings.Split(l[2], ",") {
					tx, err := blockgen.buildReturnStakingAmountTx(v, producerPrivateKey)
					if err != nil {
						Logger.log.Error("SA:", err)
						continue
					}
					resTxs = append(resTxs, tx)
				}

			}
			shardToProcess, err := strconv.Atoi(l[1])
			if err != nil {
				continue
			}
			if shardToProcess == int(shardID) {
				// metaType, err := strconv.Atoi(l[0])
				// if err != nil {
				// 	return nil, err
				// }
				// var newIns []string
				// switch metaType {
				// case metadata.BeaconSalaryRequestMeta:
				// 	txs, err := blockgen.buildBeaconSalaryRes(l[0], l[3], producerPrivateKey)
				// 	if err != nil {
				// 		return nil, err
				// 	}
				// 	resTxs = append(resTxs, txs...)
				// }

			}
			if l[0] == StakeAction || l[0] == RandomAction {
				continue
			}
			if len(l) <= 2 {
				continue
			}
		}
	}
	return resTxs, nil
}

func (blockgen *BlkTmplGenerator) buildStabilityResponseTxsAtShardOnly(
	txs []metadata.Transaction,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) ([]metadata.Transaction, error) {
	respTxs := []metadata.Transaction{}
	removeIds := []int{}
	for i, tx := range txs {
		var respTx metadata.Transaction
		var err error

		switch tx.GetMetadataType() {
		case metadata.IssuingRequestMeta:
			respTx, err = blockgen.buildIssuanceTx(tx, producerPrivateKey, shardID)
		}

		if err != nil {
			// Remove this tx if cannot create corresponding response
			removeIds = append(removeIds, i)
		} else if respTx != nil {
			respTxs = append(respTxs, respTx)
		}
	}
	return respTxs, nil
}

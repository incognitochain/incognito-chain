package blockchain

import (
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
	producerAddress *privacy.PaymentAddress,
	shardBlockHeight uint64,
	beaconBlocks []*BeaconBlock,
	beaconHeight uint64,
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

		contentStr := inst[1]
		newInst := [][]string{}
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			return [][]string{}, err
		}
		switch metaType {
		case metadata.IssuingRequestMeta, metadata.ContractingRequestMeta:
			newInst = [][]string{inst}
		case metadata.IssuingETHRequestMeta:
			newInst, err = buildInstructionsForETHIssuingReq(contentStr)
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
			// shardToProcess, err := strconv.Atoi(l[1])
			// if err != nil {
			// 	continue
			// }
			// if shardToProcess == int(shardID) {
			// 	// metaType, err := strconv.Atoi(l[0])
			// 	// if err != nil {
			// 	// 	return nil, err
			// 	// }
			// 	// var newIns []string
			// 	// switch metaType {
			// 	// case metadata.BeaconSalaryRequestMeta:
			// 	// 	txs, err := blockgen.buildBeaconSalaryRes(l[0], l[3], producerPrivateKey)
			// 	// 	if err != nil {
			// 	// 		return nil, err
			// 	// 	}
			// 	// 	resTxs = append(resTxs, txs...)
			// 	// }

			// }
			// if l[0] == StakeAction || l[0] == RandomAction {
			// 	continue
			// }
			// if len(l) <= 2 {
			// 	continue
			// }
			metaType, err := strconv.Atoi(l[0])
			if err != nil {
				return nil, err
			}
			var newTx metadata.Transaction
			switch metaType {
			case metadata.IssuingETHRequestMeta:
				bridgeShardIDStr, _ := strconv.Atoi(l[1])
				newTx, err = blockgen.buildETHIssuanceTx(l[3], producerPrivateKey, byte(bridgeShardIDStr))

			default:
				continue
			}
			if err != nil {
				return nil, err
			}
			if newTx != nil {
				resTxs = append(resTxs, newTx)
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
	var relayingRewardTx metadata.Transaction
	var maxHeaderLen int

	for i, tx := range txs {
		var respTx metadata.Transaction
		var err error

		switch tx.GetMetadataType() {
		case metadata.IssuingRequestMeta:
			respTx, err = blockgen.buildIssuanceTx(tx, producerPrivateKey, shardID)
		case metadata.ETHHeaderRelayingMeta:
			relayingRewardTx, maxHeaderLen, err = blockgen.buildETHHeaderRelayingRewardTx(tx, producerPrivateKey, relayingRewardTx, maxHeaderLen)
		}

		if err != nil {
			// Remove this tx if cannot create corresponding response
			removeIds = append(removeIds, i)
		} else if respTx != nil {
			respTxs = append(respTxs, respTx)
		}
	}
	if relayingRewardTx != nil {
		respTxs = append(respTxs, relayingRewardTx)
	}
	return respTxs, nil
}

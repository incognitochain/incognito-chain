package blockchain

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

// FlattenAndConvertStringInst receives a slice of insts; concats each inst ([]string) and converts to []byte to build merkle tree later
func FlattenAndConvertStringInst(insts [][]string) [][]byte {
	flattenInsts := [][]byte{}
	for _, inst := range insts {
		flattenInsts = append(flattenInsts, DecodeInstruction(inst))
	}
	return flattenInsts
}

// decodeInstruction appends all part of an instruction and decode them if necessary (for special instruction that needed to be decoded before submitting to Ethereum)
func DecodeInstruction(inst []string) []byte {
	flatten := []byte{}
	switch inst[0] {
	case strconv.Itoa(metadata.BeaconPubkeyRootMeta), strconv.Itoa(metadata.BridgePubkeyRootMeta):
		flatten = decodeLastFieldOnly(inst)

	case strconv.Itoa(metadata.BurningConfirmMeta):
		flatten = decodeBurningConfirmInst(inst)

	default:
		for _, part := range inst {
			flatten = append(flatten, []byte(part)...)
		}
	}
	return flatten
}

// decodeLastFieldOnly flattens all parts of an instruction, decode the last and concats them
func decodeLastFieldOnly(inst []string) []byte {
	flatten := []byte{}
	for _, part := range inst[:len(inst)-1] {
		flatten = append(flatten, []byte(part)...)
	}
	// Special case: instruction storing merkle root of beacon/bridge's committee => decode the merkle root and sign on that instead
	// We need to decode and submit the raw merkle root to Ethereum because we can't decode it on smart contract
	if pk, _, err := (base58.Base58Check{}).Decode(inst[2]); err == nil {
		flatten = append(flatten, pk...)
	} else {
		flatten = append(flatten, []byte(inst[len(inst)-1])...)
	}
	return flatten
}

// decodeBurningConfirmInst decodes and flattens a BurningConfirm instruction
func decodeBurningConfirmInst(inst []string) []byte {
	metaType := []byte(inst[0])
	shardID := []byte(inst[1])
	tokenID, _ := common.NewHashFromStr(inst[2])
	remoteAddr, _ := decodeRemoteAddr(inst[3])
	amount, _, _ := base58.Base58Check{}.Decode(inst[4])
	txID, _ := common.NewHashFromStr(inst[5])
	flatten := []byte{}
	flatten = append(flatten, metaType...)
	flatten = append(flatten, shardID...)
	flatten = append(flatten, tokenID[:]...)
	flatten = append(flatten, remoteAddr...)
	flatten = append(flatten, toBytes32BigEndian(amount)...)
	flatten = append(flatten, txID[:]...)
	return flatten
}

// decodeRemoteAddr converts address string to 32 bytes slice
func decodeRemoteAddr(addr string) ([]byte, error) {
	remoteAddr, err := hex.DecodeString(addr)
	if err != nil {
		return nil, err
	}
	addrFixedLen := [32]byte{}
	copy(addrFixedLen[32-len(remoteAddr):], remoteAddr)
	return addrFixedLen[:], nil
}

// toBytes32BigEndian converts a Big.Int bytes to uint256 for of Ethereum
func toBytes32BigEndian(b []byte) []byte {
	a := [32]byte{}
	copy(a[32-len(b):], b)
	return a[:]
}

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

// pickInstructionWithType finds all instructions of a specific type in a list
func pickInstructionWithType(
	insts [][]string,
	typeToFind string,
) [][]string {
	found := [][]string{}
	for _, inst := range insts {
		instType := inst[0]
		if instType != typeToFind {
			continue
		}
		found = append(found, inst)
	}
	return found
}

// pickInstructionFromBeaconBlocks extracts all instructions of a specific type
func pickInstructionFromBeaconBlocks(beaconBlocks []*BeaconBlock, instType string) [][]string {
	insts := [][]string{}
	for _, block := range beaconBlocks {
		found := pickInstructionWithType(block.Body.Instructions, instType)
		if len(found) > 0 {
			insts = append(insts, found...)
		}
	}
	return insts
}

// pickBeaconPubkeyRootInstruction finds all BeaconPubkeyRootMeta instructions
// These instructions contain merkle root of beacon committee's pubkey
func pickBeaconPubkeyRootInstruction(
	beaconBlocks []*BeaconBlock,
) [][]string {
	instType := strconv.Itoa(metadata.BeaconPubkeyRootMeta)
	return pickInstructionFromBeaconBlocks(beaconBlocks, instType)
}

// pickBurningConfirmInstruction finds all BurningConfirmMeta instructions
func pickBurningConfirmInstruction(
	beaconBlocks []*BeaconBlock,
) [][]string {
	instType := strconv.Itoa(metadata.BurningConfirmMeta)
	return pickInstructionFromBeaconBlocks(beaconBlocks, instType)
}

// pickBridgePubkeyRootInstruction finds all BridgePubkeyRootMeta instructions
// These instructions contain merkle root of bridge committee's pubkey
func pickBridgePubkeyRootInstruction(
	block *ShardToBeaconBlock,
) [][]string {
	shardType := strconv.Itoa(metadata.BridgePubkeyRootMeta)
	return pickInstructionWithType(block.Instructions, shardType)
}

// parsePubkeysAndBuildMerkleRoot returns the merkle root of a list of validators'pubkey stored as string
func parsePubkeysAndBuildMerkleRoot(vals []string) []byte {
	pks := [][]byte{}
	for _, val := range vals {
		pk, _, _ := base58.Base58Check{}.Decode(val)
		// TODO(@0xbunyip): handle error
		pks = append(pks, pk)
	}
	return GetKeccak256MerkleRoot(pks)
}

// build instructions at beacon chain before syncing to shards
func buildBeaconPubkeyRootInstruction(currentValidators []string) []string {
	beaconCommRoot := parsePubkeysAndBuildMerkleRoot(currentValidators)
	fmt.Printf("[db] added beaconCommRoot: %x\n", beaconCommRoot)

	shardID := byte(1) // TODO(@0xbunyip): change to bridge shardID
	instContent := base58.Base58Check{}.Encode(beaconCommRoot, 0x00)
	return []string{
		strconv.Itoa(metadata.BeaconPubkeyRootMeta),
		strconv.Itoa(int(shardID)),
		instContent,
	}
}

func buildBridgePubkeyRootInstruction(currentValidators []string) []string {
	bridgeCommRoot := parsePubkeysAndBuildMerkleRoot(currentValidators)

	shardID := byte(1) // TODO(@0xbunyip): change to bridge shardID
	instContent := base58.Base58Check{}.Encode(bridgeCommRoot, 0x00)
	return []string{
		strconv.Itoa(metadata.BridgePubkeyRootMeta),
		strconv.Itoa(int(shardID)),
		instContent,
	}
}

// build instructions at beacon chain before syncing to shards
func (blockChain *BlockChain) buildStabilityInstructions(
	shardID byte,
	shardBlockInstructions [][]string,
	beaconBestState *BestStateBeacon,
) ([][]string, error) {
	instructions := [][]string{}
	for _, inst := range shardBlockInstructions {
		if inst[0] == strconv.Itoa(metadata.BurningRequestMeta) {
			fmt.Printf("[db] shardBlockInst: %s\n", inst)
		}
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

		case metadata.BurningRequestMeta:
			fmt.Printf("[db] found BurnningRequest meta: %d\n", metaType)
			burningConfirm, err := buildBurningConfirmInst(inst)
			if err != nil {
				return [][]string{}, err
			}
			newInst = [][]string{inst, burningConfirm}

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
				newTx, err = blockgen.buildETHIssuanceTx(l[3], producerPrivateKey, byte(bridgeShardIDStr), shardID)

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

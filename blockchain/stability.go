package blockchain

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
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
	case strconv.Itoa(metadata.BeaconSwapConfirmMeta), strconv.Itoa(metadata.BridgeSwapConfirmMeta):
		flatten = decodeSwapConfirmInst(inst)

	case strconv.Itoa(metadata.BurningConfirmMeta):
		flatten = decodeBurningConfirmInst(inst)

	default:
		for _, part := range inst {
			flatten = append(flatten, []byte(part)...)
		}
	}
	return flatten
}

// decodeSwapConfirmInst flattens all parts of a swap confirm instruction, decodes and concats it
func decodeSwapConfirmInst(inst []string) []byte {
	metaType := []byte(inst[0])
	shardID := []byte(inst[1])
	height, _, _ := base58.Base58Check{}.Decode(inst[2])
	// Special case: instruction storing beacon/bridge's committee => decode and sign on that instead
	// We need to decode and then submit the pubkeys to Ethereum because we can't decode it on smart contract
	pks := []byte(inst[3])
	if d, _, err := (base58.Base58Check{}).Decode(inst[3]); err == nil {
		pks = d
	}
	flatten := []byte{}
	flatten = append(flatten, metaType...)
	flatten = append(flatten, shardID...)
	flatten = append(flatten, toBytes32BigEndian(height)...)
	flatten = append(flatten, pks...)
	return flatten
}

// decodeBurningConfirmInst decodes and flattens a BurningConfirm instruction
func decodeBurningConfirmInst(inst []string) []byte {
	metaType := []byte(inst[0])
	shardID := []byte(inst[1])
	tokenID, _, _ := base58.Base58Check{}.Decode(inst[2])
	tokenIDFixedLen := toBytes32BigEndian(tokenID)
	remoteAddr, _ := decodeRemoteAddr(inst[3])
	amount, _, _ := base58.Base58Check{}.Decode(inst[4])
	txID, _ := common.NewHashFromStr(inst[5])
	height, _, _ := base58.Base58Check{}.Decode(inst[6])
	fmt.Printf("[db] decoded BurningConfirm inst\n")
	fmt.Printf("[db]\tamount: %x\n[db]\tremoteAddr: %x\n[db]\ttokenID: %x\n", amount, remoteAddr, tokenID)
	flatten := []byte{}
	flatten = append(flatten, metaType...)
	flatten = append(flatten, shardID...)
	flatten = append(flatten, tokenIDFixedLen...)
	flatten = append(flatten, remoteAddr...)
	flatten = append(flatten, toBytes32BigEndian(amount)...)
	flatten = append(flatten, txID[:]...)
	flatten = append(flatten, toBytes32BigEndian(height)...)
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

// toBytes32BigEndian converts []byte to uint256 for of Ethereum
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

// pickBeaconPubkeyRootInstruction finds all BeaconSwapConfirmMeta instructions
// These instructions contain merkle root of beacon committee's pubkey
func pickBeaconPubkeyRootInstruction(
	beaconBlocks []*BeaconBlock,
) [][]string {
	instType := strconv.Itoa(metadata.BeaconSwapConfirmMeta)
	return pickInstructionFromBeaconBlocks(beaconBlocks, instType)
}

// pickBurningConfirmInstruction finds all BurningConfirmMeta instructions
func pickBurningConfirmInstruction(
	beaconBlocks []*BeaconBlock,
	height uint64,
) [][]string {
	// Pick
	instType := strconv.Itoa(metadata.BurningConfirmMeta)
	insts := pickInstructionFromBeaconBlocks(beaconBlocks, instType)

	// Replace beacon block height with shard's
	h := big.NewInt(0).SetUint64(height)
	for _, inst := range insts {
		inst[len(inst)-1] = base58.Base58Check{}.Encode(h.Bytes(), 0x00)
	}
	return insts
}

// pickBridgePubkeyRootInstruction finds all BridgeSwapConfirmMeta instructions
// These instructions contain merkle root of bridge committee's pubkey
func pickBridgePubkeyRootInstruction(
	block *ShardToBeaconBlock,
) [][]string {
	shardType := strconv.Itoa(metadata.BridgeSwapConfirmMeta)
	return pickInstructionWithType(block.Instructions, shardType)
}

// parseAndConcatPubkeys parse pubkeys of a commmittee stored as string and concat them
func parseAndConcatPubkeys(vals []string) []byte {
	pks := []byte{}
	for _, val := range vals {
		pk, _, _ := base58.Base58Check{}.Decode(val)
		// TODO(@0xbunyip): handle error
		pks = append(pks, pk...)
	}
	return pks
}

// buildBeaconSwapConfirmInstruction stores in an instruction the list of new beacon validators and the block that they start signing on
func buildBeaconSwapConfirmInstruction(currentValidators []string, startHeight uint64) []string {
	beaconComm := parseAndConcatPubkeys(currentValidators)
	fmt.Printf("[db] added beaconComm: %d %x\n", startHeight, beaconComm)

	// Convert startHeight to big.Int to get bytes later
	height := big.NewInt(0).SetUint64(startHeight)

	bridgeID := byte(common.BRIDGE_SHARD_ID)
	instContent := base58.Base58Check{}.Encode(beaconComm, 0x00)
	return []string{
		strconv.Itoa(metadata.BeaconSwapConfirmMeta),
		strconv.Itoa(int(bridgeID)),
		base58.Base58Check{}.Encode(height.Bytes(), 0x00),
		instContent,
	}
}

// buildBridgeSwapConfirmInstruction stores in an instruction the list of new bridge validators and the block that they start signing on
func buildBridgeSwapConfirmInstruction(currentValidators []string, startHeight uint64) []string {
	bridgeComm := parseAndConcatPubkeys(currentValidators)
	fmt.Printf("[db] added bridgeComm: %d %x\n", startHeight, bridgeComm)

	// Convert startHeight to big.Int to get bytes later
	height := big.NewInt(0).SetUint64(startHeight)

	bridgeID := byte(common.BRIDGE_SHARD_ID)
	instContent := base58.Base58Check{}.Encode(bridgeComm, 0x00)
	return []string{
		strconv.Itoa(metadata.BridgeSwapConfirmMeta),
		strconv.Itoa(int(bridgeID)),
		base58.Base58Check{}.Encode(height.Bytes(), 0x00),
		instContent,
	}
}

// convertBurningRequestToConfirm finds all BurningRequest insts in a list of beacon blocks and convert them to BurningConfirmInst
func convertBurningRequestToConfirm(beaconBlocks []*BeaconBlock, db database.DatabaseInterface) ([][]string, error) {
	insts := [][]string{}
	for _, blk := range beaconBlocks {
		for _, inst := range blk.Body.Instructions {
			if strconv.Itoa(metadata.BurningRequestMeta) == inst[0] {
				burningConfirmInst, err := buildBurningConfirmInst(inst, blk.Header.Height, db)
				if err != nil {
					return nil, err
				}
				insts = append(insts, burningConfirmInst)
			}
		}
	}
	return insts, nil
}

// buildBurningConfirmInst builds on beacon an instruction confirming a tx burning bridge-token
func buildBurningConfirmInst(inst []string, height uint64, db database.DatabaseInterface) ([]string, error) {
	fmt.Printf("[db] build BurningConfirmInst: %s\n", inst)
	// Parse action and get metadata
	var burningReqAction BurningReqAction
	err := decodeContent(inst[1], &burningReqAction)
	if err != nil {
		return nil, err
	}
	md := burningReqAction.Meta
	txID := burningReqAction.RequestedTxID // to prevent double-release token

	// Convert amount to big.Int to get bytes later
	amount := big.NewInt(0).SetUint64(md.BurningAmount)

	shardID := byte(common.BRIDGE_SHARD_ID)

	// Convert to external tokenID
	tokenID, err := db.GetBridgeExternalTokenID(md.TokenID, false)
	if err != nil {
		return nil, err
	}

	// Convert height to big.Int to get bytes later
	h := big.NewInt(0).SetUint64(height)

	return []string{
		strconv.Itoa(metadata.BurningConfirmMeta),
		strconv.Itoa(int(shardID)),
		base58.Base58Check{}.Encode(tokenID, 0x00),
		md.RemoteAddress,
		base58.Base58Check{}.Encode(amount.Bytes(), 0x00),
		txID.String(),
		base58.Base58Check{}.Encode(h.Bytes(), 0x00),
	}, nil
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
		case metadata.IssuingRequestMeta, metadata.ContractingRequestMeta, metadata.BurningRequestMeta, metadata.BurningConfirmMeta:
			newInst = [][]string{inst}

		case metadata.IssuingETHRequestMeta:
			newInst, err = buildInstructionsForETHIssuingReq(contentStr, shardID)

		// TODO(@0xbunyip): remove processing BurningRequestMeta when reverting beacon blocks
		// case metadata.BurningRequestMeta:

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
	accumulatedValues := &metadata.AccumulatedValues{
		UniqETHTxsUsed:   [][]byte{},
		DBridgeTokenPair: map[string][]byte{},
		CBridgeTokens:    []*common.Hash{},
	}
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
			metaType, err := strconv.Atoi(l[0])
			if err != nil {
				return nil, err
			}
			var newTx metadata.Transaction
			switch metaType {
			case metadata.IssuingETHRequestMeta:
				fmt.Println("haha isntruction: ", l)
				if len(l) >= 4 {
					newTx, err = blockgen.buildETHIssuanceTx(l[3], producerPrivateKey, shardID, accumulatedValues)
				}

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
	if relayingRewardTx != nil {
		respTxs = append(respTxs, relayingRewardTx)
	}
	return respTxs, nil
}

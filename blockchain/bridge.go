package blockchain

import (
	"encoding/hex"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
)

// FlattenAndConvertStringInst receives a slice of insts; concats each inst ([]string) and converts to []byte to build merkle tree later
func FlattenAndConvertStringInst(insts [][]string) ([][]byte, error) {
	flattenInsts := [][]byte{}
	for _, inst := range insts {
		d, err := DecodeInstruction(inst)
		if err != nil {
			return nil, err
		}
		flattenInsts = append(flattenInsts, d)
	}
	return flattenInsts, nil
}

// decodeInstruction appends all part of an instruction and decode them if necessary (for special instruction that needed to be decoded before submitting to Ethereum)
func DecodeInstruction(inst []string) ([]byte, error) {
	flatten := []byte{}
	switch inst[0] {
	case strconv.Itoa(metadata.BeaconSwapConfirmMeta), strconv.Itoa(metadata.BridgeSwapConfirmMeta):
		flatten = decodeSwapConfirmInst(inst)

	case strconv.Itoa(metadata.BurningConfirmMeta):
		var err error
		flatten, err = decodeBurningConfirmInst(inst)
		if err != nil {
			return nil, err
		}

	default:
		for _, part := range inst {
			flatten = append(flatten, []byte(part)...)
		}
	}
	return flatten, nil
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
func decodeBurningConfirmInst(inst []string) ([]byte, error) {
	if len(inst) < 7 {
		return nil, errors.New("invalid length of BurningConfirm inst")
	}
	metaType := []byte(inst[0])
	shardID := []byte(inst[1])
	tokenID, _, errToken := base58.Base58Check{}.Decode(inst[2])
	remoteAddr, errAddr := decodeRemoteAddr(inst[3])
	amount, _, errAmount := base58.Base58Check{}.Decode(inst[4])
	txID, errTx := common.Hash{}.NewHashFromStr(inst[5])
	height, _, errHeight := base58.Base58Check{}.Decode(inst[6])
	if err := common.CheckError(errToken, errAddr, errAmount, errTx, errHeight); err != nil {
		BLogger.log.Error(errors.WithStack(err))
		return nil, errors.WithStack(err)
	}

	BLogger.log.Infof("Decoded BurningConfirm inst")
	BLogger.log.Infof("\tamount: %d\n\tremoteAddr: %x\n\ttokenID: %x", big.NewInt(0).SetBytes(amount), remoteAddr, tokenID)
	flatten := []byte{}
	flatten = append(flatten, metaType...)
	flatten = append(flatten, shardID...)
	flatten = append(flatten, toBytes32BigEndian(tokenID)...)
	flatten = append(flatten, remoteAddr...)
	flatten = append(flatten, toBytes32BigEndian(amount)...)
	flatten = append(flatten, txID[:]...)
	flatten = append(flatten, toBytes32BigEndian(height)...)
	return flatten, nil
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

// toBytes32BigEndian converts a []byte to uint256 of Ethereum
func toBytes32BigEndian(b []byte) []byte {
	a := [32]byte{}
	copy(a[32-len(b):], b)
	return a[:]
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

// pickBeaconSwapConfirmInst finds all BeaconSwapConfirmMeta instructions in some beacon blocks
func pickBeaconSwapConfirmInst(
	beaconBlocks []*BeaconBlock,
) [][]string {
	instType := strconv.Itoa(metadata.BeaconSwapConfirmMeta)
	return pickInstructionFromBeaconBlocks(beaconBlocks, instType)
}

// pickBridgeSwapConfirmInst finds all BridgeSwapConfirmMeta instructions in a shard to beacon block
func pickBridgeSwapConfirmInst(
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
		pks = append(pks, pk...)
	}
	return pks
}

// buildBeaconSwapConfirmInstruction stores in an instruction the list of new beacon validators and the block that they start signing on
func buildBeaconSwapConfirmInstruction(currentValidators []string, startHeight uint64) []string {
	beaconComm := parseAndConcatPubkeys(currentValidators)
	BLogger.log.Infof("New beaconComm: %d %x", startHeight, beaconComm)

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
	BLogger.log.Infof("New bridgeComm: %d %x", startHeight, bridgeComm)

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

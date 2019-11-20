package blockchain

import (
	"bytes"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/core/rawdb"
	"github.com/incognitochain/incognito-chain/incdb"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
)

// build instructions at beacon chain before syncing to shards
func (blockchain *BlockChain) buildBridgeInstructions(
	shardID byte,
	shardBlockInstructions [][]string,
	beaconHeight uint64,
	db incdb.Database,
) ([][]string, error) {
	instructions := [][]string{}
	for _, inst := range shardBlockInstructions {
		if len(inst) < 2 {
			continue
		}
		if inst[0] == SetAction || inst[0] == StakeAction || inst[0] == SwapAction || inst[0] == RandomAction || inst[0] == AssignAction {
			continue
		}

		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue
		}
		contentStr := inst[1]
		newInst := [][]string{}
		switch metaType {
		case metadata.ContractingRequestMeta:
			newInst, err = blockchain.buildInstructionsForContractingReq(contentStr, shardID, metaType)

		case metadata.BurningRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(inst, beaconHeight, db)
			newInst = [][]string{burningConfirm}

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

// buildBurningConfirmInst builds on beacon an instruction confirming a tx burning bridge-token
func buildBurningConfirmInst(inst []string, height uint64, db incdb.Database) ([]string, error) {
	BLogger.log.Infof("Build BurningConfirmInst: %s", inst)
	// Parse action and get metadata
	var burningReqAction BurningReqAction
	err := decodeContent(inst[1], &burningReqAction)
	if err != nil {
		return nil, errors.Wrap(err, "invalid BurningRequest")
	}
	md := burningReqAction.Meta
	txID := burningReqAction.RequestedTxID // to prevent double-release token
	shardID := byte(common.BridgeShardID)

	// Convert to external tokenID
	tokenID, err := findExternalTokenID(&md.TokenID, db)
	if err != nil {
		return nil, err
	}

	// Convert amount to big.Int to get bytes later
	amount := big.NewInt(0).SetUint64(md.BurningAmount)
	if bytes.Equal(tokenID, rCommon.HexToAddress(common.EthAddrStr).Bytes()) {
		// Convert Gwei to Wei for Ether
		amount = amount.Mul(amount, big.NewInt(1000000000))
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
		base58.Base58Check{}.Encode(md.TokenID[:], 0x00),
		base58.Base58Check{}.Encode(h.Bytes(), 0x00),
	}, nil
}

// findExternalTokenID finds the external tokenID for a bridge token from database
func findExternalTokenID(tokenID *common.Hash, db incdb.Database) ([]byte, error) {
	allBridgeTokensBytes, err := rawdb.GetAllBridgeTokens(db)
	if err != nil {
		return nil, err
	}
	var allBridgeTokens []*rawdb.BridgeTokenInfo
	err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, token := range allBridgeTokens {
		if token.TokenID.IsEqual(tokenID) && len(token.ExternalTokenID) > 0 {
			return token.ExternalTokenID, nil
		}
	}
	return nil, errors.New("invalid tokenID")
}

package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/instruction"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/pkg/errors"
)

// build instructions at beacon chain before syncing to shards
func (blockchain *BlockChain) buildBridgeInstructions(curView *BeaconBestState, shardID byte, shardBlockInstructions [][]string, beaconHeight uint64) ([][]string, error) {
	stateDB := curView.GetBeaconFeatureStateDB()
	instructions := [][]string{}
	for _, inst := range shardBlockInstructions {
		if len(inst) < 2 {
			continue
		}
		if inst[0] == instruction.SET_ACTION || inst[0] == instruction.STAKE_ACTION || inst[0] == instruction.SWAP_ACTION || inst[0] == instruction.RANDOM_ACTION || inst[0] == instruction.ASSIGN_ACTION {
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
			burningConfirm, _, err = buildBurningConfirmInst(curView, stateDB, metadata.BurningConfirmMeta, inst, beaconHeight, "")
			newInst = [][]string{burningConfirm}

		case metadata.BurningRequestMetaV2:
			burningConfirm := []string{}
			burningConfirm, _, err = buildBurningConfirmInst(curView, stateDB, metadata.BurningConfirmMetaV2, inst, beaconHeight, "")
			newInst = [][]string{burningConfirm}

		case metadata.BurningPBSCRequestMeta:
			burningConfirm := []string{}
			burningConfirm, _, err = buildBurningConfirmInst(curView, stateDB, metadata.BurningBSCConfirmMeta, inst, beaconHeight, common.BSCPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningPBSCForDepositToSCRequestMeta:
			burningConfirm := []string{}
			burningConfirm, _, err = buildBurningConfirmInst(curView, stateDB, metadata.BurningPBSCConfirmForDepositToSCMeta, inst, beaconHeight, common.BSCPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningForDepositToSCRequestMeta:
			burningConfirm := []string{}
			burningConfirm, _, err = buildBurningConfirmInst(curView, stateDB, metadata.BurningConfirmForDepositToSCMeta, inst, beaconHeight, "")
			newInst = [][]string{burningConfirm}

		case metadata.BurningForDepositToSCRequestMetaV2:
			burningConfirm := []string{}
			burningConfirm, _, err = buildBurningConfirmInst(curView, stateDB, metadata.BurningConfirmForDepositToSCMetaV2, inst, beaconHeight, "")
			newInst = [][]string{burningConfirm}

		case metadata.BurningPRVERC20RequestMeta:
			burningConfirm := []string{}
			burningConfirm, _, err = buildBurningPRVEVMConfirmInst(curView, metadata.BurningPRVERC20ConfirmMeta, inst, beaconHeight, config.Param().PRVERC20ContractAddressStr)
			newInst = [][]string{burningConfirm}

		case metadata.BurningPRVBEP20RequestMeta:
			burningConfirm := []string{}
			burningConfirm, _, err = buildBurningPRVEVMConfirmInst(curView, metadata.BurningPRVBEP20ConfirmMeta, inst, beaconHeight, config.Param().PRVBEP20ContractAddressStr)
			newInst = [][]string{burningConfirm}

		case metadata.BurningPLGRequestMeta:
			burningConfirm := []string{}
			burningConfirm, _, err = buildBurningConfirmInst(curView, stateDB, metadata.BurningPLGConfirmMeta, inst, beaconHeight, common.PLGPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningPLGForDepositToSCRequestMeta:
			burningConfirm := []string{}
			burningConfirm, _, err = buildBurningConfirmInst(curView, stateDB, metadata.BurningPLGConfirmForDepositToSCMeta, inst, beaconHeight, common.PLGPrefix)
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
func buildBurningConfirmInst(
	curView *BeaconBestState,
	stateDB *statedb.StateDB,
	burningMetaType int,
	inst []string,
	height uint64,
	prefix string,
) ([]string, bridgeagg.UnshieldAction, error) {
	BLogger.log.Infof("Build BurningConfirmInst: %s", inst)
	// Parse action and get metadata
	var burningReqAction BurningReqAction
	err := decodeContent(inst[1], &burningReqAction)
	if err != nil {
		return nil, bridgeagg.UnshieldAction{}, errors.Wrap(err, "invalid BurningRequest")
	}
	md := burningReqAction.Meta
	txID := burningReqAction.RequestedTxID // to prevent double-release token
	shardID := byte(common.BridgeShardID)

	incTokenID := md.TokenID
	if burningMetaType == metadataCommon.BurningUnifiedTokenRequestMeta {
		if md.NetworkID != common.DefaultNetworkID {
			incTokenID, prefix, _, err = curView.bridgeAggState.GetUnshieldInfo(md.TokenID, md.NetworkID)
			if err != nil {
				return nil, bridgeagg.UnshieldAction{}, err
			}
		}
	}

	// Convert to external tokenID
	tokenID, err := findExternalTokenID(stateDB, &incTokenID)
	if err != nil {
		return nil, bridgeagg.UnshieldAction{}, err
	}

	if len(tokenID) < common.ExternalBridgeTokenLength {
		return nil, bridgeagg.UnshieldAction{}, errors.New("invalid external token id")
	}

	prefixLen := len(prefix)
	if (prefixLen > 0 && !bytes.Equal([]byte(prefix), tokenID[:prefixLen])) ||
		len(tokenID) != (common.ExternalBridgeTokenLength+prefixLen) {
		return nil, bridgeagg.UnshieldAction{}, errors.New(fmt.Sprintf("invalid BurningRequestConfirm type %v with external tokeid %v", burningMetaType, tokenID))
	}

	// Convert amount to big.Int to get bytes later
	amount := big.NewInt(0).SetUint64(md.BurningAmount)
	if bytes.Equal(tokenID, append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...)) {
		// Convert Gwei to Wei for Ether
		if burningMetaType != metadataCommon.BurningUnifiedTokenRequestMeta {
			amount = amount.Mul(amount, big.NewInt(1000000000))
		}
	}

	// Convert height to big.Int to get bytes later
	h := big.NewInt(0).SetUint64(height)

	res := []string{
		strconv.Itoa(burningMetaType),
		strconv.Itoa(int(shardID)),
		base58.Base58Check{}.Encode(tokenID, 0x00),
		md.RemoteAddress,
		base58.Base58Check{}.Encode(amount.Bytes(), 0x00),
		txID.String(),
		base58.Base58Check{}.Encode(md.TokenID[:], 0x00),
		base58.Base58Check{}.Encode(h.Bytes(), 0x00),
	}

	return res, bridgeagg.UnshieldAction{
		Content:         md,
		TxReqID:         *txID,
		ExternalTokenID: tokenID[:],
	}, nil
}

// buildBurningPRVEVMConfirmInst builds on beacon an instruction confirming a tx burning PRV-EVM-token
func buildBurningPRVEVMConfirmInst(
	curView *BeaconBestState,
	burningMetaType int,
	inst []string,
	height uint64,
	tokenIDStr string,
) ([]string, bridgeagg.UnshieldAction, error) {
	BLogger.log.Infof("PRV EVM: Build BurningConfirmInst: %s", inst)
	// Parse action and get metadata
	var burningReqAction BurningReqAction
	err := decodeContent(inst[1], &burningReqAction)
	if err != nil {
		return nil, bridgeagg.UnshieldAction{}, errors.Wrap(err, "PRV EVM: invalid BurningRequest")
	}

	md := burningReqAction.Meta
	if md.TokenID.String() != common.PRVIDStr {
		return nil, bridgeagg.UnshieldAction{}, errors.New("PRV EVM: invalid PRV token ID")
	}

	if burningMetaType == metadataCommon.BurningUnifiedTokenRequestMeta {
		if md.NetworkID != common.DefaultNetworkID {
			_, _, tokenIDStr, err = curView.bridgeAggState.GetUnshieldInfo(md.TokenID, md.NetworkID)
			if err != nil {
				return nil, bridgeagg.UnshieldAction{}, err
			}
		}
	}

	tokenID := rCommon.HexToAddress(tokenIDStr)
	txID := burningReqAction.RequestedTxID // to prevent double-release token
	shardID := byte(common.BridgeShardID)

	// Convert amount to big.Int to get bytes later
	amount := big.NewInt(0).SetUint64(md.BurningAmount)
	// Convert height to big.Int to get bytes later
	h := big.NewInt(0).SetUint64(height)

	return []string{
			strconv.Itoa(burningMetaType),
			strconv.Itoa(int(shardID)),
			base58.Base58Check{}.Encode(tokenID[:], 0x00),
			md.RemoteAddress,
			base58.Base58Check{}.Encode(amount.Bytes(), 0x00),
			txID.String(),
			base58.Base58Check{}.Encode(md.TokenID[:], 0x00),
			base58.Base58Check{}.Encode(h.Bytes(), 0x00),
		}, bridgeagg.UnshieldAction{
			Content:         md,
			TxReqID:         *txID,
			ExternalTokenID: tokenID[:],
		}, nil
}

// findExternalTokenID finds the external tokenID for a bridge token from database
func findExternalTokenID(stateDB *statedb.StateDB, tokenID *common.Hash) ([]byte, error) {
	allBridgeTokensBytes, err := statedb.GetAllBridgeTokens(stateDB)
	if err != nil {
		return nil, err
	}
	var allBridgeTokens []*rawdbv2.BridgeTokenInfo
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

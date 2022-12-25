package blockchain

import (
	"bytes"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/instruction"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	"github.com/pkg/errors"
)

// build instructions at beacon chain before syncing to shards
func (blockchain *BlockChain) buildBridgeInstructions(stateDB *statedb.StateDB, shardID byte, shardBlockInstructions [][]string, beaconHeight uint64) ([][]string, error) {
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
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningConfirmMeta, inst, beaconHeight, "")
			newInst = [][]string{burningConfirm}

		case metadata.BurningRequestMetaV2:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningConfirmMetaV2, inst, beaconHeight, "")
			newInst = [][]string{burningConfirm}

		case metadata.BurningPBSCRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningBSCConfirmMeta, inst, beaconHeight, common.BSCPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningPBSCForDepositToSCRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningPBSCConfirmForDepositToSCMeta, inst, beaconHeight, common.BSCPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningForDepositToSCRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningConfirmForDepositToSCMeta, inst, beaconHeight, "")
			newInst = [][]string{burningConfirm}

		case metadata.BurningForDepositToSCRequestMetaV2:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningConfirmForDepositToSCMetaV2, inst, beaconHeight, "")
			newInst = [][]string{burningConfirm}

		case metadata.BurningPRVERC20RequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningPRVEVMConfirmInst(metadata.BurningPRVERC20ConfirmMeta, inst, beaconHeight, config.Param().PRVERC20ContractAddressStr)
			newInst = [][]string{burningConfirm}

		case metadata.BurningPRVBEP20RequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningPRVEVMConfirmInst(metadata.BurningPRVBEP20ConfirmMeta, inst, beaconHeight, config.Param().PRVBEP20ContractAddressStr)
			newInst = [][]string{burningConfirm}

		case metadata.BurningPLGRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningPLGConfirmMeta, inst, beaconHeight, common.PLGPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningPLGForDepositToSCRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningPLGConfirmForDepositToSCMeta, inst, beaconHeight, common.PLGPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningFantomRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningFantomConfirmMeta, inst, beaconHeight, common.FTMPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningFantomForDepositToSCRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningFantomConfirmForDepositToSCMeta, inst, beaconHeight, common.FTMPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningAuroraRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningAuroraConfirmMeta, inst, beaconHeight, common.AURORAPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningAvaxRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningAvaxConfirmMeta, inst, beaconHeight, common.AVAXPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningAuroraForDepositToSCRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningAuroraConfirmForDepositToSCMeta, inst, beaconHeight, common.AURORAPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningAvaxForDepositToSCRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningAvaxConfirmForDepositToSCMeta, inst, beaconHeight, common.AVAXPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningNearRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningConfirmInst(stateDB, metadata.BurningNearConfirmMeta, inst, beaconHeight, common.NEARPrefix)
			newInst = [][]string{burningConfirm}

		case metadata.BurningPRVRequestMeta:
			burningConfirm := []string{}
			burningConfirm, err = buildBurningPRVEVMConfirmInst(metadata.BurningPRVRequestConfirmMeta, inst, beaconHeight, config.Param().PRVERC20ContractAddressStr)
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
	stateDB *statedb.StateDB,
	burningMetaType int,
	inst []string,
	height uint64,
	prefix string,
) ([]string, error) {
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
	tokenID, err := metadataBridge.FindExternalTokenID(stateDB, md.TokenID, prefix)
	if err != nil {
		return nil, err
	}

	// Convert amount to big.Int to get bytes later
	amount := big.NewInt(0).SetUint64(md.BurningAmount)
	if bytes.Equal(tokenID, append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...)) {
		// Convert Gwei to Wei for Ether
		amount = amount.Mul(amount, big.NewInt(1000000000))
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

	return res, nil
}

// buildBurningPRVEVMConfirmInst builds on beacon an instruction confirming a tx burning PRV-EVM-token
func buildBurningPRVEVMConfirmInst(
	burningMetaType int,
	inst []string,
	height uint64,
	tokenIDStr string,
) ([]string, error) {
	BLogger.log.Infof("PRV EVM: Build BurningConfirmInst: %s", inst)
	// Parse action and get metadata
	var burningReqAction BurningReqAction
	err := decodeContent(inst[1], &burningReqAction)
	if err != nil {
		return nil, errors.Wrap(err, "PRV EVM: invalid BurningRequest")
	}

	md := burningReqAction.Meta
	if md.TokenID.String() != common.PRVIDStr {
		return nil, errors.New("PRV EVM: invalid PRV token ID")
	}

	tokenID := rCommon.HexToAddress(tokenIDStr)
	txID := burningReqAction.RequestedTxID // to prevent double-release token
	shardID := byte(common.BridgeShardID)

	// Convert amount to big.Int to get bytes later
	amount := big.NewInt(0).SetUint64(md.BurningAmount)
	// Convert height to big.Int to get bytes later
	h := big.NewInt(0).SetUint64(height)

	results := []string{
		strconv.Itoa(burningMetaType),
		strconv.Itoa(int(shardID)),
		base58.Base58Check{}.Encode(tokenID[:], 0x00),
		md.RemoteAddress,
		base58.Base58Check{}.Encode(amount.Bytes(), 0x00),
		txID.String(),
		base58.Base58Check{}.Encode(md.TokenID[:], 0x00),
	}
	if burningMetaType == metadata.BurningPRVRequestConfirmMeta {
		reAddrStr, err := burningReqAction.Meta.RedepositReceiver.String()
		if err != nil {
			return nil, errors.New("PRV EVM: invalid RedepositReceiver address")
		}
		results = append(results, reAddrStr)
	}
	results = append(results, base58.Base58Check{}.Encode(h.Bytes(), 0x00))
	return results, nil
}

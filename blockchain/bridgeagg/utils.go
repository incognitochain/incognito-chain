package bridgeagg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/utils"
)

type ShieldStatusData struct {
	Amount uint64 `json:"Amount"`
	Reward uint64 `json:"Reward"`
}

type ShieldStatus struct {
	Status    byte               `json:"Status"`
	Data      []ShieldStatusData `json:"Data,omitempty"`
	ErrorCode int                `json:"ErrorCode,omitempty"`
}

type UnshieldStatusData struct {
	ReceivedAmount uint64 `json:"ReceivedAmount"`
	Fee            uint64 `json:"Fee"`
}

type UnshieldStatus struct {
	Status    byte                 `json:"Status"`
	Data      []UnshieldStatusData `json:"Data,omitempty"`
	ErrorCode int                  `json:"ErrorCode,omitempty"`
}

type ModifyParamStatus struct {
	Status               byte   `json:"Status"`
	NewPercentFeeWithDec uint64 `json:"NewPercentFeeWithDec"`
	ErrorCode            int    `json:"ErrorCode,omitempty"`
}

type ConvertStatus struct {
	Status                byte   `json:"Status"`
	ConvertPUnifiedAmount uint64 `json:"ConvertPUnifiedAmount"`
	Reward                uint64 `json:"Reward"`
	ErrorCode             int    `json:"ErrorCode,omitempty"`
}

type VaultChange struct {
	IsChanged bool
}

func NewVaultChange() *VaultChange {
	return &VaultChange{}
}

type StateChange struct {
	unifiedTokenID map[common.Hash]bool
	vaultChange    map[common.Hash]map[common.Hash]VaultChange
}

func NewStateChange() *StateChange {
	return &StateChange{
		unifiedTokenID: make(map[common.Hash]bool),
		vaultChange:    make(map[common.Hash]map[common.Hash]VaultChange),
	}
}

func GetInsertTxHashIssuedFuncByNetworkID(networkID uint8) func(*statedb.StateDB, []byte) error {
	switch networkID {
	case common.PLGNetworkID:
		return statedb.InsertPLGTxHashIssued
	case common.BSCNetworkID:
		return statedb.InsertBSCTxHashIssued
	case common.ETHNetworkID:
		return statedb.InsertETHTxHashIssued
	case common.FTMNetworkID:
		return statedb.InsertFTMTxHashIssued
	}
	return nil
}

// buildRejectedInst returns a rejected instruction
// content maybe is null
func buildRejectedInst(metaType int, shardID byte, txReqID common.Hash, errorType int, content []byte) []string {
	rejectContentStr, _ := metadataCommon.NewRejectContentWithValue(txReqID, ErrCodeMessage[errorType].Code, content).String()
	rejectedInst := metadataCommon.NewInstructionWithValue(
		metaType,
		common.RejectedStatusStr,
		shardID,
		rejectContentStr,
	)
	return rejectedInst.StringSlice()
}

// buildAcceptedInst returns accepted instructions
// contents maybe is null
func buildAcceptedInst(metaType int, shardID byte, contents [][]byte) [][]string {
	insts := [][]string{}
	for _, content := range contents {
		inst := metadataCommon.NewInstructionWithValue(
			metaType,
			common.AcceptedStatusStr,
			shardID,
			base64.StdEncoding.EncodeToString(content),
		)
		insts = append(insts, inst.StringSlice())
	}
	return insts
}

func buildRejectedConvertReqInst(meta metadataBridge.ConvertTokenToUnifiedTokenRequest, shardID byte, txReqID common.Hash, errorType int) []string {
	rejectedUnshieldRequest := metadataBridge.RejectedConvertTokenToUnifiedToken{
		TokenID:  meta.TokenID,
		Amount:   meta.Amount,
		Receiver: meta.Receiver,
	}
	rejectedContent, _ := json.Marshal(rejectedUnshieldRequest)

	rejectContentStr, _ := metadataCommon.NewRejectContentWithValue(txReqID, ErrCodeMessage[errorType].Code, rejectedContent).String()
	rejectedInst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta,
		common.RejectedStatusStr,
		shardID,
		rejectContentStr,
	)
	return rejectedInst.StringSlice()
}

func buildRejectedUnshieldReqInst(meta metadataBridge.UnshieldRequest, shardID byte, txReqID common.Hash, errorType int) []string {
	var totalBurnAmt uint64
	for _, data := range meta.Data {
		totalBurnAmt += data.BurningAmount
	}
	rejectedUnshieldRequest := metadataBridge.RejectedUnshieldRequest{
		UnifiedTokenID: meta.UnifiedTokenID,
		Amount:         totalBurnAmt,
		Receiver:       meta.Receiver,
	}
	rejectedContent, _ := json.Marshal(rejectedUnshieldRequest)

	rejectContentStr, _ := metadataCommon.NewRejectContentWithValue(txReqID, ErrCodeMessage[errorType].Code, rejectedContent).String()
	rejectedInst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BurningUnifiedTokenRequestMeta,
		common.RejectedStatusStr,
		shardID,
		rejectContentStr,
	)
	return rejectedInst.StringSlice()
}

// buildAddWaitingUnshieldInst returns processing unshield instructions
func buildUnshieldInst(unifiedTokenID common.Hash, isDepositToSC bool, waitingUnshieldReq *statedb.BridgeAggWaitingUnshieldReq, status string, shardID byte) []string {
	acceptedUnshieldInst := metadataBridge.AcceptedUnshieldRequestInst{
		UnifiedTokenID:     unifiedTokenID,
		IsDepositToSC:      isDepositToSC,
		WaitingUnshieldReq: waitingUnshieldReq,
	}
	acceptedUnshieldInstBytes, _ := json.Marshal(acceptedUnshieldInst)
	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BurningUnifiedTokenRequestMeta,
		status,
		shardID,
		base64.StdEncoding.EncodeToString(acceptedUnshieldInstBytes),
	)
	return inst.StringSlice()
}

func buildBurningConfirmInsts(waitingUnshieldReq *statedb.BridgeAggWaitingUnshieldReq, beaconHeightForConfirmInst uint64) [][]string {
	burningInsts := [][]string{}
	beaconHeightBN := big.NewInt(0).SetUint64(beaconHeightForConfirmInst)
	txID := waitingUnshieldReq.GetUnshieldID()

	for index, data := range waitingUnshieldReq.GetData() {
		// maybe there are multiple  proofs for one txID, so append index to make newTxReqID unique
		newTxReqID := common.HashH(append(txID.Bytes(), common.IntToBytes(index)...))
		burningInst := []string{
			strconv.Itoa(int(data.BurningConfirmMetaType)),
			strconv.Itoa(int(common.BridgeShardID)),
			base58.Base58Check{}.Encode(data.ExternalTokenID, 0x00),
			data.RemoteAddress,
			base58.Base58Check{}.Encode(data.ExternalReceivedAmt.Bytes(), 0x00),
			newTxReqID.String(),
			base58.Base58Check{}.Encode(data.IncTokenID[:], 0x00),
			base58.Base58Check{}.Encode(beaconHeightBN.Bytes(), 0x00),
		}
		burningInsts = append(burningInsts, burningInst)
	}

	return burningInsts
}

func buildInstruction(
	metaType int, errorType int,
	contents [][]byte, txReqID common.Hash,
	shardID byte, err error,
) ([][]string, error) {
	res := [][]string{}
	for _, content := range contents {
		inst := metadataCommon.NewInstructionWithValue(
			metaType,
			common.AcceptedStatusStr,
			shardID,
			utils.EmptyString,
		)
		if err != nil {
			rejectContent := metadataCommon.NewRejectContentWithValue(txReqID, ErrCodeMessage[errorType].Code, content)
			inst.Status = common.RejectedStatusStr
			rejectedInst := []string{}
			rejectedInst, err = inst.StringSliceWithRejectContent(rejectContent)
			if err != nil {
				return res, NewBridgeAggErrorWithValue(errorType, err)
			}
			res = append(res, rejectedInst)
		} else {
			inst.Content = base64.StdEncoding.EncodeToString(content)
			res = append(res, inst.StringSlice())
		}
	}
	return res, nil
}

func UnmarshalEVMShieldProof(proofBytes []byte, actionData []byte) (*metadataBridge.EVMProof, *types.Receipt, error) {
	proofData := metadataBridge.EVMProof{}
	err := json.Unmarshal(proofBytes, &proofData)
	if err != nil {
		return nil, nil, err
	}

	txReceipt := types.Receipt{}
	err = json.Unmarshal(actionData, &txReceipt)
	if err != nil {
		return nil, nil, err
	}
	return &proofData, &txReceipt, err
}

func IsBridgeTxHashUsedInBlock(uniqTx []byte, uniqTxsUsed [][]byte) bool {
	for _, item := range uniqTxsUsed {
		if bytes.Equal(uniqTx, item) {
			return true
		}
	}
	return false
}

func ValidateDoubleShieldProof(
	proof *metadataBridge.EVMProof,
	listTxUsed [][]byte,
	isTxHashIssued func(stateDB *statedb.StateDB, uniqTx []byte) (bool, error),
	stateDB *statedb.StateDB,
) (bool, []byte, error) {
	// NOTE: since TxHash from constructedReceipt is always '0x0000000000000000000000000000000000000000000000000000000000000000'
	// so must build unique eth tx as combination of block hash and tx index.
	uniqTx := append(proof.BlockHash[:], []byte(strconv.Itoa(int(proof.TxIndex)))...)
	isUsedInBlock := IsBridgeTxHashUsedInBlock(uniqTx, listTxUsed)
	if isUsedInBlock {
		return false, uniqTx, fmt.Errorf("WARNING: tx %v already issued for the hash in current block: ", uniqTx)
	}
	isIssued, err := isTxHashIssued(stateDB, uniqTx)
	if err != nil {
		return false, uniqTx, fmt.Errorf("WARNING: an issue occured while checking the bridge tx hash is issued or not: %v ", err)
	}
	if isIssued {
		return false, uniqTx, fmt.Errorf("WARNING: tx %v already issued for the hash in previous blocks: ", uniqTx)
	}

	return true, uniqTx, nil
}

func getBurningConfirmMetaType(networkID uint8, isDepositToSC bool) (uint, error) {
	var burningMetaType int
	switch networkID {
	case common.ETHNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningConfirmForDepositToSCMetaV2
		} else {
			burningMetaType = metadata.BurningConfirmMetaV2
		}
	case common.BSCNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningPBSCConfirmForDepositToSCMeta
		} else {
			burningMetaType = metadata.BurningBSCConfirmMeta
		}
	case common.PLGNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningPLGConfirmForDepositToSCMeta
		} else {
			burningMetaType = metadata.BurningPLGConfirmMeta
		}
	case common.FTMNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningFantomConfirmForDepositToSCMeta
		} else {
			burningMetaType = metadata.BurningFantomConfirmMeta
		}
	default:
		return 0, fmt.Errorf("Invalid networkID %v", networkID)
	}
	return uint(burningMetaType), nil

}

func buildAcceptedShieldContent(
	shieldData []metadataBridge.AcceptedShieldRequestData,
	paymentAddress privacy.PaymentAddress, unifiedTokenID, txReqID common.Hash,
) ([]byte, error) {
	acceptedContent := metadataBridge.AcceptedInstShieldRequest{
		Receiver:       paymentAddress,
		UnifiedTokenID: unifiedTokenID,
		TxReqID:        txReqID,
		Data:           shieldData,
	}
	return json.Marshal(acceptedContent)
}

func ConvertAmountByDecimal(amount *big.Int, decimal uint8, isToUnifiedDecimal bool) (*big.Int, error) {
	res := big.NewInt(0).Set(amount)
	if isToUnifiedDecimal {
		res.Mul(res, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(config.Param().BridgeAggParam.BaseDecimal)), nil))
		res.Div(res, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil))
		if !res.IsUint64() {
			return nil, errors.New("Out of range unit64")
		}
	} else {
		res.Mul(res, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil))
		res.Div(res, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(config.Param().BridgeAggParam.BaseDecimal)), nil))
	}
	return res, nil
}

func getBurningConfirmMeta(networkID int, isDepositToSC bool) (int, string, error) {
	var burningMetaType int
	var prefix string

	switch networkID {
	case common.ETHNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningConfirmForDepositToSCMetaV2
		} else {
			burningMetaType = metadata.BurningConfirmMetaV2
		}
		prefix = utils.EmptyString
	case common.BSCNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningPBSCConfirmForDepositToSCMeta
		} else {
			burningMetaType = metadata.BurningBSCConfirmMeta
		}
		prefix = common.BSCPrefix
	case common.PLGNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningPLGConfirmForDepositToSCMeta
		} else {
			burningMetaType = metadata.BurningPLGConfirmMeta
		}
		prefix = common.PLGPrefix
	case common.FTMNetworkID:
		if isDepositToSC {
			burningMetaType = metadata.BurningFantomConfirmForDepositToSCMeta
		} else {
			burningMetaType = metadata.BurningFantomConfirmMeta
		}
		prefix = common.FTMPrefix
	case common.DefaultNetworkID:
		return burningMetaType, prefix, NewBridgeAggErrorWithValue(OtherError, errors.New("Cannot get info from default networkID"))
	default:
		return burningMetaType, prefix, NewBridgeAggErrorWithValue(OtherError, errors.New("Cannot detect networkID"))
	}

	return burningMetaType, prefix, nil
}

func CalculateIncDecimal(decimal, baseDecimal uint8) uint8 {
	if decimal > baseDecimal {
		return baseDecimal
	}
	return decimal
}

func validateConfigVault(sDBs map[int]*statedb.StateDB, tokenID common.Hash, vault config.Vault) error {
	networkID := vault.NetworkID
	if networkID != common.BSCNetworkID && networkID != common.ETHNetworkID && networkID != common.PLGNetworkID && networkID != common.FTMNetworkID {
		return fmt.Errorf("Cannot find networkID %d", networkID)
	}
	if vault.ExternalDecimal == 0 {
		return fmt.Errorf("ExternalTokenID cannot be 0")
	}
	if vault.ExternalTokenID == utils.EmptyString {
		return fmt.Errorf("ExternalTokenID can not be empty")
	}
	if tokenID == common.PRVCoinID || tokenID == common.PDEXCoinID {
		return fmt.Errorf("incTokenID is prv or pdex")
	}
	bridgeTokenInfoIndex, externalTokenIDIndex, err := GetBridgeTokenIndex(sDBs[common.BeaconChainID])
	if err != nil {
		return err
	}
	externalTokenID, err := getExternalTokenIDByNetworkID(vault.ExternalTokenID, networkID)
	if err != nil {
		return err
	}
	if bridgeTokenInfoState, found := bridgeTokenInfoIndex[tokenID]; found {
		if !bytes.Equal(bridgeTokenInfoState.ExternalTokenID, externalTokenID) {
			return errors.New("ExternalTokenID is not valid with data from db")
		}
	} else {
		encodedExternalTokenID := base64.StdEncoding.EncodeToString(externalTokenID)
		if externalTokenIDIndex[encodedExternalTokenID] {
			return errors.New("ExternalTokenID has existed")
		}
		isExisted, err := statedb.CheckTokenIDExisted(sDBs, tokenID)
		if err != nil {
			return fmt.Errorf("WARNING: Error in finding tokenID %s", tokenID.String())
		}
		if isExisted {
			return fmt.Errorf("WARNING: tokenID %s has existed", tokenID.String())
		}
	}
	networkType, _ := metadataBridge.GetNetworkTypeByNetworkID(networkID)
	if networkType == common.EVMNetworkType {
		externalTokenIDStr := vault.ExternalTokenID
		if len(externalTokenIDStr) != len("0x")+common.EVMAddressLength {
			return fmt.Errorf("ExternalTokenID %s is invalid length", externalTokenIDStr)
		}
		if !bytes.Equal([]byte(externalTokenIDStr[:len("0x")]), []byte("0x")) {
			return fmt.Errorf("ExternalTokenID %s is invalid format", externalTokenIDStr)
		}
		if !rCommon.IsHexAddress(externalTokenIDStr[len("0x"):]) {
			return fmt.Errorf("ExternalTokenID %s is invalid format", externalTokenIDStr)
		}
	}
	return nil
}

func getExternalTokenIDByNetworkID(externalTokenID string, networkID uint8) ([]byte, error) {
	var res []byte
	var prefix string
	switch networkID {
	case common.ETHNetworkID:
		prefix = utils.EmptyString
	case common.BSCNetworkID:
		prefix = common.BSCPrefix
	case common.PLGNetworkID:
		prefix = common.PLGPrefix
	case common.FTMNetworkID:
		prefix = common.FTMPrefix
	default:
		return nil, fmt.Errorf("Invalid networkID %v", networkID)
	}
	networkType, err := metadataBridge.GetNetworkTypeByNetworkID(networkID)
	if err != nil {
		return nil, err
	}
	switch networkType {
	case common.EVMNetworkType:
		tokenAddr := rCommon.HexToAddress(externalTokenID)
		res = append([]byte(prefix), tokenAddr.Bytes()...)
	}
	return res, nil
}

func updateRewardReserve(lastUpdatedRewardReserve, currentRewardReserve, newRewardReserve uint64) (uint64, uint64, error) {
	if lastUpdatedRewardReserve == currentRewardReserve && lastUpdatedRewardReserve == newRewardReserve && newRewardReserve == 0 {
		return 0, 0, nil
	}
	var resLastUpdatedRewardReserve uint64
	tmp := big.NewInt(0).Sub(big.NewInt(0).SetUint64(lastUpdatedRewardReserve), big.NewInt(0).SetUint64(currentRewardReserve))
	if tmp.Cmp(big.NewInt(0).SetUint64(newRewardReserve)) >= 0 {
		return 0, 0, errors.New("deltaY is >= newRewardReserve")
	}

	resLastUpdatedRewardReserve = newRewardReserve
	tmpRewardReserve := big.NewInt(0).Sub(big.NewInt(0).SetUint64(newRewardReserve), tmp)
	if !tmpRewardReserve.IsUint64() {
		return 0, 0, errors.New("Out of range uint64")
	}
	return resLastUpdatedRewardReserve, tmpRewardReserve.Uint64(), nil
}

func GetExternalTokenIDForUnifiedToken() []byte {
	return []byte(common.UnifiedTokenPrefix)
}

func getPrefixByNetworkID(networkID uint8) (string, error) {
	var prefix string
	switch networkID {
	case common.ETHNetworkID:
		prefix = utils.EmptyString
	case common.BSCNetworkID:
		prefix = common.BSCPrefix
	case common.PLGNetworkID:
		prefix = common.PLGPrefix
	case common.FTMNetworkID:
		prefix = common.FTMPrefix
	default:
		return utils.EmptyString, errors.New("Invalid networkID")
	}
	return prefix, nil
}

func CalculateReceivedAmount(amount uint64, tokenID common.Hash, decimal uint8, networkID uint8, sDB *statedb.StateDB) (uint64, error) {
	prefix, err := getPrefixByNetworkID(networkID)
	if err != nil {
		return 0, err
	}
	externalTokenID, err := GetExternalTokenIDByIncTokenID(tokenID, sDB)
	if err != nil {
		return 0, err
	}

	if !bytes.Equal(append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), externalTokenID) {
		if decimal > config.Param().BridgeAggParam.BaseDecimal {
			decimal = config.Param().BridgeAggParam.BaseDecimal
		}
	}
	unshieldAmount, err := ConvertAmountByDecimal(big.NewInt(0).SetUint64(amount), decimal, false)
	if err != nil {
		return 0, err
	}
	if unshieldAmount.Cmp(big.NewInt(0)) == 0 {
		return 0, errors.New("Received amount is 0")
	}
	return unshieldAmount.Uint64(), nil
}

func CalculateMaxReceivedAmount(x, y uint64) (uint64, error) {
	if x <= 1 {
		return 0, nil
	}
	return x - 1, nil
}

func decreaseVaultAmount(v *statedb.BridgeAggVaultState, amount uint64) (*statedb.BridgeAggVaultState, error) {
	temp := v.Amount() - amount
	if temp > v.Amount() {
		return nil, errors.New("decrease out of range uint64")
	}
	v.SetAmount(temp)
	return v, nil
}

func increaseVaultAmount(v *statedb.BridgeAggVaultState, amount uint64) (*statedb.BridgeAggVaultState, error) {
	temp := v.Amount() + amount
	if temp < v.Amount() {
		return nil, errors.New("increase out of range uint64")
	}
	v.SetAmount(temp)
	return v, nil
}

func convertPTokenAmtToPUnifiedTokenAmt(extDec uint8, amount uint64) (uint64, error) {
	pDecimal := CalculateIncDecimal(extDec, config.Param().BridgeAggParam.BaseDecimal)
	tmpAmount, err := ConvertAmountByDecimal(big.NewInt(0).SetUint64(amount), pDecimal, true)
	if err != nil {
		return 0, err
	}
	if tmpAmount.Cmp(big.NewInt(0)) == 0 {
		return 0, fmt.Errorf("amount %d is not enough for converting", amount)
	}
	if !tmpAmount.IsUint64() {
		return 0, fmt.Errorf("convert amount %d is invalid", amount)
	}
	return tmpAmount.Uint64(), nil
}

// calculate actual received amount and actual fee
func CalUnshieldFeeByShortageBurnAmount(burnAmount uint64, percentFeeWithDec uint64) (uint64, error) {
	percentFeeDec := config.Param().BridgeAggParam.PercentFeeDecimal
	// receiveAmt + fee = shortageAmt
	// fee = (percentFee * shortageAmt) / (percentFee + 1)
	// = (percent * shortageAmt) / (percent + dec)
	x := new(big.Int).Mul(
		new(big.Int).SetUint64(percentFeeWithDec),
		new(big.Int).SetUint64(burnAmount),
	)
	y := new(big.Int).Add(
		new(big.Int).SetUint64(percentFeeWithDec),
		new(big.Int).SetUint64(percentFeeDec),
	)

	fee := new(big.Int).Div(x, y).Uint64()

	if fee == 0 && percentFeeWithDec != 0 {
		fee = 1 // at least 1
	}
	if fee > burnAmount {
		return 0, fmt.Errorf("Needed fee %v larger than shortage amount %v", fee, burnAmount)
	}
	return fee, nil
}

// calculate actual received amount and actual fee
func CalUnshieldFeeByShortageReceivedAmount(receivedAmount uint64, percentFeeWithDec uint64) (uint64, error) {
	percentFeeDec := config.Param().BridgeAggParam.PercentFeeDecimal

	// fee = percentFee * receivedAmount
	feeBN := new(big.Int).Mul(
		new(big.Int).SetUint64(receivedAmount),
		new(big.Int).SetUint64(percentFeeWithDec),
	)
	feeBN = feeBN.Div(feeBN, new(big.Int).SetUint64(percentFeeDec))
	fee := feeBN.Uint64()

	if fee == 0 && percentFeeWithDec != 0 {
		fee = 1 // at least 1
	}

	return fee, nil
}

func CalRewardForRefillVault(v *statedb.BridgeAggVaultState, shieldAmt uint64) (uint64, error) {
	// no demand for unshield
	if v.WaitingUnshieldAmount() == 0 {
		return 0, nil
	}

	if shieldAmt >= v.WaitingUnshieldAmount() {
		return v.WaitingUnshieldFee(), nil
	}

	res := new(big.Int).Mul(new(big.Int).SetUint64(shieldAmt), new(big.Int).SetUint64(v.WaitingUnshieldFee()))
	res = res.Div(res, new(big.Int).SetUint64(v.WaitingUnshieldAmount()))

	if !res.IsUint64() {
		return 0, errors.New("Out of range uint64")
	}
	return res.Uint64(), nil
}

func updateVaultForRefill(v *statedb.BridgeAggVaultState, shieldAmt, reward uint64) (*statedb.BridgeAggVaultState, error) {
	res := v.Clone()
	// increase vault amount
	newAmount := new(big.Int).Add(new(big.Int).SetUint64(v.Amount()), new(big.Int).SetUint64(shieldAmt))
	if !newAmount.IsUint64() {
		return v, errors.New("Out of range uint64")
	}
	res.SetAmount(newAmount.Uint64())

	// decrease waiting unshield amount, waiting fee
	if v.WaitingUnshieldAmount() > 0 {
		// shieldAmt is maybe greater than WaitingUnshieldAmount in Vault
		if v.WaitingUnshieldAmount() <= shieldAmt {
			res.SetWaitingUnshieldAmount(0)
			res.SetLockedAmount(v.LockedAmount() + v.WaitingUnshieldAmount())
		} else {
			res.SetWaitingUnshieldAmount(v.WaitingUnshieldAmount() - shieldAmt)
			res.SetLockedAmount(v.LockedAmount() + shieldAmt)
		}

		// reward can't be greater than WaitingUnshieldFee in Vault
		if v.WaitingUnshieldFee() < reward {
			return v, fmt.Errorf("Invalid reward %v: can't be greater than WaitingUnshieldFee in Vault %v", reward, v.WaitingUnshieldFee())
		}
		res.SetWaitingUnshieldFee(v.WaitingUnshieldFee() - reward)
	}

	return res, nil
}

func checkVaultForWaitUnshieldReq(
	vaults map[common.Hash]*statedb.BridgeAggVaultState,
	unshieldDatas []statedb.WaitingUnshieldReqData,
	lockedVaults map[common.Hash]uint64,
) (bool, map[common.Hash]uint64) {
	// check vaults are enough for process waiting unshield req
	isEnoughVault := true
	for _, data := range unshieldDatas {
		if lockedVaults[data.IncTokenID] >= vaults[data.IncTokenID].Amount() {
			isEnoughVault = false
			break
		}
		receivedAmount := data.BurningAmount - data.Fee
		if vaults[data.IncTokenID].Amount()-lockedVaults[data.IncTokenID] < receivedAmount {
			isEnoughVault = false
			break
		}
	}
	// update lockedVaults (mem) if not enough
	if !isEnoughVault {
		for _, data := range unshieldDatas {
			receivedAmount := data.BurningAmount - data.Fee
			lockedVaults[data.IncTokenID] += receivedAmount
		}
	}
	return isEnoughVault, lockedVaults
}

func CalUnshieldFeeByBurnAmount(v *statedb.BridgeAggVaultState, burningAmt uint64, percentFeeWithDec uint64) (bool, uint64, error) {
	isEnoughVault := true
	shortageAmt := uint64(0)
	fee := uint64(0)
	var err error

	// calculate shortage amount in this vault
	if v.Amount() <= v.LockedAmount() {
		// all amount in vault was locked
		shortageAmt = burningAmt
	} else {
		remainAmt := v.Amount() - v.LockedAmount()
		if remainAmt < burningAmt {
			shortageAmt = burningAmt - remainAmt
		}
	}
	if shortageAmt > 0 {
		isEnoughVault = false

		// calculate unshield fee by shortage amount
		fee, err = CalUnshieldFeeByShortageBurnAmount(shortageAmt, percentFeeWithDec)
		if err != nil {
			return false, 0, fmt.Errorf("Error when calculating unshield fee %v", err)
		}
	}

	return isEnoughVault, fee, nil
}

func CalUnshieldFeeByReceivedAmount(v *statedb.BridgeAggVaultState, receivedAmt uint64, percentFeeWithDec uint64) (bool, uint64, error) {
	isEnoughVault := true
	shortageAmt := uint64(0)
	fee := uint64(0)
	var err error

	// calculate shortage amount in this vault
	if v.Amount() <= v.LockedAmount() {
		// all amount in vault was locked
		shortageAmt = receivedAmt
	} else {
		remainAmt := v.Amount() - v.LockedAmount()
		if remainAmt < receivedAmt {
			shortageAmt = receivedAmt - remainAmt
		}
	}
	if shortageAmt > 0 {
		isEnoughVault = false

		// calculate unshield fee by shortage amount
		fee, err = CalUnshieldFeeByShortageReceivedAmount(shortageAmt, percentFeeWithDec)
		if err != nil {
			return false, 0, fmt.Errorf("Error when calculating unshield fee %v", err)
		}
	}

	return isEnoughVault, fee, nil
}

func checkVaultForNewUnshieldReq(
	vaults map[common.Hash]*statedb.BridgeAggVaultState,
	unshieldDatas []metadataBridge.UnshieldRequestData,
	isDepositToSC bool,
	percentFeeWithDec uint64,
	stateDB *statedb.StateDB,
) (bool, []statedb.WaitingUnshieldReqData, error) {
	waitingUnshieldDatas := []statedb.WaitingUnshieldReqData{}
	isEnoughVault := true

	for _, data := range unshieldDatas {
		v := vaults[data.IncTokenID]
		if v == nil {
			return false, nil, fmt.Errorf("Can not found vault with incTokenID %v", data.IncTokenID)
		}

		// calculate unshield fee
		isEnoughVaultTmp, fee, err := CalUnshieldFeeByBurnAmount(v, data.BurningAmount, percentFeeWithDec)
		if err != nil {
			return false, nil, fmt.Errorf("Error when calculating unshield fee %v", err)
		}
		// reject if vault not enough for deposit to SC
		if !isEnoughVaultTmp && isDepositToSC {
			return false, nil, fmt.Errorf("Not enough vaults for unshield to deposit to SC - IncTokenID %v", data.IncTokenID)
		}

		// update isEnoughVault = false when there is any vault is not enough
		if !isEnoughVaultTmp {
			isEnoughVault = false
		}

		// check minExpectedAmount
		actualAmt := data.BurningAmount - fee
		if actualAmt < data.MinExpectedAmount {
			return false, nil, fmt.Errorf("Min expected amount is invalid, expect not greater than %v, but get %v", actualAmt, data.MinExpectedAmount)
		}

		// find the corresponding external tokenID
		prefix, err := getPrefixByNetworkID(v.NetworkID())
		if err != nil {
			return false, nil, fmt.Errorf("Error when getting prefix external token ID by networkID %v", err)
		}
		externalTokenID, err := metadataBridge.FindExternalTokenID(stateDB, data.IncTokenID, prefix)
		if err != nil {
			return false, nil, fmt.Errorf("Error when finding external token ID with IncTokenID %v - %v", data.IncTokenID, err)
		}

		// calculate external received amount
		extDecimal := v.ExtDecimal()
		if !bytes.Equal(append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), externalTokenID) {
			if extDecimal > config.Param().BridgeAggParam.BaseDecimal {
				extDecimal = config.Param().BridgeAggParam.BaseDecimal
			}
		}
		extReceivedAmt, err := ConvertAmountByDecimal(new(big.Int).SetUint64(actualAmt), extDecimal, false)
		if err != nil {
			return false, nil, fmt.Errorf("Error when convert to external received amount %v", err)
		}
		if extReceivedAmt.Cmp(big.NewInt(0)) == 0 {
			return false, nil, errors.New("Cannot received unshield amount equal to 0")
		}

		// get burning confirm metadata type
		burningConfirmMetaType, err := getBurningConfirmMetaType(v.NetworkID(), isDepositToSC)
		if err != nil {
			return false, nil, fmt.Errorf("Error when getting burning confirm metadata type %v", err)
		}

		waitingUnshieldData := statedb.WaitingUnshieldReqData{
			IncTokenID:             data.IncTokenID,
			BurningAmount:          data.BurningAmount,
			RemoteAddress:          data.RemoteAddress,
			Fee:                    fee,
			ExternalTokenID:        externalTokenID,
			ExternalReceivedAmt:    extReceivedAmt,
			BurningConfirmMetaType: burningConfirmMetaType,
		}

		waitingUnshieldDatas = append(waitingUnshieldDatas, waitingUnshieldData)
	}
	return isEnoughVault, waitingUnshieldDatas, nil
}

func addWaitingUnshieldReq(state *State, waitingUnshieldReq *statedb.BridgeAggWaitingUnshieldReq, unifiedTokenID common.Hash) *State {
	if state.waitingUnshieldReqs[unifiedTokenID] == nil {
		state.waitingUnshieldReqs[unifiedTokenID] = []*statedb.BridgeAggWaitingUnshieldReq{}
	}
	state.waitingUnshieldReqs[unifiedTokenID] = append(state.waitingUnshieldReqs[unifiedTokenID], waitingUnshieldReq)

	if state.newWaitingUnshieldReqs[unifiedTokenID] == nil {
		state.newWaitingUnshieldReqs[unifiedTokenID] = []*statedb.BridgeAggWaitingUnshieldReq{}
	}
	state.newWaitingUnshieldReqs[unifiedTokenID] = append(state.newWaitingUnshieldReqs[unifiedTokenID], waitingUnshieldReq)

	return state
}

func deleteWaitingUnshieldReq(state *State, waitingUnshieldReq *statedb.BridgeAggWaitingUnshieldReq, unifiedTokenID common.Hash) (*State, error) {
	tmpReqs := state.waitingUnshieldReqs[unifiedTokenID]
	indexReq := -1
	for i, req := range tmpReqs {
		if bytes.Equal(req.GetUnshieldID().Bytes(), waitingUnshieldReq.GetUnshieldID().Bytes()) {
			indexReq = i
			break
		}
	}
	if indexReq == -1 {
		return state, errors.New("Can not find waiting unshield req to delete")
	}
	state.waitingUnshieldReqs[unifiedTokenID] = append(tmpReqs[:indexReq], tmpReqs[indexReq+1:]...)

	key := statedb.GenerateBridgeAggWaitingUnshieldReqObjectKey(unifiedTokenID, waitingUnshieldReq.GetUnshieldID())
	state.deletedWaitingUnshieldReqKeyHashes = append(state.deletedWaitingUnshieldReqKeyHashes, key)

	return state, nil
}

func updateStateForNewWaitingUnshieldReq(
	vaults map[common.Hash]*statedb.BridgeAggVaultState,
	waitingUnshieldReq *statedb.BridgeAggWaitingUnshieldReq,
) map[common.Hash]*statedb.BridgeAggVaultState {
	for _, data := range waitingUnshieldReq.GetData() {
		v := vaults[data.IncTokenID]
		receiveAmt := data.BurningAmount - data.Fee

		remainAmt := v.Amount() - v.LockedAmount()
		matchUnshieldAmt := receiveAmt
		if matchUnshieldAmt > remainAmt {
			matchUnshieldAmt = remainAmt
		}

		v.SetLockedAmount(v.LockedAmount() + matchUnshieldAmt)
		v.SetWaitingUnshieldAmount(v.WaitingUnshieldAmount() + receiveAmt - matchUnshieldAmt)
		v.SetWaitingUnshieldFee(v.WaitingUnshieldFee() + data.Fee)

		vaults[data.IncTokenID] = v
	}
	return vaults
}

func updateStateForNewMatchedUnshieldReq(
	vaults map[common.Hash]*statedb.BridgeAggVaultState,
	waitingUnshieldReq *statedb.BridgeAggWaitingUnshieldReq,
) map[common.Hash]*statedb.BridgeAggVaultState {
	for _, data := range waitingUnshieldReq.GetData() {
		v := vaults[data.IncTokenID]
		v.SetAmount(v.Amount() - data.BurningAmount)
		vaults[data.IncTokenID] = v
	}
	return vaults
}

func updateStateForMatchedWaitUnshieldReq(
	vaults map[common.Hash]*statedb.BridgeAggVaultState,
	waitUnshieldReq *statedb.BridgeAggWaitingUnshieldReq,
) (map[common.Hash]*statedb.BridgeAggVaultState, error) {
	for _, data := range waitUnshieldReq.GetData() {
		v := vaults[data.IncTokenID]
		actualUnshieldAmt := data.BurningAmount - data.Fee
		if v.Amount() < actualUnshieldAmt || v.LockedAmount() < actualUnshieldAmt {
			return nil, fmt.Errorf("actualUnshieldAmt %v greater than vault amount %v or vault locked amount %v",
				actualUnshieldAmt, v.Amount(), v.LockedAmount())
		}
		v.SetAmount(v.Amount() - actualUnshieldAmt)
		v.SetLockedAmount(v.LockedAmount() - actualUnshieldAmt)
		vaults[data.IncTokenID] = v
	}
	return vaults, nil
}

func updateStateForUnshield(
	state *State,
	unifiedTokenID common.Hash,
	waitingUnshieldReq *statedb.BridgeAggWaitingUnshieldReq,
	statusStr string,
) (*State, error) {
	vaults, err := state.CloneVaultsByUnifiedTokenID(unifiedTokenID)
	if err != nil {
		return state, err
	}

	switch statusStr {
	// add new unshield req to waiting list
	case common.WaitingStatusStr:
		{
			// add to waiting list
			state = addWaitingUnshieldReq(state, waitingUnshieldReq, unifiedTokenID)
			// update vault state
			state.unifiedTokenVaults[unifiedTokenID] = updateStateForNewWaitingUnshieldReq(vaults, waitingUnshieldReq)
		}

	// a unshield req in waiting list is filled
	case common.FilledStatusStr:
		{
			// delete from waiting list
			state, err = deleteWaitingUnshieldReq(state, waitingUnshieldReq, unifiedTokenID)
			if err != nil {
				return state, err
			}
			// update vault state
			updatedVaults, err := updateStateForMatchedWaitUnshieldReq(vaults, waitingUnshieldReq)
			if err != nil {
				return state, err
			}
			state.unifiedTokenVaults[unifiedTokenID] = updatedVaults
		}
	// new unshield req is accepted with current state
	case common.AcceptedStatusStr:
		{
			state.unifiedTokenVaults[unifiedTokenID] = updateStateForNewMatchedUnshieldReq(vaults, waitingUnshieldReq)
		}
	default:
		{
			return state, errors.New("Invalid unshield instruction status")
		}
	}
	return state, nil
}

func getStatusByteFromStatuStr(statusStr string) (byte, error) {
	switch statusStr {
	case common.RejectedStatusStr:
		return common.RejectedStatusByte, nil
	case common.AcceptedStatusStr:
		return common.AcceptedStatusByte, nil
	case common.WaitingStatusStr:
		return common.WaitingStatusByte, nil
	case common.FilledStatusStr:
		return common.FilledStatusByte, nil
	default:
		return 0, errors.New("Invalid status string")
	}
}

func updateStateForModifyParam(state *State, newPercentFeeWithDec uint64) *State {
	if state.param == nil {
		state.param = statedb.NewBridgeAggParamState()
	}

	state.param.SetPercentFeeWithDec(newPercentFeeWithDec)
	return state
}

type UnshieldActionForProducer struct {
	metadataCommon.Action
	ShardID      byte
	BeaconHeight uint64
}

func BuildUnshieldActionForProducerFromInsts(insts [][]string, shardID byte, beaconHeight uint64) []UnshieldActionForProducer {
	unshieldActions := []UnshieldActionForProducer{}
	for _, inst := range insts {
		if len(inst) < 2 {
			continue
		}
		contentStr := inst[1]

		// decode action from shard
		action := metadataCommon.NewAction()
		meta := &metadataBridge.UnshieldRequest{}
		action.Meta = meta
		err := action.FromString(contentStr)
		if err != nil {
			Logger.log.Warnf("Can not decode content string for unshield from shard: %v\n", err)
			continue
		}

		unshieldActions = append(unshieldActions, UnshieldActionForProducer{
			Action:       *action,
			ShardID:      shardID,
			BeaconHeight: beaconHeight,
		})
	}
	return unshieldActions
}

package bridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func IsBridgeAggMetaType(metaType int) bool {
	switch metaType {
	case metadataCommon.BridgeAggModifyListTokenMeta:
		return true
	case metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta:
		return true
	case metadataCommon.BridgeAggConvertTokenToUnifiedTokenResponseMeta:
		return true
	case metadataCommon.ShieldUnifiedTokenRequestMeta:
		return true
	case metadataCommon.ShieldUnifiedTokenResponseMeta:
		return true
	case metadataCommon.UnshieldUnifiedTokenRequestMeta:
		return true
	case metadataCommon.UnshieldUnifiedTokenResponseMeta:
		return true
	default:
		return false
	}
}

type Vault struct {
	statedb.BridgeAggConvertedTokenState
	RewardReserve uint64 `json:"RewardReserve"`
	Decimal       uint   `json:"Decimal"`
}

func (v *Vault) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		State         *statedb.BridgeAggConvertedTokenState `json:"State"`
		RewardReserve uint64                                `json:"RewardReserve"`
		Decimal       uint                                  `json:"Decimal"`
	}{
		State:         &v.BridgeAggConvertedTokenState,
		RewardReserve: v.RewardReserve,
		Decimal:       v.Decimal,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (v *Vault) UnmarshalJSON(data []byte) error {
	temp := struct {
		State         *statedb.BridgeAggConvertedTokenState `json:"State"`
		RewardReserve uint64                                `json:"RewardReserve"`
		Decimal       uint                                  `json:"Decimal"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	v.RewardReserve = temp.RewardReserve
	if temp.State != nil {
		v.BridgeAggConvertedTokenState = *temp.State
	}
	v.Decimal = temp.Decimal
	return nil
}

func IsBridgeTxHashUsedInBlock(uniqTx []byte, uniqTxsUsed [][]byte) bool {
	for _, item := range uniqTxsUsed {
		if bytes.Equal(uniqTx, item) {
			return true
		}
	}
	return false
}

func getShardIDFromPaymentAddress(addressStr string) (byte, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(addressStr)
	if err != nil {
		return byte(0), err
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return byte(0), fmt.Errorf("Payment address' public key must not be empty")
	}
	// calculate shard ID
	lastByte := keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1]
	shardID := common.GetShardIDFromLastByte(lastByte)
	return shardID, nil
}

func ExtractIssueEVMData(
	stateDB *statedb.StateDB, shardID byte, listTxUsed [][]byte, contractAddress string, prefix string,
	isTxHashIssued func(stateDB *statedb.StateDB, uniqueEthTx []byte) (bool, error),
	txReceipt *types.Receipt, blockHash rCommon.Hash, txIndex uint,
) (uint64, byte, string, []byte, []byte, error) {
	if txReceipt == nil {
		return 0, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: bridge tx receipt is null")
	}

	// NOTE: since TxHash from constructedReceipt is always '0x0000000000000000000000000000000000000000000000000000000000000000'
	// so must build unique eth tx as combination of block hash and tx index.
	uniqTx := append(blockHash[:], []byte(strconv.Itoa(int(txIndex)))...)
	isUsedInBlock := IsBridgeTxHashUsedInBlock(uniqTx, listTxUsed)
	if isUsedInBlock {
		return 0, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: already issued for the hash in current block: ", uniqTx)
	}
	isIssued, err := isTxHashIssued(stateDB, uniqTx)
	if err != nil {
		return 0, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: an issue occured while checking the bridge tx hash is issued or not: ", err)
	}
	if isIssued {
		return 0, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: already issued for the hash in previous blocks: ", uniqTx)
	}

	logMap, err := PickAndParseLogMapFromReceipt(txReceipt, contractAddress)
	if err != nil {
		return 0, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: an error occurred while parsing log map from receipt: ", err)
	}
	if logMap == nil {
		return 0, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: could not find log map out from receipt")
	}

	logMapBytes, _ := json.Marshal(logMap)
	metadataCommon.Logger.Log.Warn("INFO: eth logMap json - ", string(logMapBytes))

	// the token might be ETH/ERC20 BNB/BEP20
	tokenAddr, ok := logMap["token"].(rCommon.Address)
	if !ok {
		return 0, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: could not parse evm token id from log map.")
	}
	token := append([]byte(prefix), tokenAddr.Bytes()...)

	amount := uint64(0)
	addressStr, ok := logMap["incognitoAddress"].(string)
	if !ok {
		return 0, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: could not parse incognito address from bridge log map.")
	}
	amt, ok := logMap["amount"].(*big.Int)
	if !ok {
		return 0, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: could not parse amount from bridge log map.")
	}
	if bytes.Equal(append([]byte(prefix), rCommon.HexToAddress(common.NativeToken).Bytes()...), token) {
		// convert amt from wei (10^18) to nano eth (10^9)
		amount = big.NewInt(0).Div(amt, big.NewInt(1000000000)).Uint64()
	} else { // ERC20 / BEP20
		amount = amt.Uint64()
	}

	receivingShardID, err := getShardIDFromPaymentAddress(addressStr)
	if err != nil {
		return 0, 0, utils.EmptyString, nil, nil, fmt.Errorf("WARNING: an error occurred while getting shard id from payment address: ", err)
	}
	return amount, receivingShardID, addressStr, token, uniqTx, nil
}

func VerifyTokenPair(
	stateDBs map[int]*statedb.StateDB,
	ac *metadataCommon.AccumulatedValues,
	incTokenID common.Hash,
	token []byte,
) error {
	canProcess, err := ac.CanProcessTokenPair(token, incTokenID)
	if err != nil {
		return fmt.Errorf("WARNING: an error occurred while checking it can process for token pair on the current block or not: ", err)
	}
	if !canProcess {
		return fmt.Errorf("WARNING: pair of incognito token id & bridge's id is invalid in current block")
	}
	privacyTokenExisted, err := statedb.CheckTokenIDExisted(stateDBs, incTokenID)
	if err != nil {
		return fmt.Errorf("WARNING: Cannot find tokenID %s", incTokenID.String())
	}
	isValid, err := statedb.CanProcessTokenPair(stateDBs[common.BeaconChainID], token, incTokenID, privacyTokenExisted)
	if err != nil {
		return fmt.Errorf("WARNING: an error occured while checking it can process for token pair on the previous blocks or not: ", err)
	}
	if !isValid {
		return fmt.Errorf("WARNING: pair of incognito token id & bridge's id is invalid with previous blocks")
	}
	return nil
}

func FindExternalTokenID(stateDB *statedb.StateDB, incTokenID common.Hash, prefix string, metaType int) ([]byte, error) {
	// Convert to external tokenID
	tokenID, err := findExternalTokenID(stateDB, &incTokenID)
	if err != nil {
		return nil, err
	}

	if len(tokenID) < common.ExternalBridgeTokenLength {
		return nil, errors.New("invalid external token id")
	}

	prefixLen := len(prefix)
	if (prefixLen > 0 && !bytes.Equal([]byte(prefix), tokenID[:prefixLen])) || len(tokenID) != (common.ExternalBridgeTokenLength+prefixLen) {
		return nil, errors.New(fmt.Sprintf("metadata type %v with invalid external tokenID %v", metaType, tokenID))
	}
	return tokenID, nil
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

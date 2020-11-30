package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/basemeta"

	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

// PDEFeeWithdrawalRequest - privacy dex withdrawal request
type PDEFeeWithdrawalRequest struct {
	WithdrawerAddressStr  string
	WithdrawalToken1IDStr string
	WithdrawalToken2IDStr string
	WithdrawalFeeAmt      uint64
	basemeta.MetadataBase
}

type PDEFeeWithdrawalRequestAction struct {
	Meta    PDEFeeWithdrawalRequest
	TxReqID common.Hash
	ShardID byte
}

func NewPDEFeeWithdrawalRequest(
	withdrawerAddressStr string,
	withdrawalToken1IDStr string,
	withdrawalToken2IDStr string,
	withdrawalFeeAmt uint64,
	metaType int,
) (*PDEFeeWithdrawalRequest, error) {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	pdeFeeWithdrawalRequest := &PDEFeeWithdrawalRequest{
		WithdrawerAddressStr:  withdrawerAddressStr,
		WithdrawalToken1IDStr: withdrawalToken1IDStr,
		WithdrawalToken2IDStr: withdrawalToken2IDStr,
		WithdrawalFeeAmt:      withdrawalFeeAmt,
	}
	pdeFeeWithdrawalRequest.MetadataBase = metadataBase
	return pdeFeeWithdrawalRequest, nil
}

func (pc PDEFeeWithdrawalRequest) ValidateTxWithBlockChain(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (pc PDEFeeWithdrawalRequest) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, tx basemeta.Transaction) (bool, bool, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(pc.WithdrawerAddressStr)
	if err != nil {
		return false, false, NewMetadataTxError(PDEFeeWithdrawalRequestFromMapError, errors.New("WithdrawerAddressStr incorrect"))
	}
	withdrawerAddr := keyWallet.KeySet.PaymentAddress
	if len(withdrawerAddr.Pk) == 0 {
		return false, false, errors.New("Wrong request info's withdrawer address")
	}
	if !bytes.Equal(tx.GetSigPubKey()[:], withdrawerAddr.Pk[:]) {
		return false, false, errors.New("WithdrawerAddr incorrect")
	}
	_, err = common.Hash{}.NewHashFromStr(pc.WithdrawalToken1IDStr)
	if err != nil {
		return false, false, NewMetadataTxError(PDEFeeWithdrawalRequestFromMapError, errors.New("WithdrawalTokenID1Str incorrect"))
	}
	_, err = common.Hash{}.NewHashFromStr(pc.WithdrawalToken2IDStr)
	if err != nil {
		return false, false, NewMetadataTxError(PDEFeeWithdrawalRequestFromMapError, errors.New("WithdrawalTokenID2Str incorrect"))
	}
	if pc.WithdrawalFeeAmt == 0 {
		return false, false, NewMetadataTxError(PDEFeeWithdrawalRequestFromMapError, errors.New("WithdrawalFeeAmt should be large than 0"))
	}
	return true, true, nil
}

func (pc PDEFeeWithdrawalRequest) ValidateMetadataByItself() bool {
	return pc.Type == basemeta.PDEFeeWithdrawalRequestMeta
}

func (pc PDEFeeWithdrawalRequest) Hash() *common.Hash {
	record := pc.MetadataBase.Hash().String()
	record += pc.WithdrawerAddressStr
	record += pc.WithdrawalToken1IDStr
	record += pc.WithdrawalToken2IDStr
	record += strconv.FormatUint(pc.WithdrawalFeeAmt, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (pc *PDEFeeWithdrawalRequest) BuildReqActions(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PDEFeeWithdrawalRequestAction{
		Meta:    *pc,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(pc.Type), actionContentBase64Str}
	return [][]string{action}, nil
}

func (pc *PDEFeeWithdrawalRequest) CalculateSize() uint64 {
	return basemeta.CalculateSize(pc)
}

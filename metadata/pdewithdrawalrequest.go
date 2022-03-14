package metadata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

// PDEWithdrawalRequest - privacy dex withdrawal request
type PDEWithdrawalRequest struct {
	WithdrawerAddressStr  string
	WithdrawalToken1IDStr string
	WithdrawalToken2IDStr string
	WithdrawalShareAmt    uint64
	MetadataBaseWithSignature
}

type PDEWithdrawalRequestAction struct {
	Meta    PDEWithdrawalRequest
	TxReqID common.Hash
	ShardID byte
}

type PDEWithdrawalAcceptedContent struct {
	WithdrawalTokenIDStr string
	WithdrawerAddressStr string
	DeductingPoolValue   uint64
	DeductingShares      uint64
	PairToken1IDStr      string
	PairToken2IDStr      string
	TxReqID              common.Hash
	ShardID              byte
}

func NewPDEWithdrawalRequest(
	withdrawerAddressStr string,
	withdrawalToken1IDStr string,
	withdrawalToken2IDStr string,
	withdrawalShareAmt uint64,
	metaType int,
) (*PDEWithdrawalRequest, error) {
	metadataBase := NewMetadataBaseWithSignature(metaType)
	pdeWithdrawalRequest := &PDEWithdrawalRequest{
		WithdrawerAddressStr:  withdrawerAddressStr,
		WithdrawalToken1IDStr: withdrawalToken1IDStr,
		WithdrawalToken2IDStr: withdrawalToken2IDStr,
		WithdrawalShareAmt:    withdrawalShareAmt,
	}
	pdeWithdrawalRequest.MetadataBaseWithSignature = *metadataBase
	return pdeWithdrawalRequest, nil
}

func (pc PDEWithdrawalRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (pc PDEWithdrawalRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	addr, err := AssertPaymentAddressAndTxVersion(pc.WithdrawerAddressStr, tx.GetVersion())
	if err != nil {
		return false, false, NewMetadataTxError(PDEWithdrawalRequestFromMapError, errors.New("WithdrawerAddressStr incorrect"))
	}

	if ok, err := pc.MetadataBaseWithSignature.VerifyMetadataSignature(addr.Pk, tx); err != nil || !ok {
		fmt.Println("Check authorized sender fail:", ok, err)
		return false, false, errors.New("WithdrawerAddr unauthorized")
	}

	_, err = common.Hash{}.NewHashFromStr(pc.WithdrawalToken1IDStr)
	if err != nil {
		return false, false, NewMetadataTxError(PDEWithdrawalRequestFromMapError, errors.New("WithdrawalTokenID1Str incorrect"))
	}
	_, err = common.Hash{}.NewHashFromStr(pc.WithdrawalToken2IDStr)
	if err != nil {
		return false, false, NewMetadataTxError(PDEWithdrawalRequestFromMapError, errors.New("WithdrawalTokenID2Str incorrect"))
	}
	if pc.WithdrawalShareAmt == 0 {
		return false, false, NewMetadataTxError(PDEWithdrawalRequestFromMapError, errors.New("WithdrawalShareAmt should be large than 0"))
	}
	return true, true, nil
}

func (pc PDEWithdrawalRequest) ValidateMetadataByItself() bool {
	return pc.Type == PDEWithdrawalRequestMeta
}

func (pc PDEWithdrawalRequest) Hash() *common.Hash {
	record := pc.MetadataBase.Hash().String()
	record += pc.WithdrawerAddressStr
	record += pc.WithdrawalToken1IDStr
	record += pc.WithdrawalToken2IDStr
	record += strconv.FormatUint(pc.WithdrawalShareAmt, 10)
	if pc.Sig != nil && len(pc.Sig) != 0 {
		record += string(pc.Sig)
	}
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (pc PDEWithdrawalRequest) HashWithoutSig() *common.Hash {
	record := pc.MetadataBase.Hash().String()
	record += pc.WithdrawerAddressStr
	record += pc.WithdrawalToken1IDStr
	record += pc.WithdrawalToken2IDStr
	record += strconv.FormatUint(pc.WithdrawalShareAmt, 10)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (pc *PDEWithdrawalRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PDEWithdrawalRequestAction{
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

func (pc *PDEWithdrawalRequest) CalculateSize() uint64 {
	return calculateSize(pc)
}

func (pc *PDEWithdrawalRequest) ToCompactBytes() ([]byte, error) {
	return toCompactBytes(pc)
}

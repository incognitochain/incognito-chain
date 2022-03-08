package metadata

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/pkg/errors"
)

type IssuingSOLRequest struct {
	TxSigStr   string
	IncTokenID common.Hash
	MetadataBase
}

type IssuingSOLReqAction struct {
	Meta                IssuingSOLRequest `json:"Meta"`
	TxReqID             common.Hash       `json:"TxReqId"`
	Amount              uint64            `json:"Amount"`
	ReceivingIncAddrStr string            `json:"ReceivingIncAddr"`
	ExternalTokenID     []byte            `json:"ExternalTokenIDStr"`
}

type IssuingSOLAcceptedInst struct {
	ShardID             byte        `json:"ShardId"`
	IssuingAmount       uint64      `json:"IssuingAmount"`
	ReceivingIncAddrStr string      `json:"ReceiverAddrStr"`
	IncTokenID          common.Hash `json:"IncTokenId"`
	TxReqID             common.Hash `json:"TxReqId"`
	UniqExternalTx      []byte      `json:"UniqExternalTx"`
	ExternalTokenID     []byte      `json:"ExternalTokenId"`
}

func ParseSOLIssuingInstContent(instContentStr string) (*IssuingSOLReqAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, NewMetadataTxError(IssuingSolReqDecodeInstructionError, err)
	}
	var IssuingSOLReqAction IssuingSOLReqAction
	err = json.Unmarshal(contentBytes, &IssuingSOLReqAction)
	if err != nil {
		return nil, NewMetadataTxError(IssuingSolReqUnmarshalJsonError, err)
	}
	return &IssuingSOLReqAction, nil
}

func ParseSOLIssuingInstAcceptedContent(instAcceptedContentStr string) (*IssuingSOLAcceptedInst, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instAcceptedContentStr)
	if err != nil {
		return nil, NewMetadataTxError(IssuingSolReqDecodeInstructionError, err)
	}
	var issuingSOLAcceptedInst IssuingSOLAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingSOLAcceptedInst)
	if err != nil {
		return nil, NewMetadataTxError(IssuingSolReqUnmarshalJsonError, err)
	}
	return &issuingSOLAcceptedInst, nil
}

func NewIssuingSOLRequest(
	txSig string,
	incTokenID common.Hash,
	metaType int,
) (*IssuingSOLRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	issuingSOLReq := &IssuingSOLRequest{
		TxSigStr:   txSig,
		IncTokenID: incTokenID,
	}
	issuingSOLReq.MetadataBase = metadataBase
	return issuingSOLReq, nil
}

func (iReq IssuingSOLRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (iReq IssuingSOLRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	if len(iReq.TxSigStr) == 0 {
		return false, false, NewMetadataTxError(IssuingSolReqValidateSanityDataError, errors.New("TxSig is empty"))
	}

	return true, true, nil
}

func (iReq IssuingSOLRequest) ValidateMetadataByItself() bool {
	if iReq.Type != IssuingSOLRequestMeta {
		return false
	}
	shieldInfo, err := iReq.verifyAndParseSolTxSig()
	if err != nil {
		Logger.log.Error(NewMetadataTxError(IssuingSolReqValidateTxWithBlockChainError, err))
		return false
	}
	if shieldInfo == nil {
		Logger.log.Error(errors.Errorf("The shielding info could not be null."))
		return false
	}
	return true
}

func (iReq *IssuingSOLRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&iReq)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (iReq *IssuingSOLRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TxSigStr   string      `json:"TxSig"`
		IncTokenID common.Hash `json:"IncTokenID"`
		metadataCommon.MetadataBase
	}{
		TxSigStr:     iReq.TxSigStr,
		IncTokenID:   iReq.IncTokenID,
		MetadataBase: iReq.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (iReq *IssuingSOLRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		TxSig      string      `json:"TxSig"`
		IncTokenID common.Hash `json:"IncTokenID"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	iReq.TxSigStr = temp.TxSig
	iReq.IncTokenID = temp.IncTokenID
	iReq.MetadataBase = temp.MetadataBase

	return nil
}

func (iReq *IssuingSOLRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	shieldInfo, err := iReq.verifyAndParseSolTxSig()
	if err != nil {
		return [][]string{}, NewMetadataTxError(IssuingSolReqBuildReqActionsError, err)
	}
	if shieldInfo == nil {
		return [][]string{}, NewMetadataTxError(IssuingSolReqBuildReqActionsError, errors.Errorf("The evm proof's receipt could not be null."))
	}
	txReqID := *(tx.Hash())
	actionContent := IssuingSOLReqAction{
		Meta:                *iReq,
		TxReqID:             txReqID,
		Amount:              shieldInfo.Amount,
		ReceivingIncAddrStr: shieldInfo.ReceivingIncAddrStr,
		ExternalTokenID:     shieldInfo.ExternalTokenID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, NewMetadataTxError(IssuingSolReqBuildReqActionsError, err)
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(iReq.Type), actionContentBase64Str}

	return [][]string{action}, nil
}

func (iReq *IssuingSOLRequest) CalculateSize() uint64 {
	return calculateSize(iReq)
}

func (iReq *IssuingSOLRequest) verifyAndParseSolTxSig() (*ShieldInfo, error) {
	solParam := config.Param().SOLParam
	solParam.GetFromEnv()
	host := solParam.Host
	solClient := rpc.New(host)

	// get sol transaction by txSig
	txSig, err := solana.SignatureFromBase58(iReq.TxSigStr)
	if err != nil {
		return nil, NewMetadataTxError(IssuingSolReqVerifyAndParseTxError, fmt.Errorf("Invalid tx signature with error %v", err))
	}
	res, err := solClient.GetTransaction(
		context.TODO(),
		txSig,
		&rpc.GetTransactionOpts{
			Encoding: solana.EncodingBase64,
		},
	)
	if err != nil {
		return nil, NewMetadataTxError(IssuingSolReqVerifyAndParseTxError, fmt.Errorf("Can not get sol tx signature with error %v", err))
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(res.Transaction.GetBinary()))
	if err != nil {
		return nil, NewMetadataTxError(IssuingSolReqVerifyAndParseTxError, fmt.Errorf("Can not decode sol tx signature with error %v", err))
	}

	// find instruction shield by program ID
	insts := tx.Message.Instructions
	programIDs := tx.Message.AccountKeys
	if len(insts) < 1 {
		return nil, NewMetadataTxError(IssuingSolReqVerifyAndParseTxError, errors.New("Transaction instruction is empty"))
	}
	expectedProgramIDPk, _ := solana.PublicKeyFromBase58(config.Param().SolProgramIDStr)
	var shieldInst *solana.CompiledInstruction
	for _, inst := range insts {
		programIDIdx := inst.ProgramIDIndex
		if len(programIDs)-1 < int(programIDIdx) {
			Logger.log.Warnf("Invalid program id index: %v - %v", int(programIDIdx), len(programIDs))
			continue
		}
		programID := programIDs[programIDIdx]

		if expectedProgramIDPk.Equals(programID) {
			shieldInst = &inst
			break
		}
	}
	if shieldInst == nil {
		return nil, NewMetadataTxError(IssuingSolReqVerifyAndParseTxError, errors.New("Can not find shielding instruction in txSig"))
	}

	// check instruction data length
	instData := shieldInst.Data
	if len(instData) < SolShieldInstLen {
		return nil, NewMetadataTxError(IssuingSolReqVerifyAndParseTxError, errors.New("Invalid instruction data length"))
	}

	// check instruction tag (the first byte)
	if instData[0] != SolShieldInstTag {
		return nil, NewMetadataTxError(IssuingSolReqVerifyAndParseTxError, errors.New("Invalid shielding instruction tag"))
	}

	// extract amount and receiving address
	amountBytes := instData[1:9]
	receivingAddr := instData[9:SolShieldInstLen]
	amount := binary.LittleEndian.Uint64(amountBytes[:])
	receivingAddrStr := string(receivingAddr[:])

	// parse accounts from instruction
	accs := shieldInst.ResolveInstructionAccounts(&tx.Message)
	if len(accs) < 5 {
		return nil, NewMetadataTxError(IssuingSolReqVerifyAndParseTxError, errors.New("Invalid instruction accounts length"))
	}

	// check Incognito proxy account
	incProxyAcc := accs[2]
	expectedIncProxyPk, _ := solana.PublicKeyFromBase58(config.Param().SolContractAddressStr)
	if !expectedIncProxyPk.Equals(incProxyAcc.PublicKey) {
		return nil, NewMetadataTxError(IssuingSolReqVerifyAndParseTxError, errors.New("Invalid incognito proxy account - is not expected"))
	}

	// get solana's token id from writable account (shield maker account)
	writableAccPk := accs[0].PublicKey
	writableAccInfo, err := solClient.GetAccountInfoWithOpts(context.TODO(), writableAccPk, &rpc.GetAccountInfoOpts{Encoding: "jsonParsed"})
	if err != nil {
		return nil, NewMetadataTxError(IssuingSolReqVerifyAndParseTxError, fmt.Errorf("Can not get writable account info with error %v", err))
	}

	infoJson := writableAccInfo.Value.Data.GetRawJSON()
	accData := SolAccountData{}
	err = json.Unmarshal(infoJson[:], &accData)
	if err != nil {
		return nil, NewMetadataTxError(IssuingSolReqVerifyAndParseTxError, fmt.Errorf("Can not unmarshal writable account data with error %v", err))
	}

	externalTokenIDStr := accData.Parsed.Info.Mint
	externalTokenIDBytes, err := solana.PublicKeyFromBase58(externalTokenIDStr)
	if err != nil {
		return nil, NewMetadataTxError(IssuingSolReqVerifyAndParseTxError, fmt.Errorf("Can not unmarshal writable account data with error %v", err))
	}

	return &ShieldInfo{
		Amount:              amount,
		ReceivingIncAddrStr: receivingAddrStr,
		ExternalTokenID:     externalTokenIDBytes[:],
	}, nil
}

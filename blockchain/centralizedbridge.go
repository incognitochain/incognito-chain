package blockchain

import (
	"encoding/base64"
	"encoding/json"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

type IssuingReqAction struct {
	Meta metadata.IssuingRequest `json:"meta"`
}

type ContractingReqAction struct {
	Meta metadata.ContractingRequest `json:"meta"`
}

func (blockgen *BlkTmplGenerator) buildIssuanceTx(
	tx metadata.Transaction,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	issuingReq := tx.GetMetadata().(*metadata.IssuingRequest)

	issuingTokenID := issuingReq.TokenID
	issuingRes := metadata.NewIssuingResponse(
		*tx.Hash(),
		metadata.IssuingResponseMeta,
	)

	txTokenVout := transaction.TxTokenVout{
		Value:          issuingReq.DepositedAmount,
		PaymentAddress: issuingReq.ReceiverAddress,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], issuingTokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenParamTx{
		PropertyID:     propID.String(),
		PropertyName:   propID.String(),
		PropertySymbol: propID.String(),
		Amount:         issuingReq.DepositedAmount,
		TokenTxType:    transaction.CustomTokenInit,
		Receiver:       []transaction.TxTokenVout{txTokenVout},
		Mintable:       true,
	}

	resTx := &transaction.TxCustomToken{}
	initErr := resTx.Init(
		producerPrivateKey,
		[]*privacy.PaymentInfo{},
		nil,
		0,
		tokenParams,
		blockgen.chain.config.DataBase,
		issuingRes,
		false,
		shardID,
	)

	if initErr != nil {
		Logger.log.Error(initErr)
		return nil, initErr
	}
	return resTx, nil
}

func processIssuingReq(
	bc *BlockChain,
	actionContentStr string,
) ([][]string, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(actionContentStr)
	if err != nil {
		return [][]string{}, err
	}
	var issuingReqAction IssuingReqAction
	err = json.Unmarshal(contentBytes, &issuingReqAction)
	if err != nil {
		return [][]string{}, err
	}
	md := issuingReqAction.Meta
	err = bc.GetDatabase().CountUpDepositedAmtByTokenID(&md.TokenID, md.DepositedAmount)
	if err != nil {
		return [][]string{}, err
	}
	return [][]string{}, nil
}

func processContractingReq(
	bc *BlockChain,
	actionContentStr string,
) ([][]string, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(actionContentStr)
	if err != nil {
		return [][]string{}, err
	}
	var contractingReqAction ContractingReqAction
	err = json.Unmarshal(contentBytes, &contractingReqAction)
	if err != nil {
		return [][]string{}, err
	}
	md := contractingReqAction.Meta
	err = bc.GetDatabase().DeductAmtByTokenID(&md.TokenID, md.BurnedAmount)
	if err != nil {
		return [][]string{}, err
	}
	return [][]string{}, nil
}

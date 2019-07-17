package blockchain

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

func buildInstructionsForIssuingReq(
	contentStr string,
	shardID byte,
	metaType int,
) ([][]string, error) {
	instructions := [][]string{}
	returnedInst := []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		"accepted",
		contentStr,
	}
	instructions = append(instructions, returnedInst)
	fmt.Println("hahah returnedInst : ", returnedInst)
	return instructions, nil
}

func (blockgen *BlkTmplGenerator) buildIssuanceTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	ac *metadata.AccumulatedValues,
) (metadata.Transaction, error) {
	if shardID != common.BRIDGE_SHARD_ID {
		return nil, nil
	}

	db := blockgen.chain.GetDatabase()
	fmt.Println("haha start buildIssuanceTx")

	issuingReqAction, err := metadata.ParseIssuingInstContent(contentStr)
	if err != nil {
		return nil, err
	}
	issuingReq := issuingReqAction.Meta
	issuingTokenID := issuingReq.TokenID
	issuingTokenName := issuingReq.TokenName
	if !ac.CanProcessCIncToken(issuingTokenID) {
		fmt.Printf("WARNING: The issuing token (%s) was already used in the current block.", issuingTokenID.String())
		return nil, nil
	}

	ok, err := db.CanProcessCIncToken(issuingTokenID)
	if err != nil {
		return nil, err
	}
	if !ok {
		fmt.Printf("WARNING: The issuing token (%s) was already used in the previous blocks.", issuingTokenID.String())
		return nil, nil
	}

	issuingRes := metadata.NewIssuingResponse(
		issuingReqAction.TxReqID,
		metadata.IssuingResponseMeta,
	)

	receiver := &privacy.PaymentInfo{
		Amount:         issuingReq.DepositedAmount,
		PaymentAddress: issuingReq.ReceiverAddress,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], issuingTokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID:     propID.String(),
		PropertyName:   issuingTokenName,
		PropertySymbol: issuingTokenName,
		Amount:         issuingReq.DepositedAmount,
		TokenTxType:    transaction.CustomTokenInit,
		Receiver:       []*privacy.PaymentInfo{receiver},
		TokenInput:     []*privacy.InputCoin{},
		Mintable:       true,
	}

	resTx := &transaction.TxCustomTokenPrivacy{}
	initErr := resTx.Init(
		producerPrivateKey,
		[]*privacy.PaymentInfo{},
		nil,
		0,
		tokenParams,
		blockgen.chain.config.DataBase,
		issuingRes,
		false,
		false,
		shardID,
	)

	if initErr != nil {
		Logger.log.Error(initErr)
		return nil, initErr
	}
	ac.CBridgeTokens = append(ac.CBridgeTokens, &issuingTokenID)
	return resTx, nil
}

func (blockgen *BlkTmplGenerator) buildETHIssuanceTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	ac *metadata.AccumulatedValues,
) (metadata.Transaction, error) {
	if shardID != common.BRIDGE_SHARD_ID {
		return nil, nil
	}

	fmt.Println("haha start buildETHIssuanceTx")
	db := blockgen.chain.GetDatabase()
	issuingETHReqAction, err := metadata.ParseETHIssuingInstContent(contentStr)
	if err != nil {
		return nil, err
	}

	md := issuingETHReqAction.Meta
	constructedReceipt, err := metadata.VerifyProofAndParseReceipt(issuingETHReqAction, blockgen.chain)
	if err != nil {
		fmt.Println("WARNING: an error occured during verifying proof & parsing receipt: ", err)
		return nil, err
	}
	if constructedReceipt == nil {
		return nil, nil
	}

	// NOTE: since TxHash from constructedReceipt is always '0x0000000000000000000000000000000000000000000000000000000000000000'
	// so must build unique eth tx as combination of block hash and tx index.
	uniqETHTx := append(md.BlockHash[:], []byte(strconv.Itoa(int(md.TxIndex)))...)
	isUsedInBlock := metadata.IsETHTxHashUsedInBlock(uniqETHTx, ac.UniqETHTxsUsed)
	if isUsedInBlock {
		fmt.Println("WARNING: already issued for the hash in current block: ", uniqETHTx)
		return nil, nil
	}
	isIssued, err := db.IsETHTxHashIssued(uniqETHTx)
	if err != nil {
		return nil, err
	}
	if isIssued {
		fmt.Println("WARNING: already issued for the hash in previous blocks: ", uniqETHTx)
		return nil, nil
	}

	logMap, err := metadata.PickNParseLogMapFromReceipt(constructedReceipt)
	if err != nil {
		fmt.Println("WARNING: an error occured during parsing log map from receipt: ", err)
		return nil, err
	}
	if logMap == nil {
		fmt.Println("WARNING: could not find log map out from receipt")
		return nil, nil
	}

	// the token might be ETH/ERC20
	ethereumAddr, ok := logMap["_token"].(rCommon.Address)
	if !ok {
		return nil, nil
	}
	ethereumToken := ethereumAddr.Bytes()
	canProcess, err := ac.CanProcessTokenPair(ethereumToken, md.IncTokenID)
	if err != nil {
		return nil, err
	}
	if !canProcess {
		fmt.Println("WARNING: pair of incognito token id & ethereum's id is invalid in current block")
		return nil, nil
	}

	isValid, err := db.CanProcessTokenPair(ethereumToken, md.IncTokenID)
	if err != nil {
		return nil, err
	}
	if !isValid {
		fmt.Println("WARNING: pair of incognito token id & ethereum's id is invalid with previous blocks")
		return nil, nil
	}

	addressStr := logMap["_incognito_address"].(string)
	key, err := wallet.Base58CheckDeserialize(addressStr)
	if err != nil {
		return nil, err
	}
	amt := logMap["_amount"].(*big.Int)
	amount := uint64(0)
	if bytes.Equal(rCommon.HexToAddress(common.ETH_ADDR_STR).Bytes(), ethereumToken) {
		// convert amt from wei (10^18) to nano eth (10^9)
		amount = big.NewInt(0).Div(amt, big.NewInt(1000000000)).Uint64()
	} else { // ERC20
		amount = amt.Uint64()
	}

	receiver := &privacy.PaymentInfo{
		Amount:         amount,
		PaymentAddress: key.KeySet.PaymentAddress,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], md.IncTokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID: propID.String(),
		// PropertyName:   common.PETHTokenName,
		// PropertySymbol: common.PETHTokenName,
		Amount:      amount,
		TokenTxType: transaction.CustomTokenInit,
		Receiver:    []*privacy.PaymentInfo{receiver},
		TokenInput:  []*privacy.InputCoin{},
		Mintable:    true,
	}

	issuingETHRes := metadata.NewIssuingETHResponse(
		issuingETHReqAction.TxReqID,
		uniqETHTx,
		ethereumToken,
		metadata.IssuingETHResponseMeta,
	)
	resTx := &transaction.TxCustomTokenPrivacy{}
	initErr := resTx.Init(
		producerPrivateKey,
		[]*privacy.PaymentInfo{},
		nil,
		0,
		tokenParams,
		blockgen.chain.config.DataBase,
		issuingETHRes,
		false,
		false,
		shardID,
	)

	if initErr != nil {
		return nil, initErr
	}
	ac.UniqETHTxsUsed = append(ac.UniqETHTxsUsed, uniqETHTx)
	ac.DBridgeTokenPair[md.IncTokenID.String()] = ethereumToken
	fmt.Println("haha create tx ok")
	return resTx, nil
}

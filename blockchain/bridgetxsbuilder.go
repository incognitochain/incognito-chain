package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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

func buildInstructionsForETHIssuingReq(
	contentStr string,
	shardID byte,
) ([][]string, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return [][]string{}, err
	}
	var issuingETHReqAction metadata.IssuingETHReqAction
	err = json.Unmarshal(contentBytes, &issuingETHReqAction)
	if err != nil {
		return nil, err
	}
	instructions := [][]string{}

	returnedInst := []string{
		strconv.Itoa(metadata.IssuingETHRequestMeta),
		strconv.Itoa(int(shardID)),
		"accepted",
		contentStr,
	}
	instructions = append(instructions, returnedInst)
	fmt.Println("hahah returnedInst : ", returnedInst)
	return instructions, nil
}

func (blockgen *BlkTmplGenerator) buildIssuanceTx(
	tx metadata.Transaction,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	issuingReq := tx.GetMetadata().(*metadata.IssuingRequest)

	issuingTokenID := issuingReq.TokenID
	issuingTokenName := issuingReq.TokenName
	issuingRes := metadata.NewIssuingResponse(
		*tx.Hash(),
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

	db := blockgen.chain.GetDatabase()
	fmt.Println("haha start buildETHIssuanceTx")

	issuingETHReqAction, err := metadata.ParseInstContent(contentStr)
	if err != nil {
		return nil, err
	}

	md := issuingETHReqAction.Meta
	constructedReceipt, err := metadata.VerifyProofAndParseReceipt(issuingETHReqAction, blockgen.chain)
	if err != nil {
		fmt.Println("WARNING: an error occured during verifying proof & parsing receipt: ", err)
		return nil, err
	}
	bb, _ := json.MarshalIndent(constructedReceipt, "", "    ")
	fmt.Println("haha constructedReceipt: ", string(bb))

	// NOTE: since TxHash from constructedReceipt is always '0x0000000000000000000000000000000000000000000000000000000000000000'
	// so must build unique eth tx as combination of block hash and tx index.
	uniqETHTx := append(md.BlockHash[:], []byte(strconv.Itoa(int(md.TxIndex)))...)
	isUsedInBlock := metadata.IsETHHashUsedInBlock(uniqETHTx, ac.UniqETHTxsUsed)
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

	jj, _ := json.Marshal(logMap)
	fmt.Println("haha logMap: ", string(jj))

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

	fmt.Println("haha addressStr: ", addressStr)
	fmt.Println("haha amount: ", amount)

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
	fmt.Println("haha create res tx ok")
	return resTx, nil
}

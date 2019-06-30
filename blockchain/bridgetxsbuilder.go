package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/ethrelaying/core/types"
	"github.com/incognitochain/incognito-chain/ethrelaying/light"
	"github.com/incognitochain/incognito-chain/ethrelaying/rlp"
	"github.com/incognitochain/incognito-chain/ethrelaying/trie"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func buildInstructionsForETHIssuingReq(
	contentStr string,
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
		strconv.Itoa(int(issuingETHReqAction.BridgeShardID)),
		"accepted",
		contentStr,
	}
	instructions = append(instructions, returnedInst)
	fmt.Println("hahah shard id : ", issuingETHReqAction.BridgeShardID)
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

func (blockgen *BlkTmplGenerator) buildETHHeaderRelayingRewardTx(
	tx metadata.Transaction,
	producerPrivateKey *privacy.PrivateKey,
	relayingRewardTx metadata.Transaction,
	maxHeaderLen int,
) (metadata.Transaction, int, error) {
	ethHeaderRelaying := tx.GetMetadata().(*metadata.ETHHeaderRelaying)
	ethHeaders := ethHeaderRelaying.ETHHeaders
	insertedHeadersLen := len(ethHeaders)
	if insertedHeadersLen <= maxHeaderLen {
		return relayingRewardTx, maxHeaderLen, nil
	}

	lc := blockgen.chain.LightEthereum.GetLightChain()
	_, err := lc.ValidateHeaderChain(ethHeaders, 0)
	if err != nil {
		fmt.Printf("ETH header relaying failed: %v", err)
		return relayingRewardTx, maxHeaderLen, nil
	}
	// TODO: figure out relaying reward amt here
	reward := tx.GetTxFee() + uint64(insertedHeadersLen*1)

	ethHeaderRelayingReward := metadata.NewETHHeaderRelayingReward(
		*tx.Hash(),
		metadata.ETHHeaderRelayingRewardMeta,
	)
	resTx := &transaction.Tx{}
	err = resTx.InitTxSalary(
		reward,
		&ethHeaderRelaying.RelayerAddress,
		producerPrivateKey,
		blockgen.chain.config.DataBase,
		ethHeaderRelayingReward,
	)
	if err != nil {
		return relayingRewardTx, maxHeaderLen, err
	}
	return resTx, insertedHeadersLen, nil
}

func (blockgen *BlkTmplGenerator) buildETHIssuanceTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	if shardID != 0 { // TODO: will have dedicated bridge shard with its shardID
		return nil, nil
	}

	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return nil, err
	}
	var issuingETHReqAction metadata.IssuingETHReqAction
	err = json.Unmarshal(contentBytes, &issuingETHReqAction)
	if err != nil {
		return nil, err
	}

	md := issuingETHReqAction.Meta
	ethHeader := blockgen.chain.LightEthereum.GetLightChain().GetHeaderByHash(md.BlockHash)
	keybuf := new(bytes.Buffer)
	keybuf.Reset()
	rlp.Encode(keybuf, md.TxIndex)

	nodeList := new(light.NodeList)
	for _, proofStr := range md.ProofStrs {
		proofBytes, err := base64.StdEncoding.DecodeString(proofStr)
		if err != nil {
			return nil, err
		}
		nodeList.Put([]byte{}, proofBytes)
	}
	proof := nodeList.NodeSet()
	val, _, err := trie.VerifyProof(ethHeader.ReceiptHash, keybuf.Bytes(), proof)
	if err != nil {
		fmt.Printf("ETH issuance proof verification failed: %v", err)
		return nil, nil
	}
	// Decode value from VerifyProof into Receipt
	constructedReceipt := new(types.Receipt)
	err = rlp.DecodeBytes(val, constructedReceipt)
	if err != nil {
		return nil, err
	}
	ethTxHash := constructedReceipt.TxHash
	isIssued, err := blockgen.chain.GetDatabase().IsETHTxHashIssued(ethTxHash)
	if err != nil {
		return nil, err
	}
	if isIssued {
		fmt.Println("WARNING: already issued for the hash: ", ethTxHash)
		return nil, nil
	}

	if len(constructedReceipt.Logs) == 0 {
		return nil, nil
	}
	logData := constructedReceipt.Logs[0].Data
	logMap, err := metadata.ParseETHLogData(logData)
	if err != nil {
		return nil, err
	}
	addressStr := logMap["_incognito_address"].(string)
	key, err := wallet.Base58CheckDeserialize(addressStr)
	if err != nil {
		return nil, err
	}
	amt := logMap["_amount"].(*big.Int)
	// convert amt from wei (10^18) to nano eth (10^9)
	amount := big.NewInt(0).Div(amt, big.NewInt(1000000000)).Uint64()
	receiver := &privacy.PaymentInfo{
		Amount:         amount,
		PaymentAddress: key.KeySet.PaymentAddress,
	}
	tokenID, err := common.NewHashFromStr(common.PETHTokenID)
	if err != nil {
		return nil, errors.Errorf("TokenID incorrect")
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], tokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID:     propID.String(),
		PropertyName:   common.PETHTokenName,
		PropertySymbol: common.PETHTokenName,
		Amount:         amount,
		TokenTxType:    transaction.CustomTokenInit,
		Receiver:       []*privacy.PaymentInfo{receiver},
		TokenInput:     []*privacy.InputCoin{},
		Mintable:       true,
	}

	issuingETHRes := metadata.NewIssuingETHResponse(
		issuingETHReqAction.TxReqID,
		ethTxHash,
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
	return resTx, nil
}

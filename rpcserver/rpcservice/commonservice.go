package rpcservice

import (
	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

func NewContractingRequestMetadata(senderPrivateKeyStr string, tokenReceivers interface{}, tokenID string) (*metadata.ContractingRequest, *RPCError) {
	senderKey, err := wallet.Base58CheckDeserialize(senderPrivateKeyStr)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	err = senderKey.KeySet.InitFromPrivateKey(&senderKey.KeySet.PrivateKey)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	paymentAddr := senderKey.KeySet.PaymentAddress

	_, voutsAmount, err := transaction.CreateCustomTokenReceiverArray(tokenReceivers)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	tokenIDHash, err := common.Hash{}.NewHashFromStr(tokenID)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}

	meta, _ := metadata.NewContractingRequest(
		paymentAddr,
		uint64(voutsAmount),
		*tokenIDHash,
		metadata.ContractingRequestMeta,
	)

	return meta, nil
}

func NewBurningRequestMetadata(senderPrivateKeyStr string, tokenReceivers interface{}, tokenID string, tokenName string, remoteAddress string) (*metadata.BurningRequest, *RPCError) {
	senderKey, err := wallet.Base58CheckDeserialize(senderPrivateKeyStr)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	err = senderKey.KeySet.InitFromPrivateKey(&senderKey.KeySet.PrivateKey)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	paymentAddr := senderKey.KeySet.PaymentAddress

	_, voutsAmount, err := transaction.CreateCustomTokenReceiverArray(tokenReceivers)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	tokenIDHash, err := common.Hash{}.NewHashFromStr(tokenID)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}

	meta, _ := metadata.NewBurningRequest(
		paymentAddr,
		uint64(voutsAmount),
		*tokenIDHash,
		tokenName,
		remoteAddress,
		metadata.BurningRequestMeta,
	)

	return meta, nil
}

func GetETHHeaderByHash(ethBlockHash string) (*types.Header, error) {
	return metadata.GetETHHeader(rCommon.HexToHash(ethBlockHash))
}

package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ninjadotorg/constant/database"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
)

func iPlusPlus(x *int) int {
	*x += 1
	return *x - 1
}

func (rpcServer RpcServer) handleGetAmountVoteToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	paymentAddressData := arrayParams[0].(string)
	paymentAddress, err := rpcServer.GetPaymentAddressFromSenderKeyParams(paymentAddressData)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	db := *rpcServer.config.Database
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}

	// For DCB voting token
	result.PaymentAddress = string(paymentAddressData)
	item := jsonresult.CustomTokenBalance{}
	item.Name = "DCB voting token"
	item.Symbol = "DCB Voting Token"
	TokenID := &common.Hash{}
	TokenID.SetBytes(common.DCBVotingTokenID[:])
	item.TokenID = TokenID.String()
	item.TokenImage = common.Render([]byte(item.TokenID))
	amount, err := db.GetVoteTokenAmount(metadata.DCBBoard.BoardTypeDB(), rpcServer.config.BlockChain.GetCurrentBoardIndex(blockchain.DCBConstitutionHelper{}), *paymentAddress)
	if err != nil {
		Logger.log.Error(err)
	}
	item.Amount = uint64(amount)
	result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)

	// For GOV voting token
	item = jsonresult.CustomTokenBalance{}
	item.Name = "GOV voting token"
	item.Symbol = "GOV Voting Token"
	TokenID = &common.Hash{}
	TokenID.SetBytes(common.GOVVotingTokenID[:])
	item.TokenID = TokenID.String()
	item.TokenImage = common.Render([]byte(item.TokenID))
	amount, err = db.GetVoteTokenAmount(metadata.GOVBoard.BoardTypeDB(), rpcServer.config.BlockChain.GetCurrentBoardIndex(blockchain.GOVConstitutionHelper{}), *paymentAddress)
	if err != nil {
		Logger.log.Error(err)
	}
	item.Amount = uint64(amount)
	result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)

	return result, nil
}

func (rpcServer RpcServer) handleSetAmountVoteToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	paymentAddressSenderKey := arrayParams[0].(string)
	paymentAddress, err1 := rpcServer.GetPaymentAddressFromSenderKeyParams(paymentAddressSenderKey)
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	p2, _ := rpcServer.GetPaymentAddressFromSenderKeyParams(string(paymentAddressSenderKey))
	_ = p2
	db := *rpcServer.config.Database

	amountDCBVote := uint32(arrayParams[1].(float64))
	amountGOVVote := uint32(arrayParams[2].(float64))

	err := db.SetVoteTokenAmount(metadata.DCBBoard.BoardTypeDB(), rpcServer.config.BlockChain.GetCurrentBoardIndex(blockchain.DCBConstitutionHelper{}), *paymentAddress, amountDCBVote)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	err = db.SetVoteTokenAmount(metadata.GOVBoard.BoardTypeDB(), rpcServer.config.BlockChain.GetCurrentBoardIndex(blockchain.DCBConstitutionHelper{}), *paymentAddress, amountGOVVote)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return nil, nil
}

// ============================== VOTE PROPOSAL

func (rpcServer RpcServer) handleGetEncryptionFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	db := *rpcServer.config.Database
	dcbEncryptionFlag, _ := db.GetEncryptFlag(metadata.DCBBoard.BoardTypeDB())
	govEncryptionFlag, _ := db.GetEncryptFlag(metadata.GOVBoard.BoardTypeDB())
	return jsonresult.GetEncryptionFlagResult{
		DCBFlag: dcbEncryptionFlag,
		GOVFlag: govEncryptionFlag,
	}, nil
}
func (rpcServer RpcServer) handleSetEncryptionFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	fmt.Print("delete me, use only for test purpose!!!")
	db := *rpcServer.config.Database
	dcbEncryptionFlag, _ := db.GetEncryptFlag(metadata.DCBBoard.BoardTypeDB())
	govEncryptionFlag, _ := db.GetEncryptFlag(metadata.GOVBoard.BoardTypeDB())
	db.SetEncryptFlag(metadata.DCBBoard.BoardTypeDB(), (dcbEncryptionFlag+1)%4)
	db.SetEncryptFlag(metadata.GOVBoard.BoardTypeDB(), (govEncryptionFlag+1)%4)
	return dcbEncryptionFlag, nil
}

func (rpcServer RpcServer) handleGetEncryptionLastBlockHeightFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	boardType := database.BoardTypeDB(arrayParams[0].([]byte)[0])
	db := *rpcServer.config.Database
	blockHeight, _ := db.GetEncryptionLastBlockHeight(boardType)
	return jsonresult.GetEncryptionLastBlockHeightResult{blockHeight}, nil
}

func CreateSealLv3Data(data *metadata.VoteProposalData, pubKeys [][]byte) []byte {
	SealLv3 := common.Encrypt(common.Encrypt(common.Encrypt(data.ToBytes(), pubKeys[0]), pubKeys[1]), pubKeys[2])
	return SealLv3
}

func (rpcServer RpcServer) buildRawSealLv3VoteProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	index := len(arrayParams) - 3

	boardType := database.BoardTypeDB(arrayParams[iPlusPlus(&index)].([]byte)[0])
	voteProposalData := metadata.NewVoteProposalDataFromJson(arrayParams[iPlusPlus(&index)])

	threeSenderKey := common.SliceInterfaceToSliceString(arrayParams[iPlusPlus(&index)].([]interface{}))
	pubKeys, err := rpcServer.ListPubKeyFromListSenderKey(threeSenderKey)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	threePaymentAddress := rpcServer.ListPaymentAddressFromListSenderKey(threeSenderKey)

	Seal3Data := CreateSealLv3Data(voteProposalData, pubKeys)
	meta := NewSealedLv3VoteProposalMetadata(boardType, Seal3Data, threePaymentAddress)

	params = setBuildRawBurnTransactionParams(params, FeeVoteProposal)
	tx, err1 := rpcServer.buildRawTransaction(params, meta)
	if err1 != nil {
		return tx, err1
	}
	return tx, nil
}

func (rpcServer RpcServer) ListPaymentAddressFromListSenderKey(listSenderKey []string) []privacy.PaymentAddress {
	paymentAddresses := make([]privacy.PaymentAddress, 0)
	for i := 0; i < 3; i++ {
		new, _ := rpcServer.GetPaymentAddressFromSenderKeyParams(listSenderKey[i])
		paymentAddresses = append(paymentAddresses, *new)
	}
	return paymentAddresses
}

func (rpcServer RpcServer) ListPubKeyFromListSenderKey(threePaymentAddress []string) ([][]byte, error) {
	pubKeys := make([][]byte, len(threePaymentAddress))
	for i := 0; i < len(threePaymentAddress); i++ {
		paymentAddress, err := rpcServer.GetPaymentAddressFromSenderKeyParams(threePaymentAddress[i])
		if err != nil {
			return nil, err
		}
		pubKeys[i] = paymentAddress.Pk
	}
	return pubKeys, nil
}

func NewSealedLv3VoteProposalMetadata(boardType database.BoardTypeDB, Seal3Data []byte, paymentAddresses []privacy.PaymentAddress) metadata.Metadata {
	var meta metadata.Metadata
	if boardType == metadata.DCBBoard.BoardTypeDB() {
		meta = metadata.NewSealedLv3DCBVoteProposalMetadata(Seal3Data, paymentAddresses)
	} else {
		meta = metadata.NewSealedLv3GOVVoteProposalMetadata(Seal3Data, paymentAddresses)
	}
	return meta
}

func (rpcServer RpcServer) handleCreateRawSealLv3VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := rpcServer.buildRawSealLv3VoteProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}

	byteArrays, err1 := json.Marshal(tx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

//create lv3 vote by 3 layer encrypt
func (rpcServer RpcServer) handleCreateAndSendSealLv3VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawSealLv3VoteProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (rpcServer RpcServer) buildRawSealLv2VoteProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	index := len(arrayParams) - 3

	boardType := database.BoardTypeDB(arrayParams[iPlusPlus(&index)].([]byte)[0])

	firstPrivateKey := []byte(arrayParams[iPlusPlus(&index)].(string))

	lv3txID, err := common.NewHashFromStr(arrayParams[iPlusPlus(&index)].(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	_, _, _, lv3Tx, err := rpcServer.config.BlockChain.GetTransactionByHash(lv3txID)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	SealLv3Data, err := GetSealLv3Data(lv3Tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	pubKeys := GetLockerPaymentAddress(lv3Tx)
	Seal2Data := common.Decrypt(SealLv3Data, firstPrivateKey)

	meta := NewSealedLv2VoteProposalMetadata(
		database.BoardTypeDB(boardType),
		Seal2Data,
		pubKeys,
		*lv3txID,
	)
	params = setBuildRawBurnTransactionParams(params, FeeVoteProposal)
	tx, err1 := rpcServer.buildRawTransaction(params, meta)
	return tx, err1
}

func NewSealedLv2VoteProposalMetadata(boardType database.BoardTypeDB, Seal2Data []byte, paymentAddresses []privacy.PaymentAddress, pointer common.Hash) metadata.Metadata {
	var meta metadata.Metadata
	if boardType == metadata.DCBBoard.BoardTypeDB() {
		meta = metadata.NewSealedLv2DCBVoteProposalMetadata(
			Seal2Data,
			paymentAddresses,
			pointer,
		)
	} else {
		meta = metadata.NewSealedLv2GOVVoteProposalMetadata(
			Seal2Data,
			paymentAddresses,
			pointer,
		)
	}
	return meta
}

func GetLockerPaymentAddress(tx metadata.Transaction) []privacy.PaymentAddress {
	meta := tx.GetMetadata()
	if meta.GetType() == metadata.SealedLv3DCBVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv3DCBVoteProposalMetadata)
		return newMeta.SealedLv3VoteProposalMetadata.SealedVoteProposal.LockerPaymentAddress
	} else {
		newMeta := meta.(*metadata.SealedLv3GOVVoteProposalMetadata)
		return newMeta.SealedLv3VoteProposalMetadata.SealedVoteProposal.LockerPaymentAddress
	}
}

func GetSealLv3Data(tx metadata.Transaction) ([]byte, error) {
	meta := tx.GetMetadata()
	if meta.GetType() == metadata.SealedLv3DCBVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv3DCBVoteProposalMetadata)
		return newMeta.SealedLv3VoteProposalMetadata.SealedVoteProposal.SealVoteProposalData, nil
	} else if meta.GetType() == metadata.SealedLv3GOVVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv3GOVVoteProposalMetadata)
		return newMeta.SealedLv3VoteProposalMetadata.SealedVoteProposal.SealVoteProposalData, nil
	}
	return nil, errors.New("wrong type")
}

func (rpcServer RpcServer) handleCreateRawSealLv2VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := rpcServer.buildRawSealLv2VoteProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}

	byteArrays, err1 := json.Marshal(tx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

//create lv2 vote by decrypt A layer
func (rpcServer RpcServer) handleCreateAndSendSealLv2VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawSealLv2VoteProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (rpcServer RpcServer) buildRawSealLv1VoteProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	index := len(arrayParams) - 4

	boardType := database.BoardTypeDB(arrayParams[iPlusPlus(&index)].([]byte)[0])

	secondPrivateKey := []byte(arrayParams[iPlusPlus(&index)].(string))

	lv3TxID, err1 := common.NewHashFromStr(arrayParams[iPlusPlus(&index)].(string))
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	lv2TxID, err1 := common.NewHashFromStr(arrayParams[iPlusPlus(&index)].(string))
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	_, _, _, lv2tx, _ := rpcServer.config.BlockChain.GetTransactionByHash(lv2TxID)
	SealLv2Data, err1 := GetSealLv2Data(lv2tx)
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	_, _, _, lv3tx, _ := rpcServer.config.BlockChain.GetTransactionByHash(lv3TxID)
	pubKeys := GetLockerPaymentAddress(lv3tx)

	Seal1Data := common.Decrypt(SealLv2Data, secondPrivateKey)

	meta := NewSealedLv1VoteProposalMetadata(
		database.BoardTypeDB(boardType),
		Seal1Data,
		pubKeys,
		*lv2TxID,
		*lv3TxID,
	)
	params = setBuildRawBurnTransactionParams(params, FeeVoteProposal)
	tx, err := rpcServer.buildRawTransaction(params, meta)
	return tx, err
}

func NewSealedLv1VoteProposalMetadata(boardType database.BoardTypeDB, sealLv1Data []byte, listPaymentAddress []privacy.PaymentAddress, lv2TxID common.Hash, lv3TxID common.Hash) metadata.Metadata {
	var meta metadata.Metadata
	if boardType == metadata.DCBBoard.BoardTypeDB() {
		meta = metadata.NewSealedLv1DCBVoteProposalMetadata(
			sealLv1Data,
			listPaymentAddress,
			lv2TxID,
			lv3TxID,
		)
	} else {
		meta = metadata.NewSealedLv1GOVVoteProposalMetadata(
			sealLv1Data,
			listPaymentAddress,
			lv2TxID,
			lv3TxID,
		)
	}
	return meta
}

func GetSealLv2Data(lv2tx metadata.Transaction) ([]byte, error) {
	meta := lv2tx.GetMetadata()
	if meta.GetType() == metadata.SealedLv2DCBVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv2DCBVoteProposalMetadata)
		return newMeta.SealedLv2VoteProposalMetadata.SealedVoteProposal.SealVoteProposalData, nil
	} else if meta.GetType() == metadata.SealedLv2GOVVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv2GOVVoteProposalMetadata)
		return newMeta.SealedLv2VoteProposalMetadata.SealedVoteProposal.SealVoteProposalData, nil
	}
	return nil, errors.New("wrong type")
}

func GetSealLv1Data(lv1tx metadata.Transaction) ([]byte, error) {
	meta := lv1tx.GetMetadata()
	if meta.GetType() == metadata.SealedLv1DCBVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv1DCBVoteProposalMetadata)
		return newMeta.SealedLv1VoteProposalMetadata.SealedVoteProposal.SealVoteProposalData, nil
	} else if meta.GetType() == metadata.SealedLv1GOVVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv1GOVVoteProposalMetadata)
		return newMeta.SealedLv1VoteProposalMetadata.SealedVoteProposal.SealVoteProposalData, nil
	}
	return nil, errors.New("wrong type")
}

func (rpcServer RpcServer) handleCreateRawSealLv1VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := rpcServer.buildRawSealLv1VoteProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(tx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateAndSendSealLv1VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawSealLv1VoteProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (rpcServer RpcServer) buildRawNormalVoteProposalTransactionFromOwner(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	index := len(arrayParams) - 3

	boardType := database.BoardTypeDB(arrayParams[iPlusPlus(&index)].([]byte)[0])

	lv3TxID, err1 := common.NewHashFromStr(arrayParams[iPlusPlus(&index)].(string))
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	_, _, _, lv3tx, _ := rpcServer.config.BlockChain.GetTransactionByHash(lv3TxID)
	paymentAddresses := GetLockerPaymentAddress(lv3tx)

	voteProposalData := metadata.NewVoteProposalDataFromJson(arrayParams[iPlusPlus(&index)])

	meta := NewNormalVoteProposalFromOwnerMetadata(
		database.BoardTypeDB(boardType),
		voteProposalData,
		paymentAddresses,
		*lv3TxID,
	)
	params = setBuildRawBurnTransactionParams(params, FeeVoteProposal)
	tx, err := rpcServer.buildRawTransaction(params, meta)
	return tx, err
}

func NewNormalVoteProposalFromOwnerMetadata(boardType database.BoardTypeDB, voteProposalData *metadata.VoteProposalData, listPaymentAddress []privacy.PaymentAddress, lv3TxID common.Hash) metadata.Metadata {
	var meta metadata.Metadata
	if boardType == metadata.DCBBoard.BoardTypeDB() {
		meta = metadata.NewNormalDCBVoteProposalFromOwnerMetadata(
			*voteProposalData,
			listPaymentAddress,
			lv3TxID,
		)
	} else {
		meta = metadata.NewNormalGOVVoteProposalFromOwnerMetadata(
			*voteProposalData,
			listPaymentAddress,
			lv3TxID,
		)
	}
	return meta
}

func (rpcServer RpcServer) handleCreateRawNormalVoteProposalTransactionFromOwner(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err1 := rpcServer.buildRawNormalVoteProposalTransactionFromOwner(params)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateAndSendNormalVoteProposalFromOwnerTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawNormalVoteProposalTransactionFromOwner(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (rpcServer RpcServer) buildRawNormalVoteProposalTransactionFromSealer(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	index := len(arrayParams) - 4

	boardType := database.BoardTypeDB(arrayParams[iPlusPlus(&index)].([]byte)[0])

	lv3TxID, err1 := common.NewHashFromStr(arrayParams[iPlusPlus(&index)].(string))
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	lv1TxID, err1 := common.NewHashFromStr(arrayParams[iPlusPlus(&index)].(string))
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	_, _, _, lv1tx, _ := rpcServer.config.BlockChain.GetTransactionByHash(lv1TxID)
	SealLv1Data, err1 := GetSealLv1Data(lv1tx)
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	_, _, _, lv3tx, _ := rpcServer.config.BlockChain.GetTransactionByHash(lv3TxID)
	paymentAddresses := GetLockerPaymentAddress(lv3tx)

	thirdPrivateKey := []byte(arrayParams[iPlusPlus(&index)].(string))

	normalVoteProposalData := common.Decrypt(SealLv1Data, thirdPrivateKey)
	voteProposalData := metadata.NewVoteProposalDataFromBytes(normalVoteProposalData)

	meta := NewNormalVoteProposalFromSealerMetadata(
		database.BoardTypeDB(boardType),
		*voteProposalData,
		paymentAddresses,
		*lv1TxID,
		*lv3TxID,
	)
	params = setBuildRawBurnTransactionParams(params, FeeVoteProposal)
	tx, err := rpcServer.buildRawTransaction(params, meta)
	return tx, err
}

func NewNormalVoteProposalFromSealerMetadata(boardType database.BoardTypeDB, voteProposalData metadata.VoteProposalData, paymentAddresses []privacy.PaymentAddress, lv1TxID common.Hash, lv3TxID common.Hash) metadata.Metadata {
	var meta metadata.Metadata
	if boardType == metadata.DCBBoard.BoardTypeDB() {
		meta = metadata.NewNormalDCBVoteProposalFromSealerMetadata(
			voteProposalData,
			paymentAddresses,
			lv1TxID,
			lv3TxID,
		)
	} else {
		meta = metadata.NewNormalGOVVoteProposalFromSealerMetadata(
			voteProposalData,
			paymentAddresses,
			lv1TxID,
			lv3TxID,
		)
	}
	return meta
}

func (rpcServer RpcServer) handleCreateRawNormalVoteProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := rpcServer.buildRawNormalVoteProposalTransactionFromSealer(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err1 := json.Marshal(tx)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateAndSendNormalVoteProposalFromSealerTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawNormalVoteProposalTransactionFromSealer(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

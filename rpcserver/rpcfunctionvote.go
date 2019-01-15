package rpcserver

import (
	"encoding/json"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
)

func iPlusPlus(x *int) int {
	*x += 1
	return *x - 1
}

func (self RpcServer) handleGetAmountVoteToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	paymentAddress := arrayParams[0].(string)
	pubKey := wallet.GetPubKeyFromPaymentAddress(paymentAddress)
	db := *self.config.Database
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}

	// For DCB voting token
	result.PaymentAddress = paymentAddress
	item := jsonresult.CustomTokenBalance{}
	item.Name = "DCB voting token"
	item.Symbol = "DCB Voting Token"
	TokenID := &common.Hash{}
	TokenID.SetBytes(common.DCBVotingTokenID[:])
	item.TokenID = TokenID.String()
	item.TokenImage = common.Render([]byte(item.TokenID))
	amount, err := db.GetVoteTokenAmount("dcb", self.config.BlockChain.GetCurrentBoardIndex(blockchain.DCBConstitutionHelper{}), pubKey)
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
	amount, err = db.GetVoteTokenAmount("gov", self.config.BlockChain.GetCurrentBoardIndex(blockchain.GOVConstitutionHelper{}), pubKey)
	if err != nil {
		Logger.log.Error(err)
	}
	item.Amount = uint64(amount)
	result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)

	return result, nil
}

// ============================== VOTE PROPOSAL

func (self RpcServer) handleGetEncryptionFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	db := *self.config.Database
	dcbEncryptionFlag, _ := db.GetEncryptFlag("dcb")
	govEncryptionFlag, _ := db.GetEncryptFlag("gov")
	return jsonresult.GetEncryptionFlagResult{
		DCBFlag: dcbEncryptionFlag,
		GOVFlag: govEncryptionFlag,
	}, nil
}

func (self RpcServer) handleGetEncryptionLastBlockHeightFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	boardType := arrayParams[0].(string)
	db := *self.config.Database
	blockHeight, _ := db.GetEncryptionLastBlockHeight(boardType)
	return jsonresult.GetEncryptionLastBlockHeightResult{blockHeight}, nil
}

func CreateSealLv3Data(data *metadata.VoteProposalData, pubKeys [][]byte) []byte {
	SealLv3 := common.Encrypt(common.Encrypt(common.Encrypt(data.ToBytes(), pubKeys[0]), pubKeys[1]), pubKeys[2])
	return SealLv3
}

func (self RpcServer) buildRawSealLv3VoteProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	index := len(arrayParams) - 3

	boardType := arrayParams[iPlusPlus(&index)].(string)
	voteProposalData := metadata.NewVoteProposalDataFromJson(arrayParams[iPlusPlus(&index)])

	threePaymentAddress := common.SliceInterfaceToSliceSliceByte(arrayParams[iPlusPlus(&index)].([]interface{}))
	pubKeys := ListPubKeyFromListSenderKey(threePaymentAddress)

	Seal3Data := CreateSealLv3Data(voteProposalData, pubKeys)
	meta := NewSealedLv3VoteProposalMetadata(boardType, Seal3Data, pubKeys)

	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func ListPubKeyFromListSenderKey(threePaymentAddress [][]byte) [][]byte {
	pubKeys := make([][]byte, 3)
	for i := 0; i < 3; i++ {
		pubKeys[i], _ = GetPubKeyFromSenderKeyParams(string(threePaymentAddress[i]))
	}
	return pubKeys
}

func NewSealedLv3VoteProposalMetadata(boardType string, Seal3Data []byte, pubKeys [][]byte) metadata.Metadata {
	var meta metadata.Metadata
	if boardType == "dcb" {
		meta = metadata.NewSealedLv3DCBVoteProposalMetadata(Seal3Data, pubKeys)
	} else {
		meta = metadata.NewSealedLv3GOVVoteProposalMetadata(Seal3Data, pubKeys)
	}
	return meta
}

func (self RpcServer) handleCreateRawSealLv3VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := self.buildRawSealLv3VoteProposalTransaction(params)
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

//create lv3 vote by 3 layer encrypt
func (self RpcServer) handleCreateAndSendSealLv3VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSealLv3VoteProposalTransaction(params, closeChan)
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
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawSealLv2VoteProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	index := len(arrayParams) - 3

	boardType := arrayParams[iPlusPlus(&index)].(string)

	firstPrivateKey := []byte(arrayParams[iPlusPlus(&index)].(string))

	lv3txID := common.NewHash([]byte(arrayParams[iPlusPlus(&index)].(string)))
	_, _, _, lv3Tx, _ := self.config.BlockChain.GetTransactionByHash(&lv3txID)
	SealLv3Data := GetSealLv3Data(lv3Tx)
	pubKeys := GetLockerPubKeys(lv3Tx)
	Seal2Data := common.Decrypt(SealLv3Data, firstPrivateKey)

	meta := NewSealedLv2VoteProposalMetadata(
		boardType,
		Seal2Data,
		pubKeys,
		lv3txID,
	)
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func NewSealedLv2VoteProposalMetadata(boardType string, Seal2Data []byte, pubKeys [][]byte, pointer common.Hash) metadata.Metadata {
	var meta metadata.Metadata
	if boardType == "dcb" {
		meta = metadata.NewSealedLv2DCBVoteProposalMetadata(
			Seal2Data,
			pubKeys,
			pointer,
		)
	} else {
		meta = metadata.NewSealedLv2GOVVoteProposalMetadata(
			Seal2Data,
			pubKeys,
			pointer,
		)
	}
	return meta
}

func GetLockerPubKeys(tx metadata.Transaction) [][]byte {
	meta := tx.GetMetadata()
	if meta.GetType() == metadata.SealedLv3DCBVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv3DCBVoteProposalMetadata)
		return newMeta.LockerPubKeys
	} else {
		newMeta := meta.(*metadata.SealedLv3DCBVoteProposalMetadata)
		return newMeta.LockerPubKeys
	}
}

func GetSealLv3Data(tx metadata.Transaction) []byte {
	meta := tx.GetMetadata()
	if meta.GetType() == metadata.SealedLv3DCBVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv3DCBVoteProposalMetadata)
		return newMeta.SealedVoteProposal.SealVoteProposalData
	} else {
		newMeta := meta.(*metadata.SealedLv3DCBVoteProposalMetadata)
		return newMeta.SealedVoteProposal.SealVoteProposalData
	}
}

func (self RpcServer) handleCreateRawSealLv2VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := self.buildRawSealLv2VoteProposalTransaction(params)
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

//create lv2 vote by decrypt A layer
func (self RpcServer) handleCreateAndSendSealLv2VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSealLv2VoteProposalTransaction(params, closeChan)
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
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawSealLv1VoteProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	index := len(arrayParams) - 4

	boardType := arrayParams[iPlusPlus(&index)].(string)

	secondPrivateKey := []byte(arrayParams[iPlusPlus(&index)].(string))

	lv3TxID := common.NewHash([]byte(arrayParams[iPlusPlus(&index)].(string)))

	lv2TxID := common.NewHash([]byte(arrayParams[iPlusPlus(&index)].(string)))
	_, _, _, lv2tx, _ := self.config.BlockChain.GetTransactionByHash(&lv2TxID)
	SealLv2Data := GetSealLv2Data(lv2tx)

	_, _, _, lv3tx, _ := self.config.BlockChain.GetTransactionByHash(&lv3TxID)
	pubKeys := GetLockerPubKeys(lv3tx)

	Seal1Data := common.Decrypt(SealLv2Data, secondPrivateKey)

	meta := NewSealedLv1VoteProposalMetadata(
		boardType,
		Seal1Data,
		pubKeys,
		lv2TxID,
		lv3TxID,
	)
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func NewSealedLv1VoteProposalMetadata(boardType string, sealLv1Data []byte, pubKeys [][]byte, lv2TxID common.Hash, lv3TxID common.Hash) metadata.Metadata {
	var meta metadata.Metadata
	if boardType == "dcb" {
		meta = metadata.NewSealedLv1DCBVoteProposalMetadata(
			sealLv1Data,
			pubKeys,
			lv2TxID,
			lv3TxID,
		)
	} else {
		meta = metadata.NewSealedLv1GOVVoteProposalMetadata(
			sealLv1Data,
			pubKeys,
			lv2TxID,
			lv3TxID,
		)
	}
	return meta
}

func GetSealLv2Data(lv2tx metadata.Transaction) []byte {
	meta := lv2tx.GetMetadata()
	if meta.GetType() == metadata.SealedLv2DCBVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv3DCBVoteProposalMetadata)
		return newMeta.SealVoteProposalData
	} else if meta.GetType() == metadata.SealedLv2GOVVoteProposalMeta {
		newMeta := meta.(*metadata.SealedLv3GOVVoteProposalMetadata)
		return newMeta.SealVoteProposalData
	}
	return nil
}

func (self RpcServer) handleCreateRawSealLv1VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := self.buildRawSealLv1VoteProposalTransaction(params)
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

func (self RpcServer) handleCreateAndSendSealLv1VoteProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSealLv1VoteProposalTransaction(params, closeChan)
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
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawNormalVoteProposalTransactionFromOwner(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	index := len(arrayParams) - 3

	boardType := arrayParams[iPlusPlus(&index)].(string)

	lv3TxID := common.NewHash([]byte(arrayParams[iPlusPlus(&index)].(string)))

	_, _, _, lv3tx, _ := self.config.BlockChain.GetTransactionByHash(&lv3TxID)
	pubKeys := GetLockerPubKeys(lv3tx)

	voteProposalData := metadata.NewVoteProposalDataFromJson(arrayParams[iPlusPlus(&index)])

	meta := NewNormalVoteProposalFromOwnerMetadata(
		boardType,
		voteProposalData,
		pubKeys,
		lv3TxID,
	)
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func NewNormalVoteProposalFromOwnerMetadata(boardType string, voteProposalData *metadata.VoteProposalData, pubKeys [][]byte, lv3TxID common.Hash) metadata.Metadata {
	var meta metadata.Metadata
	if boardType == "dcb" {
		meta = metadata.NewNormalDCBVoteProposalFromOwnerMetadata(
			*voteProposalData,
			pubKeys,
			lv3TxID,
		)
	} else {
		meta = metadata.NewNormalGOVVoteProposalFromOwnerMetadata(
			*voteProposalData,
			pubKeys,
			lv3TxID,
		)
	}
	return meta
}

func (self RpcServer) handleCreateRawNormalVoteProposalTransactionFromOwner(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err1 := self.buildRawNormalVoteProposalTransactionFromOwner(params)
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

func (self RpcServer) handleCreateAndSendNormalVoteProposalFromOwnerTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawNormalVoteProposalTransactionFromOwner(params, closeChan)
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
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawNormalVoteProposalTransactionFromSealer(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	index := len(arrayParams) - 4

	boardType := arrayParams[iPlusPlus(&index)].(string)

	lv3TxID := common.NewHash([]byte(arrayParams[iPlusPlus(&index)].(string)))

	lv1TxID := common.NewHash([]byte(arrayParams[iPlusPlus(&index)].(string)))
	_, _, _, lv1tx, _ := self.config.BlockChain.GetTransactionByHash(&lv1TxID)
	SealLv1Data := GetSealLv2Data(lv1tx)

	_, _, _, lv3tx, _ := self.config.BlockChain.GetTransactionByHash(&lv3TxID)
	pubKeys := GetLockerPubKeys(lv3tx)

	thirdPrivateKey := []byte(arrayParams[iPlusPlus(&index)].(string))

	normalVoteProposalData := common.Decrypt(SealLv1Data, thirdPrivateKey)
	voteProposalData := metadata.NewVoteProposalDataFromBytes(normalVoteProposalData)

	meta := NewNormalVoteProposalFromSealerMetadata(
		boardType,
		*voteProposalData,
		pubKeys,
		lv1TxID,
		lv3TxID,
	)
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func NewNormalVoteProposalFromSealerMetadata(boardType string, voteProposalData metadata.VoteProposalData, pubKeys [][]byte, lv1TxID common.Hash, lv3TxID common.Hash) metadata.Metadata {
	var meta metadata.Metadata
	if boardType == "dcb" {
		meta = metadata.NewNormalDCBVoteProposalFromSealerMetadata(
			voteProposalData,
			pubKeys,
			lv1TxID,
			lv3TxID,
		)
	} else {
		meta = metadata.NewNormalGOVVoteProposalFromSealerMetadata(
			voteProposalData,
			pubKeys,
			lv1TxID,
			lv3TxID,
		)
	}
	return meta
}

func (self RpcServer) handleCreateRawNormalVoteProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tx, err := self.buildRawNormalVoteProposalTransactionFromSealer(params)
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

func (self RpcServer) handleCreateAndSendNormalVoteProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawNormalVoteProposalTransactionFromSealer(params, closeChan)
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
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

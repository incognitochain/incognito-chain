package rpcserver

import (
	"encoding/hex"
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
)

func (self RpcServer) buildRawSealLv3VoteGOVProposalTransaction(
	params interface{},
) (*transaction.Tx, error) {
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	voteInfo := arrayParams[len(arrayParams)-4]
	firstPubKey := arrayParams[len(arrayParams)-3] // firstPubKey is pubkey of itself
	secondPubKey := arrayParams[len(arrayParams)-2]
	thirdPubKey := arrayParams[len(arrayParams)-1]
	Seal3Data := common.Encrypt(common.Encrypt(common.Encrypt(voteInfo, thirdPubKey), secondPubKey), firstPubKey)
	tx.Metadata = metadata.NewSealedLv3GOVBallotMetadata(
		map[string]interface{}{
			"SealedBallot": []byte(Seal3Data.(string)),
			"LockerPubKey": [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))},
		})
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv3VoteGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	tx, err := self.buildRawSealLv3VoteGOVProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	hexData := hex.EncodeToString(byteArrays)
	result := jsonresult.CreateTransactionResult{
		HexData: hexData,
	}
	return result, nil
}

//create lv3 vote by 3 layer encrypt
func (self RpcServer) handleCreateAndSendSealLv3VoteGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawSealLv3VoteGOVProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	hexStrOfTx := tx.HexData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawSealLv2VoteGOVProposalTransaction(
	params interface{},
) (*transaction.Tx, error) {
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	PointerToLv3Ballot := arrayParams[len(arrayParams)-6]
	Seal3Data := arrayParams[len(arrayParams)-5]
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	firstPrivateKey := arrayParams[len(arrayParams)-1]
	Seal2Data := common.Decrypt(Seal3Data, firstPrivateKey)
	Pointer := common.Hash{}
	copy(Pointer[:], []byte(PointerToLv3Ballot.(string)))
	tx.Metadata = metadata.NewSealedLv2GOVBallotMetadata(
		map[string]interface{}{
			"SealedBallot":       []byte(Seal2Data.(string)),
			"LockerPubKey":       [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))},
			"PointerToLv3Ballot": &Pointer,
		},
	)
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv2VoteGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	tx, err := self.buildRawSealLv2VoteGOVProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	hexData := hex.EncodeToString(byteArrays)
	result := jsonresult.CreateTransactionResult{
		HexData: hexData,
	}
	return result, nil
}

//create lv2 vote by decrypt A layer
func (self RpcServer) handleCreateAndSendSealLv2VoteGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawSealLv2VoteGOVProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	hexStrOfTx := tx.HexData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawSealLv1VoteGOVProposalTransaction(
	params interface{},
) (*transaction.Tx, error) {
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)

	PointerToLv3Ballot := arrayParams[len(arrayParams)-7]
	Pointer3 := common.Hash{}
	copy(Pointer3[:], []byte(PointerToLv3Ballot.(string)))

	PointerToLv2Ballot := arrayParams[len(arrayParams)-6]
	Pointer2 := common.Hash{}
	copy(Pointer2[:], []byte(PointerToLv2Ballot.(string)))

	Seal2Data := arrayParams[len(arrayParams)-5]
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	secondPrivateKey := arrayParams[len(arrayParams)-1]
	Seal1Data := common.Decrypt(Seal2Data, secondPrivateKey)
	tx.Metadata = metadata.NewSealedLv1GOVBallotMetadata(
		map[string]interface{}{
			"SealedBallot":       []byte(Seal1Data.(string)),
			"LockerPubKey":       [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))},
			"PointerToLv2Ballot": &Pointer2,
			"PointerToLv3Ballot": &Pointer3,
		})
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv1VoteGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	tx, err := self.buildRawSealLv1VoteGOVProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	hexData := hex.EncodeToString(byteArrays)
	result := jsonresult.CreateTransactionResult{
		HexData: hexData,
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendSealLv1VoteGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawSealLv1VoteGOVProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	hexStrOfTx := tx.HexData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawNormalVoteGOVProposalTransactionFromSealer(
	params interface{},
) (*transaction.Tx, error) {
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)

	PointerToLv3Ballot := arrayParams[len(arrayParams)-7]
	Pointer3 := common.Hash{}
	copy(Pointer3[:], []byte(PointerToLv3Ballot.(string)))

	PointerToLv1Ballot := arrayParams[len(arrayParams)-6]
	Pointer1 := common.Hash{}
	copy(Pointer1[:], []byte(PointerToLv1Ballot.(string)))

	Seal1Data := arrayParams[len(arrayParams)-5]
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	thirdPrivateKey := arrayParams[len(arrayParams)-1]
	normalBallot := common.Decrypt(Seal1Data, thirdPrivateKey)
	tx.Metadata = metadata.NewNormalGOVBallotFromSealerMetadata(
		map[string]interface{}{
			"Ballot":             normalBallot.([]byte),
			"LockerPubKey":       [][]byte{firstPubKey.([]byte), secondPubKey.([]byte), thirdPubKey.([]byte)},
			"PointerToLv1Ballot": &Pointer1,
			"PointerToLv3Ballot": &Pointer3,
		})
	return tx, err
}

func (self RpcServer) handleCreateRawNormalVoteGOVProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	tx, err := self.buildRawNormalVoteGOVProposalTransactionFromSealer(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	hexData := hex.EncodeToString(byteArrays)
	result := jsonresult.CreateTransactionResult{
		HexData: hexData,
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendNormalVoteGOVProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawNormalVoteGOVProposalTransactionFromSealer(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	hexStrOfTx := tx.HexData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawNormalVoteGOVProposalTransactionFromOwner(
	params interface{},
) (*transaction.Tx, error) {
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	PointerToLv3Ballot := arrayParams[len(arrayParams)-5]
	Pointer := common.Hash{}
	copy(Pointer[:], []byte(PointerToLv3Ballot.(string)))
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	normalBallot := arrayParams[len(arrayParams)-1]
	tx.Metadata = metadata.NewNormalGOVBallotFromOwnerMetadata(
		map[string]interface{}{
			"Ballot":             normalBallot.([]byte),
			"LockerPubKey":       [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))},
			"PointerToLv3Ballot": &Pointer,
		})
	return tx, err
}

func (self RpcServer) handleCreateRawNormalVoteGOVProposalTransactionFromOwner(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	tx, err := self.buildRawNormalVoteGOVProposalTransactionFromOwner(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	hexData := hex.EncodeToString(byteArrays)
	result := jsonresult.CreateTransactionResult{
		HexData: hexData,
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendNormalVoteGOVProposalTransactionFromOwner(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	data, err := self.handleCreateRawNormalVoteGOVProposalTransactionFromOwner(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	hexStrOfTx := tx.HexData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

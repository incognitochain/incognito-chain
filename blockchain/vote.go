package blockchain

import (
	"encoding/binary"
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/syndtr/goleveldb/leveldb/util"
)

//Todo: @0xjackalope count by database
func (blockgen *BlkTmplGenerator) createAcceptConstitutionTx(
	chainID byte,
	ConstitutionHelper ConstitutionHelper,
) (*metadata.Transaction, error) {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock

	// count vote from lastConstitution.StartedBlockHeight to Bestblock height
	CountVote := make(map[common.Hash]int64)
	Transaction := make(map[common.Hash]*metadata.Transaction)
	for blockHeight := ConstitutionHelper.GetStartedBlockHeight(blockgen, chainID); blockHeight < BestBlock.Header.Height; blockHeight += 1 {
		//retrieve block from block's height
		hashBlock, err := blockgen.chain.config.DataBase.GetBlockByIndex(blockHeight, chainID)
		if err != nil {
			return nil, err
		}
		blockBytes, err := blockgen.chain.config.DataBase.FetchBlock(hashBlock)
		if err != nil {
			return nil, err
		}
		block := Block{}
		err = json.Unmarshal(blockBytes, &block)
		if err != nil {
			return nil, err
		}
		//count vote of this block
		for _, tx := range block.Transactions {
			_, exist := CountVote[*tx.Hash()]
			if ConstitutionHelper.CheckSubmitProposalType(tx) {
				if exist {
					return nil, err
				}
				CountVote[*tx.Hash()] = 0
				Transaction[*tx.Hash()] = &tx
			} else {
				if ConstitutionHelper.CheckVotingProposalType(tx) {
					if !exist {
						return nil, err
					}
					CountVote[*tx.Hash()] += int64(ConstitutionHelper.GetAmountVoteToken(tx))
				}
			}
		}
	}

	// get transaction and create transaction desc
	var maxVote int64
	var res common.Hash
	for key, value := range CountVote {
		if value > maxVote {
			maxVote = value
			res = key
		}
	}

	acceptedSubmitProposalTransaction := ConstitutionHelper.TxAcceptProposal(*Transaction[res])

	return &acceptedSubmitProposalTransaction, nil
}

func (blockgen *BlkTmplGenerator) createSingleSendDCBVoteTokenTx(chainID byte, pubKey []byte, amount uint64) (metadata.Transaction, error) {

	paymentAddress := privacy.PaymentAddress{
		Pk: pubKey,
	}
	txTokenVout := transaction.TxTokenVout{
		Value:          amount,
		PaymentAddress: paymentAddress,
	}
	txTokenData := transaction.TxTokenData{
		Type:       transaction.InitVoteDCBToken,
		Amount:     amount,
		PropertyID: VoteDCBTokenID,
		Vins:       []transaction.TxTokenVin{},
		Vouts:      []transaction.TxTokenVout{txTokenVout},
	}
	sendDCBVoteTokenTransaction := transaction.TxSendInitDCBVoteToken{
		TxCustomToken: transaction.TxCustomToken{
			TxTokenData: txTokenData,
		},
	}
	return &sendDCBVoteTokenTransaction, nil
}

func (blockgen *BlkTmplGenerator) createSingleSendGOVVoteTokenTx(chainID byte, pubKey []byte, amount uint64) (metadata.Transaction, error) {

	paymentAddress := privacy.PaymentAddress{
		Pk: pubKey,
	}
	txTokenVout := transaction.TxTokenVout{
		Value:          amount,
		PaymentAddress: paymentAddress,
	}
	txTokenData := transaction.TxTokenData{
		Type:       transaction.InitVoteGOVToken,
		Amount:     amount,
		PropertyID: VoteGOVTokenID,
		Vins:       []transaction.TxTokenVin{},
		Vouts:      []transaction.TxTokenVout{txTokenVout},
	}
	sendGOVVoteTokenTransaction := transaction.TxSendInitGOVVoteToken{
		TxCustomToken: transaction.TxCustomToken{
			TxTokenData: txTokenData,
		},
	}
	return &sendGOVVoteTokenTransaction, nil
}

func getAmountOfVoteToken(sumAmount uint64, voteAmount uint64) uint64 {
	return voteAmount * common.SumOfVoteDCBToken / sumAmount
}

func (blockgen *BlkTmplGenerator) CreateSendDCBVoteTokenToGovernorTx(chainID byte, newDCBList database.CandidateList, sumAmountDCB uint64) []metadata.Transaction {
	var SendVoteTx []metadata.Transaction
	var newTx metadata.Transaction
	for i := 0; i <= NumberOfDCBGovernors; i++ {
		newTx, _ = blockgen.createSingleSendDCBVoteTokenTx(chainID, newDCBList[i].PubKey, getAmountOfVoteToken(sumAmountDCB, newDCBList[i].VoteAmount))
		SendVoteTx = append(SendVoteTx, newTx)
	}
	return SendVoteTx
}

func (blockgen *BlkTmplGenerator) CreateSendGOVVoteTokenToGovernorTx(chainID byte, newGOVList database.CandidateList, sumAmountGOV uint64) []metadata.Transaction {
	var SendVoteTx []metadata.Transaction
	var newTx metadata.Transaction
	for i := 0; i <= NumberOfGOVGovernors; i++ {
		newTx, _ = blockgen.createSingleSendGOVVoteTokenTx(chainID, newGOVList[i].PubKey, getAmountOfVoteToken(sumAmountGOV, newGOVList[i].VoteAmount))
		SendVoteTx = append(SendVoteTx, newTx)
	}
	return SendVoteTx
}

func (blockgen *BlkTmplGenerator) createAcceptDCBBoardTx(DCBBoardPubKeys [][]byte, sumOfVote uint64) metadata.Transaction {
	return &transaction.Tx{
		Metadata: &metadata.AcceptDCBBoardMetadata{
			DCBBoardPubKeys:     DCBBoardPubKeys,
			StartAmountDCBToken: sumOfVote,
		},
	}
}

func (blockgen *BlkTmplGenerator) createAcceptGOVBoardTx(DCBBoardPubKeys [][]byte, sumOfVote uint64) metadata.Transaction {
	return &transaction.Tx{
		Metadata: &metadata.AcceptGOVBoardMetadata{
			GOVBoardPubKeys:     DCBBoardPubKeys,
			StartAmountGOVToken: sumOfVote,
		},
	}
}

func (block *Block) UpdateDCBBoard(thisTx metadata.Transaction) error {
	tx := thisTx.(transaction.TxAcceptDCBBoard)
	block.Header.DCBGovernor.DCBBoardPubKeys = tx.DCBBoardPubKeys
	block.Header.DCBGovernor.StartedBlock = uint32(block.Header.Height)
	block.Header.DCBGovernor.EndBlock = block.Header.DCBGovernor.StartedBlock + common.DurationOfTermDCB
	block.Header.DCBGovernor.StartAmountDCBToken = tx.StartAmountDCBToken
	return nil
}

func (block *Block) UpdateGOVBoard(thisTx metadata.Transaction) error {
	tx := thisTx.(transaction.TxAcceptGOVBoard)
	block.Header.GOVGovernor.GOVBoardPubKeys = tx.GOVBoardPubKeys
	block.Header.GOVGovernor.StartedBlock = uint32(block.Header.Height)
	block.Header.GOVGovernor.EndBlock = block.Header.GOVGovernor.StartedBlock + common.DurationOfTermGOV
	block.Header.GOVGovernor.StartAmountGOVToken = tx.StartAmountGOVToken
	return nil
}

// startblock, pubkey, index
func (blockgen *BlkTmplGenerator) parseVoteDCBBoardListKey(key []byte) (int32, []byte, uint32) {
	keyWithoutPrefixI, _ := blockgen.chain.config.DataBase.ReverseGetKey(string(blockgen.chain.config.DataBase.GetVoteDCBBoardListPrefix()), key)
	keyWithoutPrefix := keyWithoutPrefixI.([]byte)
	startedBlock := int32(binary.LittleEndian.Uint32(keyWithoutPrefix[:4]))
	pubKey := keyWithoutPrefix[4 : 4+common.HashSize]
	currentIndex := binary.LittleEndian.Uint32(keyWithoutPrefix[4+common.HashSize:])
	return startedBlock, pubKey, currentIndex
}

// startblock, pubkey, index
func (blockgen *BlkTmplGenerator) parseVoteGOVBoardListKey(key []byte) (int32, []byte, uint32) {
	keyWithoutPrefixI, _ := blockgen.chain.config.DataBase.ReverseGetKey(string(blockgen.chain.config.DataBase.GetVoteGOVBoardListPrefix()), key)
	keyWithoutPrefix := keyWithoutPrefixI.([]byte)
	startedBlock := int32(binary.LittleEndian.Uint32(keyWithoutPrefix[:4]))
	pubKey := keyWithoutPrefix[4 : 4+common.HashSize]
	currentIndex := binary.LittleEndian.Uint32(keyWithoutPrefix[4+common.HashSize:])
	return startedBlock, pubKey, currentIndex
}

func parseVoteDCBBoardListValue(value []byte) ([]byte, uint64) {
	voterPubKey := value[:common.HashSize]
	amount := binary.LittleEndian.Uint64(value[common.HashSize:])
	return voterPubKey, amount
}

func parseVoteGOVBoardListValue(value []byte) ([]byte, uint64) {
	voterPubKey := value[:common.HashSize]
	amount := binary.LittleEndian.Uint64(value[common.HashSize:])
	return voterPubKey, amount
}

func createSingleSendDCBVoteTokenFail(pubKey []byte, amount uint64) metadata.Transaction {
	paymentAddress := privacy.PaymentAddress{
		Pk: pubKey,
	}
	txTokenVout := transaction.TxTokenVout{
		Value:          amount,
		PaymentAddress: paymentAddress,
	}
	newTx := transaction.TxCustomToken{
		TxTokenData: transaction.TxTokenData{
			Type:       transaction.SendBackDCBTokenVoteFail,
			Amount:     amount,
			PropertyID: DCBTokenID,
			Vins:       []transaction.TxTokenVin{},
			Vouts:      []transaction.TxTokenVout{txTokenVout},
		},
	}
	return &newTx
}

func createSingleSendGOVVoteTokenFail(pubKey []byte, amount uint64) metadata.Transaction {
	paymentAddress := privacy.PaymentAddress{
		Pk: pubKey,
	}
	txTokenVout := transaction.TxTokenVout{
		Value:          amount,
		PaymentAddress: paymentAddress,
	}
	newTx := transaction.TxCustomToken{
		TxTokenData: transaction.TxTokenData{
			Type:       transaction.SendBackGOVTokenVoteFail,
			Amount:     amount,
			PropertyID: GOVTokenID,
			Vins:       []transaction.TxTokenVin{},
			Vouts:      []transaction.TxTokenVout{txTokenVout},
		},
	}
	return &newTx
}

//Send back vote token to voters who have vote to lose candidate
func (blockgen *BlkTmplGenerator) CreateSendBackDCBTokenAfterVoteFail(chainID byte, newDCBList [][]byte) []metadata.Transaction {
	setOfNewDCB := make(map[string]bool, 0)
	for _, i := range newDCBList {
		setOfNewDCB[string(i)] = true
	}
	currentHeight := blockgen.chain.BestState[chainID].Height
	db := blockgen.chain.config.DataBase
	begin := db.GetKey(string(blockgen.chain.config.DataBase.GetVoteDCBBoardListPrefix()), string(0))
	end := db.GetKey(string(blockgen.chain.config.DataBase.GetVoteDCBBoardListPrefix()), currentHeight)
	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}

	iter := blockgen.chain.config.DataBase.NewIterator(&searchRange, nil)
	listNewTx := make([]metadata.Transaction, 0)
	for iter.Next() {
		key := iter.Key()
		startedBlock, PubKey, _ := blockgen.parseVoteDCBBoardListKey(key)
		value := iter.Value()
		senderPubkey, amountOfDCBToken := parseVoteDCBBoardListValue(value)
		_, found := setOfNewDCB[string(PubKey)]
		if startedBlock < currentHeight || !found {
			listNewTx = append(listNewTx, createSingleSendDCBVoteTokenFail(senderPubkey, amountOfDCBToken))
		}
	}
	return listNewTx
}

func (blockgen *BlkTmplGenerator) CreateSendBackGOVTokenAfterVoteFail(chainID byte, newGOVList [][]byte) []metadata.Transaction {
	setOfNewGOV := make(map[string]bool, 0)
	for _, i := range newGOVList {
		setOfNewGOV[string(i)] = true
	}
	currentHeight := blockgen.chain.BestState[chainID].Height
	db := blockgen.chain.config.DataBase
	begin := db.GetKey(string(blockgen.chain.config.DataBase.GetVoteGOVBoardListPrefix()), string(0))
	end := db.GetKey(string(blockgen.chain.config.DataBase.GetVoteGOVBoardListPrefix()), currentHeight)
	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}

	iter := blockgen.chain.config.DataBase.NewIterator(&searchRange, nil)
	listNewTx := make([]metadata.Transaction, 0)
	for iter.Next() {
		key := iter.Key()
		startedBlock, PubKey, _ := blockgen.parseVoteGOVBoardListKey(key)
		value := iter.Value()
		senderPubkey, amountOfGOVToken := parseVoteGOVBoardListValue(value)
		_, found := setOfNewGOV[string(PubKey)]
		if startedBlock < currentHeight || !found {
			listNewTx = append(listNewTx, createSingleSendGOVVoteTokenFail(senderPubkey, amountOfGOVToken))
		}
	}
	return listNewTx
}

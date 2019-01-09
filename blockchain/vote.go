package blockchain

import (
	"encoding/binary"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/database/lvdb"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (blockgen *BlkTmplGenerator) createRewardProposalWinnerTx(chainID byte, constitutionHelper ConstitutionHelper,
) (metadata.Transaction, error) {
	pubKey, _ := constitutionHelper.GetPubKeyVoter(blockgen, chainID)
	prize := constitutionHelper.GetPrizeProposal()
	meta := metadata.NewRewardProposalWinnerMetadata(pubKey, prize)
	tx := transaction.Tx{
		Metadata: meta,
	}
	return &tx, nil
}

func (blockgen *BlkTmplGenerator) createAcceptConstitutionAndPunishTxAndRewardSubmitter(
	chainID byte,
	helper ConstitutionHelper,
	minerPrivateKey *privacy.SpendingKey,
) ([]metadata.Transaction, error) {
	resTx := make([]metadata.Transaction, 0)
	SumVote := make(map[common.Hash]uint64)
	CountVote := make(map[common.Hash]uint32)
	VoteTable := make(map[common.Hash]map[string]int32)
	NextConstitutionIndex := blockgen.chain.GetCurrentBoardIndex(helper)

	db := blockgen.chain.config.DataBase
	boardType := helper.GetLowerCaseBoardType()
	begin := lvdb.GetKeyThreePhraseCryptoSealer(boardType, 0, nil)
	// +1 to search in that range
	end := lvdb.GetKeyThreePhraseCryptoSealer(boardType, 1+NextConstitutionIndex, nil)

	searchrange := util.Range{
		Start: begin,
		Limit: end,
	}
	iter := db.NewIterator(&searchrange, nil)
	rightIndex := blockgen.chain.GetConstitutionIndex(helper) + 1
	for iter.Next() {
		key := iter.Key()
		_, constitutionIndex, transactionID, err := lvdb.ParseKeyThreePhraseCryptoSealer(key)
		if err != nil {
			return nil, err
		}
		if constitutionIndex != uint32(rightIndex) {
			//@todo 0xjackalope delete all relevant thing
			db.Delete(key)
			continue
		}
		//Punish owner if he don't send decrypted message
		keyOwner := lvdb.GetKeyThreePhraseCryptoOwner(boardType, constitutionIndex, transactionID)
		valueOwnerInByte, err := db.Get(keyOwner)
		if err != nil {
			return nil, err
		}
		valueOwner, err := lvdb.ParseValueThreePhraseCryptoOwner(valueOwnerInByte)
		if err != nil {
			return nil, err
		}

		_, _, _, lv3Tx, _ := blockgen.chain.GetTransactionByHash(transactionID)
		sealerPubKeyList := helper.GetSealerPubKey(lv3Tx)
		if valueOwner != 1 {
			newTx := transaction.Tx{
				Metadata: helper.CreatePunishDecryptTx(sealerPubKeyList[0]),
			}
			resTx = append(resTx, &newTx)
		}
		//Punish sealer if he don't send decrypted message
		keySealer := lvdb.GetKeyThreePhraseCryptoSealer(boardType, constitutionIndex, transactionID)
		valueSealerInByte, err := db.Get(keySealer)
		if err != nil {
			return nil, err
		}
		valueSealer := binary.LittleEndian.Uint32(valueSealerInByte)
		if valueSealer != 3 {
			//Count number of time she don't send encrypted message if number==2 create punish transaction
			newTx := transaction.Tx{
				Metadata: helper.CreatePunishDecryptTx(sealerPubKeyList[valueSealer]),
			}
			resTx = append(resTx, &newTx)
		}

		//Accumulate count vote
		voter := sealerPubKeyList[0]
		keyVote := lvdb.GetKeyThreePhraseVoteValue(boardType, constitutionIndex, transactionID)
		valueVote, err := db.Get(keyVote)
		if err != nil {
			return nil, err
		}
		txId, voteAmount, err := lvdb.ParseValueThreePhraseVoteValue(valueVote)
		if err != nil {
			return nil, err
		}

		SumVote[*txId] += uint64(voteAmount)
		if VoteTable[*txId] == nil {
			VoteTable[*txId] = make(map[string]int32)
		}
		VoteTable[*txId][string(voter)] += voteAmount
		CountVote[*txId] += 1
	}

	bestProposal := metadata.ProposalVote{
		TxId:         common.Hash{},
		AmountOfVote: 0,
		NumberOfVote: 0,
	}
	bestVoterAll := metadata.Voter{
		PubKey:       make([]byte, 0),
		AmountOfVote: 0,
	}
	// Get most vote proposal
	for txId, listVoter := range VoteTable {
		bestVoterThisProposal := metadata.Voter{
			PubKey:       make([]byte, 0),
			AmountOfVote: 0,
		}
		amountOfThisProposal := int64(0)
		countOfThisProposal := uint32(0)
		for voterPubKey, amount := range listVoter {
			voterToken, _ := db.GetAmountVoteToken(boardType, NextConstitutionIndex, []byte(voterPubKey))
			if int32(voterToken) < amount || amount < 0 {
				listVoter[voterPubKey] = 0
				// can change listvoter because it is a pointer
				continue
			} else {
				tVoter := metadata.Voter{
					PubKey:       []byte(voterPubKey),
					AmountOfVote: amount,
				}
				if tVoter.Greater(bestVoterThisProposal) {
					bestVoterThisProposal = tVoter
				}
				amountOfThisProposal += int64(tVoter.AmountOfVote)
				countOfThisProposal += 1
			}
		}
		amountOfThisProposal -= int64(bestVoterThisProposal.AmountOfVote)
		tProposalVote := metadata.ProposalVote{
			TxId:         txId,
			AmountOfVote: amountOfThisProposal,
			NumberOfVote: countOfThisProposal,
		}
		if tProposalVote.Greater(bestProposal) {
			bestProposal = tProposalVote
			bestVoterAll = bestVoterThisProposal
		}
	}
	acceptedSubmitProposalTransaction := helper.TxAcceptProposal(&bestProposal.TxId, bestVoterAll)
	_, _, _, bestSubmittedProposal, _ := blockgen.chain.GetTransactionByHash(&bestProposal.TxId)
	submitterPaymentAddress := helper.GetPaymentAddressFromSubmitProposalMetadata(bestSubmittedProposal)

	// If submitterPaymentAdress use don't use privacy for
	if submitterPaymentAddress == nil {
		rewardForProposalSubmitter, err := helper.NewTxRewardProposalSubmitter(blockgen, submitterPaymentAddress, minerPrivateKey)
		if err != nil {
			return nil, err
		}
		resTx = append(resTx, rewardForProposalSubmitter)
	}

	resTx = append(resTx, acceptedSubmitProposalTransaction)

	return resTx, nil
}

func (blockgen *BlkTmplGenerator) createSingleSendDCBVoteTokenTx(chainID byte, pubKey []byte, amount uint32) (metadata.Transaction, error) {
	sendDCBVoteTokenTransaction := transaction.Tx{
		Metadata: metadata.NewSendInitDCBVoteTokenMetadata(amount, pubKey),
	}
	return &sendDCBVoteTokenTransaction, nil
}

func (blockgen *BlkTmplGenerator) createSingleSendGOVVoteTokenTx(chainID byte, pubKey []byte, amount uint32) (metadata.Transaction, error) {
	sendGOVVoteTokenTransaction := transaction.Tx{
		Metadata: metadata.NewSendInitGOVVoteTokenMetadata(amount, pubKey),
	}
	return &sendGOVVoteTokenTransaction, nil
}

func getAmountOfVoteToken(sumAmount uint64, voteAmount uint64) uint64 {
	return voteAmount * common.SumOfVoteDCBToken / sumAmount
}

func (blockgen *BlkTmplGenerator) CreateSendDCBVoteTokenToGovernorTx(chainID byte, newDCBList database.CandidateList, sumAmountDCB uint64) []metadata.Transaction {
	var SendVoteTx []metadata.Transaction
	var newTx metadata.Transaction
	for i := 0; i <= common.NumberOfDCBGovernors; i++ {
		newTx, _ = blockgen.createSingleSendDCBVoteTokenTx(chainID, newDCBList[i].PubKey, uint32(getAmountOfVoteToken(sumAmountDCB, newDCBList[i].VoteAmount)))
		SendVoteTx = append(SendVoteTx, newTx)
	}
	return SendVoteTx
}

func (blockgen *BlkTmplGenerator) CreateSendGOVVoteTokenToGovernorTx(chainID byte, newGOVList database.CandidateList, sumAmountGOV uint64) []metadata.Transaction {
	var SendVoteTx []metadata.Transaction
	var newTx metadata.Transaction
	for i := 0; i <= common.NumberOfGOVGovernors; i++ {
		newTx, _ = blockgen.createSingleSendGOVVoteTokenTx(chainID, newGOVList[i].PubKey, uint32(getAmountOfVoteToken(sumAmountGOV, newGOVList[i].VoteAmount)))
		SendVoteTx = append(SendVoteTx, newTx)
	}
	return SendVoteTx
}

func (blockgen *BlkTmplGenerator) createAcceptDCBBoardTx(DCBBoardPubKeys [][]byte, sumOfVote uint64) metadata.Transaction {
	return &transaction.Tx{
		Metadata: metadata.NewAcceptDCBBoardMetadata(DCBBoardPubKeys, sumOfVote),
	}
}

func (blockgen *BlkTmplGenerator) createAcceptGOVBoardTx(DCBBoardPubKeys [][]byte, sumOfVote uint64) metadata.Transaction {
	return &transaction.Tx{
		Metadata: metadata.NewAcceptGOVBoardMetadata(DCBBoardPubKeys, sumOfVote),
	}
}

func (block *Block) UpdateDCBBoard(thisTx metadata.Transaction) error {
	meta := thisTx.GetMetadata().(*metadata.AcceptDCBBoardMetadata)
	block.Header.DCBGovernor.BoardPubKeys = meta.DCBBoardPubKeys
	block.Header.DCBGovernor.StartedBlock = uint32(block.Header.Height)
	block.Header.DCBGovernor.EndBlock = block.Header.DCBGovernor.StartedBlock + common.DurationOfTermDCB
	block.Header.DCBGovernor.StartAmountToken = meta.StartAmountDCBToken
	return nil
}

func (block *Block) UpdateGOVBoard(thisTx metadata.Transaction) error {
	meta := thisTx.GetMetadata().(*metadata.AcceptGOVBoardMetadata)
	block.Header.GOVGovernor.BoardPubKeys = meta.GOVBoardPubKeys
	block.Header.GOVGovernor.StartedBlock = uint32(block.Header.Height)
	block.Header.GOVGovernor.EndBlock = block.Header.GOVGovernor.StartedBlock + common.DurationOfTermGOV
	block.Header.GOVGovernor.StartAmountToken = meta.StartAmountGOVToken
	return nil
}

func (block *Block) UpdateDCBFund(tx metadata.Transaction) error {
	block.Header.BankFund -= common.RewardProposalSubmitter
	return nil
}

func (block *Block) UpdateGOVFund(tx metadata.Transaction) error {
	block.Header.SalaryFund -= common.RewardProposalSubmitter
	return nil
}

func parseVoteDCBBoardListValue(value []byte) ([]byte, uint64) {
	voterPubKey := value[:common.PubKeyLength]
	amount := binary.LittleEndian.Uint64(value[common.PubKeyLength:])
	return voterPubKey, amount
}

func parseVoteGOVBoardListValue(value []byte) ([]byte, uint64) {
	voterPubKey := value[:common.PubKeyLength]
	amount := binary.LittleEndian.Uint64(value[common.PubKeyLength:])
	return voterPubKey, amount
}

func createSingleSendDCBVoteTokenFail(paymentAddress privacy.PaymentAddress, amount uint64) metadata.Transaction {
	txTokenVout := transaction.TxTokenVout{
		Value:          amount,
		PaymentAddress: paymentAddress,
	}
	newTx := transaction.TxCustomToken{
		TxTokenData: transaction.TxTokenData{
			Type:       transaction.SendBackDCBTokenVoteFail,
			Amount:     amount,
			PropertyID: common.DCBTokenID,
			Vins:       []transaction.TxTokenVin{},
			Vouts:      []transaction.TxTokenVout{txTokenVout},
		},
	}
	return &newTx
}

func createSingleSendGOVVoteTokenFail(paymentAddress privacy.PaymentAddress, amount uint64) metadata.Transaction {
	txTokenVout := transaction.TxTokenVout{
		Value:          amount,
		PaymentAddress: paymentAddress,
	}
	newTx := transaction.TxCustomToken{
		TxTokenData: transaction.TxTokenData{
			Type:       transaction.SendBackGOVTokenVoteFail,
			Amount:     amount,
			PropertyID: common.GOVTokenID,
			Vins:       []transaction.TxTokenVin{},
			Vouts:      []transaction.TxTokenVout{txTokenVout},
		},
	}
	return &newTx
}

//Send back vote token to voters who have vote to lose candidate
func (blockgen *BlkTmplGenerator) CreateSendBackTokenAfterVoteFail(boardType string, chainID byte, newDCBList [][]byte) []metadata.Transaction {
	setOfNewDCB := make(map[string]bool, 0)
	for _, i := range newDCBList {
		setOfNewDCB[string(i)] = true
	}
	currentBoardIndex := blockgen.chain.GetCurrentBoardIndex(DCBConstitutionHelper{})
	db := blockgen.chain.config.DataBase
	begin := lvdb.GetKeyVoteBoardList(boardType, 0, make([]byte, common.PubKeyLength), make([]byte, common.PubKeyLength))
	end := lvdb.GetKeyVoteBoardList(boardType, currentBoardIndex+1, make([]byte, common.PubKeyLength), make([]byte, common.PubKeyLength))
	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}

	iter := blockgen.chain.config.DataBase.NewIterator(&searchRange, nil)
	listNewTx := make([]metadata.Transaction, 0)
	for iter.Next() {
		key := iter.Key()
		_, boardIndex, PubKey, _, _ := lvdb.ParseKeyVoteBoardList(key)
		value := iter.Value()
		senderPubKey, amountOfDCBToken := parseVoteDCBBoardListValue(value)
		_, found := setOfNewDCB[string(PubKey)]
		if boardIndex < uint32(currentBoardIndex) || !found {
			paymentAddressByte := db.GetPaymentAddressFromPubKey(senderPubKey)
			paymentAddress := privacy.PaymentAddress{}
			paymentAddress.SetBytes(paymentAddressByte)
			listNewTx = append(listNewTx, createSingleSendDCBVoteTokenFail(paymentAddress, amountOfDCBToken))
		}
	}
	return listNewTx
}

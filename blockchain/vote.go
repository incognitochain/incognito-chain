package blockchain

import (
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/database/lvdb"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type ConstitutionHelper interface {
	GetStartedNormalVote(generator *BlkTmplGenerator, chainID byte) uint64
	CheckSubmitProposalType(tx metadata.Transaction) bool
	CheckVotingProposalType(tx metadata.Transaction) bool
	GetAmountVoteTokenOfTx(tx metadata.Transaction) uint64
	TxAcceptProposal(txId *common.Hash, voter metadata.Voter) metadata.Transaction
	GetBoardType() string
	GetConstitutionEndedBlockHeight(generator *BlkTmplGenerator, chainID byte) uint64
	CreatePunishDecryptTx([]byte) metadata.Metadata
	GetSealerPubKey(metadata.Transaction) [][]byte
	NewTxRewardProposalSubmitter(blockgen *BlkTmplGenerator, receiverAddress *privacy.PaymentAddress, minerPrivateKey *privacy.SpendingKey) (metadata.Transaction, error)
	GetPaymentAddressFromSubmitProposalMetadata(tx metadata.Transaction) *privacy.PaymentAddress
	GetPubKeyVoter(blockgen *BlkTmplGenerator, chainID byte) ([]byte, error)
	GetPrizeProposal() uint32
	GetTopMostVoteGovernor(blockgen *BlkTmplGenerator) (database.CandidateList, error)
	GetCurrentBoardPubKeys(blockgen *BlkTmplGenerator) [][]byte
	GetAmountVoteTokenOfBoard(blockgen *BlkTmplGenerator, pubKey []byte, boardIndex uint32) uint64
	GetBoardSumToken(blockgen *BlkTmplGenerator) uint64
	GetBoardFund(blockgen *BlkTmplGenerator) uint64
	GetTokenID() *common.Hash
	GetAmountOfVoteToBoard(blockgen *BlkTmplGenerator, candidatePubKey []byte, voterPubKey []byte, boardIndex uint32) uint64
	GetBoard(chain *BlockChain) Governor
	GetConstitutionInfo(chain *BlockChain) ConstitutionInfo
	GetCurrentNationalWelfare(chain *BlockChain) int32
	GetThresholdRatioOfCrisis() int32
	GetOldNationalWelfare(chain *BlockChain) int32
}

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
	boardType := helper.GetBoardType()
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
		proposalData := metadata.NewVoteProposalDataFromBytes(valueVote)
		txId, voteAmount := &proposalData.ProposalTxID, proposalData.AmountOfVote
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

func GetOracleDCBNationalWelfare() int32 {
	fmt.Print("Get national welfare. It is constant now. Need to change !!!")
	return 1234
}
func GetOracleGOVNationalWelfare() int32 {
	fmt.Print("Get national welfare. It is constant now. Need to change !!!")
	return 1234
}

func (blockgen *BlkTmplGenerator) CreateSingleShareRewardOldBoard(
	helper ConstitutionHelper,
	chairPubKey []byte,
	voterPubKey []byte,
	amountOfCoin uint64,
	amountOfToken uint64,
	minerPrivateKey *privacy.SpendingKey,
) metadata.Transaction {
	paymentAddressByte := blockgen.chain.config.DataBase.GetPaymentAddressFromPubKey(voterPubKey)
	paymentAddress := privacy.PaymentAddress{}
	paymentAddress.SetBytes(paymentAddressByte)
	tx := transaction.Tx{}
	rewardShareOldBoardMeta := metadata.NewRewardShareOldBoardMetadata(chairPubKey, voterPubKey, helper.GetBoardType())
	tx.InitTxSalary(amountOfCoin, &paymentAddress, minerPrivateKey, blockgen.chain.config.DataBase, rewardShareOldBoardMeta)
	txTokenData := transaction.TxTokenData{
		Type:       transaction.CustomTokenInit,
		Amount:     amountOfToken,
		PropertyID: *helper.GetTokenID(),
		Vins:       []transaction.TxTokenVin{},
		Vouts:      []transaction.TxTokenVout{},
	}

	txCustomToken := transaction.TxCustomToken{
		Tx:          tx,
		TxTokenData: txTokenData,
	}
	return &txCustomToken
}

func (blockgen *BlkTmplGenerator) CreateShareRewardOldBoard(
	helper ConstitutionHelper,
	chairPubKey []byte,
	totalAmountCoinReward uint64,
	totalAmountTokenReward uint64,
	totalVoteAmount uint64,
	minerPrivateKey *privacy.SpendingKey,
) []metadata.Transaction {
	txs := make([]metadata.Transaction, 0)

	voterList := blockgen.chain.config.DataBase.GetBoardVoterList(helper.GetBoardType(), chairPubKey, blockgen.chain.GetCurrentBoardIndex(helper))
	boardIndex := blockgen.chain.GetCurrentBoardIndex(helper)
	for _, pubKey := range voterList {
		amountOfVote := helper.GetAmountOfVoteToBoard(blockgen, chairPubKey, pubKey, boardIndex)
		amountOfCoin := amountOfVote * totalAmountCoinReward / totalVoteAmount
		amountOfToken := amountOfVote * totalAmountTokenReward / totalVoteAmount
		blockgen.CreateSingleShareRewardOldBoard(helper, chairPubKey, pubKey, amountOfCoin, amountOfToken, minerPrivateKey)
	}
	return txs
}

func (blockgen *BlkTmplGenerator) GetCoinTermReward(helper ConstitutionHelper) uint64 {
	return helper.GetBoardFund(blockgen) * common.PercentageBoardSalary / common.BasePercentage
}

func (blockgen *BlkTmplGenerator) CreateSendRewardOldBoard(helper ConstitutionHelper, minerPrivateKey *privacy.SpendingKey) []metadata.Transaction {
	txs := make([]metadata.Transaction, 0)
	voteTokenAmount := make(map[string]uint64)
	sumVoteTokenAmount := uint64(0)
	pubKeys := helper.GetCurrentBoardPubKeys(blockgen)
	totalAmountOfTokenReward := helper.GetBoardSumToken(blockgen) // Total amount of token
	totalAmountOfCoinReward := blockgen.GetCoinTermReward(helper)
	helper.GetBoardFund(blockgen)
	//reward for each by voteDCBList
	for _, i := range pubKeys {
		amount := helper.GetAmountVoteTokenOfBoard(blockgen, i, blockgen.chain.GetCurrentBoardIndex(helper))
		voteTokenAmount[string(i)] = amount
		sumVoteTokenAmount += amount
	}
	for pubKey, voteAmount := range voteTokenAmount {
		percentageReward := voteAmount * common.BasePercentage / sumVoteTokenAmount
		amountTokenReward := totalAmountOfTokenReward * uint64(percentageReward) / common.BasePercentage
		amountCoinReward := totalAmountOfCoinReward * uint64(percentageReward) / common.BasePercentage
		txs = append(txs, blockgen.CreateShareRewardOldBoard(helper, []byte(pubKey), amountCoinReward, amountTokenReward, voteAmount, minerPrivateKey)...)
		//todo @0xjackalope: reward for chair
	}
	return txs
}

func (blockgen *BlkTmplGenerator) UpdateNewGovernor(helper ConstitutionHelper, chainID byte, minerPrivateKey *privacy.SpendingKey) []metadata.Transaction {
	txs := make([]metadata.Transaction, 0)
	newBoardList, _ := helper.GetTopMostVoteGovernor(blockgen)
	sort.Sort(newBoardList)
	sumOfVote := uint64(0)
	var newDCBBoardPubKey [][]byte
	for _, i := range newBoardList {
		newDCBBoardPubKey = append(newDCBBoardPubKey, i.PubKey)
		sumOfVote += i.VoteAmount
	}

	txs = append(txs, blockgen.createAcceptDCBBoardTx(newDCBBoardPubKey, sumOfVote))
	txs = append(txs, blockgen.CreateSendDCBVoteTokenToGovernorTx(chainID, newBoardList, sumOfVote)...)

	txs = append(txs, blockgen.CreateSendBackTokenAfterVoteFail(helper.GetBoardType(), chainID, newDCBBoardPubKey)...)

	txs = append(txs, blockgen.CreateSendRewardOldBoard(helper, minerPrivateKey)...)

	return txs
}

func (blockgen *BlkTmplGenerator) neededNewDCBGovernor(chainID byte) bool {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	return int32(BestBlock.Header.DCBGovernor.EndBlock) == BestBlock.Header.Height+2
}
func (blockgen *BlkTmplGenerator) neededNewGOVGovernor(chainID byte) bool {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	return int32(BestBlock.Header.GOVGovernor.EndBlock) == BestBlock.Header.Height+2
}

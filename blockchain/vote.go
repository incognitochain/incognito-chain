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
	GetStartedNormalVote(chain *BlockChain) uint64
	CheckSubmitProposalType(tx metadata.Transaction) bool
	GetAmountVoteTokenOfTx(tx metadata.Transaction) uint64
	TxAcceptProposal(txId *common.Hash, voter metadata.Voter, minerPrivateKey *privacy.SpendingKey, db database.DatabaseInterface) metadata.Transaction
	GetBoardType() byte
	GetConstitutionEndedBlockHeight(chain *BlockChain) uint64
	CreatePunishDecryptTx(address privacy.PaymentAddress) metadata.Metadata
	GetSealerPaymentAddress(metadata.Transaction) []privacy.PaymentAddress
	NewTxRewardProposalSubmitter(blockgen *BlockChain, receiverAddress *privacy.PaymentAddress, minerPrivateKey *privacy.SpendingKey) (metadata.Transaction, error)
	GetPaymentAddressFromSubmitProposalMetadata(tx metadata.Transaction) *privacy.PaymentAddress
	GetPaymentAddressVoter(chain *BlockChain) (privacy.PaymentAddress, error)
	GetPrizeProposal() uint32
	GetCurrentBoardPaymentAddress(chain *BlockChain) []privacy.PaymentAddress
	GetAmountVoteTokenOfBoard(chain *BlockChain, paymentAddress privacy.PaymentAddress, boardIndex uint32) uint64
	GetBoardSumToken(chain *BlockChain) uint64
	GetBoardFund(chain *BlockChain) uint64
	GetTokenID() *common.Hash
	GetAmountOfVoteToBoard(chain *BlockChain, candidatePaymentAddress privacy.PaymentAddress, voterPaymentAddress privacy.PaymentAddress, boardIndex uint32) uint64
	GetBoard(chain *BlockChain) Governor
	GetConstitutionInfo(chain *BlockChain) ConstitutionInfo
	GetCurrentNationalWelfare(chain *BlockChain) int32
	GetThresholdRatioOfCrisis() int32
	GetOldNationalWelfare(chain *BlockChain) int32
	GetNumberOfGovernor() int32
}

func (blockgen *BlkTmplGenerator) createRewardProposalWinnerTx(
	chainID byte,
	constitutionHelper ConstitutionHelper,
	minerPrivateKey *privacy.SpendingKey,
) (metadata.Transaction, error) {
	paymentAddress, err := constitutionHelper.GetPaymentAddressVoter(blockgen.chain)
	if err != nil {
		return nil, err
	}
	prize := constitutionHelper.GetPrizeProposal()
	meta := metadata.NewRewardProposalWinnerMetadata(paymentAddress, prize)
	tx := transaction.NewEmptyTx(minerPrivateKey, blockgen.chain.config.DataBase, meta)
	return tx, nil
}

func (self *BlockChain) BuildVoteTableAndPunishTransaction(
	helper ConstitutionHelper,
	minerPrivateKey *privacy.SpendingKey,
) (
	resTx []metadata.Transaction,
	VoteTable map[common.Hash]map[string]int32,
	err error,
) {
	resTx = make([]metadata.Transaction, 0)
	VoteTable = make(map[common.Hash]map[string]int32)
	SumVote := make(map[common.Hash]uint64)
	CountVote := make(map[common.Hash]uint32)
	NextConstitutionIndex := self.GetCurrentBoardIndex(helper)

	db := self.config.DataBase
	boardType := helper.GetBoardType()
	begin := lvdb.GetKeyThreePhraseCryptoSealer(boardType, 0, nil)
	// +1 to search in that range
	end := lvdb.GetKeyThreePhraseCryptoSealer(boardType, 1+NextConstitutionIndex, nil)

	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}
	iter := db.NewIterator(&searchRange, nil)
	rightIndex := self.GetConstitutionIndex(helper) + 1
	for iter.Next() {
		key := iter.Key()
		_, constitutionIndex, transactionID, err := lvdb.ParseKeyThreePhraseCryptoSealer(key)
		if err != nil {
			return nil, nil, err
		}
		if constitutionIndex != uint32(rightIndex) {
			db.Delete(key)
			continue
		}
		//Punish owner if he don't send decrypted message
		keyOwner := lvdb.GetKeyThreePhraseCryptoOwner(boardType, constitutionIndex, transactionID)
		valueOwnerInByte, err := db.Get(keyOwner)
		if err != nil {
			return nil, nil, err
		}
		valueOwner, err := lvdb.ParseValueThreePhraseCryptoOwner(valueOwnerInByte)
		if err != nil {
			return nil, nil, err
		}

		_, _, _, lv3Tx, _ := self.GetTransactionByHash(transactionID)
		sealerPaymentAddressList := helper.GetSealerPaymentAddress(lv3Tx)
		if valueOwner != 1 {
			meta := helper.CreatePunishDecryptTx(sealerPaymentAddressList[0])
			newTx := transaction.NewEmptyTx(minerPrivateKey, db, meta)
			resTx = append(resTx, newTx)
		}
		//Punish sealer if he don't send decrypted message
		keySealer := lvdb.GetKeyThreePhraseCryptoSealer(boardType, constitutionIndex, transactionID)
		valueSealerInByte, err := db.Get(keySealer)
		if err != nil {
			return nil, nil, err
		}
		valueSealer := binary.LittleEndian.Uint32(valueSealerInByte)
		if valueSealer != 3 {
			//Count number of time she don't send encrypted message if number==2 create punish transaction
			meta := helper.CreatePunishDecryptTx(sealerPaymentAddressList[valueSealer])
			newTx := transaction.NewEmptyTx(minerPrivateKey, self.config.DataBase, meta)
			resTx = append(resTx, newTx)
		}

		//Accumulate count vote
		voter := sealerPaymentAddressList[0]
		keyVote := lvdb.GetKeyThreePhraseVoteValue(boardType, constitutionIndex, transactionID)
		valueVote, err := db.Get(keyVote)
		if err != nil {
			return nil, nil, err
		}
		proposalData := metadata.NewVoteProposalDataFromBytes(valueVote)
		txId, voteAmount := &proposalData.ProposalTxID, proposalData.AmountOfVote
		if err != nil {
			return nil, nil, err
		}

		SumVote[*txId] += uint64(voteAmount)
		if VoteTable[*txId] == nil {
			VoteTable[*txId] = make(map[string]int32)
		}
		VoteTable[*txId][string(voter.Bytes())] += voteAmount
		CountVote[*txId] += 1
	}
	return
}

func (self *BlockChain) createAcceptConstitutionAndPunishTxAndRewardSubmitter(
	helper ConstitutionHelper,
	minerPrivateKey *privacy.SpendingKey,
) ([]metadata.Transaction, error) {
	resTx, VoteTable, err := self.BuildVoteTableAndPunishTransaction(helper, minerPrivateKey)
	NextConstitutionIndex := self.GetCurrentBoardIndex(helper)
	bestProposal := metadata.ProposalVote{
		TxId:         common.Hash{},
		AmountOfVote: 0,
		NumberOfVote: 0,
	}
	var bestVoterAll metadata.Voter
	// Get most vote proposal
	db := self.config.DataBase
	for txId, listVoter := range VoteTable {
		var bestVoterThisProposal metadata.Voter
		amountOfThisProposal := int64(0)
		countOfThisProposal := uint32(0)
		for voterPaymentAddressBytes, amount := range listVoter {
			voterPaymentAddress := privacy.NewPaymentAddressFromByte([]byte(voterPaymentAddressBytes))
			voterToken, _ := db.GetVoteTokenAmount(helper.GetBoardType(), NextConstitutionIndex, *voterPaymentAddress)
			if int32(voterToken) < amount || amount < 0 {
				listVoter[string(voterPaymentAddress.Bytes())] = 0
				// can change listvoter because it is a pointer
				continue
			} else {
				tVoter := metadata.Voter{
					PaymentAddress: *voterPaymentAddress,
					AmountOfVote:   amount,
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
	acceptedSubmitProposalTransaction := helper.TxAcceptProposal(&bestProposal.TxId, bestVoterAll, minerPrivateKey, db)
	_, _, _, bestSubmittedProposal, err := self.GetTransactionByHash(&bestProposal.TxId)
	if err != nil {
		return nil, err
	}
	submitterPaymentAddress := helper.GetPaymentAddressFromSubmitProposalMetadata(bestSubmittedProposal)

	// If submitterPaymentAdress use don't use privacy for
	if submitterPaymentAddress == nil {
		rewardForProposalSubmitter, err := helper.NewTxRewardProposalSubmitter(self, submitterPaymentAddress, minerPrivateKey)
		if err != nil {
			return nil, err
		}
		resTx = append(resTx, rewardForProposalSubmitter)
	}

	resTx = append(resTx, acceptedSubmitProposalTransaction)

	return resTx, nil
}

func (self *BlockChain) createSingleSendVoteTokenTx(
	helper ConstitutionHelper,
	paymentAddress privacy.PaymentAddress,
	amount uint32,
	minerPrivateKey *privacy.SpendingKey,
) metadata.Transaction {
	var meta metadata.Metadata
	if helper.GetBoardType() == common.DCBBoard {
		meta = metadata.NewSendInitDCBVoteTokenMetadata(amount, paymentAddress)
	} else if helper.GetBoardType() == common.GOVBoard {
		meta = metadata.NewSendInitGOVVoteTokenMetadata(amount, paymentAddress)
	}
	sendVoteTokenTransaction := transaction.NewEmptyTx(minerPrivateKey, self.config.DataBase, meta)
	return sendVoteTokenTransaction
}

func getAmountOfVoteToken(sumAmount uint64, voteAmount uint64) uint64 {
	// TODO: 0xjackalop
	// not check sumAmount = 0
	return voteAmount * common.SumOfVoteDCBToken / sumAmount
}

func (self *BlockChain) CreateSendVoteTokenToGovernorTx(
	helper ConstitutionHelper,
	newBoardList database.CandidateList,
	sumAmountToken uint64,
	minerPrivateKey *privacy.SpendingKey,
) []metadata.Transaction {
	var SendVoteTx []metadata.Transaction
	var newTx metadata.Transaction
	for i := int32(0); i < helper.GetNumberOfGovernor(); i++ {
		newTx = self.createSingleSendVoteTokenTx(
			helper,
			newBoardList[i].PaymentAddress,
			uint32(getAmountOfVoteToken(sumAmountToken, newBoardList[i].VoteAmount)),
			minerPrivateKey,
		)
		SendVoteTx = append(SendVoteTx, newTx)
	}
	return SendVoteTx
}

func (self *BlockChain) createAcceptBoardTx(
	boardType byte,
	DCBBoardPaymentAddress []privacy.PaymentAddress,
	sumOfVote uint64,
	minerPrivateKey *privacy.SpendingKey,
) metadata.Transaction {
	var meta metadata.Metadata
	if boardType == common.DCBBoard {
		meta = metadata.NewAcceptDCBBoardMetadata(DCBBoardPaymentAddress, sumOfVote)
	} else {
		meta = metadata.NewAcceptGOVBoardMetadata(DCBBoardPaymentAddress, sumOfVote)
	}
	tx := transaction.NewEmptyTx(minerPrivateKey, self.config.DataBase, meta)
	return tx
}

func (block *BeaconBlock) UpdateDCBBoard(thisTx metadata.Transaction) error {
	// meta := thisTx.GetMetadata().(*metadata.AcceptDCBBoardMetadata)
	// block.Header.DCBGovernor.BoardIndex += 1
	// block.Header.DCBGovernor.BoardPaymentAddress = meta.DCBBoardPaymentAddress
	// block.Header.DCBGovernor.StartedBlock = uint32(block.Header.Height)
	// block.Header.DCBGovernor.EndBlock = block.Header.DCBGovernor.StartedBlock + common.DurationOfTermDCB
	// block.Header.DCBGovernor.StartAmountToken = meta.StartAmountDCBToken
	return nil
}

func (block *BeaconBlock) UpdateGOVBoard(thisTx metadata.Transaction) error {
	// meta := thisTx.GetMetadata().(*metadata.AcceptGOVBoardMetadata)
	// block.Header.GOVGovernor.BoardPaymentAddress = meta.GOVBoardPaymentAddress
	// block.Header.GOVGovernor.StartedBlock = uint32(block.Header.Height)
	// block.Header.GOVGovernor.EndBlock = block.Header.GOVGovernor.StartedBlock + common.DurationOfTermGOV
	// block.Header.GOVGovernor.StartAmountToken = meta.StartAmountGOVToken
	return nil
}

func (block *BeaconBlock) UpdateDCBFund(tx metadata.Transaction) error {
	// block.Header.BankFund -= common.RewardProposalSubmitter
	return nil
}

func (block *BeaconBlock) UpdateGOVFund(tx metadata.Transaction) error {
	// block.Header.SalaryFund -= common.RewardProposalSubmitter
	return nil
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
func (self *BlockChain) CreateSendBackTokenAfterVoteFail(
	boardType byte,
	newDCBList []privacy.PaymentAddress,
) []metadata.Transaction {
	setOfNewDCB := make(map[string]bool)
	for _, i := range newDCBList {
		setOfNewDCB[string(i.Bytes())] = true
	}
	currentBoardIndex := self.GetCurrentBoardIndex(DCBConstitutionHelper{})
	begin := lvdb.GetKeyVoteBoardList(boardType, 0, nil, nil)
	end := lvdb.GetKeyVoteBoardList(boardType, currentBoardIndex+1, nil, nil)
	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}

	iter := self.config.DataBase.NewIterator(&searchRange, nil)
	listNewTx := make([]metadata.Transaction, 0)
	for iter.Next() {
		key := iter.Key()
		_, boardIndex, candidatePubKey, voterPaymentAddress, _ := lvdb.ParseKeyVoteBoardList(key)
		value := iter.Value()
		amountOfDCBToken := lvdb.ParseValueVoteBoardList(value)

		_, found := setOfNewDCB[string(candidatePubKey)]
		if boardIndex < uint32(currentBoardIndex) || !found {
			listNewTx = append(listNewTx, createSingleSendDCBVoteTokenFail(*voterPaymentAddress, amountOfDCBToken))
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

func (chain *BlockChain) CreateSingleShareRewardOldBoard(
	helper ConstitutionHelper,
	chairPaymentAddress privacy.PaymentAddress,
	voterPaymentAddress privacy.PaymentAddress,
	amountOfCoin uint64,
	amountOfToken uint64,
	minerPrivateKey *privacy.SpendingKey,
) metadata.Transaction {
	tx := transaction.Tx{}
	rewardShareOldBoardMeta := metadata.NewRewardShareOldBoardMetadata(chairPaymentAddress, voterPaymentAddress, helper.GetBoardType())
	tx.InitTxSalary(amountOfCoin, &voterPaymentAddress, minerPrivateKey, chain.config.DataBase, rewardShareOldBoardMeta)
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

func (chain *BlockChain) CreateShareRewardOldBoard(
	helper ConstitutionHelper,
	chairPaymentAddress privacy.PaymentAddress,
	totalAmountCoinReward uint64,
	totalAmountTokenReward uint64,
	totalVoteAmount uint64,
	minerPrivateKey *privacy.SpendingKey,
) []metadata.Transaction {
	txs := make([]metadata.Transaction, 0)

	voterList := chain.config.DataBase.GetBoardVoterList(helper.GetBoardType(), chairPaymentAddress, chain.GetCurrentBoardIndex(helper))
	boardIndex := chain.GetCurrentBoardIndex(helper)
	for _, pubKey := range voterList {
		amountOfVote := helper.GetAmountOfVoteToBoard(chain, chairPaymentAddress, pubKey, boardIndex)
		amountOfCoin := amountOfVote * totalAmountCoinReward / totalVoteAmount
		amountOfToken := amountOfVote * totalAmountTokenReward / totalVoteAmount
		chain.CreateSingleShareRewardOldBoard(helper, chairPaymentAddress, pubKey, amountOfCoin, amountOfToken, minerPrivateKey)
	}
	return txs
}

func (chain *BlockChain) GetCoinTermReward(helper ConstitutionHelper) uint64 {
	return helper.GetBoardFund(chain) * common.PercentageBoardSalary / common.BasePercentage
}

func (self *BlockChain) CreateSendRewardOldBoard(helper ConstitutionHelper, minerPrivateKey *privacy.SpendingKey) []metadata.Transaction {
	txs := make([]metadata.Transaction, 0)
	voteTokenAmount := make(map[string]uint64)
	sumVoteTokenAmount := uint64(0)
	paymentAddresses := helper.GetCurrentBoardPaymentAddress(self)
	totalAmountOfTokenReward := helper.GetBoardSumToken(self) // Total amount of token
	totalAmountOfCoinReward := self.GetCoinTermReward(helper)
	helper.GetBoardFund(self)
	//reward for each by voteDCBList
	for _, i := range paymentAddresses {
		amount := helper.GetAmountVoteTokenOfBoard(self, i, self.GetCurrentBoardIndex(helper))
		voteTokenAmount[string(i.Bytes())] = amount
		sumVoteTokenAmount += amount
	}
	if sumVoteTokenAmount == 0 {
		sumVoteTokenAmount = 1
	}
	for payment, voteAmount := range voteTokenAmount {
		percentageReward := voteAmount * common.BasePercentage / sumVoteTokenAmount
		amountTokenReward := totalAmountOfTokenReward * uint64(percentageReward) / common.BasePercentage
		amountCoinReward := totalAmountOfCoinReward * uint64(percentageReward) / common.BasePercentage
		txs = append(txs, self.CreateShareRewardOldBoard(helper, *privacy.NewPaymentAddressFromByte([]byte(payment)), amountCoinReward, amountTokenReward, voteAmount, minerPrivateKey)...)
		//todo @0xjackalope: reward for chair
	}
	return txs
}

func (self *BlockChain) CreateUpdateNewGovernorInstruction(
	helper ConstitutionHelper,
	minerPrivateKey *privacy.SpendingKey,
) [][]string {
	instructions := make([][]string, 0)
	newBoardList, err := self.config.DataBase.GetTopMostVoteGovernor(helper.GetBoardType(), self.GetCurrentBoardIndex(helper)+1)

	if err != nil || len(newBoardList) == 0 {
		Logger.log.Error(err)
		// return empty array
		return instructions
	}

	sort.Sort(newBoardList)
	sumOfVote := uint64(0)
	var newBoardPaymentAddress []privacy.PaymentAddress
	for _, i := range newBoardList {
		newBoardPaymentAddress = append(newBoardPaymentAddress, i.PaymentAddress)
		sumOfVote += i.VoteAmount
	}
	if sumOfVote == 0 {
		sumOfVote = 1
	}

	createAcceptBoardTx := []metadata.Transaction{self.createAcceptBoardTx(
		helper.GetBoardType(),
		newBoardPaymentAddress,
		sumOfVote,
		minerPrivateKey,
	)}
	instructions = append(instructions, ListTxToListIns(createAcceptBoardTx)...)
	CreateSendVoteTokenToGovernorTx := self.CreateSendVoteTokenToGovernorTx(
		helper,
		newBoardList,
		sumOfVote,
		minerPrivateKey,
	)
	instructions = append(instructions, ListTxToListIns(CreateSendVoteTokenToGovernorTx)...)
	CreateSendBackTokenAfterVoteFailTx := self.CreateSendBackTokenAfterVoteFail(
		helper.GetBoardType(),
		newBoardPaymentAddress,
	)
	instructions = append(instructions, ListTxToListIns(CreateSendBackTokenAfterVoteFailTx)...)
	CreateSendRewardOldBoardTx := self.CreateSendRewardOldBoard(helper, minerPrivateKey)
	instructions = append(instructions, ListTxToListIns(CreateSendRewardOldBoardTx)...)

	return instructions
}

func ListTxToListIns(listTx []metadata.Transaction) [][]string {
	listIns := make([][]string, 0)
	for _, tx := range listTx {
		listIns = append(listIns, transaction.TxToIns(tx))
	}
	return listIns
}

func (chain *BlockChain) neededNewGovernor(boardType byte) bool {
	BestBlock := chain.BestState.Beacon.BestBlock
	var endGovernorBlock uint64
	if boardType == common.DCBBoard {
		endGovernorBlock = chain.BestState.Beacon.StabilityInfo.DCBGovernor.EndBlock
	} else {
		endGovernorBlock = chain.BestState.Beacon.StabilityInfo.GOVGovernor.EndBlock
	}
	currentHeight := BestBlock.Header.Height + 1
	return endGovernorBlock == currentHeight
}

func (self *BlockChain) generateVotingInstruction(minerPrivateKey *privacy.SpendingKey) ([][]string, error) {
	//todo 0xjackalope

	// 	prevBlock := blockgen.chain.BestState[shardID].BestBlock
	dcbHelper := DCBConstitutionHelper{}
	govHelper := GOVConstitutionHelper{}
	db := self.config.DataBase
	instruction := make([][]string, 0)

	//============================ VOTE PROPOSAL
	//coinbases := []metadata.Transaction{salaryTx}
	// 	// Voting transaction
	// 	// Check if it is the case we need to apply a new proposal
	// 	// 1. newNW < lastNW * 0.9
	// 	// 2. current block height == last Constitution start time + last Constitution execute duration
	if self.readyNewConstitution(dcbHelper) {
		db.SetEncryptionLastBlockHeight(
			dcbHelper.GetBoardType(),
			uint32(self.BestState.Beacon.BestBlock.Header.Height+1),
		)
		db.SetEncryptFlag(dcbHelper.GetBoardType(), uint32(common.Lv3EncryptionFlag))
		//tx, err := self.createAcceptConstitutionAndPunishTxAndRewardSubmitter(DCBConstitutionHelper{}, minerPrivateKey)
		//coinbases = append(coinbases, tx...)
		//if err != nil {
		//	Logger.log.Error(err)
		//	return nil, err
		//}
		//rewardTx, err := blockgen.createRewardProposalWinnerTx(shardID, DCBConstitutionHelper{})
		//coinbases = append(coinbases, rewardTx)
	}
	if self.readyNewConstitution(govHelper) {
		self.config.DataBase.SetEncryptionLastBlockHeight(
			govHelper.GetBoardType(),
			uint32(self.BestState.Beacon.BestBlock.Header.Height+1),
		)
		self.config.DataBase.SetEncryptFlag(govHelper.GetBoardType(), uint32(common.Lv3EncryptionFlag))
		//tx, err := self.createAcceptConstitutionAndPunishTxAndRewardSubmitter(
		//	GOVConstitutionHelper{},
		//	minerPrivateKey,
		//)
		//coinbases = append(coinbases, tx...)
		//if err != nil {
		//	Logger.log.Error(err)
		//	return nil, er:
		//}
		//rewardTx, err := blockgen.createRewardProposalWinnerTx(shardID, GOVConstitutionHelper{})
		//coinbases = append(coinbases, rewardTx)
	}

	//============================ VOTE BOARD

	if self.neededNewGovernor(common.DCBBoard) {
		instruction = append(instruction, self.CreateUpdateNewGovernorInstruction(DCBConstitutionHelper{}, minerPrivateKey)...)
	}
	if self.neededNewGovernor(common.GOVBoard) {
		instruction = append(instruction, self.CreateUpdateNewGovernorInstruction(GOVConstitutionHelper{}, minerPrivateKey)...)
	}
	return instruction, nil
}

func (self *BlockChain) readyNewConstitution(helper ConstitutionHelper) bool {
	db := self.config.DataBase
	bestBlock := self.BestState.Beacon.BestBlock
	thisBlockHeight := bestBlock.Header.Height + 1
	lastEncryptBlockHeight, _ := db.GetEncryptionLastBlockHeight(helper.GetBoardType())
	encryptFlag, _ := db.GetEncryptFlag(helper.GetBoardType())
	if uint32(thisBlockHeight) == lastEncryptBlockHeight+common.EncryptionOnePhraseDuration &&
		encryptFlag == common.NormalEncryptionFlag {
		return true
	}
	return false
}

package blockchain

import (
	"encoding/binary"
	"fmt"
	"github.com/ninjadotorg/constant/metadata/toshardins"
	"github.com/pkg/errors"
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

func getAmountOfVoteToken(sumAmount uint64, voteAmount uint64) uint64 {
	// TODO: 0xjackalop
	// not check sumAmount = 0
	return voteAmount * common.SumOfVoteDCBToken / sumAmount
}

func (self *BlockChain) createListSendInitVoteTokenTxIns(
	helper ConstitutionHelper,
	newBoardList database.CandidateList,
	sumAmountToken uint64,
) ([]toshardins.Instruction, error) {

	var SendVoteTx []toshardins.Instruction
	for i := int32(0); i < int32(newBoardList.Len()); i++ {
		newTx := toshardins.NewTxSendInitVoteTokenMetadataIns(
			helper.GetBoardType(),
			uint32(getAmountOfVoteToken(sumAmountToken, newBoardList[i].VoteAmount)),
			newBoardList[i].PaymentAddress,
		)
		SendVoteTx = append(SendVoteTx, newTx)
	}
	return SendVoteTx, nil
}

func (self *BlockChain) createAcceptBoardTxIns(
	boardType byte,
	BoardPaymentAddress []privacy.PaymentAddress,
	sumOfVote uint64,
) ([]toshardins.Instruction, error) {
	txAcceptBoardIns := toshardins.NewTxAcceptBoardIns(boardType, BoardPaymentAddress, sumOfVote)
	return []toshardins.Instruction{txAcceptBoardIns}, nil
}

func (block *BeaconBlock) UpdateDCBBoard(thisTx metadata.Transaction) error {
	// meta := thisTx.GetMetadata().(*metadata.AcceptDCBBoardMetadata)
	// block.Header.DCBGovernor.BoardIndex += 1
	// block.Header.DCBGovernor.BoardPaymentAddress = meta.BoardPaymentAddress
	// block.Header.DCBGovernor.StartedBlock = uint32(block.Header.Height)
	// block.Header.DCBGovernor.EndBlock = block.Header.DCBGovernor.StartedBlock + common.DurationOfTermDCB
	// block.Header.DCBGovernor.StartAmountToken = meta.StartAmountToken
	return nil
}

func (block *BeaconBlock) UpdateGOVBoard(thisTx metadata.Transaction) error {
	// meta := thisTx.GetMetadata().(*metadata.AcceptGOVBoardMetadata)
	// block.Header.GOVGovernor.BoardPaymentAddress = meta.BoardPaymentAddress
	// block.Header.GOVGovernor.StartedBlock = uint32(block.Header.Height)
	// block.Header.GOVGovernor.EndBlock = block.Header.GOVGovernor.StartedBlock + common.DurationOfTermGOV
	// block.Header.GOVGovernor.StartAmountToken = meta.StartAmountToken
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

func createSendBackTokenVoteFailIns(
	boardType byte,
	paymentAddress privacy.PaymentAddress,
	amount uint64,
) toshardins.Instruction {
	var propertyID common.Hash
	if boardType == common.DCBBoard {
		propertyID = common.DCBTokenID
	} else {
		propertyID = common.GOVTokenID
	}
	return toshardins.NewSendBackTokenVoteFailIns(
		paymentAddress,
		amount,
		propertyID,
	)
}

func (self *BlockChain) createSendBackTokenAfterVoteFailIns(
	boardType byte,
	newDCBList []privacy.PaymentAddress,
	shardID byte,
) ([]toshardins.Instruction, error) {
	var propertyID common.Hash
	if boardType == common.DCBBoard {
		propertyID = common.DCBTokenID
	} else {
		propertyID = common.GOVTokenID
	}
	setOfNewDCB := make(map[string]bool, 0)
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
	listNewIns := make([]toshardins.Instruction, 0)
	for iter.Next() {
		key := iter.Key()
		_, boardIndex, candidatePubKey, voterPaymentAddress, _ := lvdb.ParseKeyVoteBoardList(key)
		value := iter.Value()
		amountOfDCBToken := lvdb.ParseValueVoteBoardList(value)

		_, found := setOfNewDCB[string(candidatePubKey)]
		if boardIndex < uint32(currentBoardIndex) || !found {
			listNewIns = append(
				listNewIns,
				toshardins.NewSendBackTokenVoteFailIns(
					*voterPaymentAddress,
					amountOfDCBToken,
					propertyID,
				),
			)
		}
	}
	return listNewIns, nil
}

func GetOracleDCBNationalWelfare() int32 {
	fmt.Print("Get national welfare. It is constant now. Need to change !!!")
	return 1234
}
func GetOracleGOVNationalWelfare() int32 {
	fmt.Print("Get national welfare. It is constant now. Need to change !!!")
	return 1234
}

func (chain *BlockChain) CreateSingleShareRewardOldBoardIns(
	helper ConstitutionHelper,
	chairPaymentAddress privacy.PaymentAddress,
	voterPaymentAddress privacy.PaymentAddress,
	amountOfCoin uint64,
	amountOfToken uint64,
	minerPrivateKey *privacy.SpendingKey,
) toshardins.Instruction {
	return toshardins.NewShareRewardOldBoardMetadataIns(
		chairPaymentAddress, voterPaymentAddress, helper.GetBoardType(), amountOfCoin, amountOfToken,
	)
}

func (chain *BlockChain) CreateShareRewardOldBoardIns(
	helper ConstitutionHelper,
	chairPaymentAddress privacy.PaymentAddress,
	totalAmountCoinReward uint64,
	totalAmountTokenReward uint64,
	totalVoteAmount uint64,
	minerPrivateKey *privacy.SpendingKey,
) []toshardins.Instruction {
	Ins := make([]toshardins.Instruction, 0)

	voterList := chain.config.DataBase.GetBoardVoterList(helper.GetBoardType(), chairPaymentAddress, chain.GetCurrentBoardIndex(helper))
	boardIndex := chain.GetCurrentBoardIndex(helper)
	for _, pubKey := range voterList {
		amountOfVote := helper.GetAmountOfVoteToBoard(chain, chairPaymentAddress, pubKey, boardIndex)
		amountOfCoin := amountOfVote * totalAmountCoinReward / totalVoteAmount
		amountOfToken := amountOfVote * totalAmountTokenReward / totalVoteAmount
		Ins = append(Ins, chain.CreateSingleShareRewardOldBoardIns(
			helper,
			chairPaymentAddress,
			pubKey,
			amountOfCoin,
			amountOfToken,
			minerPrivateKey,
		))
	}
	return Ins
}

func (chain *BlockChain) GetCoinTermReward(helper ConstitutionHelper) uint64 {
	return helper.GetBoardFund(chain) * common.PercentageBoardSalary / common.BasePercentage
}

func (self *BlockChain) createSendRewardOldBoardIns(
	helper ConstitutionHelper,
	minerPrivateKey *privacy.SpendingKey,
	shardID byte,
) ([]toshardins.Instruction, error) {
	Ins := make([]toshardins.Instruction, 0)
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
		Ins = append(Ins, self.CreateShareRewardOldBoardIns(helper, *privacy.NewPaymentAddressFromByte([]byte(payment)), amountCoinReward, amountTokenReward, voteAmount, minerPrivateKey)...)
		//todo @0xjackalope: reward for chair
	}
	return Ins, nil
}

//todo @0xjackalope reward for chair
func (self *BlockChain) CreateUpdateNewGovernorInstruction(
	helper ConstitutionHelper,
	minerPrivateKey *privacy.SpendingKey,
	shardID byte,
) ([]toshardins.Instruction, error) {
	instructions := make([]toshardins.Instruction, 0)
	newBoardList, err := self.config.DataBase.GetTopMostVoteGovernor(helper.GetBoardType(), self.GetCurrentBoardIndex(helper)+1)

	if err != nil {
		return nil, err
	}
	if len(newBoardList) == 0 {
		return nil, errors.New("not enough candidate")
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

	acceptBoardIns, err := self.createAcceptBoardTxIns(
		helper.GetBoardType(),
		newBoardPaymentAddress,
		sumOfVote,
	)
	if err != nil {
		return nil, err
	}
	instructions = append(instructions, acceptBoardIns...)

	sendVoteTokenToGovernorIns, err := self.createListSendInitVoteTokenTxIns(
		helper,
		newBoardList,
		sumOfVote,
	)
	if err != nil {
		return nil, err
	}
	instructions = append(instructions, sendVoteTokenToGovernorIns...)

	sendBackTokenAfterVoteFailIns, err := self.createSendBackTokenAfterVoteFailIns(
		helper.GetBoardType(),
		newBoardPaymentAddress,
		shardID,
	)
	if err != nil {
		return nil, err
	}
	instructions = append(instructions, sendBackTokenAfterVoteFailIns...)

	sendRewardOldBoardIns, err := self.createSendRewardOldBoardIns(helper, minerPrivateKey, shardID)
	if err != nil {
		return nil, err
	}
	instructions = append(instructions, sendRewardOldBoardIns...)

	return instructions, nil
}

func TxToIns(tx metadata.Transaction, chain *BlockChain, shardID byte) ([][]string, error) {
	meta := tx.GetMetadata()
	action, err := meta.BuildReqActions(tx, chain, shardID)
	if err != nil {
		return nil, err
	}
	return action, nil
}

func ListTxToListIns(listTx []metadata.Transaction, chain *BlockChain, shardID byte) ([][]string, error) {
	listIns := make([][]string, 0)
	for _, tx := range listTx {
		meta := tx.GetMetadata()
		action, err := meta.BuildReqActions(tx, chain, shardID)
		if err != nil {
			return nil, err
		}
		listIns = append(listIns, action...)
	}
	return listIns, nil
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

func (self *BlockChain) generateVotingInstruction(minerPrivateKey *privacy.SpendingKey, shardID byte) ([][]string, error) {
	//todo 0xjackalope

	// 	prevBlock := blockgen.chain.BestState[shardID].BestBlock
	dcbHelper := DCBConstitutionHelper{}
	govHelper := GOVConstitutionHelper{}
	db := self.config.DataBase
	instruction := make([]toshardins.Instruction, 0)

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
		updateGovernorInstruction, err := self.CreateUpdateNewGovernorInstruction(DCBConstitutionHelper{}, minerPrivateKey, shardID)
		if err != nil {
			return nil, err
		}
		instruction = append(instruction, updateGovernorInstruction...)
	}
	if self.neededNewGovernor(common.GOVBoard) {
		updateGovernorInstruction, err := self.CreateUpdateNewGovernorInstruction(GOVConstitutionHelper{}, minerPrivateKey, shardID)
		if err != nil {
			return nil, err
		}
		instruction = append(instruction, updateGovernorInstruction...)
	}
	instructionString := make([][]string, 0)
	for _, i := range instruction {
		instructionString = append(instructionString, i.GetStringFormat())
	}
	return instructionString, nil
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

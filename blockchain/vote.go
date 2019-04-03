package blockchain

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database/lvdb"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/metadata/frombeaconins"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type ConstitutionHelper interface {
	CheckSubmitProposalType(tx metadata.Transaction) bool
	NewAcceptProposalIns(txId *common.Hash, voter []privacy.PaymentAddress, shardID byte) frombeaconins.InstructionFromBeacon
	GetBoardType() common.BoardType
	GetConstitutionEndedBlockHeight(chain *BlockChain) uint64
	NewRewardProposalSubmitterIns(blockgen *BlockChain, receiverAddress *privacy.PaymentAddress) (instruction frombeaconins.InstructionFromBeacon, err error)
	GetPaymentAddressFromSubmitProposalMetadata(tx metadata.Transaction) *privacy.PaymentAddress
	GetPaymentAddressVoter(chain *BlockChain) ([]privacy.PaymentAddress, error)
	GetPrizeProposal() uint32
	GetCurrentBoardPaymentAddress(chain *BlockChain) []privacy.PaymentAddress
	GetBoardSumToken(chain *BlockChain) uint64
	GetBoardFund(chain *BlockChain) uint64
	GetTokenID() *common.Hash
	GetAmountOfVoteToBoard(chain *BlockChain, candidatePaymentAddress privacy.PaymentAddress, voterPaymentAddress privacy.PaymentAddress, boardIndex uint32) uint64
	GetBoard(chain *BlockChain) metadata.GovernorInterface
	GetConstitutionInfo(chain *BlockChain) ConstitutionInfo
	GetCurrentNationalWelfare(chain *BlockChain) int32
	GetThresholdRatioOfCrisis() int32
	GetOldNationalWelfare(chain *BlockChain) int32
	GetSubmitProposalInfo(tx metadata.Transaction) (*component.SubmitProposalInfo, error)
	GetProposalTxID(tx metadata.Transaction) (hash *common.Hash)
	SetNewConstitution(bc *BlockChain, constitutionInfo *ConstitutionInfo, welfare int32, submitProposalTx metadata.Transaction)
}

// func (chain *BlockChain) createRewardProposalWinnerIns(
// 	constitutionHelper ConstitutionHelper,
// ) ([]frombeaconins.InstructionFromBeacon, error) {
// 	paymentAddresses, err := constitutionHelper.GetPaymentAddressVoter(chain)
// 	var resIns frombeaconins
// 	if err != nil {
// 		return nil, err
// 	}
// 	prize := constitutionHelper.GetPrizeProposal()
// 	for _, paymentAddress := range paymentAddresses {
// 		ins := frombeaconins.NewRewardProposalWinnerIns(paymentAddresses, prize)
// 	}
// 	return ins, nil
// }

func (self *BlockChain) BuildVoteTableAndPunishTransaction(
	helper ConstitutionHelper,
	nextConstitutionIndex uint32,
) (
	VoteTable map[common.Hash][]privacy.PaymentAddress,
	CountVote map[common.Hash]uint32,
	err error,
) {
	VoteTable = make(map[common.Hash][]privacy.PaymentAddress)
	CountVote = make(map[common.Hash]uint32)

	db := self.config.DataBase
	gg := lvdb.ViewDBByPrefix(db, lvdb.VoteProposalPrefix)
	_ = gg
	boardType := helper.GetBoardType()
	begin := lvdb.GetKeyVoteProposal(boardType, 0, nil)
	// +1 to search in that range
	end := lvdb.GetKeyVoteProposal(boardType, 1+nextConstitutionIndex, nil)

	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}

	iter := db.NewIterator(&searchRange, nil)
	rightIndex := self.GetConstitutionIndex(helper) + 1
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		_, constitutionIndex, voterPayment, err := lvdb.ParseKeyVoteProposal(key)
		if err != nil {
			return nil, nil, err
		}
		if constitutionIndex != uint32(rightIndex) {
			db.Delete(key)
			continue
		}

		//Accumulate count vote
		proposalTxID, err := lvdb.ParseValueVoteProposal(value)
		if err != nil {
			return nil, nil, err
		}

		if VoteTable[*proposalTxID] == nil {
			VoteTable[*proposalTxID] = make([]privacy.PaymentAddress, 0)
		}
		VoteTable[*proposalTxID] = append(VoteTable[*proposalTxID], *voterPayment)
		CountVote[*proposalTxID] += 1
	}
	return
}

func (self *BlockChain) createAcceptConstitutionAndRewardSubmitter(
	helper ConstitutionHelper,
) ([]frombeaconins.InstructionFromBeacon, error) {
	nextConstitutionIndex := self.GetConstitutionIndex(DCBConstitutionHelper{}) + 1
	resIns := make([]frombeaconins.InstructionFromBeacon, 0)
	VoteTable, CountVote, err := self.BuildVoteTableAndPunishTransaction(helper, nextConstitutionIndex)
	if err != nil {
		return nil, err
	}
	bestProposal := metadata.ProposalVote{
		TxId:         common.Hash{},
		NumberOfVote: 0,
	}
	db := self.config.DataBase
	for txId, _ := range VoteTable {
		if CountVote[txId] > bestProposal.NumberOfVote {
			bestProposal.TxId = txId
			bestProposal.NumberOfVote = CountVote[txId]
		}
	}
	// panic("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n\n\n\n\n\n")
	if bestProposal.NumberOfVote == 0 {
		return resIns, nil
	}
	byteTemp, err0 := db.GetSubmitProposalDB(helper.GetBoardType(), helper.GetConstitutionInfo(self).ConstitutionIndex+1, bestProposal.TxId.GetBytes())
	gg := lvdb.ViewDBByPrefix(db, lvdb.SubmitProposalPrefix)
	_ = gg
	if err0 != nil {
		return resIns, nil
	}
	submitterPaymentAddress := privacy.NewPaymentAddressFromByte(byteTemp)
	if submitterPaymentAddress != nil {
		rewardForProposalSubmitterIns, err := helper.NewRewardProposalSubmitterIns(self, submitterPaymentAddress)
		if err != nil {
			return nil, err
		}
		resIns = append(resIns, rewardForProposalSubmitterIns)
	}
	shardID := frombeaconins.GetShardIDFromPaymentAddressBytes(*submitterPaymentAddress)
	acceptedProposalIns := helper.NewAcceptProposalIns(&bestProposal.TxId, VoteTable[bestProposal.TxId], shardID)
	resIns = append(resIns, acceptedProposalIns)
	boardType := helper.GetBoardType()
	boardIndex := helper.GetBoard(self).GetBoardIndex()
	totalReward := uint64(helper.GetCurrentNationalWelfare(self)) * BaseSalaryBoard
	listVotersOfCurrentProposal, err := db.GetCurrentProposalWinningVoter(boardType, helper.GetConstitutionInfo(self).ConstitutionIndex)
	if err == nil {
		voterAndSupporters := make([][]privacy.PaymentAddress, len(listVotersOfCurrentProposal))
		voterAndAmount := make([]uint64, len(listVotersOfCurrentProposal))
		for i, voter := range listVotersOfCurrentProposal {
			listSupporters := self.config.DataBase.GetBoardVoterList(boardType, voter, boardIndex)
			voterAndSupporters[i] = listSupporters
			voterAndAmount[i] = 0
			for _, supporter := range listSupporters {
				voterAndAmount[i] += helper.GetAmountOfVoteToBoard(self, voter, supporter, boardIndex)
			}
			shareRewardIns := self.CreateShareRewardOldBoardIns(helper, listVotersOfCurrentProposal[i], totalReward, voterAndAmount[i])
			resIns = append(resIns, shareRewardIns...)
		}
	}
	return resIns, nil
}

func (self *BlockChain) createAcceptBoardIns(
	boardType common.BoardType,
	BoardPaymentAddress []privacy.PaymentAddress,
	sumOfVote uint64,
) ([]frombeaconins.InstructionFromBeacon, error) {
	acceptBoardIns := frombeaconins.NewAcceptBoardIns(boardType, BoardPaymentAddress, sumOfVote)
	inst, _ := acceptBoardIns.GetStringFormat()
	fmt.Println("[voting] -  SendBackToken to vote failed acceptBoardIns", inst)
	return []frombeaconins.InstructionFromBeacon{acceptBoardIns}, nil
}

func (stateBeacon *BestStateBeacon) UpdateDCBBoard(ins frombeaconins.AcceptDCBBoardIns) error {
	stateBeacon.StabilityInfo.DCBGovernor.BoardIndex += 1
	stateBeacon.StabilityInfo.DCBGovernor.BoardPaymentAddress = ins.BoardPaymentAddress
	stateBeacon.StabilityInfo.DCBGovernor.StartedBlock = stateBeacon.BestBlock.Header.Height
	stateBeacon.StabilityInfo.DCBGovernor.EndBlock = stateBeacon.StabilityInfo.DCBGovernor.StartedBlock + common.DurationOfDCBBoard
	Logger.log.Info("New DCBGovernor.EndBlock is: ", stateBeacon.StabilityInfo.DCBGovernor.EndBlock, "\n")
	stateBeacon.StabilityInfo.DCBGovernor.StartAmountToken = ins.StartAmountToken
	return nil
}

func (stateBeacon *BestStateBeacon) UpdateGOVBoard(ins frombeaconins.AcceptGOVBoardIns) error {
	stateBeacon.StabilityInfo.GOVGovernor.BoardIndex += 1
	stateBeacon.StabilityInfo.GOVGovernor.BoardPaymentAddress = ins.BoardPaymentAddress
	stateBeacon.StabilityInfo.GOVGovernor.StartedBlock = stateBeacon.BestBlock.Header.Height
	stateBeacon.StabilityInfo.GOVGovernor.EndBlock = stateBeacon.StabilityInfo.GOVGovernor.StartedBlock + common.DurationOfGOVBoard
	Logger.log.Info("New DCBGovernor.EndBlock is: ", stateBeacon.StabilityInfo.GOVGovernor.EndBlock, "\n")
	stateBeacon.StabilityInfo.GOVGovernor.StartAmountToken = ins.StartAmountToken
	return nil
}

//????
func (blockchain *BlockChain) UpdateDCBFund(tx metadata.Transaction) {
	blockchain.BestState.Beacon.StabilityInfo.BankFund -= common.RewardProposalSubmitter
}

func (blockchain *BlockChain) UpdateGOVFund(tx metadata.Transaction) {
	blockchain.BestState.Beacon.StabilityInfo.BankFund -= common.RewardProposalSubmitter
}

func createSendBackTokenVoteFailIns(
	boardType common.BoardType,
	paymentAddress privacy.PaymentAddress,
	amount uint64,
) frombeaconins.InstructionFromBeacon {
	var propertyID common.Hash
	if boardType == common.DCBBoard {
		propertyID = common.DCBTokenID
	} else {
		propertyID = common.GOVTokenID
	}
	return frombeaconins.NewSendBackTokenVoteFailIns(
		paymentAddress,
		amount,
		propertyID,
	)
}

func (self *BlockChain) createSendBackTokenAfterVoteFailIns(
	boardType common.BoardType,
	newDCBList []privacy.PaymentAddress,
) ([]frombeaconins.InstructionFromBeacon, error) {
	fmt.Println("[voting]- Enter createSendBackTokenAfterVoteFailIns")
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
	listNewIns := make([]frombeaconins.InstructionFromBeacon, 0)
	for iter.Next() {
		key := iter.Key()
		_, boardIndex, candidatePubKey, voterPaymentAddress, _ := lvdb.ParseKeyVoteBoardList(key)
		value := iter.Value()
		amountOfDCBToken := lvdb.ParseValueVoteBoardList(value)
		_, found := setOfNewDCB[string(candidatePubKey)]
		if boardIndex < uint32(currentBoardIndex) || !found {
			inst, _ := frombeaconins.NewSendBackTokenVoteFailIns(
				*voterPaymentAddress,
				amountOfDCBToken,
				propertyID,
			).GetStringFormat()
			listNewIns = append(
				listNewIns,
				frombeaconins.NewSendBackTokenVoteFailIns(
					*voterPaymentAddress,
					amountOfDCBToken,
					propertyID,
				),
			)
			fmt.Println("[voting]-SendBackIns: ", inst)
		}
		err := self.config.DataBase.Delete(key)
		if err != nil {
			return nil, err
		}
	}
	fmt.Println("[voting]- createSendBackTokenAfterVoteFailIns ok")
	return listNewIns, nil
}

func GetOracleDCBNationalWelfare() int32 {
	fmt.Print("Get national welfare. It is constant now. Need to change !!!\n")
	return 1234
}
func GetOracleGOVNationalWelfare() int32 {
	fmt.Print("Get national welfare. It is constant now. Need to change !!!\n")
	return 1234
}

func (chain *BlockChain) CreateSingleShareRewardOldBoardIns(
	helper ConstitutionHelper,
	chairPaymentAddress privacy.PaymentAddress,
	voterPaymentAddress privacy.PaymentAddress,
	amountOfCoin uint64,
) frombeaconins.InstructionFromBeacon {
	return frombeaconins.NewShareRewardOldBoardMetadataIns(
		chairPaymentAddress, voterPaymentAddress, helper.GetBoardType(), amountOfCoin,
	)
}

func (chain *BlockChain) CreateShareRewardOldBoardIns(
	helper ConstitutionHelper,
	chairPaymentAddress privacy.PaymentAddress,
	totalAmountCoinReward uint64,
	totalVoteAmount uint64,
) []frombeaconins.InstructionFromBeacon {
	Ins := make([]frombeaconins.InstructionFromBeacon, 0)
	boardIndex := chain.GetCurrentBoardIndex(helper)
	voterList := chain.config.DataBase.GetBoardVoterList(helper.GetBoardType(), chairPaymentAddress, boardIndex)
	for _, voter := range voterList {
		amountOfVote := helper.GetAmountOfVoteToBoard(chain, chairPaymentAddress, voter, boardIndex)
		amountOfCoin := amountOfVote * totalAmountCoinReward / totalVoteAmount
		Ins = append(Ins, chain.CreateSingleShareRewardOldBoardIns(
			helper,
			chairPaymentAddress,
			voter,
			amountOfCoin,
		))
	}
	return Ins
}

func (chain *BlockChain) GetCoinTermReward(helper ConstitutionHelper) uint64 {
	return helper.GetBoardFund(chain) * common.PercentageBoardSalary / common.BasePercentage
}

//todo @0xjackalope reward for chair
func (self *BlockChain) CreateUpdateNewGovernorInstruction(
	helper ConstitutionHelper,
) ([]frombeaconins.InstructionFromBeacon, error) {
	instructions := make([]frombeaconins.InstructionFromBeacon, 0)
	newBoardList, err := self.config.DataBase.GetTopMostVoteGovernor(helper.GetBoardType(), self.GetCurrentBoardIndex(helper)+1)
	fmt.Println("[voting] - SendBackToken to vote failed Enter function")
	if err != nil {
		fmt.Println("[voting] - Error 1", err)
		if reflect.TypeOf(err).String() == "*database.DatabaseError" {
			return nil, nil
		}
		return nil, err
	}
	if len(newBoardList) == 0 {
		fmt.Println("[voting] - not enough candidate")
		return nil, errors.New("not enough candidate")
	}
	fmt.Println("[voting] -  SendBackToken to vote failed Get top most vote governor ok")
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

	acceptBoardIns, err := self.createAcceptBoardIns(
		helper.GetBoardType(),
		newBoardPaymentAddress,
		sumOfVote,
	)
	if err != nil {
		return nil, err
	}
	fmt.Println("[voting]-acceptBoardIns ok")
	instructions = append(instructions, acceptBoardIns...)

	sendBackTokenAfterVoteFailIns, err := self.createSendBackTokenAfterVoteFailIns(
		helper.GetBoardType(),
		newBoardPaymentAddress,
	)
	if err != nil {
		fmt.Println("[voting]- err", err)
		return nil, err
	}
	instructions = append(instructions, sendBackTokenAfterVoteFailIns...)
	fmt.Println("[voting]-Update new governor inst ok", sendBackTokenAfterVoteFailIns)
	// hyyyyyyyyyyyyyyyyyyyyy
	// send back dcbtoken after board
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

func (chain *BlockChain) neededNewGovernor(helper ConstitutionHelper) bool {
	constitutionIndex := chain.GetConstitutionIndex(helper)
	return constitutionIndex == ConstitutionPerBoard
}

func (chain *BlockChain) neededFirstNewGovernor(helper ConstitutionHelper) bool {
	fmt.Println("[voting] - chain BeaconHeight of BestState", chain.BestState.Beacon.BeaconHeight)
	if EndOfFirstBoard == chain.BestState.Beacon.BeaconHeight {
		fmt.Println("[voting] -  EndOfFirstBoard vs BeaconHeight", EndOfFirstBoard, chain.BestState.Beacon.BeaconHeight)
		return true
	}
	return false
}

func (chain *BlockChain) neededNewConstitution(helper ConstitutionHelper) bool {
	// todo: hyyyyyyyyyyyy
	endBlock := helper.GetConstitutionEndedBlockHeight(chain)
	fmt.Println("[voting] - neededNewConstitution: ", endBlock, chain.BestState.Beacon.BeaconHeight)
	if chain.BestState.Beacon.BeaconHeight >= endBlock {
		return true
	}
	return false
}

func (self *BlockChain) generateVotingInstructionWOIns(helper ConstitutionHelper) ([][]string, error) {
	// panic("[voting] aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	// 	prevBlock := blockgen.chain.BestState[shardID].BestBlock
	instructions := make([]frombeaconins.InstructionFromBeacon, 0)
	fmt.Println("[voting]-Enter generateVotingInstructionWOIns")
	if self.neededFirstNewGovernor(helper) {
		fmt.Println("[voting]-[neededNewGovernor]-Create first instruction")
		updateGovernorInstruction, err := self.CreateUpdateNewGovernorInstruction(helper)
		if err != nil {
			fmt.Println("[voting] - error", err)
			return nil, err
		}
		for _, inst := range updateGovernorInstruction {
			instString, _ := inst.GetStringFormat()
			fmt.Println("[voting]-[neededNewGovernor] - Created ", instString)
		}
		instructions = append(instructions, updateGovernorInstruction...)
	}

	if self.neededNewConstitution(helper) {
		//============================ VOTE PROPOSAL
		// 	// Voting transaction
		// 	// Check if it is the case we need to apply a new proposal
		// 	// 1. newNW < lastNW * 0.9
		// 	// 2. current block height == last Constitution start time + last Constitution execute duration

		// //Hyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy
		// // step 2 Hyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy
		updateProposalInstruction, err := self.createAcceptConstitutionAndRewardSubmitter(helper)
		if err != nil {
			return nil, err
		}
		fmt.Println("[voting] - updateProposalInstruction ok: ", updateProposalInstruction)
		instructions = append(instructions, updateProposalInstruction...)

		//============================ VOTE BOARD
		if self.neededNewGovernor(helper) {
			updateGovernorInstruction, err := self.CreateUpdateNewGovernorInstruction(helper)
			if err != nil {
				return nil, err
			}
			instructions = append(instructions, updateGovernorInstruction...)
		}
	}

	instructionsString := make([][]string, 0)
	for _, instruction := range instructions {
		newIns, err := instruction.GetStringFormat()
		if err != nil {
			return nil, err
		}
		fmt.Println("[voting] - ", newIns)
		instructionsString = append(instructionsString, newIns)
	}
	return instructionsString, nil
}

//func (self *BlockChain) readyNewConstitution(helper ConstitutionHelper) bool {
//	db := self.config.DataBase
//	bestBlock := self.BestState.Beacon.BestBlock
//	thisBlockHeight := bestBlock.Header.Height + 1
//	lastEncryptBlockHeight, _ := db.GetEncryptionLastBlockHeight(helper.GetBoardType())
//	encryptFlag, _ := db.GetEncryptFlag(helper.GetBoardType())
//	if thisBlockHeight == lastEncryptBlockHeight+common.EncryptionOnePhraseDuration &&
//		encryptFlag == common.NormalEncryptionFlag {
//		return true
//	}
//	return false
//}

func (self *BlockChain) GetListVoterOfProposalDB(helper ConstitutionHelper, proposalTxID []byte) ([]privacy.PaymentAddress, error) {
	var res []privacy.PaymentAddress
	// currentBoardIndex := self.GetCurrentBoardIndex(helper)
	boardType := helper.GetBoardType()
	begin := lvdb.GetKeyListVoterOfProposal(boardType, helper.GetConstitutionInfo(self).ConstitutionIndex, proposalTxID, nil)
	end := lvdb.GetKeyListVoterOfProposal(boardType, helper.GetConstitutionInfo(self).ConstitutionIndex, common.BytesPlusOne(proposalTxID), nil)
	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}
	iter := self.config.DataBase.NewIterator(&searchRange, nil)
	for iter.Next() {
		key := iter.Key()
		_, _, _, voterPaymentAddress, err := lvdb.ParseKeyListVoterOfProposal(key)
		if err != nil {
			return nil, err
		}
		res = append(res, *voterPaymentAddress)
		err = self.config.DataBase.Delete(key)
		if err != nil {
			return nil, err
		}
	}
	//use prefix  as in file lvdb/block.go FetchChain
	return nil, nil
}

func (bc *BlockChain) GetGovernor(boardType common.BoardType) metadata.GovernorInterface {
	if boardType == common.DCBBoard {
		return bc.BestState.Beacon.StabilityInfo.DCBGovernor
	} else {
		return bc.BestState.Beacon.StabilityInfo.GOVGovernor
	}
}

func (bc *BlockChain) GetConstitution(boardType common.BoardType) metadata.ConstitutionInterface {
	if boardType == common.DCBBoard {
		return bc.BestState.Beacon.StabilityInfo.DCBConstitution
	} else {
		return bc.BestState.Beacon.StabilityInfo.GOVConstitution
	}
}

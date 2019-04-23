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
	NewKeepOldProposalIns() frombeaconins.InstructionFromBeacon
	GetBoardType() common.BoardType
	GetConstitutionEndedBlockHeight(chain *BlockChain) uint64
	NewRewardProposalSubmitterIns(blockgen *BlockChain, receiverAddress *privacy.PaymentAddress, amount uint64) (instruction frombeaconins.InstructionFromBeacon, err error)
	GetPaymentAddressFromSubmitProposalMetadata(tx metadata.Transaction) *privacy.PaymentAddress
	GetPaymentAddressVoter(chain *BlockChain) ([]privacy.PaymentAddress, error)
	GetPrizeProposal() uint32
	GetCurrentBoardPaymentAddress(chain *BlockChain) []privacy.PaymentAddress
	GetBoardSumToken(chain *BlockChain) uint64
	GetBoardFund(chain *BlockChain) uint64
	GetBoardReward(chain *BlockChain) uint64
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
	fmt.Println("[ndh] [--- Build vote table ---]")
	VoteTable = make(map[common.Hash][]privacy.PaymentAddress)
	CountVote = make(map[common.Hash]uint32)

	db := self.config.DataBase
	gg := lvdb.ViewDBByPrefix(db, lvdb.VoteProposalPrefix)
	_ = gg
	boardType := helper.GetBoardType()
	begin := lvdb.GetKeyVoteProposal(boardType, nextConstitutionIndex, nil)
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
		fmt.Printf("[ndh] ==> constitution Index: %+v; voterPayment: %+v; rightIndex: %+v\n", constitutionIndex, voterPayment, rightIndex)
		if err != nil {
			return nil, nil, err
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
	fmt.Println("[ndh] [+++ Build vote table +++]")
	return
}

func (self *BlockChain) createAcceptConstitutionAndRewardSubmitter(
	helper ConstitutionHelper,
) ([]frombeaconins.InstructionFromBeacon, error) {
	currentConstitutionIndex := helper.GetConstitutionInfo(self).ConstitutionIndex
	nextConstitutionIndex := currentConstitutionIndex + 1
	fmt.Println("[ndh] - create accept constitution nextConstitutionIndex: ", nextConstitutionIndex)
	resIns := make([]frombeaconins.InstructionFromBeacon, 0)
	VoteTable, CountVote, err := self.BuildVoteTableAndPunishTransaction(helper, nextConstitutionIndex)
	if err != nil {
		fmt.Printf("[ndh] - error here %+v\n", err)
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
	if (bestProposal.NumberOfVote == 0) && (helper.GetConstitutionInfo(self).ConstitutionIndex == 0) {
		fmt.Printf("[ndh] - Number of vote is zero \n")
		return resIns, nil
	}
	if bestProposal.NumberOfVote == 0 {
		currentProposalTxID, err := db.GetProposalTXIDByConstitutionIndex(helper.GetBoardType(), currentConstitutionIndex)
		if err != nil {
			fmt.Printf("[ndh] - - can not get proposal txId by constitution index! Error: %+v\n", err.Error())
			return resIns, err
		}
		submitterOfCurrentProposal, err := db.GetSubmitProposalDB(helper.GetBoardType(), currentConstitutionIndex, currentProposalTxID)
		if err != nil {
			fmt.Printf("[ndh] - - can not get submitter of current proposal! Error: %+v\n", err.Error())
			return resIns, err
		}
		err = db.AddSubmitProposalDB(helper.GetBoardType(), currentConstitutionIndex+1, currentProposalTxID, submitterOfCurrentProposal)
		if err != nil {
			fmt.Printf("[ndh] - - can not add submit proposal! Error: %+v\n", err.Error())
			return resIns, err
		}
		resIns = append(resIns, helper.NewKeepOldProposalIns())
	} else {
		err = db.DeleteAnyProposalButThisDB(helper.GetBoardType(), nextConstitutionIndex, bestProposal.TxId.GetBytes())
		byteTemp, err0 := db.GetProposalSubmitterByConstitutionIndexDB(helper.GetBoardType(), nextConstitutionIndex)
		if err0 != nil {
			return resIns, nil
		}
		submitterPaymentAddress := privacy.NewPaymentAddressFromByte(byteTemp)
		shardID := frombeaconins.GetShardIDFromPaymentAddressBytes(*submitterPaymentAddress)
		acceptedProposalIns := helper.NewAcceptProposalIns(&bestProposal.TxId, VoteTable[bestProposal.TxId], shardID)
		resIns = append(resIns, acceptedProposalIns)
	}
	totalReward := helper.GetBoardReward(self)
	fmt.Printf("[ndh] - totalReward: %+v\n", totalReward)
	if (totalReward > 0) && (currentConstitutionIndex > 0) {
		boardType := helper.GetBoardType()
		boardIndex := helper.GetBoard(self).GetBoardIndex()
		listVotersOfCurrentProposal, err := db.GetCurrentProposalWinningVoter(boardType, currentConstitutionIndex)
		// }
		if err == nil {
			byteTemp, err0 := db.GetProposalSubmitterByConstitutionIndexDB(helper.GetBoardType(), currentConstitutionIndex)
			if (err0 != nil) && (currentConstitutionIndex > 0) {
				return resIns, nil
			}
			numberOfReceiver := len(listVotersOfCurrentProposal) + 1
			submitterRewardAmount := common.BoardRewardPercent * (totalReward / uint64(numberOfReceiver)) / 100
			supporterRewardAmount := (totalReward / uint64(numberOfReceiver)) - submitterRewardAmount
			submitterPaymentAddress := privacy.NewPaymentAddressFromByte(byteTemp)
			if submitterPaymentAddress != nil {
				rewardForProposalSubmitterIns, err1 := helper.NewRewardProposalSubmitterIns(self, submitterPaymentAddress, submitterRewardAmount)
				stringTest, _ := rewardForProposalSubmitterIns.GetStringFormat()
				fmt.Printf("[ndh] - reward for submitter instruction: %+v\n", stringTest)
				if err1 == nil {
					resIns = append(resIns, rewardForProposalSubmitterIns)
				}
			}
			voterAndAmount := make([]uint64, len(listVotersOfCurrentProposal))
			for i, voter := range listVotersOfCurrentProposal {
				listSupporters := self.config.DataBase.GetBoardVoterList(boardType, voter, boardIndex)
				voterAndAmount[i] = 0
				for _, supporter := range listSupporters {
					fmt.Printf("[ndh] Voter: %+v; Supporter: %+v\n", voter.Bytes(), supporter.Bytes())
					voterAndAmount[i] += helper.GetAmountOfVoteToBoard(self, voter, supporter, boardIndex)
				}
				shareRewardIns := self.CreateShareRewardOldBoardIns(helper, listVotersOfCurrentProposal[i], supporterRewardAmount, voterAndAmount[i])
				resIns = append(resIns, shareRewardIns...)
				rewardForVoter := frombeaconins.NewRewardProposalVoterIns(&voter, submitterRewardAmount, helper.GetBoardType())
				stringTest, _ := rewardForVoter.GetStringFormat()
				fmt.Printf("[ndh] - reward for voter instruction: %+v\n", stringTest)
				if rewardForVoter != nil {
					resIns = append(resIns, rewardForVoter)
				}
			}
		}
		fmt.Println("[ndh] - - - - Fund Update - ", helper.GetBoardFund(self))
	}
	fmt.Println("[ndh] - - - - - - end createAcceptConstitutionAndRewardSubmitter", resIns)
	return resIns, nil
}

func (self *BlockChain) createAcceptBoardIns(
	boardType common.BoardType,
	BoardPaymentAddress []privacy.PaymentAddress,
	sumOfVote uint64,
) ([]frombeaconins.InstructionFromBeacon, error) {
	acceptBoardIns := frombeaconins.NewAcceptBoardIns(boardType, BoardPaymentAddress, sumOfVote)
	inst, _ := acceptBoardIns.GetStringFormat()
	fmt.Println("[ndh] -  SendBackToken to vote failed acceptBoardIns", inst)
	if len(inst) != 0 {
		return []frombeaconins.InstructionFromBeacon{acceptBoardIns}, nil
	} else {
		return nil, nil
	}
}

func (stateBeacon *BestStateBeacon) UpdateDCBBoard(ins frombeaconins.AcceptDCBBoardIns) error {
	stateBeacon.StabilityInfo.DCBGovernor.BoardIndex += 1
	stateBeacon.StabilityInfo.DCBGovernor.BoardPaymentAddress = ins.BoardPaymentAddress
	stateBeacon.StabilityInfo.DCBGovernor.StartedBlock = stateBeacon.BestBlock.Header.Height
	stateBeacon.StabilityInfo.DCBGovernor.EndBlock = stateBeacon.StabilityInfo.DCBGovernor.StartedBlock + common.DurationOfDCBBoard
	stateBeacon.StabilityInfo.DCBGovernor.StartAmountToken = ins.StartAmountToken
	return nil
}

func (stateBeacon *BestStateBeacon) UpdateGOVBoard(ins frombeaconins.AcceptGOVBoardIns) error {
	stateBeacon.StabilityInfo.GOVGovernor.BoardIndex += 1
	stateBeacon.StabilityInfo.GOVGovernor.BoardPaymentAddress = ins.BoardPaymentAddress
	stateBeacon.StabilityInfo.GOVGovernor.StartedBlock = stateBeacon.BestBlock.Header.Height
	stateBeacon.StabilityInfo.GOVGovernor.EndBlock = stateBeacon.StabilityInfo.GOVGovernor.StartedBlock + common.DurationOfGOVBoard
	stateBeacon.StabilityInfo.GOVGovernor.StartAmountToken = ins.StartAmountToken
	return nil
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
		boardType,
		paymentAddress,
		amount,
		propertyID,
	)
}

func (self *BlockChain) createSendBackTokenAfterVoteFailIns(
	boardType common.BoardType,
	newGovernorList []privacy.PaymentAddress,
	helper ConstitutionHelper,
) ([]frombeaconins.InstructionFromBeacon, error) {
	fmt.Println("[ndh]- Enter createSendBackTokenAfterVoteFailIns")
	var propertyID common.Hash
	if boardType == common.DCBBoard {
		propertyID = common.DCBTokenID
	} else {
		propertyID = common.GOVTokenID
	}
	setOfNewGovernor := make(map[string]bool, 0)
	for _, i := range newGovernorList {
		setOfNewGovernor[string(i.Bytes())] = true
	}
	currentBoardIndex := self.GetCurrentBoardIndex(helper)
	fmt.Println("[ndh] - Current board index: ", currentBoardIndex)
	// db := self.config.DataBase
	// gg := lvdb.ViewDBByPrefix(db, lvdb.VoteBoardListPrefix)
	// fmt.Println("[ndh] - START watch db when send back token:")
	// for key, value := range gg {
	// 	fmt.Println("[ndh] - ", key, value)
	// }
	// fmt.Println("[ndh] - END watch db when send back token:")
	begin := lvdb.GetKeyVoteBoardList(boardType, 0, nil, nil)
	end := lvdb.GetKeyVoteBoardList(boardType, currentBoardIndex+2, nil, nil)
	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}

	iter := self.config.DataBase.NewIterator(&searchRange, nil)
	listNewIns := make([]frombeaconins.InstructionFromBeacon, 0)
	for iter.Next() {
		key := iter.Key()
		_, boardIndex, candidatePayment, voterPaymentAddress, _ := lvdb.ParseKeyVoteBoardList(key)
		value := iter.Value()
		amountOfToken := lvdb.ParseValueVoteBoardList(value)
		_, found := setOfNewGovernor[string(candidatePayment)]
		if (boardIndex == currentBoardIndex+1) && (!found) {
			inst := frombeaconins.NewSendBackTokenVoteFailIns(
				boardType,
				*voterPaymentAddress,
				amountOfToken,
				propertyID,
			)
			listNewIns = append(
				listNewIns,
				inst,
			)
			instString, _ := inst.GetStringFormat()
			fmt.Println("[ndh]-SendBackIns: ", instString)
			err := self.config.DataBase.Delete(key)
			if err != nil {
				return nil, err
			}
		}
	}
	fmt.Println("[ndh]- createSendBackTokenAfterVoteFailIns ok", listNewIns)
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
	fmt.Printf("[ndh] ------------------- Create Share Reward Old Board Ins to %+v %+v\n", chairPaymentAddress, totalAmountCoinReward)
	voterList := chain.config.DataBase.GetBoardVoterList(helper.GetBoardType(), chairPaymentAddress, boardIndex)
	for _, voter := range voterList {
		amountOfVote := helper.GetAmountOfVoteToBoard(chain, chairPaymentAddress, voter, boardIndex)
		amountOfCoin := amountOfVote * totalAmountCoinReward / totalVoteAmount
		if amountOfCoin > 0 {
			Ins = append(Ins, chain.CreateSingleShareRewardOldBoardIns(
				helper,
				chairPaymentAddress,
				voter,
				amountOfCoin,
			))
		}
	}
	return Ins
}

// func (chain *BlockChain) GetCoinTermReward(helper ConstitutionHelper) uint64 {
// 	return helper.GetBoardFund(chain) * common.PercentageBoardSalary / common.BasePercentage
// }

func (self *BlockChain) CreateUpdateNewGovernorInstruction(
	helper ConstitutionHelper,
) ([]frombeaconins.InstructionFromBeacon, error) {
	instructions := make([]frombeaconins.InstructionFromBeacon, 0)
	newBoardList, err := self.config.DataBase.GetTopMostVoteGovernor(helper.GetBoardType(), self.GetCurrentBoardIndex(helper)+1)
	fmt.Println("[ndh] - SendBackToken to vote failed Enter function")
	if err != nil {
		fmt.Println("[ndh] - Error 1", err)
		if reflect.TypeOf(err).String() == "*database.DatabaseError" {
			return nil, nil
		}
		return nil, err
	}
	if len(newBoardList) == 0 {
		fmt.Println("[ndh] - not enough candidate")
		return nil, errors.New("not enough candidate")
	}
	fmt.Println("[ndh] -  SendBackToken to vote failed Get top most vote governor ok")
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
	fmt.Println("[ndh]-acceptBoardIns ok")
	if acceptBoardIns != nil {
		instructions = append(instructions, acceptBoardIns...)
	}
	sendBackTokenAfterVoteFailIns, err := self.createSendBackTokenAfterVoteFailIns(
		helper.GetBoardType(),
		newBoardPaymentAddress,
		helper,
	)
	if err != nil {
		fmt.Println("[ndh]- err", err)
		return nil, err
	}
	instructions = append(instructions, sendBackTokenAfterVoteFailIns...)
	fmt.Println("[ndh]-Update new governor inst ok", sendBackTokenAfterVoteFailIns)
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
	if chain.GetCurrentBoardIndex(helper) == 1 {
		return false
	}
	constitutionIndex := chain.GetConstitutionIndex(helper)
	if constitutionIndex == 0 {
		return false
	}
	fmt.Println("[ndh] - neededNewGovernor", chain.GetCurrentBoardIndex(helper), constitutionIndex, ConstitutionPerBoard, (constitutionIndex%ConstitutionPerBoard) == 0)
	return (constitutionIndex % ConstitutionPerBoard) == 0
}

func (chain *BlockChain) neededFirstNewGovernor(helper ConstitutionHelper) bool {
	fmt.Println("[ndh] - chain BeaconHeight of BestState", chain.BestState.Beacon.BeaconHeight, helper.GetBoardType())
	fmt.Println("[ndh] - Current board index ", chain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardIndex, chain.BestState.Beacon.StabilityInfo.GOVGovernor.BoardIndex)
	if helper.GetBoardType() == common.DCBBoard {
		if chain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardIndex > 1 {
			return false
		}
	} else {
		if chain.BestState.Beacon.StabilityInfo.GOVGovernor.BoardIndex > 1 {
			return false
		}
	}
	if EndOfFirstBoard == chain.BestState.Beacon.BeaconHeight {
		fmt.Println("[ndh] -  EndOfFirstBoard vs BeaconHeight", EndOfFirstBoard, chain.BestState.Beacon.BeaconHeight)
		return true
	}
	if (EndOfFirstBoard < chain.BestState.Beacon.BeaconHeight) && ((chain.BestState.Beacon.BeaconHeight-EndOfFirstBoard)%ExtendDurationForFirstBoard == 0) {
		return true
	}
	return false
}

func (chain *BlockChain) neededNewConstitution(helper ConstitutionHelper) bool {
	currentBoardIndex := chain.GetCurrentBoardIndex(helper)
	if currentBoardIndex == 1 {
		return false
	}
	endBlock := helper.GetConstitutionEndedBlockHeight(chain)
	fmt.Println("[ndh] - neededNewConstitution: ", endBlock, chain.BestState.Beacon.BeaconHeight)
	if chain.BestState.Beacon.BeaconHeight < endBlock {
		return false
	}
	if (chain.BestState.Beacon.BeaconHeight-endBlock)%ExtendDurationForFirstBoard == 0 {
		fmt.Println("[ndh] - neededNewConstitution!!!!!!!!!!!!!!!!!!")
		return true
	}
	return false
}

func (self *BlockChain) generateVotingInstructionWOIns(helper ConstitutionHelper) ([][]string, error) {

	instructions := make([]frombeaconins.InstructionFromBeacon, 0)
	fmt.Println("[ndh]-Enter generateVotingInstructionWOIns")
	if self.neededFirstNewGovernor(helper) {
		fmt.Println("[ndh]-[neededNewGovernor]-Create first instruction")
		updateGovernorInstruction, err := self.CreateUpdateNewGovernorInstruction(helper)
		if err != nil {
			return nil, err
		}
		for _, inst := range updateGovernorInstruction {
			instString, _ := inst.GetStringFormat()
			fmt.Println("[ndh]-[neededNewGovernor] - Created ", instString)
		}
		if len(updateGovernorInstruction) != 0 {
			instructions = append(instructions, updateGovernorInstruction...)
		}
	} else {
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
			if len(updateProposalInstruction) != 0 {
				fmt.Println("[ndh] - updateProposalInstruction ok.")
				instructions = append(instructions, updateProposalInstruction...)
			}
			//============================ VOTE BOARD
			if self.neededNewGovernor(helper) {
				updateGovernorInstruction, err := self.CreateUpdateNewGovernorInstruction(helper)
				if err != nil {
					return nil, err
				}
				if len(updateGovernorInstruction) != 0 {
					instructions = append(instructions, updateGovernorInstruction...)
				}
			}
		}
	}
	instructionsString := make([][]string, 0)
	for _, instruction := range instructions {
		newIns, err := instruction.GetStringFormat()
		if err != nil {
			return nil, err
		}
		fmt.Println("[ndh] - ", newIns)
		if len(newIns) != 0 {
			instructionsString = append(instructionsString, newIns)
		}
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

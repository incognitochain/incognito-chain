package blockchain

import (
	"encoding/binary"
	"fmt"
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
	GetStartedNormalVote(chain *BlockChain) uint64
	CheckSubmitProposalType(tx metadata.Transaction) bool
	NewAcceptProposalIns(txId *common.Hash, voter component.Voter, shardID byte) frombeaconins.InstructionFromBeacon
	GetBoardType() common.BoardType
	GetConstitutionEndedBlockHeight(chain *BlockChain) uint64
	GetSealerPaymentAddress(metadata.Transaction) []privacy.PaymentAddress
	NewRewardProposalSubmitterIns(blockgen *BlockChain, receiverAddress *privacy.PaymentAddress) (instruction frombeaconins.InstructionFromBeacon, err error)
	GetPaymentAddressFromSubmitProposalMetadata(tx metadata.Transaction) *privacy.PaymentAddress
	GetPaymentAddressVoter(chain *BlockChain) (privacy.PaymentAddress, error)
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
	GetNumberOfGovernor() int32
	GetSubmitProposalInfo(tx metadata.Transaction) (*component.SubmitProposalInfo, error)
	GetProposalTxID(tx metadata.Transaction) (hash *common.Hash)
	SetNewConstitution(bc *BlockChain, constitutionInfo *ConstitutionInfo, welfare int32, submitProposalTx metadata.Transaction)
	CreatePunishDecryptIns(paymentAddress *privacy.PaymentAddress) frombeaconins.InstructionFromBeacon
}

func (chain *BlockChain) createRewardProposalWinnerIns(
	constitutionHelper ConstitutionHelper,
) (frombeaconins.InstructionFromBeacon, error) {
	paymentAddress, err := constitutionHelper.GetPaymentAddressVoter(chain)
	if err != nil {
		return nil, err
	}
	prize := constitutionHelper.GetPrizeProposal()
	ins := frombeaconins.NewRewardProposalWinnerIns(paymentAddress, prize)
	return ins, nil
}

func (self *BlockChain) BuildVoteTableAndPunishTransaction(
	helper ConstitutionHelper,
) (
	resIns []frombeaconins.InstructionFromBeacon,
	VoteTable map[common.Hash]map[string]int32,
	err error,
) {
	resIns = make([]frombeaconins.InstructionFromBeacon, 0)
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
			punishDecryptIns := helper.CreatePunishDecryptIns(&sealerPaymentAddressList[0])
			resIns = append(resIns, punishDecryptIns)
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
			punishDecryptIns := helper.CreatePunishDecryptIns(&sealerPaymentAddressList[valueSealer])
			resIns = append(resIns, punishDecryptIns)
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
) ([]frombeaconins.InstructionFromBeacon, error){
	//todo @big0t
	return nil, nil
}

func (self *BlockChain) createAcceptBoardIns(
	boardType common.BoardType,
	BoardPaymentAddress []privacy.PaymentAddress,
	sumOfVote uint64,
) ([]frombeaconins.InstructionFromBeacon, error) {
	acceptBoardIns := frombeaconins.NewAcceptBoardIns(boardType, BoardPaymentAddress, sumOfVote)
	return []frombeaconins.InstructionFromBeacon{acceptBoardIns}, nil
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
	shardID byte,
) ([]frombeaconins.InstructionFromBeacon, error) {
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
			listNewIns = append(
				listNewIns,
				frombeaconins.NewSendBackTokenVoteFailIns(
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
	amountOfToken uint64,
) frombeaconins.InstructionFromBeacon {
	return frombeaconins.NewShareRewardOldBoardMetadataIns(
		chairPaymentAddress, voterPaymentAddress, helper.GetBoardType(), amountOfCoin, amountOfToken,
	)
}

func (chain *BlockChain) CreateShareRewardOldBoardIns(
	helper ConstitutionHelper,
	chairPaymentAddress privacy.PaymentAddress,
	totalAmountCoinReward uint64,
	totalAmountTokenReward uint64,
	totalVoteAmount uint64,
) []frombeaconins.InstructionFromBeacon {
	Ins := make([]frombeaconins.InstructionFromBeacon, 0)

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
		))
	}
	return Ins
}

func (chain *BlockChain) GetCoinTermReward(helper ConstitutionHelper) uint64 {
	return helper.GetBoardFund(chain) * common.PercentageBoardSalary / common.BasePercentage
}

func (self *BlockChain) createSendRewardOldBoardIns(
	helper ConstitutionHelper,
	shardID byte,
) ([]frombeaconins.InstructionFromBeacon, error) {
	//reward for each by voteDCBList
	//todo @big0t
	return nil, nil
}

//todo @0xjackalope reward for chair
func (self *BlockChain) CreateUpdateNewGovernorInstruction(
	helper ConstitutionHelper,
	shardID byte,
) ([]frombeaconins.InstructionFromBeacon, error) {
	instructions := make([]frombeaconins.InstructionFromBeacon, 0)
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

	acceptBoardIns, err := self.createAcceptBoardIns(
		helper.GetBoardType(),
		newBoardPaymentAddress,
		sumOfVote,
	)
	if err != nil {
		return nil, err
	}
	instructions = append(instructions, acceptBoardIns...)

	sendBackTokenAfterVoteFailIns, err := self.createSendBackTokenAfterVoteFailIns(
		helper.GetBoardType(),
		newBoardPaymentAddress,
		shardID,
	)
	if err != nil {
		return nil, err
	}
	instructions = append(instructions, sendBackTokenAfterVoteFailIns...)

	sendRewardOldBoardIns, err := self.createSendRewardOldBoardIns(helper, shardID)
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

func (chain *BlockChain) neededNewGovernor(boardType common.BoardType) bool {
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

func (self *BlockChain) generateVotingInstructionWOIns(shardID byte) ([][]string, error) {
	//todo 0xjackalope

	// 	prevBlock := blockgen.chain.BestState[shardID].BestBlock
	instructions := make([]frombeaconins.InstructionFromBeacon, 0)

	//============================ VOTE BOARD

	if self.neededNewGovernor(common.DCBBoard) {
		updateGovernorInstruction, err := self.CreateUpdateNewGovernorInstruction(DCBConstitutionHelper{}, shardID)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, updateGovernorInstruction...)
	}
	if self.neededNewGovernor(common.GOVBoard) {
		updateGovernorInstruction, err := self.CreateUpdateNewGovernorInstruction(GOVConstitutionHelper{}, shardID)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, updateGovernorInstruction...)
	}
	instructionsString := make([][]string, 0)
	for _, instruction := range instructions {
		instructionString, err := instruction.GetStringFormat()
		if err != nil {
			return nil, err
		}
		instructionsString = append(instructionsString, instructionString)
	}

	//============================ VOTE PROPOSAL
	// 	// Voting transaction
	// 	// Check if it is the case we need to apply a new proposal
	// 	// 1. newNW < lastNW * 0.9
	// 	// 2. current block height == last Constitution start time + last Constitution execute duration
	updateDCBEncryptPhraseInstruction, err := self.CreateUpdateEncryptPhraseAndRewardConstitutionIns(DCBConstitutionHelper{})
	if err != nil {
		return nil, err
	}
	instructions = append(instructions, updateDCBEncryptPhraseInstruction...)
	updateGOVEncryptPhraseInstruction, err := self.CreateUpdateEncryptPhraseAndRewardConstitutionIns(GOVConstitutionHelper{})
	if err != nil {
		return nil, err
	}
	instructions = append(instructions, updateGOVEncryptPhraseInstruction...)

	return instructionsString, nil
}

func (self *BlockChain) readyNewConstitution(helper ConstitutionHelper) bool {
	db := self.config.DataBase
	bestBlock := self.BestState.Beacon.BestBlock
	thisBlockHeight := bestBlock.Header.Height + 1
	lastEncryptBlockHeight, _ := db.GetEncryptionLastBlockHeight(helper.GetBoardType())
	encryptFlag, _ := db.GetEncryptFlag(helper.GetBoardType())
	if thisBlockHeight == lastEncryptBlockHeight+common.EncryptionOnePhraseDuration &&
		encryptFlag == common.NormalEncryptionFlag {
		return true
	}
	return false
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

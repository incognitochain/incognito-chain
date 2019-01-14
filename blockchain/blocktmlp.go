package blockchain

import (
	"fmt"
	"sort"
	"time"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type BlkTmplGenerator struct {
	txPool      TxPool
	chain       *BlockChain
	rewardAgent RewardAgent
}

type ConstitutionHelper interface {
	GetStartedNormalVote(generator *BlkTmplGenerator, chainID byte) uint32
	CheckSubmitProposalType(tx metadata.Transaction) bool
	CheckVotingProposalType(tx metadata.Transaction) bool
	GetAmountVoteTokenOfTx(tx metadata.Transaction) uint64
	TxAcceptProposal(txId *common.Hash, voter metadata.Voter) metadata.Transaction
	GetBoardType() string
	GetConstitutionEndedBlockHeight(generator *BlkTmplGenerator, chainID byte) uint32
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

// txPool represents a source of transactions to consider for inclusion in
// new blocks.
//
// The interface contract requires that all of these methods are safe for
// concurrent access with respect to the source.
type TxPool interface {
	// LastUpdated returns the last time a transaction was added to or
	// removed from the source pool.
	LastUpdated() time.Time

	// MiningDescs returns a slice of mining descriptors for all the
	// transactions in the source pool.
	MiningDescs() []*metadata.TxDesc

	// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
	HaveTransaction(hash *common.Hash) bool

	// RemoveTx remove tx from tx resource
	RemoveTx(tx metadata.Transaction) error

	//CheckTransactionFee
	// CheckTransactionFee(tx metadata.Transaction) (uint64, error)

	// Check tx validate by it self
	// ValidateTxByItSelf(tx metadata.Transaction) bool
}

type RewardAgent interface {
	GetBasicSalary(chainID byte) uint64
	GetSalaryPerTx(chainID byte) uint64
}

type buyBackFromInfo struct {
	paymentAddress privacy.PaymentAddress
	buyBackPrice   uint64
	value          uint64
	requestedTxID  *common.Hash
}

func (self BlkTmplGenerator) Init(txPool TxPool, chain *BlockChain, rewardAgent RewardAgent) (*BlkTmplGenerator, error) {
	return &BlkTmplGenerator{
		txPool:      txPool,
		chain:       chain,
		rewardAgent: rewardAgent,
	}, nil
}

func (blockgen *BlkTmplGenerator) buildCoinbases(
	chainID byte,
	privatekey *privacy.SpendingKey,
	txGroups map[string][]metadata.Transaction,
	salaryTx metadata.Transaction,
) ([]metadata.Transaction, error) {

	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	dcbHelper := DCBConstitutionHelper{}
	govHelper := GOVConstitutionHelper{}
	coinbases := []metadata.Transaction{salaryTx}
	// Voting transaction
	// Check if it is the case we need to apply a new proposal
	// 1. newNW < lastNW * 0.9
	// 2. current block height == last Constitution start time + last Constitution execute duration
	if blockgen.chain.readyNewConstitution(dcbHelper) {
		blockgen.chain.config.DataBase.SetEncryptionLastBlockHeight(dcbHelper.GetBoardType(), uint32(prevBlock.Header.Height+1))
		blockgen.chain.config.DataBase.SetEncryptFlag(dcbHelper.GetBoardType(), uint32(common.Lv3EncryptionFlag))
		tx, err := blockgen.createAcceptConstitutionAndPunishTxAndRewardSubmitter(chainID, DCBConstitutionHelper{}, privatekey)
		coinbases = append(coinbases, tx...)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		rewardTx, err := blockgen.createRewardProposalWinnerTx(chainID, DCBConstitutionHelper{})
		coinbases = append(coinbases, rewardTx)
	}
	if blockgen.chain.readyNewConstitution(govHelper) {
		blockgen.chain.config.DataBase.SetEncryptionLastBlockHeight(govHelper.GetBoardType(), uint32(prevBlock.Header.Height+1))
		blockgen.chain.config.DataBase.SetEncryptFlag(govHelper.GetBoardType(), uint32(common.Lv3EncryptionFlag))
		tx, err := blockgen.createAcceptConstitutionAndPunishTxAndRewardSubmitter(chainID, GOVConstitutionHelper{}, privatekey)
		coinbases = append(coinbases, tx...)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		rewardTx, err := blockgen.createRewardProposalWinnerTx(chainID, GOVConstitutionHelper{})
		coinbases = append(coinbases, rewardTx)
	}

	if blockgen.neededNewDCBGovernor(chainID) {
		coinbases = append(coinbases, blockgen.UpdateNewGovernor(DCBConstitutionHelper{}, chainID, privatekey)...)
	}
	if blockgen.neededNewGOVGovernor(chainID) {
		coinbases = append(coinbases, blockgen.UpdateNewGovernor(GOVConstitutionHelper{}, chainID, privatekey)...)
	}

	for _, tx := range txGroups["unlockTxs"] {
		coinbases = append(coinbases, tx)
	}
	for _, resTx := range txGroups["buySellResTxs"] {
		coinbases = append(coinbases, resTx)
	}
	for _, resTx := range txGroups["buyBackResTxs"] {
		coinbases = append(coinbases, resTx)
	}
	for _, resTx := range txGroups["issuingResTxs"] {
		coinbases = append(coinbases, resTx)
	}
	for _, refundTx := range txGroups["refundTxs"] {
		coinbases = append(coinbases, refundTx)
	}
	for _, oracleRewardTx := range txGroups["oracleRewardTxs"] {
		coinbases = append(coinbases, oracleRewardTx)
	}
	return coinbases, nil
}

func (blockgen *BlkTmplGenerator) NewBlockTemplate(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey, chainID byte) (*Block, error) {

	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	prevBlockHash := blockgen.chain.BestState[chainID].BestBlock.Hash()
	//prevCmTree := blockgen.chain.BestState[chainID].CmTree.MakeCopy()
	sourceTxns := blockgen.txPool.MiningDescs()
	txGroups, accumulativeValues, buyBackFromInfos, err := blockgen.checkAndGroupTxs(sourceTxns, chainID, privatekey)
	if err != nil {
		return nil, err
	}

	// Get salary per tx
	salaryPerTx := blockgen.rewardAgent.GetSalaryPerTx(chainID)
	// Get basic salary on block
	basicSalary := blockgen.rewardAgent.GetBasicSalary(chainID)

	if len(sourceTxns) < common.MinTxsInBlock {
		// if len of sourceTxns < MinTxsInBlock -> wait for more transactions
		Logger.log.Info("not enough transactions. Wait for more...")
		<-time.Tick(common.MinBlockWaitTime * time.Second)
		sourceTxns = blockgen.txPool.MiningDescs()
		if len(sourceTxns) == 0 {
			<-time.Tick(common.MaxBlockWaitTime * time.Second)
			sourceTxns = blockgen.txPool.MiningDescs()
			if len(sourceTxns) == 0 {
				// return nil, errors.Zero("No TxNormal")
				Logger.log.Info("Creating empty block...")
				goto concludeBlock
			}
		}
	}

	// check len of txs in block
	if len(txGroups["txsToAdd"]) == 0 {
		// return nil, errors.Zero("no transaction available for this chain")
		Logger.log.Info("Creating empty block...")
	}

concludeBlock:

	// Get blocksalary fund from txs
	salaryMULTP := uint64(0) //salary multiplier
	for _, blockTx := range txGroups["txsToAdd"] {
		if blockTx.GetTxFee() > 0 {
			salaryMULTP++
		}
	}

	// ------------------------ HOW to GET salary on a block-------------------
	// total salary = tx * (salary per tx) + (basic salary on block)
	// ------------------------------------------------------------------------
	totalSalary := salaryMULTP*salaryPerTx + basicSalary
	// create salary tx to pay constant for block producer
	salaryTx := new(transaction.Tx)
	err = salaryTx.InitTxSalary(totalSalary, payToAddress, privatekey, blockgen.chain.config.DataBase, nil)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	accumulativeValues["totalSalary"] = totalSalary
	txGroups, accumulativeValues, updatedOracleValues, err := blockgen.buildResponseTxs(chainID, sourceTxns, privatekey, txGroups, accumulativeValues, buyBackFromInfos)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}

	coinbases, err := blockgen.buildCoinbases(chainID, privatekey, txGroups, salaryTx)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	txGroups["txsToAdd"] = append(coinbases, txGroups["txsToAdd"]...)

	for _, tx := range txGroups["txToRemove"] {
		blockgen.txPool.RemoveTx(tx)
	}

	// Check for final balance of DCB and GOV
	if accumulativeValues["currentSalaryFund"]+accumulativeValues["totalFee"]+accumulativeValues["incomeFromBonds"] < accumulativeValues["totalSalary"]+accumulativeValues["govPayoutAmount"]+accumulativeValues["buyBackCoins"]+accumulativeValues["totalRefundAmt"]+accumulativeValues["totalOracleRewards"] {
		return nil, fmt.Errorf("Gov fund is not enough for salary and dividend payout")
	}

	currentBankFund := prevBlock.Header.BankFund
	if currentBankFund < accumulativeValues["bankPayoutAmount"] { // Can't spend loan payment just received in this block
		return nil, fmt.Errorf("Bank fund is not enough for dividend payout")
	}

	merkleRoots := Merkle{}.BuildMerkleTreeStore(txGroups["txsToAdd"])
	merkleRoot := merkleRoots[len(merkleRoots)-1]

	block := Block{
		Transactions: make([]metadata.Transaction, 0),
	}

	block.Header = BlockHeader{
		Height:        prevBlock.Header.Height + 1,
		Version:       BlockVersion,
		PrevBlockHash: *prevBlockHash,
		MerkleRoot:    *merkleRoot,
		// MerkleRootCommitments: common.Hash{},
		Timestamp:          time.Now().Unix(),
		BlockCommitteeSigs: make([]string, common.TotalValidators),
		Committee:          make([]string, common.TotalValidators),
		ChainID:            chainID,
		SalaryFund:         accumulativeValues["currentSalaryFund"] + accumulativeValues["incomeFromBonds"] + accumulativeValues["totalFee"] - accumulativeValues["totalSalary"] - accumulativeValues["govPayoutAmount"] - accumulativeValues["buyBackCoins"] - accumulativeValues["totalRefundAmt"] - accumulativeValues["totalOracleRewards"],
		BankFund:           prevBlock.Header.BankFund + accumulativeValues["loanPaymentAmount"] - accumulativeValues["bankPayoutAmount"],
		GOVConstitution:    prevBlock.Header.GOVConstitution, // TODO: 0xbunyip need get from gov-params tx
		DCBConstitution:    prevBlock.Header.DCBConstitution, // TODO: 0xbunyip need get from dcb-params tx
		Oracle:             prevBlock.Header.Oracle,
	}

	err = (&block).updateBlockHeader(blockgen, txGroups, accumulativeValues, updatedOracleValues)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}

	// Add new commitments to merkle tree and save the root
	/*newTree := prevCmTree
	err = UpdateMerkleTreeForBlock(newTree, &block)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	rt = newTree.GetRoot(common.IncMerkleTreeHeight)
	copy(block.Header.MerkleRootCommitments[:], rt)*/

	return &block, nil
}

func GetOracleDCBNationalWelfare() int32 {
	fmt.Print("Get national welfare. It is constant now. Need to change !!!")
	return 1234
}
func GetOracleGOVNationalWelfare() int32 {
	fmt.Print("Get national welfare. It is constant now. Need to change !!!")
	return 1234
}

func (blockgen *BlkTmplGenerator) neededNewDCBGovernor(chainID byte) bool {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	return int32(BestBlock.Header.DCBGovernor.EndBlock) == BestBlock.Header.Height+2
}
func (blockgen *BlkTmplGenerator) neededNewGOVGovernor(chainID byte) bool {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	return int32(BestBlock.Header.GOVGovernor.EndBlock) == BestBlock.Header.Height+2
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

package blockchain

import (
	"time"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
)

type BlkTmplGenerator struct {
	blockpool   BlockPool
	txPool      TxPool
	chain       *BlockChain
	rewardAgent RewardAgent
}

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

type buyBackFromInfo struct {
	paymentAddress privacy.PaymentAddress
	buyBackPrice   uint64
	value          uint64
	requestedTxID  *common.Hash
}

type RewardAgent interface {
	GetBasicSalary(shardID byte) uint64
	GetSalaryPerTx(shardID byte) uint64
}

func (self BlkTmplGenerator) Init(txPool TxPool, chain *BlockChain, rewardAgent RewardAgent) (*BlkTmplGenerator, error) {
	return &BlkTmplGenerator{
		txPool:      txPool,
		chain:       chain,
		rewardAgent: rewardAgent,
	}, nil
}

type BlockPool interface {
	RemoveBlock(shard int, blockHeight int) error
	GetNewShardBlock() map[byte]([]common.Hash)
}

type BlockPoolImp struct {
	//blocks map[common.Hash]
}

func (self BlockPoolImp) GetNewShardBlock() map[byte]([]common.Hash) {
	//TODO: implementation
	return nil
}

func (self *BlkTmplGenerator) NewBlockShard(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey, shardID byte) (*ShardBlock, error) {
	prevBlock := self.chain.BestState.Shard[shardID].BestBlock
	prevBlockHash := self.chain.BestState.Shard[shardID].BestBlock.Hash()
	sourceTxns := self.txPool.MiningDescs()

	var txsToAdd []metadata.Transaction
	var txToRemove []metadata.Transaction
	// var buySellReqTxs []metadata.Transaction
	// var issuingReqTxs []metadata.Transaction
	// var buyBackFromInfos []*buyBackFromInfo
	// bondsSold := uint64(0)
	// dcbTokensSold := uint64(0)
	// incomeFromBonds := uint64(0)
	// buyBackCoins := uint64(0)
	totalFee := uint64(0)

	// Get salary per tx
	salaryPerTx := self.rewardAgent.GetSalaryPerTx(shardID)
	// Get basic salary on block
	basicSalary := self.rewardAgent.GetBasicSalary(shardID)
	// currentBlockHeight := prevBlock.Header.Height + 1

	if len(sourceTxns) < common.MinTxsInBlock {
		// if len of sourceTxns < MinTxsInBlock -> wait for more transactions
		Logger.log.Info("not enough transactions. Wait for more...")
		<-time.Tick(common.MinBlockWaitTime * time.Second)
		sourceTxns = self.txPool.MiningDescs()
		if len(sourceTxns) == 0 {
			<-time.Tick(common.MaxBlockWaitTime * time.Second)
			sourceTxns = self.txPool.MiningDescs()
			if len(sourceTxns) == 0 {
				// return nil, errors.Zero("No TxNormal")
				Logger.log.Info("Creating empty block...")
				goto concludeBlock
			}
		}
	}

	for _, txDesc := range sourceTxns {
		tx := txDesc.Tx
		txShardID, _ := common.GetTxSenderChain(tx.GetSenderAddrLastByte())
		if txShardID != shardID {
			continue
		}
		// ValidateTransaction vote and propose transaction

		// TODO: need to determine a tx is in privacy format or not
		if !tx.ValidateTxByItself(tx.IsPrivacy(), self.chain.config.DataBase, self.chain, shardID) {
			txToRemove = append(txToRemove, metadata.Transaction(tx))
			continue
		}

		// switch tx.GetMetadataType() {
		// case metadata.BuyFromGOVRequestMeta:
		// 	{
		// 		income, soldAmt, addable := self.checkBuyFromGOVReqTx(shardID, tx, bondsSold)
		// 		if !addable {
		// 			txToRemove = append(txToRemove, tx)
		// 			continue
		// 		}
		// 		bondsSold += soldAmt
		// 		incomeFromBonds += income
		// 		buySellReqTxs = append(buySellReqTxs, tx)
		// 	}
		// case metadata.BuyBackRequestMeta:
		// 	{
		// 		buyBackFromInfo, addable := self.checkBuyBackReqTx(shardID, tx, buyBackCoins)
		// 		if !addable {
		// 			txToRemove = append(txToRemove, tx)
		// 			continue
		// 		}
		// 		buyBackCoins += (buyBackFromInfo.buyBackPrice + buyBackFromInfo.value)
		// 		buyBackFromInfos = append(buyBackFromInfos, buyBackFromInfo)
		// 	}
		// case metadata.NormalDCBBallotMetaFromSealer:
		// 	if !(currentBlockHeight < endedDCBPivot && currentBlockHeight >= lv1DCBPivot) {
		// 		continue
		// 	}
		// case metadata.NormalDCBBallotMetaFromOwner:
		// 	if !(currentBlockHeight < endedDCBPivot && currentBlockHeight >= lv1DCBPivot) {
		// 		continue
		// 	}
		// case metadata.SealedLv1DCBBallotMeta:
		// 	if !(currentBlockHeight < lv1DCBPivot && currentBlockHeight >= lv2DCBPivot) {
		// 		continue
		// 	}
		// case metadata.SealedLv2DCBBallotMeta:
		// 	if !(currentBlockHeight < lv2DCBPivot && currentBlockHeight >= lv3DCBPivot) {
		// 		continue
		// 	}
		// case metadata.SealedLv3DCBBallotMeta:
		// 	if !(currentBlockHeight < lv3DCBPivot && currentBlockHeight >= startedDCBPivot) {
		// 		continue
		// 	}
		// case metadata.NormalGOVBallotMetaFromSealer:
		// 	if !(currentBlockHeight < endedGOVPivot && currentBlockHeight >= lv1GOVPivot) {
		// 		continue
		// 	}
		// case metadata.NormalGOVBallotMetaFromOwner:
		// 	if !(currentBlockHeight < endedGOVPivot && currentBlockHeight >= lv1GOVPivot) {
		// 		continue
		// 	}
		// case metadata.SealedLv1GOVBallotMeta:
		// 	if !(currentBlockHeight < lv1GOVPivot && currentBlockHeight >= lv2GOVPivot) {
		// 		continue
		// 	}
		// case metadata.SealedLv2GOVBallotMeta:
		// 	if !(currentBlockHeight < lv2GOVPivot && currentBlockHeight >= lv3GOVPivot) {
		// 		continue
		// 	}
		// case metadata.SealedLv3GOVBallotMeta:
		// 	if !(currentBlockHeight < lv3GOVPivot && currentBlockHeight >= startedGOVPivot) {
		// 		continue
		// 	}
		// case metadata.IssuingRequestMeta:
		// 	addable, newDCBTokensSold := blockgen.checkIssuingReqTx(shardID, tx, dcbTokensSold)
		// 	dcbTokensSold = newDCBTokensSold
		// 	if !addable {
		// 		txToRemove = append(txToRemove, tx)
		// 		continue
		// 	}
		// 	issuingReqTxs = append(issuingReqTxs, tx)
		// }

		totalFee += tx.GetTxFee()
		txsToAdd = append(txsToAdd, tx)
		if len(txsToAdd) == common.MaxTxsInBlock {
			break
		}
	}

	// check len of txs in block
	if len(txsToAdd) == 0 {
		// return nil, errors.Zero("no transaction available for this chain")
		Logger.log.Info("Creating empty block...")
	}

concludeBlock:
	// rt := prevBlock.Header.MerkleRootCommitments.CloneBytes()

	// TODO(@0xbunyip): cap #tx to common.MaxTxsInBlock
	// Process dividend payout for DCB if needed
	// bankDivTxs, bankPayoutAmount, err := blockgen.processBankDividend(blockHeight, privatekey)
	// if err != nil {
	// 	return nil, err
	// }
	// for _, tx := range bankDivTxs {
	// 	txsToAdd = append(txsToAdd, tx)
	// }

	// // Process dividend payout for GOV if needed
	// govDivTxs, govPayoutAmount, err := blockgen.processGovDividend(blockHeight, privatekey)
	// if err != nil {
	// 	return nil, err
	// }
	// for _, tx := range govDivTxs {
	// 	txsToAdd = append(txsToAdd, tx)
	// }

	// // Process crowdsale for DCB
	// dcbSaleTxs, removableTxs, err := blockgen.processCrowdsale(sourceTxns, rt, shardID, privatekey)
	// if err != nil {
	// 	return nil, err
	// }
	// for _, tx := range dcbSaleTxs {
	// 	txsToAdd = append(txsToAdd, tx)
	// }
	// for _, tx := range removableTxs {
	// 	txToRemove = append(txToRemove, tx)
	// }

	// Get blocksalary fund from txs
	salaryFundAdd := uint64(0)
	salaryMULTP := uint64(0) //salary multiplier
	for _, blockTx := range txsToAdd {
		if blockTx.GetTxFee() > 0 {
			salaryMULTP++
		}
	}

	// ------------------------ HOW to GET salary on a block-------------------
	// total salary = tx * (salary per tx) + (basic salary on block)
	// ------------------------------------------------------------------------
	totalSalary := salaryMULTP*salaryPerTx + basicSalary
	// create salary tx to pay constant for block producer
	salaryTx, err := transaction.CreateTxSalary(totalSalary, payToAddress, privatekey, self.chain.config.DataBase)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	// create buy/sell response txs to distribute bonds/govs to requesters
	// buySellResTxs, err := blockgen.buildBuySellResponsesTx(
	// 	buySellReqTxs,
	// 	blockgen.chain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.SellingBonds,
	// )
	// if err != nil {
	// 	Logger.log.Error(err)
	// 	return nil, err
	// }
	// // create buy-back response txs to distribute constants to buy-back requesters
	// buyBackResTxs, err := blockgen.buildBuyBackResponsesTx(buyBackFromInfos, shardID, privatekey)
	// if err != nil {
	// 	Logger.log.Error(err)
	// 	return nil, err
	// }

	// create refund txs
	currentSalaryFund := prevBlock.Header.SalaryFund
	remainingFund := currentSalaryFund + totalFee + salaryFundAdd - totalSalary
	// remainingFund := currentSalaryFund + totalFee + salaryFundAdd + incomeFromBonds - (totalSalary + buyBackCoins)
	// refundTxs, totalRefundAmt := blockgen.buildRefundTxs(shardID, remainingFund, privatekey)

	// issuingResTxs, err := blockgen.buildIssuingResTxs(shardID, issuingReqTxs, privatekey)
	// if err != nil {
	// 	Logger.log.Error(err)
	// 	return nil, err
	// }

	// // Get loan payment amount to add to DCB fund
	// loanPaymentAmount, unlockTxs, removableTxs := blockgen.processLoan(sourceTxns, privatekey)
	// for _, tx := range removableTxs {
	// 	txToRemove = append(txToRemove, tx)
	// }

	coinbases := []metadata.Transaction{salaryTx}
	// Voting transaction
	// Check if it is the case we need to apply a new proposal
	// 1. newNW < lastNW * 0.9
	// 2. current block height == last Constitution start time + last Constitution execute duration
	// if blockgen.neededNewDCBConstitution(shardID) {
	// 	tx, err := blockgen.createAcceptConstitutionAndPunishTx(shardID, DCBConstitutionHelper{})
	// 	coinbases = append(coinbases, tx...)
	// 	if err != nil {
	// 		Logger.log.Error(err)
	// 		return nil, err
	// 	}
	// }
	// if blockgen.neededNewGOVConstitution(shardID) {
	// 	tx, err := blockgen.createAcceptConstitutionAndPunishTx(shardID, GOVConstitutionHelper{})
	// 	coinbases = append(coinbases, tx...)
	// 	if err != nil {
	// 		Logger.log.Error(err)
	// 		return nil, err
	// 	}
	// }

	// if int32(prevBlock.Header.DCBGovernor.EndBlock) == prevBlock.Header.Height+1 {
	// 	newBoardList, _ := blockgen.chain.config.DataBase.GetTopMostVoteDCBGovernor(common.NumberOfDCBGovernors)
	// 	sort.Sort(newBoardList)
	// 	sumOfVote := uint64(0)
	// 	var newDCBBoardPubKey [][]byte
	// 	for _, i := range newBoardList {
	// 		newDCBBoardPubKey = append(newDCBBoardPubKey, i.PubKey)
	// 		sumOfVote += i.VoteAmount
	// 	}

	// 	coinbases = append(coinbases, blockgen.createAcceptDCBBoardTx(newDCBBoardPubKey, sumOfVote))
	// 	coinbases = append(coinbases, blockgen.CreateSendDCBVoteTokenToGovernorTx(shardID, newBoardList, sumOfVote)...)

	// 	coinbases = append(coinbases, blockgen.CreateSendBackDCBTokenAfterVoteFail(shardID, newDCBBoardPubKey)...)
	// 	// Todo @0xjackalope: send reward to old board and delete them from database before send back token to new board
	// 	//xxx add to pool
	// }

	// if int32(prevBlock.Header.GOVGovernor.EndBlock) == prevBlock.Header.Height+1 {
	// 	newBoardList, _ := blockgen.chain.config.DataBase.GetTopMostVoteGOVGovernor(common.NumberOfGOVGovernors)
	// 	sort.Sort(newBoardList)
	// 	sumOfVote := uint64(0)
	// 	var newGOVBoardPubKey [][]byte
	// 	for _, i := range newBoardList {
	// 		newGOVBoardPubKey = append(newGOVBoardPubKey, i.PubKey)
	// 		sumOfVote += i.VoteAmount
	// 	}

	// 	coinbases = append(coinbases, blockgen.createAcceptGOVBoardTx(newGOVBoardPubKey, sumOfVote))
	// 	coinbases = append(coinbases, blockgen.CreateSendGOVVoteTokenToGovernorTx(shardID, newBoardList, sumOfVote)...)

	// 	coinbases = append(coinbases, blockgen.CreateSendBackGOVTokenAfterVoteFail(shardID, newGOVBoardPubKey)...)
	// 	// Todo @0xjackalope: send reward to old board and delete them from database before send back token to new board
	// 	//xxx add to pool
	// }

	// for _, tx := range unlockTxs {
	// 	coinbases = append(coinbases, tx)
	// }
	// for _, resTx := range buySellResTxs {
	// 	coinbases = append(coinbases, resTx)
	// }
	// for _, resTx := range buyBackResTxs {
	// 	coinbases = append(coinbases, resTx)
	// }
	// for _, resTx := range issuingResTxs {
	// 	coinbases = append(coinbases, resTx)
	// }
	// for _, refundTx := range refundTxs {
	// 	coinbases = append(coinbases, refundTx)
	// }

	txsToAdd = append(coinbases, txsToAdd...)

	for _, tx := range txToRemove {
		self.txPool.RemoveTx(tx)
	}

	// // Check for final balance of DCB and GOV
	// if currentSalaryFund+totalFee+salaryFundAdd+incomeFromBonds < totalSalary+govPayoutAmount+buyBackCoins+totalRefundAmt {
	// 	return nil, fmt.Errorf("Gov fund is not enough for salary and dividend payout")
	// }

	// currentBankFund := prevBlock.Header.BankFund
	// if currentBankFund < bankPayoutAmount { // Can't spend loan payment just received in this block
	// 	return nil, fmt.Errorf("Bank fund is not enough for dividend payout")
	// }

	merkleRoots := Merkle{}.BuildMerkleTreeStore(txsToAdd)
	merkleRoot := merkleRoots[len(merkleRoots)-1]

	block := &ShardBlock{
		Body: ShardBody{
			Transactions: make([]metadata.Transaction, 0),
		},
	}

	block.Header = ShardHeader{
		Height:        prevBlock.Header.Height + 1,
		Version:       BlockVersion,
		PrevBlockHash: *prevBlockHash,
		MerkleRoot:    *merkleRoot,
		// MerkleRootCommitments: common.Hash{},
		Timestamp:  time.Now().Unix(),
		ShardID:    shardID,
		SalaryFund: remainingFund,
		// BankFund:           prevBlock.Header.BankFund + loanPaymentAmount - bankPayoutAmount,
		// GOVConstitution:    prevBlock.Header.GOVConstitution, // TODO: need get from gov-params tx
		// DCBConstitution:    prevBlock.Header.DCBConstitution, // TODO: need get from dcb-params tx
	}
	// if block.Header.GOVConstitution.GOVParams.SellingBonds != nil {
	// 	block.Header.GOVConstitution.GOVParams.SellingBonds.BondsToSell -= bondsSold
	// }
	// if block.Header.DCBConstitution.DCBParams.SaleDBCTOkensByUSDData != nil {
	// 	block.Header.DCBConstitution.DCBParams.SaleDBCTOkensByUSDData.Amount -= dcbTokensSold
	// }

	for _, tx := range txsToAdd {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
		// Handle if this transaction change something in block header
		// if tx.GetMetadataType() == metadata.AcceptDCBProposalMeta {
		// 	block.updateDCBConstitution(tx, blockgen)
		// }
		// if tx.GetMetadataType() == metadata.AcceptGOVProposalMeta {
		// 	block.updateGOVConstitution(tx, blockgen)
		// }
		// if tx.GetMetadataType() == metadata.AcceptDCBBoardMeta {
		// 	block.UpdateDCBBoard(tx)
		// }
		// if tx.GetMetadataType() == metadata.AcceptGOVBoardMeta {
		// 	block.UpdateGOVBoard(tx)
		// }
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

	return block, nil
}

func (self *BlkTmplGenerator) NewBlockBeacon() (*BeaconBlock, error) {
	block := &BeaconBlock{}
	// block.ProducerSig = ""
	// block.AggregatedSig = ""
	// block.ValidatorsIdx = nil

	// //bodyBlk := BeaconBlockBody{}
	// //shardBlock := blockPool.GetNewShardBlock()
	// //TODO: get hash from shardBlock & build shard state
	// //bodyBlk.ShardState = shardState

	// // TODO: build param from shardBlock

	// //block.Body = bodyBlk
	// // TODO: build header
	return block, nil
}

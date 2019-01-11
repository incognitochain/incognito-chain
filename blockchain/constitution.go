package blockchain

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/database/lvdb"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type ConstitutionInfo struct {
	ConstitutionIndex  uint32
	StartedBlockHeight uint32
	ExecuteDuration    uint32
	AcceptProposalTXID common.Hash
}

func NewConstitutionInfo(constitutionIndex uint32, startedBlockHeight uint32, executeDuration uint32, proposalTXID common.Hash) *ConstitutionInfo {
	return &ConstitutionInfo{
		ConstitutionIndex:  constitutionIndex,
		StartedBlockHeight: startedBlockHeight,
		ExecuteDuration:    executeDuration,
		AcceptProposalTXID: proposalTXID,
	}
}

func (constitutionInfo ConstitutionInfo) GetConstitutionIndex() uint32 {
	return constitutionInfo.ConstitutionIndex
}

type GOVConstitution struct {
	ConstitutionInfo
	CurrentGOVNationalWelfare int32
	GOVParams                 params.GOVParams
}

func NewGOVConstitution(constitutionInfo *ConstitutionInfo, currentGOVNationalWelfare int32, GOVParams *params.GOVParams) *GOVConstitution {
	return &GOVConstitution{
		ConstitutionInfo:          *constitutionInfo,
		CurrentGOVNationalWelfare: currentGOVNationalWelfare,
		GOVParams:                 *GOVParams,
	}
}

func (dcbConstitution DCBConstitution) GetEndedBlockHeight() uint32 {
	return dcbConstitution.StartedBlockHeight + dcbConstitution.ExecuteDuration
}

func (govConstitution GOVConstitution) GetEndedBlockHeight() uint32 {
	return govConstitution.StartedBlockHeight + govConstitution.ExecuteDuration
}

type DCBConstitution struct {
	ConstitutionInfo
	CurrentDCBNationalWelfare int32
	DCBParams                 params.DCBParams
}

func NewDCBConstitution(constitutionInfo *ConstitutionInfo, currentDCBNationalWelfare int32, DCBParams *params.DCBParams) *DCBConstitution {
	return &DCBConstitution{
		ConstitutionInfo:          *constitutionInfo,
		CurrentDCBNationalWelfare: currentDCBNationalWelfare,
		DCBParams:                 *DCBParams,
	}
}

type DCBConstitutionHelper struct{}
type GOVConstitutionHelper struct{}

func (helper DCBConstitutionHelper) GetConstitutionEndedBlockHeight(blockgen *BlkTmplGenerator, chainID byte) uint32 {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastDCBConstitution := BestBlock.Header.DCBConstitution
	return lastDCBConstitution.StartedBlockHeight + lastDCBConstitution.ExecuteDuration
}

func (helper GOVConstitutionHelper) GetConstitutionEndedBlockHeight(blockgen *BlkTmplGenerator, chainID byte) uint32 {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastGOVConstitution := BestBlock.Header.GOVConstitution
	return lastGOVConstitution.StartedBlockHeight + lastGOVConstitution.ExecuteDuration
}

func (helper DCBConstitutionHelper) GetStartedNormalVote(blockgen *BlkTmplGenerator, chainID byte) uint32 {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastDCBConstitution := BestBlock.Header.DCBConstitution
	return lastDCBConstitution.StartedBlockHeight - common.EncryptionOnePhraseDuration
}

func (helper DCBConstitutionHelper) CheckSubmitProposalType(tx metadata.Transaction) bool {
	return tx.GetMetadataType() == metadata.SubmitDCBProposalMeta
}

func (helper DCBConstitutionHelper) CheckVotingProposalType(tx metadata.Transaction) bool {
	return tx.GetMetadataType() == metadata.VoteDCBProposalMeta
}

func (helper DCBConstitutionHelper) GetAmountVoteTokenOfTx(tx metadata.Transaction) uint64 {
	return tx.(*transaction.TxCustomToken).GetAmountOfVote()
}

func (helper GOVConstitutionHelper) GetStartedNormalVote(blockgen *BlkTmplGenerator, chainID byte) uint32 {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastGOVConstitution := BestBlock.Header.GOVConstitution
	return lastGOVConstitution.StartedBlockHeight - common.EncryptionOnePhraseDuration
}

func (helper GOVConstitutionHelper) CheckSubmitProposalType(tx metadata.Transaction) bool {
	return tx.GetMetadataType() == metadata.SubmitGOVProposalMeta
}

func (helper GOVConstitutionHelper) CheckVotingProposalType(tx metadata.Transaction) bool {
	return tx.GetMetadataType() == metadata.VoteGOVProposalMeta
}

func (helper GOVConstitutionHelper) GetAmountVoteTokenOfTx(tx metadata.Transaction) uint64 {
	return tx.(*transaction.TxCustomToken).GetAmountOfVote()
}

func (helper DCBConstitutionHelper) TxAcceptProposal(txId *common.Hash, voter metadata.Voter) metadata.Transaction {
	acceptTx := transaction.Tx{
		Metadata: metadata.NewAcceptDCBProposalMetadata(*txId, voter),
	}
	return &acceptTx
}

func (helper GOVConstitutionHelper) TxAcceptProposal(txId *common.Hash, voter metadata.Voter) metadata.Transaction {
	acceptTx := transaction.Tx{
		Metadata: metadata.NewAcceptGOVProposalMetadata(*txId, voter),
	}
	return &acceptTx
}

func (helper DCBConstitutionHelper) GetBoardType() string {
	return "dcb"
}

func (helper GOVConstitutionHelper) GetBoardType() string {
	return "gov"
}

func (helper DCBConstitutionHelper) CreatePunishDecryptTx(pubKey []byte) metadata.Metadata {
	return metadata.NewPunishDCBDecryptMetadata(pubKey)
}

func (helper GOVConstitutionHelper) CreatePunishDecryptTx(pubKey []byte) metadata.Metadata {
	return metadata.NewPunishGOVDecryptMetadata(pubKey)
}

func (helper DCBConstitutionHelper) GetSealerPubKey(tx metadata.Transaction) [][]byte {
	meta := tx.GetMetadata().(*metadata.SealedLv3DCBVoteProposalMetadata)
	return meta.SealedDCBVoteProposal.LockerPubKeys
}

func (helper GOVConstitutionHelper) GetSealerPubKey(tx metadata.Transaction) [][]byte {
	meta := tx.GetMetadata().(*metadata.SealedLv3GOVVoteProposalMetadata)
	return meta.SealedGOVVoteProposal.LockerPubKeys
}

func (helper DCBConstitutionHelper) NewTxRewardProposalSubmitter(blockgen *BlkTmplGenerator, receiverAddress *privacy.PaymentAddress, minerPrivateKey *privacy.SpendingKey) (metadata.Transaction, error) {
	meta := metadata.NewRewardDCBProposalSubmitterMetadata()
	tx := transaction.Tx{}
	err := tx.InitTxSalary(common.RewardProposalSubmitter, receiverAddress, minerPrivateKey, blockgen.chain.config.DataBase, meta)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (helper GOVConstitutionHelper) NewTxRewardProposalSubmitter(blockgen *BlkTmplGenerator, receiverAddress *privacy.PaymentAddress, minerPrivateKey *privacy.SpendingKey) (metadata.Transaction, error) {
	meta := metadata.NewRewardGOVProposalSubmitterMetadata()
	tx := transaction.Tx{}
	err := tx.InitTxSalary(common.RewardProposalSubmitter, receiverAddress, minerPrivateKey, blockgen.chain.config.DataBase, meta)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (helper DCBConstitutionHelper) GetPaymentAddressFromSubmitProposalMetadata(tx metadata.Transaction) *privacy.PaymentAddress {
	meta := tx.GetMetadata().(*metadata.SubmitDCBProposalMetadata)
	return &meta.PaymentAddress
}
func (helper GOVConstitutionHelper) GetPaymentAddressFromSubmitProposalMetadata(tx metadata.Transaction) *privacy.PaymentAddress {
	meta := tx.GetMetadata().(*metadata.SubmitGOVProposalMetadata)
	return &meta.PaymentAddress
}

func (helper DCBConstitutionHelper) GetPubKeyVoter(blockgen *BlkTmplGenerator, chainID byte) ([]byte, error) {
	bestBlock := blockgen.chain.BestState[chainID].BestBlock
	_, _, _, tx, _ := blockgen.chain.GetTransactionByHash(&bestBlock.Header.DCBConstitution.AcceptProposalTXID)
	meta := tx.GetMetadata().(*metadata.AcceptDCBProposalMetadata)
	return meta.Voter.PubKey, nil
}
func (helper GOVConstitutionHelper) GetPubKeyVoter(blockgen *BlkTmplGenerator, chainID byte) ([]byte, error) {
	bestBlock := blockgen.chain.BestState[chainID].BestBlock
	_, _, _, tx, _ := blockgen.chain.GetTransactionByHash(&bestBlock.Header.GOVConstitution.AcceptProposalTXID)
	meta := tx.GetMetadata().(*metadata.AcceptGOVProposalMetadata)
	return meta.Voter.PubKey, nil
}

func (helper DCBConstitutionHelper) GetPrizeProposal() uint32 {
	return uint32(common.Maxint32(GetOracleDCBNationalWelfare(), int32(0)))
}

func (helper GOVConstitutionHelper) GetPrizeProposal() uint32 {
	return uint32(common.Maxint32(GetOracleGOVNationalWelfare(), int32(0)))
}

func (helper DCBConstitutionHelper) GetTopMostVoteGovernor(blockgen *BlkTmplGenerator) (database.CandidateList, error) {
	return blockgen.chain.config.DataBase.GetTopMostVoteGovernor(helper.GetBoardType(), blockgen.chain.GetCurrentBoardIndex(helper))
}
func (helper GOVConstitutionHelper) GetTopMostVoteGovernor(blockgen *BlkTmplGenerator) (database.CandidateList, error) {
	return blockgen.chain.config.DataBase.GetTopMostVoteGovernor(helper.GetBoardType(), blockgen.chain.GetCurrentBoardIndex(helper))
}

func (helper DCBConstitutionHelper) GetBoardSumToken(blockgen *BlkTmplGenerator) uint64 {
	return blockgen.chain.BestState[0].BestBlock.Header.DCBGovernor.StartAmountToken
}

func (helper GOVConstitutionHelper) GetBoardSumToken(blockgen *BlkTmplGenerator) uint64 {
	return blockgen.chain.BestState[0].BestBlock.Header.GOVGovernor.StartAmountToken
}

func (helper DCBConstitutionHelper) GetBoardFund(blockgen *BlkTmplGenerator) uint64 {
	return blockgen.chain.BestState[0].BestBlock.Header.BankFund
}

func (helper GOVConstitutionHelper) GetBoardFund(blockgen *BlkTmplGenerator) uint64 {
	return blockgen.chain.BestState[0].BestBlock.Header.SalaryFund
}

func (helper DCBConstitutionHelper) GetTokenID() *common.Hash {
	id := common.Hash(common.DCBTokenID)
	return &id
}

func (helper GOVConstitutionHelper) GetTokenID() *common.Hash {
	id := common.Hash(common.GOVTokenID)
	return &id
}

func (helper DCBConstitutionHelper) GetBoard(chain *BlockChain) Governor {
	return chain.BestState[0].BestBlock.Header.DCBGovernor
}

func (helper GOVConstitutionHelper) GetBoard(chain *BlockChain) Governor {
	return chain.BestState[0].BestBlock.Header.GOVGovernor
}

func (helper DCBConstitutionHelper) GetAmountVoteTokenOfBoard(blockgen *BlkTmplGenerator, pubKey []byte, boardIndex uint32) uint64 {
	value, _ := blockgen.chain.config.DataBase.GetVoteTokenAmount(helper.GetBoardType(), boardIndex, pubKey)
	return uint64(value)
}
func (helper GOVConstitutionHelper) GetAmountVoteTokenOfBoard(blockgen *BlkTmplGenerator, pubKey []byte, boardIndex uint32) uint64 {
	value, _ := blockgen.chain.config.DataBase.GetVoteTokenAmount(helper.GetBoardType(), boardIndex, pubKey)
	return uint64(value)
}

func (helper DCBConstitutionHelper) GetAmountOfVoteToBoard(blockgen *BlkTmplGenerator, candidatePubKey []byte, voterPubKey []byte, boardIndex uint32) uint64 {
	key := lvdb.GetKeyVoteBoardList(helper.GetBoardType(), boardIndex, candidatePubKey, voterPubKey)
	value, _ := blockgen.chain.config.DataBase.Get(key)
	amount := lvdb.ParseValueVoteBoardList(value)
	return amount
}
func (helper GOVConstitutionHelper) GetAmountOfVoteToBoard(blockgen *BlkTmplGenerator, candidatePubKey []byte, voterPubKey []byte, boardIndex uint32) uint64 {
	key := lvdb.GetKeyVoteBoardList(helper.GetBoardType(), boardIndex, candidatePubKey, voterPubKey)
	value, _ := blockgen.chain.config.DataBase.Get(key)
	amount := lvdb.ParseValueVoteBoardList(value)
	return amount
}

func (helper DCBConstitutionHelper) GetCurrentBoardPubKeys(blockgen *BlkTmplGenerator) [][]byte {
	return blockgen.chain.BestState[0].BestBlock.Header.DCBGovernor.BoardPubKeys
}

func (helper GOVConstitutionHelper) GetCurrentBoardPubKeys(blockgen *BlkTmplGenerator) [][]byte {
	return blockgen.chain.BestState[0].BestBlock.Header.GOVGovernor.BoardPubKeys
}

func (helper DCBConstitutionHelper) GetConstitutionInfo(chain *BlockChain) ConstitutionInfo {
	return chain.BestState[0].BestBlock.Header.DCBConstitution.ConstitutionInfo
}

func (helper GOVConstitutionHelper) GetConstitutionInfo(chain *BlockChain) ConstitutionInfo {
	return chain.BestState[0].BestBlock.Header.GOVConstitution.ConstitutionInfo
}

func (helper DCBConstitutionHelper) GetCurrentNationalWelfare(chain *BlockChain) int32 {
	return GetOracleDCBNationalWelfare()
}

func (helper GOVConstitutionHelper) GetCurrentNationalWelfare(chain *BlockChain) int32 {
	return GetOracleGOVNationalWelfare()
}

func (helper DCBConstitutionHelper) GetThresholdRatioOfCrisis() int32 {
	return ThresholdRatioOfDCBCrisis
}

func (helper GOVConstitutionHelper) GetThresholdRatioOfCrisis() int32 {
	return ThresholdRatioOfGOVCrisis
}

func (helper DCBConstitutionHelper) GetOldNationalWelfare(chain *BlockChain) int32 {
	return chain.BestState[0].BestBlock.Header.DCBConstitution.CurrentDCBNationalWelfare
}

func (helper GOVConstitutionHelper) GetOldNationalWelfare(chain *BlockChain) int32 {
	return chain.BestState[0].BestBlock.Header.GOVConstitution.CurrentGOVNationalWelfare
}

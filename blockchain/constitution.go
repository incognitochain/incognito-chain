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

func (DCBConstitutionHelper) GetConstitutionEndedBlockHeight(blockgen *BlkTmplGenerator, chainID byte) uint32 {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastDCBConstitution := BestBlock.Header.DCBConstitution
	return lastDCBConstitution.StartedBlockHeight + lastDCBConstitution.ExecuteDuration
}

func (GOVConstitutionHelper) GetConstitutionEndedBlockHeight(blockgen *BlkTmplGenerator, chainID byte) uint32 {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastGOVConstitution := BestBlock.Header.GOVConstitution
	return lastGOVConstitution.StartedBlockHeight + lastGOVConstitution.ExecuteDuration
}

func (DCBConstitutionHelper) GetStartedNormalVote(blockgen *BlkTmplGenerator, chainID byte) uint32 {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastDCBConstitution := BestBlock.Header.DCBConstitution
	return lastDCBConstitution.StartedBlockHeight - common.EncryptionPhaseDuration
}

func (DCBConstitutionHelper) CheckSubmitProposalType(tx metadata.Transaction) bool {
	return tx.GetMetadataType() == metadata.SubmitDCBProposalMeta
}

func (DCBConstitutionHelper) CheckVotingProposalType(tx metadata.Transaction) bool {
	return tx.GetMetadataType() == metadata.VoteDCBProposalMeta
}

func (DCBConstitutionHelper) GetAmountVoteTokenOfTx(tx metadata.Transaction) uint64 {
	return tx.(*transaction.TxCustomToken).GetAmountOfVote()
}

func (GOVConstitutionHelper) GetStartedNormalVote(blockgen *BlkTmplGenerator, chainID byte) uint32 {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastGOVConstitution := BestBlock.Header.GOVConstitution
	return lastGOVConstitution.StartedBlockHeight - common.EncryptionPhaseDuration
}

func (GOVConstitutionHelper) CheckSubmitProposalType(tx metadata.Transaction) bool {
	return tx.GetMetadataType() == metadata.SubmitGOVProposalMeta
}

func (GOVConstitutionHelper) CheckVotingProposalType(tx metadata.Transaction) bool {
	return tx.GetMetadataType() == metadata.VoteGOVProposalMeta
}

func (GOVConstitutionHelper) GetAmountVoteTokenOfTx(tx metadata.Transaction) uint64 {
	return tx.(*transaction.TxCustomToken).GetAmountOfVote()
}

func (DCBConstitutionHelper) TxAcceptProposal(txId *common.Hash, voter metadata.Voter) metadata.Transaction {
	acceptTx := transaction.Tx{
		Metadata: metadata.NewAcceptDCBProposalMetadata(*txId, voter),
	}
	return &acceptTx
}

func (GOVConstitutionHelper) TxAcceptProposal(txId *common.Hash, voter metadata.Voter) metadata.Transaction {
	acceptTx := transaction.Tx{
		Metadata: metadata.NewAcceptGOVProposalMetadata(*txId, voter),
	}
	return &acceptTx
}

func (DCBConstitutionHelper) GetLowerCaseBoardType() string {
	return "dcb"
}

func (GOVConstitutionHelper) GetLowerCaseBoardType() string {
	return "gov"
}

func (DCBConstitutionHelper) CreatePunishDecryptTx(pubKey []byte) metadata.Metadata {
	return metadata.NewPunishDCBDecryptMetadata(pubKey)
}

func (GOVConstitutionHelper) CreatePunishDecryptTx(pubKey []byte) metadata.Metadata {
	return metadata.NewPunishGOVDecryptMetadata(pubKey)
}

func (DCBConstitutionHelper) GetSealerPubKey(tx metadata.Transaction) [][]byte {
	meta := tx.GetMetadata().(*metadata.SealedLv3DCBBallotMetadata)
	return meta.SealedDCBBallot.LockerPubKeys
}

func (GOVConstitutionHelper) GetSealerPubKey(tx metadata.Transaction) [][]byte {
	meta := tx.GetMetadata().(*metadata.SealedLv3GOVBallotMetadata)
	return meta.SealedGOVBallot.LockerPubKeys
}

func (DCBConstitutionHelper) NewTxRewardProposalSubmitter(blockgen *BlkTmplGenerator, receiverAddress *privacy.PaymentAddress, minerPrivateKey *privacy.SpendingKey) (metadata.Transaction, error) {
	meta := metadata.NewRewardDCBProposalSubmitterMetadata()
	tx := transaction.Tx{}
	err := tx.InitTxSalary(common.RewardProposalSubmitter, receiverAddress, minerPrivateKey, blockgen.chain.config.DataBase, meta)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (GOVConstitutionHelper) NewTxRewardProposalSubmitter(blockgen *BlkTmplGenerator, receiverAddress *privacy.PaymentAddress, minerPrivateKey *privacy.SpendingKey) (metadata.Transaction, error) {
	meta := metadata.NewRewardGOVProposalSubmitterMetadata()
	tx := transaction.Tx{}
	err := tx.InitTxSalary(common.RewardProposalSubmitter, receiverAddress, minerPrivateKey, blockgen.chain.config.DataBase, meta)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (DCBConstitutionHelper) GetPaymentAddressFromSubmitProposalMetadata(tx metadata.Transaction) *privacy.PaymentAddress {
	meta := tx.GetMetadata().(*metadata.SubmitDCBProposalMetadata)
	return &meta.PaymentAddress
}
func (GOVConstitutionHelper) GetPaymentAddressFromSubmitProposalMetadata(tx metadata.Transaction) *privacy.PaymentAddress {
	meta := tx.GetMetadata().(*metadata.SubmitGOVProposalMetadata)
	return &meta.PaymentAddress
}

func (DCBConstitutionHelper) GetPubKeyVoter(blockgen *BlkTmplGenerator, chainID byte) ([]byte, error) {
	bestBlock := blockgen.chain.BestState[chainID].BestBlock
	_, _, _, tx, _ := blockgen.chain.GetTransactionByHash(&bestBlock.Header.DCBConstitution.AcceptProposalTXID)
	meta := tx.GetMetadata().(*metadata.AcceptDCBProposalMetadata)
	return meta.Voter.PubKey, nil
}
func (GOVConstitutionHelper) GetPubKeyVoter(blockgen *BlkTmplGenerator, chainID byte) ([]byte, error) {
	bestBlock := blockgen.chain.BestState[chainID].BestBlock
	_, _, _, tx, _ := blockgen.chain.GetTransactionByHash(&bestBlock.Header.GOVConstitution.AcceptProposalTXID)
	meta := tx.GetMetadata().(*metadata.AcceptGOVProposalMetadata)
	return meta.Voter.PubKey, nil
}

func (DCBConstitutionHelper) GetPrizeProposal() uint32 {
	return uint32(common.Maxint32(GetOracleDCBNationalWelfare(), int32(0)))
}

func (GOVConstitutionHelper) GetPrizeProposal() uint32 {
	return uint32(common.Maxint32(GetOracleGOVNationalWelfare(), int32(0)))
}

func (helper DCBConstitutionHelper) GetTopMostVoteGovernor(blockgen *BlkTmplGenerator) (database.CandidateList, error) {
	return blockgen.chain.config.DataBase.GetTopMostVoteDCBGovernor(blockgen.chain.GetCurrentBoardIndex(helper))
}
func (helper GOVConstitutionHelper) GetTopMostVoteGovernor(blockgen *BlkTmplGenerator) (database.CandidateList, error) {
	return blockgen.chain.config.DataBase.GetTopMostVoteGOVGovernor(blockgen.chain.GetCurrentBoardIndex(helper))
}

func (DCBConstitutionHelper) GetBoardSumToken(blockgen *BlkTmplGenerator) uint64 {
	return blockgen.chain.BestState[0].BestBlock.Header.DCBGovernor.StartAmountToken
}

func (GOVConstitutionHelper) GetBoardSumToken(blockgen *BlkTmplGenerator) uint64 {
	return blockgen.chain.BestState[0].BestBlock.Header.GOVGovernor.StartAmountToken
}

func (DCBConstitutionHelper) GetBoardFund(blockgen *BlkTmplGenerator) uint64 {
	return blockgen.chain.BestState[0].BestBlock.Header.BankFund
}

func (GOVConstitutionHelper) GetBoardFund(blockgen *BlkTmplGenerator) uint64 {
	return blockgen.chain.BestState[0].BestBlock.Header.SalaryFund
}

func (DCBConstitutionHelper) GetTokenID() *common.Hash {
	id := common.Hash(common.DCBTokenID)
	return &id
}

func (GOVConstitutionHelper) GetTokenID() *common.Hash {
	id := common.Hash(common.GOVTokenID)
	return &id
}

func (DCBConstitutionHelper) GetBoard(chain BlockChain) Governor {
	return chain.BestState[0].BestBlock.Header.DCBGovernor
}

func (GOVConstitutionHelper) GetBoard(chain BlockChain) Governor {
	return chain.BestState[0].BestBlock.Header.GOVGovernor
}

func (DCBConstitutionHelper) GetAmountVoteTokenOfBoard(blockgen *BlkTmplGenerator, pubKey []byte, boardIndex uint32) uint64 {
	value, _ := blockgen.chain.config.DataBase.GetDCBVoteTokenAmount(boardIndex, pubKey)
	return uint64(value)
}
func (GOVConstitutionHelper) GetAmountVoteTokenOfBoard(blockgen *BlkTmplGenerator, pubKey []byte, boardIndex uint32) uint64 {
	value, _ := blockgen.chain.config.DataBase.GetGOVVoteTokenAmount(boardIndex, pubKey)
	return uint64(value)
}

func (DCBConstitutionHelper) GetAmountOfVoteToBoard(blockgen *BlkTmplGenerator, candidatePubKey []byte, voterPubKey []byte, boardIndex uint32) uint64 {
	key := lvdb.GetKeyVoteDCBBoardList(boardIndex, candidatePubKey, voterPubKey)
	value, _ := blockgen.chain.config.DataBase.Get(key)
	amount := lvdb.ParseValueVoteDCBBoardList(value)
	return amount
}
func (GOVConstitutionHelper) GetAmountOfVoteToBoard(blockgen *BlkTmplGenerator, candidatePubKey []byte, voterPubKey []byte, boardIndex uint32) uint64 {
	key := lvdb.GetKeyVoteGOVBoardList(boardIndex, candidatePubKey, voterPubKey)
	value, _ := blockgen.chain.config.DataBase.Get(key)
	amount := lvdb.ParseValueVoteGOVBoardList(value)
	return amount
}

func (DCBConstitutionHelper) GetCurrentBoardPubKeys(blockgen *BlkTmplGenerator) [][]byte {
	return blockgen.chain.BestState[0].BestBlock.Header.DCBGovernor.BoardPubKeys
}

func (GOVConstitutionHelper) GetCurrentBoardPubKeys(blockgen *BlkTmplGenerator) [][]byte {
	return blockgen.chain.BestState[0].BestBlock.Header.GOVGovernor.BoardPubKeys
}

func (DCBConstitutionHelper) GetConstitutionInfo(chain BlockChain) ConstitutionInfo {
	return chain.BestState[0].BestBlock.Header.DCBConstitution.ConstitutionInfo
}

func (GOVConstitutionHelper) GetConstitutionInfo(chain BlockChain) ConstitutionInfo {
	return chain.BestState[0].BestBlock.Header.GOVConstitution.ConstitutionInfo
}

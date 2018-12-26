package blockchain

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/voting"
)

type ConstitutionInfo struct {
	StartedBlockHeight uint32
	ExecuteDuration    uint32
	AcceptProposalTXID common.Hash
}

func NewConstitutionInfo(startedBlockHeight uint32, executeDuration uint32, proposalTXID common.Hash) *ConstitutionInfo {
	return &ConstitutionInfo{
		StartedBlockHeight: startedBlockHeight,
		ExecuteDuration:    executeDuration,
		AcceptProposalTXID: proposalTXID,
	}
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

func (DCBConstitutionHelper) GetAmountVoteToken(tx metadata.Transaction) uint64 {
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

func (GOVConstitutionHelper) GetAmountVoteToken(tx metadata.Transaction) uint64 {
	return tx.(*transaction.TxCustomToken).GetAmountOfVote()
}

func (DCBConstitutionHelper) TxAcceptProposal(txId *common.Hash, voter voting.Voter) metadata.Transaction {
	acceptTx := transaction.Tx{
		Metadata: metadata.NewAcceptDCBProposalMetadata(*txId, voter),
	}
	return &acceptTx
}

func (GOVConstitutionHelper) TxAcceptProposal(txId *common.Hash, voter voting.Voter) metadata.Transaction {
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
	tx := transaction.Tx{}
	err := tx.InitTxSalary(common.RewardProposalSubmitter, receiverAddress, minerPrivateKey, blockgen.chain.config.DataBase)
	if err != nil {
		return nil, err
	}
	meta := metadata.NewRewardDCBProposalSubmitterMetadata()
	tx.SetMetadata(meta)
	return &tx, nil
}

func (GOVConstitutionHelper) NewTxRewardProposalSubmitter(blockgen *BlkTmplGenerator, receiverAddress *privacy.PaymentAddress, minerPrivateKey *privacy.SpendingKey) (metadata.Transaction, error) {
	tx := transaction.Tx{}
	err := tx.InitTxSalary(common.RewardProposalSubmitter, receiverAddress, minerPrivateKey, blockgen.chain.config.DataBase)
	if err != nil {
		return nil, err
	}
	meta := metadata.NewRewardGOVProposalSubmitterMetadata()
	tx.SetMetadata(meta)
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

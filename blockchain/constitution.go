package blockchain

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database/lvdb"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/metadata/frombeaconins"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type ConstitutionInfo struct {
	ConstitutionIndex  uint32
	StartedBlockHeight uint64
	ExecuteDuration    uint64
	Explanation        string
	AcceptProposalTXID common.Hash
}

func NewConstitutionInfo(constitutionIndex uint32, startedBlockHeight uint64, executeDuration uint64, explanation string, proposalTXID common.Hash) *ConstitutionInfo {
	return &ConstitutionInfo{
		ConstitutionIndex:  constitutionIndex,
		StartedBlockHeight: startedBlockHeight,
		ExecuteDuration:    executeDuration,
		Explanation:        explanation,
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

func (dcbConstitution DCBConstitution) GetEndedBlockHeight() uint64 {
	return dcbConstitution.StartedBlockHeight + dcbConstitution.ExecuteDuration
}

func (govConstitution GOVConstitution) GetEndedBlockHeight() uint64 {
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

func (helper DCBConstitutionHelper) GetConstitutionEndedBlockHeight(chain *BlockChain) uint64 {
	info := chain.BestState.Beacon.StabilityInfo
	lastDCBConstitution := info.DCBConstitution
	return lastDCBConstitution.StartedBlockHeight + lastDCBConstitution.ExecuteDuration
}

func (helper GOVConstitutionHelper) GetConstitutionEndedBlockHeight(chain *BlockChain) uint64 {
	info := chain.BestState.Beacon.StabilityInfo
	lastGOVConstitution := info.GOVConstitution
	return lastGOVConstitution.StartedBlockHeight + lastGOVConstitution.ExecuteDuration
}

func (helper DCBConstitutionHelper) GetStartedNormalVote(chain *BlockChain) uint64 {
	info := chain.BestState.Beacon.StabilityInfo
	lastDCBConstitution := info.DCBConstitution
	return uint64(lastDCBConstitution.StartedBlockHeight) - uint64(common.EncryptionOnePhraseDuration)
}

func (helper DCBConstitutionHelper) CheckSubmitProposalType(tx metadata.Transaction) bool {
	return tx.GetMetadataType() == metadata.SubmitDCBProposalMeta
}

func (helper DCBConstitutionHelper) GetAmountVoteTokenOfTx(tx metadata.Transaction) uint64 {
	return tx.(*transaction.TxCustomToken).GetAmountOfVote()
}

func (helper GOVConstitutionHelper) GetStartedNormalVote(chain *BlockChain) uint64 {
	info := chain.BestState.Beacon.StabilityInfo
	lastGOVConstitution := info.GOVConstitution
	return uint64(lastGOVConstitution.StartedBlockHeight) - uint64(common.EncryptionOnePhraseDuration)
}

func (helper GOVConstitutionHelper) CheckSubmitProposalType(tx metadata.Transaction) bool {
	return tx.GetMetadataType() == metadata.SubmitGOVProposalMeta
}

func (helper GOVConstitutionHelper) GetAmountVoteTokenOfTx(tx metadata.Transaction) uint64 {
	return tx.(*transaction.TxCustomToken).GetAmountOfVote()
}

func (helper DCBConstitutionHelper) NewAcceptProposalIns(
	txId *common.Hash,
	voter metadata.Voter,
) frombeaconins.InstructionFromBeacon {
	ins := frombeaconins.NewAcceptProposalIns(helper.GetBoardType(), *txId, voter)
	return ins
}

func (helper GOVConstitutionHelper) NewAcceptProposalIns(
	txId *common.Hash,
	voter metadata.Voter,
) frombeaconins.InstructionFromBeacon {
	ins := frombeaconins.NewAcceptProposalIns(helper.GetBoardType(), *txId, voter)
	return ins
}

func (helper DCBConstitutionHelper) GetBoardType() byte {
	return common.DCBBoard
}

func (helper GOVConstitutionHelper) GetBoardType() byte {
	return common.GOVBoard
}

func (helper DCBConstitutionHelper) CreatePunishDecryptIns(paymentAddress *privacy.PaymentAddress) frombeaconins.InstructionFromBeacon {
	return frombeaconins.NewPunishDecryptIns(helper.GetBoardType(), *paymentAddress)
}

func (helper GOVConstitutionHelper) CreatePunishDecryptIns(paymentAddress *privacy.PaymentAddress) frombeaconins.InstructionFromBeacon {
	return frombeaconins.NewPunishDecryptIns(helper.GetBoardType(), *paymentAddress)
}

func (helper DCBConstitutionHelper) GetSealerPaymentAddress(tx metadata.Transaction) []privacy.PaymentAddress {
	meta := tx.GetMetadata().(*metadata.SealedLv3DCBVoteProposalMetadata)
	return meta.SealedLv3VoteProposalMetadata.SealedVoteProposal.LockerPaymentAddress
}

func (helper GOVConstitutionHelper) GetSealerPaymentAddress(tx metadata.Transaction) []privacy.PaymentAddress {
	meta := tx.GetMetadata().(*metadata.SealedLv3GOVVoteProposalMetadata)
	return meta.SealedLv3VoteProposalMetadata.SealedVoteProposal.LockerPaymentAddress
}

func (helper DCBConstitutionHelper) NewRewardProposalSubmitterIns(chain *BlockChain, receiverAddress *privacy.PaymentAddress) (frombeaconins.InstructionFromBeacon, error) {
	return frombeaconins.NewRewardProposalSubmitterIns(receiverAddress, common.RewardProposalSubmitter), nil
}

func (helper GOVConstitutionHelper) NewRewardProposalSubmitterIns(chain *BlockChain, receiverAddress *privacy.PaymentAddress) (frombeaconins.InstructionFromBeacon, error) {
	ins := frombeaconins.NewRewardProposalSubmitterIns(receiverAddress, common.RewardProposalSubmitter)
	return ins, nil
}

func (helper DCBConstitutionHelper) GetPaymentAddressFromSubmitProposalMetadata(tx metadata.Transaction) *privacy.PaymentAddress {
	meta := tx.GetMetadata().(*metadata.SubmitDCBProposalMetadata)
	return &meta.SubmitProposalInfo.PaymentAddress
}
func (helper GOVConstitutionHelper) GetPaymentAddressFromSubmitProposalMetadata(tx metadata.Transaction) *privacy.PaymentAddress {
	meta := tx.GetMetadata().(*metadata.SubmitGOVProposalMetadata)
	return &meta.SubmitProposalInfo.PaymentAddress
}

func (helper DCBConstitutionHelper) GetPaymentAddressVoter(chain *BlockChain) (privacy.PaymentAddress, error) {
	info := chain.BestState.Beacon.StabilityInfo
	_, _, _, tx, _ := chain.GetTransactionByHash(&info.DCBConstitution.AcceptProposalTXID)
	meta := tx.GetMetadata().(*metadata.AcceptDCBProposalMetadata)
	return meta.AcceptProposalMetadata.Voter.PaymentAddress, nil
}
func (helper GOVConstitutionHelper) GetPaymentAddressVoter(chain *BlockChain) (privacy.PaymentAddress, error) {
	info := chain.BestState.Beacon.StabilityInfo
	_, _, _, tx, _ := chain.GetTransactionByHash(&info.GOVConstitution.AcceptProposalTXID)
	meta := tx.GetMetadata().(*metadata.AcceptGOVProposalMetadata)
	return meta.AcceptProposalMetadata.Voter.PaymentAddress, nil
}

func (helper DCBConstitutionHelper) GetPrizeProposal() uint32 {
	return uint32(common.Maxint32(GetOracleDCBNationalWelfare(), int32(0)))
}

func (helper GOVConstitutionHelper) GetPrizeProposal() uint32 {
	return uint32(common.Maxint32(GetOracleGOVNationalWelfare(), int32(0)))
}

func (helper DCBConstitutionHelper) GetBoardSumToken(chain *BlockChain) uint64 {
	return chain.BestState.Beacon.StabilityInfo.DCBGovernor.StartAmountToken
}

func (helper GOVConstitutionHelper) GetBoardSumToken(chain *BlockChain) uint64 {
	return chain.BestState.Beacon.StabilityInfo.GOVGovernor.StartAmountToken
}

func (helper DCBConstitutionHelper) GetBoardFund(chain *BlockChain) uint64 {
	return chain.BestState.Beacon.StabilityInfo.BankFund
}
func (helper GOVConstitutionHelper) GetBoardFund(chain *BlockChain) uint64 {
	return chain.BestState.Beacon.StabilityInfo.BankFund
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
	return chain.BestState.Beacon.StabilityInfo.DCBGovernor
}

func (helper GOVConstitutionHelper) GetBoard(chain *BlockChain) Governor {
	return chain.BestState.Beacon.StabilityInfo.GOVGovernor
}

func (helper DCBConstitutionHelper) GetAmountVoteTokenOfBoard(chain *BlockChain, paymentAddress privacy.PaymentAddress, boardIndex uint32) uint64 {
	value, _ := chain.config.DataBase.GetVoteTokenAmount(helper.GetBoardType(), boardIndex, paymentAddress)
	return uint64(value)
}
func (helper GOVConstitutionHelper) GetAmountVoteTokenOfBoard(chain *BlockChain, paymentAddress privacy.PaymentAddress, boardIndex uint32) uint64 {
	value, _ := chain.config.DataBase.GetVoteTokenAmount(helper.GetBoardType(), boardIndex, paymentAddress)
	return uint64(value)
}

func (helper DCBConstitutionHelper) GetAmountOfVoteToBoard(chain *BlockChain, candidatePaymentAddress privacy.PaymentAddress, voterPaymentAddress privacy.PaymentAddress, boardIndex uint32) uint64 {
	key := lvdb.GetKeyVoteBoardList(helper.GetBoardType(), boardIndex, &candidatePaymentAddress, &voterPaymentAddress)
	value, _ := chain.config.DataBase.Get(key)
	amount := lvdb.ParseValueVoteBoardList(value)
	return amount
}
func (helper GOVConstitutionHelper) GetAmountOfVoteToBoard(chain *BlockChain, candidatePaymentAddress privacy.PaymentAddress, voterPaymentAddress privacy.PaymentAddress, boardIndex uint32) uint64 {
	key := lvdb.GetKeyVoteBoardList(helper.GetBoardType(), boardIndex, &candidatePaymentAddress, &voterPaymentAddress)
	value, _ := chain.config.DataBase.Get(key)
	amount := lvdb.ParseValueVoteBoardList(value)
	return amount
}

func (helper DCBConstitutionHelper) GetCurrentBoardPaymentAddress(chain *BlockChain) []privacy.PaymentAddress {
	return chain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress
}
func (helper GOVConstitutionHelper) GetCurrentBoardPaymentAddress(chain *BlockChain) []privacy.PaymentAddress {
	return chain.BestState.Beacon.StabilityInfo.GOVGovernor.BoardPaymentAddress
}

func (helper DCBConstitutionHelper) GetConstitutionInfo(chain *BlockChain) ConstitutionInfo {
	return chain.BestState.Beacon.StabilityInfo.DCBConstitution.ConstitutionInfo
}
func (helper GOVConstitutionHelper) GetConstitutionInfo(chain *BlockChain) ConstitutionInfo {
	return chain.BestState.Beacon.StabilityInfo.GOVConstitution.ConstitutionInfo
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
	return chain.BestState.Beacon.StabilityInfo.DCBConstitution.CurrentDCBNationalWelfare
}

func (helper GOVConstitutionHelper) GetOldNationalWelfare(chain *BlockChain) int32 {
	return chain.BestState.Beacon.StabilityInfo.GOVConstitution.CurrentGOVNationalWelfare
}

func (helper DCBConstitutionHelper) GetNumberOfGovernor() int32 {
	return common.NumberOfDCBGovernors
}

func (helper GOVConstitutionHelper) GetNumberOfGovernor() int32 {
	return common.NumberOfGOVGovernors
}

func (helper DCBConstitutionHelper) GetProposalTxID(tx metadata.Transaction) (hash *common.Hash) {
	metadataAcceptProposal := tx.GetMetadata().(*metadata.AcceptDCBProposalMetadata)
	proposalTxID := &metadataAcceptProposal.AcceptProposalMetadata.ProposalTXID
	return proposalTxID
}

func (helper DCBConstitutionHelper) GetSubmitProposalInfo(
	tx metadata.Transaction,
) (*metadata.SubmitProposalInfo, error) {
	SubmitProposal := tx.GetMetadata().(*metadata.SubmitDCBProposalMetadata)
	return &SubmitProposal.SubmitProposalInfo, nil
}

func (helper GOVConstitutionHelper) GetProposalTxID(tx metadata.Transaction) (hash *common.Hash) {
	metadataAcceptProposal := tx.GetMetadata().(*metadata.AcceptGOVProposalMetadata)
	proposalTxID := &metadataAcceptProposal.AcceptProposalMetadata.ProposalTXID
	return proposalTxID
}

func (helper GOVConstitutionHelper) GetSubmitProposalInfo(
	tx metadata.Transaction,
) (*metadata.SubmitProposalInfo, error) {
	SubmitProposal := tx.GetMetadata().(*metadata.SubmitGOVProposalMetadata)
	return &SubmitProposal.SubmitProposalInfo, nil
}

func (helper DCBConstitutionHelper) SetNewConstitution(constitutionInfo *ConstitutionInfo, welfare int32, submitProposalTx metadata.Transaction) {
	params := submitProposalTx.GetMetadata().(*metadata.SubmitDCBProposalMetadata).DCBParams
	NewDCBConstitution(constitutionInfo, welfare, &params)
}

func (helper GOVConstitutionHelper) SetNewConstitution(constitutionInfo *ConstitutionInfo, welfare int32, submitProposalTx metadata.Transaction) {
	params := submitProposalTx.GetMetadata().(*metadata.SubmitGOVProposalMetadata).GOVParams
	NewGOVConstitution(constitutionInfo, welfare, &params)
}

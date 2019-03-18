package metadata

import (
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata/fromshardins"
	"github.com/constant-money/constant-chain/privacy"
)

func (normalDCBVoteProposalMetadata *NormalDCBVoteProposalMetadata) GetBoardType() common.BoardType {
	return common.DCBBoard
}

type NormalDCBVoteProposalMetadata struct {
	NormalVoteProposalMetadata component.VoteProposalData
	MetadataBase
}

func (normalDCBVoteProposalMetadata *NormalDCBVoteProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	//return normalDCBVoteProposalMetadata.NormalVoteProposalMetadata.ValidateSanityData(bcr, tx)
	return true, true, nil
}

func (normalDCBVoteProposalMetadata *NormalDCBVoteProposalMetadata) ValidateMetadataByItself() bool {
	//return normalDCBVoteProposalMetadata.NormalVoteProposalMetadata.ValidateMetadataByItself()
	return true
}

func NewNormalDCBVoteProposalMetadata(
	voteProposal component.VoteProposalData,
	lockerPaymentAddress []privacy.PaymentAddress,
	pointerToLv3VoteProposal common.Hash,
) *NormalDCBVoteProposalMetadata {
	return &NormalDCBVoteProposalMetadata{
		NormalVoteProposalMetadata: voteProposal,
		MetadataBase:               *NewMetadataBase(NormalDCBVoteProposalMeta),
	}
}

func (normalDCBVoteProposalMetadata *NormalDCBVoteProposalMetadata) Hash() *common.Hash {
	record := normalDCBVoteProposalMetadata.NormalVoteProposalMetadata.ToBytes()

	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (normalDCBVoteProposalMetadata *NormalDCBVoteProposalMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	//boardType := common.DCBBoard
	//return normalDCBVoteProposalMetadata.NormalVoteProposalMetadata.ValidateTxWithBlockChain(
	//	boardType,
	//	tx,
	//	bcr,
	//	shardID,
	//	db,
	//)
	return true, nil
}

func (normalDCBVoteProposalMetadata *NormalDCBVoteProposalMetadata) BuildReqActions(
	//Hyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy
	//Step 1 hyyyyyyyyyyyyyyyyyyyyyyyy
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	voteProposal := normalDCBVoteProposalMetadata.NormalVoteProposalMetadata
	inst := fromshardins.NewNormalVoteProposalIns(common.DCBBoard, voteProposal)

	instStr, err := inst.GetStringFormat()
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}

type PunishDCBDecryptMetadata struct {
	PunishDecryptMetadata PunishDecryptMetadata
	MetadataBase
}

func NewPunishDCBDecryptMetadata(paymentAddress privacy.PaymentAddress) *PunishDCBDecryptMetadata {
	return &PunishDCBDecryptMetadata{
		PunishDecryptMetadata: PunishDecryptMetadata{
			PaymentAddress: paymentAddress,
		},
		MetadataBase: *NewMetadataBase(PunishDCBDecryptMeta),
	}
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) Hash() *common.Hash {
	record := string(punishDCBDecryptMetadata.PunishDecryptMetadata.ToBytes())
	record += punishDCBDecryptMetadata.MetadataBase.Hash().String()

	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	return true, true, nil
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) ValidateMetadataByItself() bool {
	return true
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) CalculateSize() uint64 {
	return calculateSize(punishDCBDecryptMetadata)
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) BuildReqActions(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	inst := fromshardins.NewPunishDeryptIns(common.DCBBoard)

	instStr, err := inst.GetStringFormat()
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}

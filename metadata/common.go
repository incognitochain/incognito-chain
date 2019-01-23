package metadata

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func ParseMetadata(meta interface{}) (Metadata, error) {
	if meta == nil {
		return nil, nil
	}

	mtTemp := map[string]interface{}{}
	metaInBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metaInBytes, &mtTemp)
	if err != nil {
		return nil, err
	}
	var md Metadata
	switch int(mtTemp["Type"].(float64)) {
	case BuyFromGOVRequestMeta:
		md = &BuySellRequest{}

	case BuyBackRequestMeta:
		md = &BuyBackRequest{}

	case BuyFromGOVResponseMeta:
		md = &BuySellResponse{}

	case BuyBackResponseMeta:
		md = &BuyBackResponse{}

	case IssuingRequestMeta:
		md = &IssuingRequest{}

	case IssuingResponseMeta:
		md = &IssuingResponse{}

	case ContractingRequestMeta:
		md = &ContractingRequest{}

	case OracleFeedMeta:
		md = &OracleFeed{}

	case OracleRewardMeta:
		md = &OracleReward{}

	case RefundMeta:
		md = &Refund{}

	case UpdatingOracleBoardMeta:
		md = &UpdatingOracleBoard{}

	case MultiSigsRegistrationMeta:
		md = &MultiSigsRegistration{}

	case MultiSigsSpendingMeta:
		md = &MultiSigsSpending{}

	case WithSenderAddressMeta:
		md = &WithSenderAddress{}

	case LoanRequestMeta:
		md = &LoanRequest{}

	case LoanResponseMeta:
		md = &LoanResponse{}

	case LoanWithdrawMeta:
		md = &LoanWithdraw{}

	case LoanPaymentMeta:
		md = &LoanPayment{}

	case LoanUnlockMeta:
		md = &LoanUnlock{}

	case VoteDCBBoardMeta:
		md = &VoteDCBBoardMetadata{}

	case VoteGOVBoardMeta:
		md = &VoteGOVBoardMetadata{}
	case SubmitDCBProposalMeta:
		md = &SubmitDCBProposalMetadata{}
	case SubmitGOVProposalMeta:
		md = &SubmitGOVProposalMetadata{}

	case SealedLv3DCBVoteProposalMeta:
		md = &SealedLv3DCBVoteProposalMetadata{}
	case SealedLv2DCBVoteProposalMeta:
		md = &SealedLv2DCBVoteProposalMetadata{}
	case SealedLv1DCBVoteProposalMeta:
		md = &SealedLv1DCBVoteProposalMetadata{}
	case SealedLv3GOVVoteProposalMeta:
		md = &SealedLv3GOVVoteProposalMetadata{}
	case SealedLv2GOVVoteProposalMeta:
		md = &SealedLv2GOVVoteProposalMetadata{}
	case SealedLv1GOVVoteProposalMeta:
		md = &SealedLv1GOVVoteProposalMetadata{}
	case NormalDCBVoteProposalFromOwnerMeta:
		md = &NormalDCBVoteProposalFromOwnerMetadata{}
	case NormalDCBVoteProposalFromSealerMeta:
		md = &NormalDCBVoteProposalFromSealerMetadata{}
	case NormalGOVVoteProposalFromOwnerMeta:
		md = &NormalGOVVoteProposalFromOwnerMetadata{}
	case NormalGOVVoteProposalFromSealerMeta:
		md = &NormalGOVVoteProposalFromSealerMetadata{}
	case AcceptDCBProposalMeta:
		md = &AcceptDCBProposalMetadata{}
	case AcceptGOVProposalMeta:
		md = &AcceptGOVProposalMetadata{}
	case AcceptDCBBoardMeta:
		md = &AcceptDCBBoardMetadata{}
	case AcceptGOVBoardMeta:
		md = &AcceptGOVBoardMetadata{}

	default:
		return nil, errors.Errorf("Could not parse metadata with type: %d", int(mtTemp["Type"].(float64)))
	}

	err = json.Unmarshal(metaInBytes, &md)
	if err != nil {
		return nil, err
	}
	return md, nil
}

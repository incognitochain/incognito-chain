package metadata

import (
	"encoding/json"
	"fmt"

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

	case BuyGOVTokenRequestMeta:
		md = &BuyGOVTokenRequest{}

	case ResponseBaseMeta:
		md = &ResponseBase{}

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

	case CrowdsaleRequestMeta:
		md = &CrowdsaleRequest{}
	case CrowdsalePaymentMeta:
		md = &CrowdsalePayment{}

	case SubmitDCBProposalMeta:
		md = &SubmitDCBProposalMetadata{}
	case VoteDCBBoardMeta:
		md = &VoteDCBBoardMetadata{}
	case AcceptDCBProposalMeta:
		md = &AcceptDCBProposalMetadata{}
	case AcceptDCBBoardMeta:
		md = &AcceptDCBBoardMetadata{}

	case SubmitGOVProposalMeta:
		md = &SubmitGOVProposalMetadata{}
	case VoteGOVBoardMeta:
		md = &VoteGOVBoardMetadata{}
	case AcceptGOVProposalMeta:
		md = &AcceptGOVProposalMetadata{}
	case AcceptGOVBoardMeta:
		md = &AcceptGOVBoardMetadata{}

	case SendInitDCBVoteTokenMeta:
		md = &SendInitDCBVoteTokenMetadata{}
	case SendInitGOVVoteTokenMeta:
		md = &SendInitGOVVoteTokenMetadata{}
	case SealedLv1DCBVoteProposalMeta:
		md = &SealedLv1DCBVoteProposalMetadata{}
	case SealedLv2DCBVoteProposalMeta:
		md = &SealedLv2DCBVoteProposalMetadata{}
	case SealedLv3DCBVoteProposalMeta:
		md = &SealedLv3DCBVoteProposalMetadata{}
	case NormalDCBVoteProposalFromSealerMeta:
		md = &NormalDCBVoteProposalFromSealerMetadata{}
	case NormalDCBVoteProposalFromOwnerMeta:
		md = &NormalDCBVoteProposalFromOwnerMetadata{}
	case SealedLv1GOVVoteProposalMeta:
		md = &SealedLv1GOVVoteProposalMetadata{}
	case SealedLv2GOVVoteProposalMeta:
		md = &SealedLv2GOVVoteProposalMetadata{}
	case SealedLv3GOVVoteProposalMeta:
		md = &SealedLv3GOVVoteProposalMetadata{}
	case NormalGOVVoteProposalFromSealerMeta:
		md = &NormalGOVVoteProposalFromSealerMetadata{}
	case NormalGOVVoteProposalFromOwnerMeta:
		md = &NormalGOVVoteProposalFromOwnerMetadata{}
	case RewardProposalWinnerMeta:
		md = &RewardProposalWinnerMetadata{}
	case RewardDCBProposalSubmitterMeta:
		md = &RewardDCBProposalSubmitterMetadata{}
	case RewardGOVProposalSubmitterMeta:
		md = &RewardGOVProposalSubmitterMetadata{}
	case RewardShareOldDCBBoardMeta:
		md = &ShareRewardOldBoardMetadata{}
	case RewardShareOldGOVBoardMeta:
		md = &ShareRewardOldBoardMetadata{}
	case PunishDCBDecryptMeta:
		md = &PunishDCBDecryptMetadata{}
	case PunishGOVDecryptMeta:
		md = &PunishGOVDecryptMetadata{}

	case ShardStakingMeta:
		md = &StakingMetadata{}
	case BeaconStakingMeta:
		md = &StakingMetadata{}

	default:
		fmt.Printf("[db] meta: %+v\n", meta)
		return nil, errors.Errorf("Could not parse metadata with type: %d", int(mtTemp["Type"].(float64)))
	}

	err = json.Unmarshal(metaInBytes, &md)
	if err != nil {
		return nil, err
	}
	return md, nil
}

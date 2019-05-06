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

	case ContractingResponseMeta:
		md = &ResponseBase{}

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

	case ShardBlockSalaryResponseMeta:
		md = &ShardBlockSalaryRes{}

	case CrowdsaleRequestMeta:
		md = &CrowdsaleRequest{}
	case CrowdsalePaymentMeta:
		md = &CrowdsalePayment{}

	case TradeActivationMeta:
		md = &TradeActivation{}

	case SubmitDCBProposalMeta:
		md = &SubmitDCBProposalMetadata{}
	case VoteDCBBoardMeta:
		md = &VoteDCBBoardMetadata{}
	case SubmitGOVProposalMeta:
		md = &SubmitGOVProposalMetadata{}
	case VoteGOVBoardMeta:
		md = &VoteGOVBoardMetadata{}
	case RewardProposalWinnerMeta:
		md = &RewardProposalWinnerMetadata{}
	case ShareRewardOldDCBBoardMeta:
		md = &ShareRewardOldBoardMetadata{}
	case ShareRewardOldGOVBoardMeta:
		md = &ShareRewardOldBoardMetadata{}
	case DCBVoteProposalMeta:
		md = &DCBVoteProposalMetadata{}
	case GOVVoteProposalMeta:
		md = &GOVVoteProposalMetadata{}
	case SendBackTokenVoteBoardFailMeta:
		md = &SendBackTokenVoteBoardFailMetadata{}

	case ShardStakingMeta:
		md = &StakingMetadata{}
	case BeaconStakingMeta:
		md = &StakingMetadata{}

	case RewardDCBProposalSubmitterMeta:
		md = &RewardDCBProposalSubmitterMetadata{}

	case RewardGOVProposalSubmitterMeta:
		md = &RewardGOVProposalSubmitterMetadata{}

	case RewardDCBProposalVoterMeta:
		md = &RewardDCBProposalVoterMetadata{}

	case RewardGOVProposalVoterMeta:
		md = &RewardGOVProposalVoterMetadata{}

	default:
		fmt.Printf("[db] parse meta err: %+v\n", meta)
		return nil, errors.Errorf("Could not parse metadata with type: %d", int(mtTemp["Type"].(float64)))
	}

	err = json.Unmarshal(metaInBytes, &md)
	if err != nil {
		return nil, err
	}
	return md, nil
}

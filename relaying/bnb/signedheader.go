package bnb

import (
	"encoding/hex"
	"errors"
	"github.com/tendermint/tendermint/types"
	"strings"
)

func NewSignedHeader (h *types.Header, lastCommit *types.Commit) *types.SignedHeader{
	return &types.SignedHeader{
		Header: h,
		Commit: lastCommit,
	}
}

func VerifySignature(sh *types.SignedHeader, chainID string) error {
	validatorMap := map[string]*types.Validator{}
	totalVotingPowerParam := uint64(0)
	if chainID == TestnetBNBChainID {
		validatorMap = validatorsTestnet
		totalVotingPowerParam = TestnetTotalVotingPowers
	} else if chainID == MainnetBNBChainID {
		validatorMap = validatorsMainnet
		totalVotingPowerParam = MainnetTotalVotingPowers
	}

	signedValidators := map[string]bool{}
	sigs := sh.Commit.Precommits
	totalVotingPower := int64(0)
	// get vote from commit sig
	for i, sig := range sigs {
		if sig == nil {
			continue
		}
		vote := sh.Commit.GetVote(i)
		if vote != nil {
			validateAddressStr := strings.ToUpper(hex.EncodeToString(vote.ValidatorAddress))
			// check duplicate vote
			if !signedValidators[validateAddressStr] {
				signedValidators[validateAddressStr] = true
				err := vote.Verify(chainID, validatorMap[validateAddressStr].PubKey)
				if err != nil {
					Logger.log.Errorf("Invalid signature index %v - %v\n", i, err)
					continue
				}
				totalVotingPower += validatorMap[validateAddressStr].VotingPower
			} else {
				Logger.log.Errorf("Duplicate signature from the same validator %v\n", validateAddressStr)
			}
		}
	}

	// not greater than 2/3 voting power
	if totalVotingPower <= int64(totalVotingPowerParam) * 2 / 3 {
		return NewBNBRelayingError(InvalidSignatureSignedHeaderErr, errors.New("not greater than 2/3 voting power"))
	}

	return nil
}

func VerifySignedHeader(sh *types.SignedHeader, chainID string) (bool, error){
	err := sh.ValidateBasic(chainID)
	if err != nil {
		return false, NewBNBRelayingError(InvalidBasicSignedHeaderErr, err)
	}

	err2 := VerifySignature(sh, chainID)
	if err2 != nil {
		return false, err2
	}

	return true, nil
}
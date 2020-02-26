package relaying

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/incognitochain/incognito-chain/relaying"
	"github.com/tendermint/tendermint/types"
	"strings"
)

type HeaderChain struct {
	HeaderChain []*types.Header
	// UnconfirmedHeaders contains header blocks that aren't committed by validator set in the next header block
	UnconfirmedHeaders []*types.Header
}

// ReceiveNewHeader receives new header and last commit for the previous header block
func (hc *HeaderChain) ReceiveNewHeader(h *types.Header, lastCommit *types.Commit) (bool, *BNBRelayingError) {
	// h is the first header block
	if len(hc.HeaderChain) == 0 && lastCommit == nil {
		// just append into hc.UnconfirmedHeaders
		hc.UnconfirmedHeaders = append(hc.UnconfirmedHeaders, h)
		return true, nil
	}

	if len(hc.HeaderChain) > 0 && lastCommit == nil {
		return false, NewBNBRelayingError(InvalidNewHeaderErr, errors.New("last commit is nil"))
	}

	// verify lastCommit
	if !bytes.Equal(h.LastCommitHash , lastCommit.Hash()){
		return false, NewBNBRelayingError(InvalidBasicSignedHeaderErr, errors.New("invalid last commit hash"))
	}

	// case 1: h is the next block header of the latest block header in HeaderChain
	if len(hc.HeaderChain) > 0 {
		// get the latest committed block header
		latestHeader := hc.HeaderChain[len(hc.HeaderChain) - 1]
		latestHeaderBlockID := latestHeader.Hash()

		// check last blockID
		if bytes.Equal(h.LastBlockID.Hash.Bytes(), latestHeaderBlockID) && h.Height == latestHeader.Height + 1{
			// create new signed header and verify
			// add to UnconfirmedHeaders list
			newSignedHeader := NewSignedHeader(latestHeader, lastCommit)
			isValid, err := VerifySignedHeader(newSignedHeader)
			if isValid && err == nil{
				hc.UnconfirmedHeaders = append(hc.UnconfirmedHeaders, h)
				return true, nil
			}

			return false, NewBNBRelayingError(InvalidNewHeaderErr, err)
		}
	}

	// case2 : h is the next block header of one of block headers in UnconfirmedHeaders
	if len(hc.UnconfirmedHeaders) > 0 {
		for _, uh := range hc.UnconfirmedHeaders {
			if bytes.Equal(h.LastBlockID.Hash.Bytes(), uh.Hash())  && h.Height == uh.Height + 1 {
				// create new signed header and verify
				// append uh to hc.HeaderChain,
				// clear all UnconfirmedHeaders => append h to UnconfirmedHeaders
				newSignedHeader := NewSignedHeader(uh, lastCommit)
				isValid, err := VerifySignedHeader(newSignedHeader)
				if isValid && err == nil{
					hc.HeaderChain = append(hc.HeaderChain, uh)
					hc.UnconfirmedHeaders = []*types.Header{h}
					return true, nil
				}
				return false, NewBNBRelayingError(InvalidNewHeaderErr, err)
			}
		}
	}

	return true, nil
}

func NewSignedHeader (h *types.Header, lastCommit *types.Commit) *types.SignedHeader{
	sh := new(types.SignedHeader)
	sh.Header = h
	sh.Commit = lastCommit

	return sh
}

func VerifySignature(sh *types.SignedHeader, chainID string) *BNBRelayingError {
	signedValidator := map[string]bool{}
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
			if !signedValidator[validateAddressStr] {
				signedValidator[validateAddressStr] = true
				err := vote.Verify(chainID, validatorMap[validateAddressStr].PubKey)
				if err != nil {
					relaying.Logger.Log.Errorf("Invalid signature index %v\n", i)
					continue
				}
				totalVotingPower += ValidatorVotingPowers[i]
			} else {
				relaying.Logger.Log.Errorf("Duplicate signature from the same validator %v\n", validateAddressStr)
			}
		}
	}

	// not greater than 2/3 voting power
	if totalVotingPower <= TotalVotingPowers * 2 / 3 {
		return NewBNBRelayingError(InvalidSignatureSignedHeaderErr, errors.New("not greater than 2/3 voting power"))
	}

	return nil
}

func VerifySignedHeader(sh *types.SignedHeader) (bool, *BNBRelayingError){
	err := sh.ValidateBasic(BNBChainID)
	if err != nil {
		return false, NewBNBRelayingError(InvalidBasicSignedHeaderErr, err)
	}

	err2 := VerifySignature(sh, BNBChainID)
	if err2 != nil {
		return false, err2
	}

	return true, nil
}


type LatestHeaderChain struct {
	LatestHeader *types.Header
	// UnconfirmedHeaders contains header blocks that aren't committed by validator set in the next header block
	UnconfirmedHeaders []*types.Header
}

// ReceiveNewHeader receives new header and last commit for the previous header block
func (hc *LatestHeaderChain) ReceiveNewHeader(h *types.Header, lastCommit *types.Commit) (bool, *BNBRelayingError) {
	// h is the first header block
	if hc.LatestHeader == nil && lastCommit == nil {
		// just append into hc.UnconfirmedHeaders
		hc.UnconfirmedHeaders = append(hc.UnconfirmedHeaders, h)
		return true, nil
	}

	if hc.LatestHeader != nil && lastCommit == nil {
		return false, NewBNBRelayingError(InvalidNewHeaderErr, errors.New("last commit is nil"))
	}

	// verify lastCommit
	if !bytes.Equal(h.LastCommitHash , lastCommit.Hash()){
		return false, NewBNBRelayingError(InvalidBasicSignedHeaderErr, errors.New("invalid last commit hash"))
	}

	// case 1: h is the next block header of the latest block header in HeaderChain
	if hc.LatestHeader != nil {
		// get the latest committed block header
		latestHeader := hc.LatestHeader
		latestHeaderBlockID := latestHeader.Hash()

		// check last blockID
		if bytes.Equal(h.LastBlockID.Hash.Bytes(), latestHeaderBlockID) && h.Height == latestHeader.Height + 1{
			// create new signed header and verify
			// add to UnconfirmedHeaders list
			newSignedHeader := NewSignedHeader(latestHeader, lastCommit)
			isValid, err := VerifySignedHeader(newSignedHeader)
			if isValid && err == nil{
				// todo: avoid duplicating unconfirmed headers
				hc.UnconfirmedHeaders = append(hc.UnconfirmedHeaders, h)
				return true, nil
			}

			return false, NewBNBRelayingError(InvalidNewHeaderErr, err)
		}
	}

	// case2 : h is the next block header of one of block headers in UnconfirmedHeaders
	if len(hc.UnconfirmedHeaders) > 0 {
		for _, uh := range hc.UnconfirmedHeaders {
			if bytes.Equal(h.LastBlockID.Hash.Bytes(), uh.Hash())  && h.Height == uh.Height + 1 {
				// create new signed header and verify
				// append uh to hc.HeaderChain,
				// clear all UnconfirmedHeaders => append h to UnconfirmedHeaders
				newSignedHeader := NewSignedHeader(uh, lastCommit)
				isValid, err := VerifySignedHeader(newSignedHeader)
				if isValid && err == nil{
					hc.LatestHeader = uh
					hc.UnconfirmedHeaders = []*types.Header{h}
					return true, nil
				}
				return false, NewBNBRelayingError(InvalidNewHeaderErr, err)
			}
		}
	}

	return true, nil
}


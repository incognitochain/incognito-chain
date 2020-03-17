package relaying

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/tendermint/tendermint/types"
	"strings"
)

type HeaderChain struct {
	HeaderChain []*types.Header
	// UnconfirmedHeaders contains header blocks that aren't committed by validator set in the next header block
	UnconfirmedHeaders []*types.Header
}

// ReceiveNewHeader receives new header and last commit for the previous header block
func (hc *HeaderChain) ReceiveNewHeader(h *types.Header, lastCommit *types.Commit, chainID string) (*HeaderChain, bool, *BNBRelayingError) {
	latestHeaderChain := new(LatestHeaderChain)

	latestHeaderChain.UnconfirmedHeaders = hc.UnconfirmedHeaders
	if len(hc.HeaderChain) > 0 {
		latestHeaderChain.LatestHeader = hc.HeaderChain[len(hc.HeaderChain) - 1]
	}

	latestHeaderChain, isValid, err := latestHeaderChain.ReceiveNewHeader(h, lastCommit, chainID)
	if isValid {
		hc.UnconfirmedHeaders = latestHeaderChain.UnconfirmedHeaders
		if latestHeaderChain.LatestHeader != nil && latestHeaderChain.LatestHeader.Height == int64(len(hc.HeaderChain)) + 1 {
			hc.HeaderChain = append(hc.HeaderChain, latestHeaderChain.LatestHeader)
		}
	}

	return hc, isValid, err
}

func NewSignedHeader (h *types.Header, lastCommit *types.Commit) *types.SignedHeader{
	return &types.SignedHeader{
		Header: h,
		Commit: lastCommit,
	}
}

func VerifySignature(sh *types.SignedHeader, chainID string) *BNBRelayingError {
	validatorMap := validatorsMainnet
	validatorVotingPowers := MainnetValidatorVotingPowers
	totalVotingPowerParam := MainnetTotalVotingPowers

	if chainID == TestnetBNBChainID {
		validatorMap = validatorsTestnet
		validatorVotingPowers = ValidatorVotingPowersTestnet
		totalVotingPowerParam = TestnetTotalVotingPowers
	}

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
				Logger.log.Errorf("validatorMap: %v\n", validatorMap)
				Logger.log.Errorf("validateAddressStr: %v\n", validateAddressStr)
				Logger.log.Errorf("validatorMap[validateAddressStr]: %v\n", validatorMap[validateAddressStr])
				Logger.log.Errorf("ChainId : %v\n", chainID)
				//fmt.Printf("validatorMap[validateAddressStr].PubKey: %v\n", validatorMap[validateAddressStr].PubKey)
				err := vote.Verify(chainID, validatorMap[validateAddressStr].PubKey)
				if err != nil {
					Logger.log.Errorf("Invalid signature index %v\n", i)
					continue
				}
				totalVotingPower += validatorVotingPowers[i]
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

func VerifySignedHeader(sh *types.SignedHeader, chainID string) (bool, *BNBRelayingError){
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


type LatestHeaderChain struct {
	LatestHeader *types.Header
	// UnconfirmedHeaders contains header blocks that aren't committed by validator set in the next header block
	UnconfirmedHeaders []*types.Header
}

func appendHeaderToUnconfirmedHeaders (header *types.Header, unconfirmedHeaders []*types.Header) ([]*types.Header, error){
	hHash := header.Hash().Bytes()
	for _, unconfirmedHeader := range unconfirmedHeaders {
		if bytes.Equal(unconfirmedHeader.Hash().Bytes(), hHash) {
			Logger.log.Errorf("Header is existed %v\n", hHash)
			return unconfirmedHeaders, fmt.Errorf("Header is existed %v\n", hHash)
		}
	}
	unconfirmedHeaders = append(unconfirmedHeaders, header)
	return unconfirmedHeaders, nil
}

// ReceiveNewHeader receives new header and last commit for the previous header block
func (hc *LatestHeaderChain) ReceiveNewHeader(h *types.Header, lastCommit *types.Commit, chainID string) (*LatestHeaderChain, bool, *BNBRelayingError) {
	genesisBlockHeight := int64(0)
	if chainID == MainnetBNBChainID {
		genesisBlockHeight = MainnetGenesisBlockHeight
	} else if chainID == TestnetBNBChainID {
		genesisBlockHeight = TestnetGenesisBlockHeight
	}

	var err2 error
	// h is the first header block
	if hc.LatestHeader == nil && len(hc.UnconfirmedHeaders) == 0 {
		// check h.height = 1
		if h.Height != genesisBlockHeight {
			Logger.log.Errorf("[ReceiveNewHeader] the genesis block height must be %v", genesisBlockHeight)
			return hc, false, NewBNBRelayingError(InvalidNewHeaderErr, errors.New("the height of header must be 1"))
		}

		// just append into hc.UnconfirmedHeaders
		hc.UnconfirmedHeaders, err2 = appendHeaderToUnconfirmedHeaders(h, hc.UnconfirmedHeaders)
		if err2 != nil {
			return hc, false, NewBNBRelayingError(InvalidNewHeaderErr, err2)
		}
		return hc, true, nil
	}

	if hc.LatestHeader != nil && lastCommit == nil {
		Logger.log.Errorf("[ReceiveNewHeader] last commit is nil")
		return hc, false, NewBNBRelayingError(InvalidNewHeaderErr, errors.New("last commit is nil"))
	}

	// verify lastCommit
	if !bytes.Equal(h.LastCommitHash , lastCommit.Hash()){
		Logger.log.Errorf("[ReceiveNewHeader] invalid last commit hash")
		return hc, false, NewBNBRelayingError(InvalidBasicSignedHeaderErr, errors.New("invalid last commit hash"))
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
			isValid, err := VerifySignedHeader(newSignedHeader, chainID)
			if isValid && err == nil{
				Logger.log.Errorf("[ReceiveNewHeader] Case 1 new confirmed header %v\n", h.Height)
				hc.UnconfirmedHeaders, err2 = appendHeaderToUnconfirmedHeaders(h, hc.UnconfirmedHeaders)
				if err2 != nil {
					return hc, false, NewBNBRelayingError(InvalidNewHeaderErr, err2)
				}
				return hc, true, nil
			}

			Logger.log.Errorf("[ReceiveNewHeader] invalid new signed header %v", err)
			return hc, false, NewBNBRelayingError(InvalidNewHeaderErr, err)
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
				isValid, err := VerifySignedHeader(newSignedHeader, chainID)
				if isValid && err == nil{
					hc.LatestHeader = uh
					hc.UnconfirmedHeaders = []*types.Header{h}
					Logger.log.Errorf("[ReceiveNewHeader] Case 2 new unconfirmed header %v\n", h.Height)
					return hc, true, nil
				}

				Logger.log.Errorf("[ReceiveNewHeader] invalid new signed header %v", err)
				return hc, false, NewBNBRelayingError(InvalidNewHeaderErr, err)
			}
		}
	}

	Logger.log.Errorf("New header is invalid")
	return hc, false, NewBNBRelayingError(InvalidNewHeaderErr, nil)
}

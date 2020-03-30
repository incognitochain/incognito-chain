package bnb

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tendermint/tendermint/types"
	"strings"
)

var TestnetGenesisHeaderStr = "eyJIZWFkZXIiOnsidmVyc2lvbiI6eyJibG9jayI6MTAsImFwcCI6MH0sImNoYWluX2lkIjoiQmluYW5jZS1EZXYiLCJoZWlnaHQiOjEwMDAsInRpbWUiOiIyMDIwLTAzLTIwVDEwOjE5OjE4LjE2MDYxWiIsIm51bV90eHMiOjAsInRvdGFsX3R4cyI6MiwibGFzdF9ibG9ja19pZCI6eyJoYXNoIjoiRjVBNjNFOEJBODg5NjNGRTU0RDQwRjVEMzRBRkM5Q0JENUUzNTI3NEY0QTNCNTZBMTZBOEY2ODg2RUU4MENCQyIsInBhcnRzIjp7InRvdGFsIjoxLCJoYXNoIjoiMTUzRDM4MzBEMkE2MzkzODM4QTFBOTBBMjhENTQ2QzEyRTc4QjZGNTAzMjQ1RDlERjIxRkFCNkUwRUMyNDlDNSJ9fSwibGFzdF9jb21taXRfaGFzaCI6IkZFRUQzMzRDRTA2RjZGODNFMzgxQTUzQUUxRkZDNDk5NDY2RDZBMzVCN0NDQjAwQjgyNDgxMURBMDBDMTU5MjMiLCJkYXRhX2hhc2giOiIiLCJ2YWxpZGF0b3JzX2hhc2giOiI0NUVENUJFQTlCRTY1Q0M4NjUwQUNCRTlDQ0I0NzhDNENERDMwMkJEMjY1OEFBQjE4RUM3MzE5NDE3ODI4MEU1IiwibmV4dF92YWxpZGF0b3JzX2hhc2giOiI0NUVENUJFQTlCRTY1Q0M4NjUwQUNCRTlDQ0I0NzhDNENERDMwMkJEMjY1OEFBQjE4RUM3MzE5NDE3ODI4MEU1IiwiY29uc2Vuc3VzX2hhc2giOiIyOTREOEZCRDBCOTRCNzY3QTdFQkE5ODQwRjI5OUEzNTg2REE3RkU2QjVERUFEM0I3RUVDQkExOTNDNDAwRjkzIiwiYXBwX2hhc2giOiJENkU4QTAyRUFCQzQ3NDQ1RTJEMUI1MUU3NDYwOTc2N0MxNTBBMDBFMDM2QkVGNEI0QzA4RkJDODhGQzA3QzIwIiwibGFzdF9yZXN1bHRzX2hhc2giOiIiLCJldmlkZW5jZV9oYXNoIjoiIiwicHJvcG9zZXJfYWRkcmVzcyI6IjYxNDkzN0E4NUFFMUQyQjRBQzkwNTkzODA5NzYyQTc0MkVFOTEyRUEifSwiTGFzdENvbW1pdCI6eyJibG9ja19pZCI6eyJoYXNoIjoiRjVBNjNFOEJBODg5NjNGRTU0RDQwRjVEMzRBRkM5Q0JENUUzNTI3NEY0QTNCNTZBMTZBOEY2ODg2RUU4MENCQyIsInBhcnRzIjp7InRvdGFsIjoxLCJoYXNoIjoiMTUzRDM4MzBEMkE2MzkzODM4QTFBOTBBMjhENTQ2QzEyRTc4QjZGNTAzMjQ1RDlERjIxRkFCNkUwRUMyNDlDNSJ9fSwicHJlY29tbWl0cyI6W3sidHlwZSI6MiwiaGVpZ2h0Ijo5OTksInJvdW5kIjowLCJibG9ja19pZCI6eyJoYXNoIjoiRjVBNjNFOEJBODg5NjNGRTU0RDQwRjVEMzRBRkM5Q0JENUUzNTI3NEY0QTNCNTZBMTZBOEY2ODg2RUU4MENCQyIsInBhcnRzIjp7InRvdGFsIjoxLCJoYXNoIjoiMTUzRDM4MzBEMkE2MzkzODM4QTFBOTBBMjhENTQ2QzEyRTc4QjZGNTAzMjQ1RDlERjIxRkFCNkUwRUMyNDlDNSJ9fSwidGltZXN0YW1wIjoiMjAyMC0wMy0yMFQxMDoxOToxOC4xNjA2MVoiLCJ2YWxpZGF0b3JfYWRkcmVzcyI6IjYxNDkzN0E4NUFFMUQyQjRBQzkwNTkzODA5NzYyQTc0MkVFOTEyRUEiLCJ2YWxpZGF0b3JfaW5kZXgiOjAsInNpZ25hdHVyZSI6ImRTYVNhcllsRVpUS3pEeWIvRFZPRWV1cUpialplYVZhU3JWZFgzeE93RUFQRVdJd0ROa1Z5eWNEU0FMdGJoNmhUMm1Xd0xFTmNDUUV0RSs3aFRjakNnPT0ifV19fQ=="
var MainnetGenesisHeaderStr = ""

func createGenesisHeaderChain(chainID string) (*types.Block, error) {
	genesisHeaderStr := ""
	if chainID == MainnetBNBChainID {
		genesisHeaderStr = MainnetGenesisHeaderStr
	} else if chainID == TestnetBNBChainID {
		genesisHeaderStr = TestnetGenesisHeaderStr
	} else {
		return nil, errors.New("Invalid network chainID")
	}

	genesisHeaderBytes, err := base64.StdEncoding.DecodeString(genesisHeaderStr)
	if err != nil {
		return nil, errors.New("Can not decode genesis header string")
	}
	var bnbHHeader types.Block
	err = json.Unmarshal(genesisHeaderBytes, &bnbHHeader)
	if err != nil {
		return nil, errors.New("Can not unmarshal genesis header bytes")
	}

	return &bnbHHeader, nil
}

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
	// create genesis header before appending new header
	if hc.LatestHeader == nil && len(hc.UnconfirmedHeaders) == 0 {
		genesisHeader, _ := createGenesisHeaderChain(chainID)
		Logger.log.Errorf("genesisHeader.Header %v\n", genesisHeader.Header)
		hc.LatestHeader = &genesisHeader.Header
		Logger.log.Errorf("hc.LatestHeader %v\n", hc.LatestHeader)
	}

	var err2 error
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
		Logger.log.Errorf("h.LastBlockID.Hash.Bytes(): %v\n", h.LastBlockID.Hash.Bytes())
		Logger.log.Errorf("latestHeader.Hash(): %v\n", latestHeader.Hash().Bytes())

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
					Logger.log.Errorf("[ReceiveNewHeader] Error when append header to unconfirmed headers %v\n", err2)
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

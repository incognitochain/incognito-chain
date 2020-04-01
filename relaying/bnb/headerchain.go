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

func createGenesisHeaderChain(chainID string) (*types.Block, error) {
	genesisBlockStr, err := GetGenesisBNBHeaderStr(chainID)
	if err != nil {
		return nil, err
	}
	genesisBlockBytes, err := base64.StdEncoding.DecodeString(genesisBlockStr)
	if err != nil {
		return nil, errors.New("Can not decode genesis header string")
	}
	var bnbBlock types.Block
	err = json.Unmarshal(genesisBlockBytes, &bnbBlock)
	if err != nil {
		return nil, errors.New("Can not unmarshal genesis header bytes")
	}

	return &bnbBlock, nil
}

func NewSignedHeader (h *types.Header, lastCommit *types.Commit) *types.SignedHeader{
	return &types.SignedHeader{
		Header: h,
		Commit: lastCommit,
	}
}

func VerifySignature(sh *types.SignedHeader, chainID string) *BNBRelayingError {
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
	LatestBlock *types.Block
	// UnconfirmedBlocks contains header blocks that aren't committed by validator set in the next header block
	UnconfirmedBlocks []*types.Block
}

func appendBlockToUnconfirmedBlocks(block *types.Block, unconfirmedBlocks []*types.Block) ([]*types.Block, error){
	hHash := block.Header.Hash().Bytes()
	for _, unconfirmedHeader := range unconfirmedBlocks {
		if bytes.Equal(unconfirmedHeader.Hash().Bytes(), hHash) {
			Logger.log.Errorf("Block is existed %v\n", hHash)
			return unconfirmedBlocks, fmt.Errorf("Header is existed %v\n", hHash)
		}
	}
	unconfirmedBlocks = append(unconfirmedBlocks, block)
	return unconfirmedBlocks, nil
}

// AppendBlock receives new header and last commit for the previous header block
func (hc *LatestHeaderChain) AppendBlock(h *types.Block, chainID string) (*LatestHeaderChain, bool, *BNBRelayingError) {
	// create genesis header before appending new header
	if hc.LatestBlock == nil && len(hc.UnconfirmedBlocks) == 0 {
		genesisBlock, _ := createGenesisHeaderChain(chainID)
		Logger.log.Errorf("genesisBlock: %v\n", genesisBlock)
		hc.LatestBlock = genesisBlock
	}
	Logger.log.Errorf("h: %v\n", h)

	var err2 error
	if hc.LatestBlock != nil && h.LastCommit == nil {
		Logger.log.Errorf("[AppendBlock] last commit is nil")
		return hc, false, NewBNBRelayingError(InvalidNewHeaderErr, errors.New("last commit is nil"))
	}

	// verify lastCommit
	if !bytes.Equal(h.LastCommitHash , h.LastCommit.Hash()){
		Logger.log.Errorf("[AppendBlock] invalid last commit hash")
		return hc, false, NewBNBRelayingError(InvalidBasicSignedHeaderErr, errors.New("invalid last commit hash"))
	}

	// case 1: h is the next block header of the latest block header in HeaderChain
	if hc.LatestBlock != nil {
		// get the latest committed block header
		latestHeader := hc.LatestBlock
		latestHeaderBlockID := latestHeader.Hash()
		Logger.log.Errorf("latestHeader.Hash(): %v\n", latestHeader.Hash().Bytes())
		Logger.log.Errorf("h.LastBlockID.Hash.Bytes(): %v\n", h.LastBlockID.Hash.Bytes())

		// check last blockID
		if bytes.Equal(h.LastBlockID.Hash.Bytes(), latestHeaderBlockID) && h.Height == latestHeader.Height + 1{
			// create new signed header and verify
			// add to UnconfirmedBlocks list
			newSignedHeader := NewSignedHeader(&latestHeader.Header, h.LastCommit)
			isValid, err := VerifySignedHeader(newSignedHeader, chainID)
			if isValid && err == nil{
				Logger.log.Errorf("[AppendBlock] Case 1: Receive new confirmed header %v\n", h.Height)
				hc.UnconfirmedBlocks, err2 = appendBlockToUnconfirmedBlocks(h, hc.UnconfirmedBlocks)
				if err2 != nil {
					Logger.log.Errorf("[AppendBlock] Error when append header to unconfirmed headers %v\n", err2)
					return hc, false, NewBNBRelayingError(InvalidNewHeaderErr, err2)
				}
				return hc, true, nil
			}

			Logger.log.Errorf("[AppendBlock] invalid new signed header %v", err)
			return hc, false, NewBNBRelayingError(InvalidNewHeaderErr, err)
		}
	}

	// case2 : h is the next block header of one of block headers in UnconfirmedBlocks
	if len(hc.UnconfirmedBlocks) > 0 {
		for _, uh := range hc.UnconfirmedBlocks {
			if bytes.Equal(h.LastBlockID.Hash.Bytes(), uh.Hash())  && h.Height == uh.Height + 1 {
				// create new signed header and verify
				// append uh to hc.HeaderChain,
				// clear all UnconfirmedBlocks => append h to UnconfirmedBlocks
				newSignedHeader := NewSignedHeader(&uh.Header, h.LastCommit)
				isValid, err := VerifySignedHeader(newSignedHeader, chainID)
				if isValid && err == nil{
					hc.LatestBlock = uh
					hc.UnconfirmedBlocks = []*types.Block{h}
					Logger.log.Errorf("[AppendBlock] Case 2 new unconfirmed block %v\n", h.Height)
					return hc, true, nil
				}

				Logger.log.Errorf("[AppendBlock] invalid new signed header %v", err)
				return hc, false, NewBNBRelayingError(InvalidNewHeaderErr, err)
			}
		}
	}

	Logger.log.Errorf("New header is invalid")
	return hc, false, NewBNBRelayingError(InvalidNewHeaderErr, nil)
}

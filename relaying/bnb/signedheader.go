package relaying

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	tdmtypes "github.com/tendermint/tendermint/types"
	"strings"
)

type BNBSignedHeader struct {
	Header *BNBHeader
	Commit *tdmtypes.Commit
}

func (sh *BNBSignedHeader) Init (header *BNBHeader, commit *tdmtypes.Commit)(*BNBSignedHeader, error){
	if sh == nil {
		sh = new(BNBSignedHeader)
	}
	sh.Header = header
	sh.Commit = commit
	return sh, nil
}

func (sh *BNBSignedHeader) ValidateBasic(chainID string) error {
	// Make sure the header is consistent with the commit.
	if sh.Header == nil {
		return errors.New("signedHeader missing header")
	}
	if sh.Commit == nil {
		return errors.New("signedHeader missing commit (precommit votes)")
	}

	// Check ChainID.
	if sh.Header.ChainID != chainID {
		return fmt.Errorf("signedHeader belongs to another chain '%s' not '%s'",
			sh.Header.ChainID, chainID)
	}
	// Check Height.
	if sh.Commit.Height != sh.Header.Height {
		return fmt.Errorf("signedHeader header and commit height mismatch: %v vs %v",
			sh.Header.Height, sh.Commit.Height)
	}
	// Check Hash.
	hhash := sh.Header.Hash()
	chash := sh.Commit.BlockID.Hash
	if !bytes.Equal(hhash, chash) {
		return fmt.Errorf("signedHeader commit signs block %X, header is block %X",
			chash, hhash)
	}
	// ValidateBasic on the Commit.
	err := sh.Commit.ValidateBasic()
	if err != nil {
		return errors.Wrap(err, "commit.ValidateBasic failed during SignedHeader.ValidateBasic")
	}
	return nil
}

func (sh *BNBSignedHeader) VerifySignature(chainID string) error {
	signedValidator := map[string]bool{}
	sigs := sh.Commit.Signatures
	totalVotingPower := int64(0)
	// get vote from commit sig
	for i, sig := range sigs {
		if len(sig.Signature) == 0 {
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
					return err
				}
				totalVotingPower += ValidatorVotingPowers[i]
			}
		}
	}

	// not enough 2/3 voting power
	if totalVotingPower < 11000000000000 * 2 / 3 {
		return errors.New("not enough 2/3 voting power")
	}

	return nil
}

func(sh *BNBSignedHeader) Verify() (bool, error){
	err := sh.ValidateBasic(BNBChainID)
	if err != nil {
		return false, err
	}

	err2 := sh.VerifySignature(BNBChainID)
	if err2 != nil {
		return false, err2
	}

	return true, nil
}
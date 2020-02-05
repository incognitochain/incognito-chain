package relaying
//
//import (
//	"bytes"
//	"encoding/hex"
//	"fmt"
//	"github.com/incognitochain/incognito-chain/relaying"
//	"github.com/pkg/errors"
//	tdmtypes "github.com/tendermint/tendermint/types"
//	"strings"
//)
//
//type BNBSignedHeader struct {
//	Header *BNBHeader
//	Commit *tdmtypes.Commit
//}
//
//func (sh *BNBSignedHeader) Init (header *BNBHeader, commit *tdmtypes.Commit)(*BNBSignedHeader, *BNBRelayingError){
//	if sh == nil {
//		sh = new(BNBSignedHeader)
//	}
//	sh.Header = header
//	sh.Commit = commit
//	return sh, nil
//}
//
//func (sh *BNBSignedHeader) ValidateBasic(chainID string) *BNBRelayingError {
//	// Make sure the header is consistent with the commit.
//	if sh.Header == nil {
//		return NewBNBRelayingError(InvalidBasicSignedHeaderErr, errors.New("signedHeader missing header"))
//	}
//	if sh.Commit == nil {
//		return NewBNBRelayingError(InvalidBasicSignedHeaderErr, errors.New("signedHeader missing commit (precommit votes)"))
//	}
//
//	// Check ChainID.
//	if sh.Header.ChainID != chainID {
//		return NewBNBRelayingError(InvalidBasicSignedHeaderErr,
//			fmt.Errorf("signedHeader belongs to another chain '%s' not '%s'",
//			sh.Header.ChainID, chainID))
//	}
//	// Check Height.
//	if sh.Commit.Height != sh.Header.Height {
//		return NewBNBRelayingError(InvalidBasicSignedHeaderErr,
//			fmt.Errorf("signedHeader header and commit height mismatch: %v vs %v",
//				sh.Header.Height, sh.Commit.Height))
//	}
//	// Check Hash.
//	hhash := sh.Header.Hash()
//	chash := sh.Commit.BlockID.Hash
//	if !bytes.Equal(hhash, chash) {
//		return NewBNBRelayingError(InvalidBasicSignedHeaderErr,
//			fmt.Errorf("signedHeader commit signs block %X, header is block %X",
//			chash, hhash))
//	}
//
//	// validateBasic on the Header.
//	err := sh.Header.ValidateBasic()
//	if err != nil {
//		return NewBNBRelayingError(InvalidBasicSignedHeaderErr, err)
//	}
//
//	// ValidateBasic on the Commit.
//	err2 := sh.Commit.ValidateBasic()
//	if err2 != nil {
//		return NewBNBRelayingError(InvalidBasicSignedHeaderErr, err)
//	}
//	return nil
//}
//
//func (sh *BNBSignedHeader) VerifySignature(chainID string) *BNBRelayingError {
//	signedValidator := map[string]bool{}
//	sigs := sh.Commit.Signatures
//	totalVotingPower := int64(0)
//	// get vote from commit sig
//	for i, sig := range sigs {
//		if len(sig.Signature) == 0 {
//			continue
//		}
//		vote := sh.Commit.GetVote(i)
//		if vote != nil {
//			validateAddressStr := strings.ToUpper(hex.EncodeToString(vote.ValidatorAddress))
//			// check duplicate vote
//			if !signedValidator[validateAddressStr] {
//				signedValidator[validateAddressStr] = true
//				err := vote.Verify(chainID, validatorMap[validateAddressStr].PubKey)
//				if err != nil {
//					relaying.Logger.Log.Errorf("Invalid signature index %v\n", i)
//					continue
//				}
//				totalVotingPower += ValidatorVotingPowers[i]
//			} else {
//				relaying.Logger.Log.Errorf("Duplicate signature from the same validator %v\n", validateAddressStr)
//			}
//		}
//	}
//
//	// not greater than 2/3 voting power
//	if totalVotingPower <= TotalVotingPowers * 2 / 3 {
//		return NewBNBRelayingError(InvalidSignatureSignedHeaderErr, errors.New("not greater than 2/3 voting power"))
//	}
//
//	return nil
//}
//
//func(sh *BNBSignedHeader) Verify() (bool, *BNBRelayingError){
//	err := sh.ValidateBasic(BNBChainID)
//	if err != nil {
//		return false, err
//	}
//
//	err2 := sh.VerifySignature(BNBChainID)
//	if err2 != nil {
//		return false, err2
//	}
//
//	return true, nil
//}
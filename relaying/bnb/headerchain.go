package relaying

import (
	"bytes"
	"fmt"
	"github.com/tendermint/tendermint/types"
)

type HeaderChain struct {
	HeaderChain []*BNBHeader
	// prevHeader contains header blocks that aren't committed by validator set in the next header block
	prevHeader []*BNBHeader
}

// ReceiveNewHeader receives new header and last commit for the previous header block
func (hc *HeaderChain) ReceiveNewHeader(h *BNBHeader, lastCommit *types.Commit) (bool, *BNBRelayingError) {
	// h is the first header block
	if len(hc.HeaderChain) == 0 && len(hc.prevHeader) == 0 && lastCommit == nil {
		// just append into hc.prevHeader
		hc.prevHeader = append(hc.prevHeader, h)
		return true, nil
	}

	// todo:
	// verify lastCommit
	if lastCommit != nil {
		tmp1 := h.LastCommitHash
		tmp2 := lastCommit.Hash().Bytes()

		fmt.Printf("tmp1 %v\n", tmp1)
		fmt.Printf("tmp2 %v\n", tmp2)

		//if !bytes.Equal(tmp1 , tmp2){
		//	return false, errors.New("invalid last commit hash")
		//}
	}

	// case 1: h is the next block header of the latest block header in HeaderChain
	if len(hc.HeaderChain) > 0 {
		// get the latest committed block header
		latestHeader := hc.HeaderChain[len(hc.HeaderChain) - 1]
		latestHeaderBlockID := latestHeader.Hash()

		// check last blockID
		if bytes.Equal(h.LastBlockID.Hash.Bytes(), latestHeaderBlockID) && h.Height == latestHeader.Height + 1{
			// create new signed header and verify
			// add to prevHeader list
			sh := new(BNBSignedHeader)
			newSignedHeader, err := sh.Init(latestHeader, lastCommit)
			if err != nil{
				return false, NewBNBRelayingError(InvalidNewHeaderErr, err)
			}
			isValid, err := newSignedHeader.Verify()
			if isValid && err == nil{
				hc.prevHeader = append(hc.prevHeader, h)
				return true, nil
			}

			return false, NewBNBRelayingError(InvalidNewHeaderErr, err)
		}
	}

	// case2 : h is the next block header of one of block headers in prevHeader
	if len(hc.prevHeader) > 0 {
		for _, ph := range hc.prevHeader{
			if bytes.Equal(h.LastBlockID.Hash.Bytes(), ph.Hash())  && h.Height == ph.Height + 1 {
				// create new signed header and verify
				// append ph to hc.HeaderChain,
				// clear all prevHeader => append h to prevHeader
				sh := new(BNBSignedHeader)
				newSignedHeader, err := sh.Init(ph, lastCommit)
				if err != nil{
					return false, NewBNBRelayingError(InvalidNewHeaderErr, err)
				}
				isValid, err := newSignedHeader.Verify()
				if isValid && err == nil{
					hc.HeaderChain = append(hc.HeaderChain, ph)
					hc.prevHeader = []*BNBHeader{h}
					return true, nil
				}
				return false, NewBNBRelayingError(InvalidNewHeaderErr, err)
			}
		}
	}

	return true, nil
}


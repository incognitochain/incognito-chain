package relaying

import (
	"errors"
	"fmt"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/merkle"
	tdmtypes "github.com/tendermint/tendermint/types"

	"time"
)

type Version struct {
	Block uint64 `json:"block,string"`
	App   uint64 `json:"app,string"`
}

type BNBHeader struct {
	// basic block info
	Version  Version   `json:"version"`
	ChainID  string    `json:"chain_id"`
	Height   int64     `json:"height,string"`
	Time     time.Time `json:"time,string"`
	NumTxs   int64     `json:"num_txs,string"`
	TotalTxs int64     `json:"total_txs,string"`
	// prev block info
	LastBlockID tdmtypes.BlockID `json:"last_block_id"`

	// hashes of block data
	LastCommitHash []byte `json:"last_commit_hash"`
	// MerkleRoot of transaction hashes
	DataHash []byte `json:"data_hash"`

	// hashes from the app output from the prev block
	// validators for the current block
	ValidatorsHash []byte `json:"validators_hash"`
	// validators for the next block
	NextValidatorsHash []byte `json:"next_validators_hash"`
	// consensus params for current block
	ConsensusHash []byte `json:"consensus_hash"`
	// state after txs from the previous block
	AppHash []byte `json:"app_hash"`
	// root hash of all results from the txs from the previous block
	LastResultsHash []byte `json:"last_results_hash"`

	// consensus info
	EvidenceHash    []byte `json:"evidence_hash"`    // evidence included in the block
	ProposerAddress []byte `json:"proposer_address"` // original proposer of the block// basic block info
}

func (h *BNBHeader) Hash() []byte {
	if h == nil || len(h.ValidatorsHash) == 0 {
		return nil
	}
	return merkle.SimpleHashFromByteSlices([][]byte{
		cdcEncode(h.Version),
		cdcEncode(h.ChainID),
		cdcEncode(h.Height),
		cdcEncode(h.Time),
		cdcEncode(h.NumTxs),
		cdcEncode(h.TotalTxs),
		cdcEncode(h.LastBlockID),
		cdcEncode(h.LastCommitHash),
		cdcEncode(h.DataHash),
		cdcEncode(h.ValidatorsHash),
		cdcEncode(h.NextValidatorsHash),
		cdcEncode(h.ConsensusHash),
		cdcEncode(h.AppHash),
		cdcEncode(h.LastResultsHash),
		cdcEncode(h.EvidenceHash),
		cdcEncode(h.ProposerAddress),
	})
}

func (h *BNBHeader) ValidateBasic() *BNBRelayingError {
	if h == nil {
		return NewBNBRelayingError(InvalidBasicHeaderErr, errors.New("nil block"))
	}

	if len(h.ChainID) > tdmtypes.MaxChainIDLen {
		return NewBNBRelayingError(InvalidBasicHeaderErr,
			fmt.Errorf("chainID is too long. Max is %d, got %d", tdmtypes.MaxChainIDLen, len(h.ChainID)))
	}

	if h.Height < 0 {
		return NewBNBRelayingError(InvalidBasicHeaderErr, errors.New("negative Header.Height"))
	} else if h.Height == 0 {
		return NewBNBRelayingError(InvalidBasicHeaderErr, errors.New("zero Header.Height"))
	}

	// NOTE: Timestamp validation is subtle and handled elsewhere.

	if err := h.LastBlockID.ValidateBasic(); err != nil {
		return NewBNBRelayingError(InvalidBasicHeaderErr, fmt.Errorf("wrong Header.LastBlockID: %v", err))
	}

	// Validate the last commit and its hash.
	if h.Height > 1 {
		if len(h.LastCommitHash) == 0 {
			return NewBNBRelayingError(InvalidBasicHeaderErr, errors.New("nil LastCommit"))
		}
	}
	if err := tdmtypes.ValidateHash(h.LastCommitHash); err != nil {
		return NewBNBRelayingError(InvalidBasicHeaderErr, fmt.Errorf("wrong Header.LastCommitHash: %v", err))
	}

	// Validate the hash of the transactions.
	// NOTE: b.Data.Txs may be nil, but b.Data.Hash()
	// still works fine
	if err := tdmtypes.ValidateHash(h.DataHash); err != nil {
		return NewBNBRelayingError(InvalidBasicHeaderErr, fmt.Errorf("wrong Header.DataHash: %v", err))
	}

	// Basic validation of hashes related to application data.
	// Will validate fully against state in state#ValidateBlock.
	if err := tdmtypes.ValidateHash(h.ValidatorsHash); err != nil {
		return NewBNBRelayingError(InvalidBasicHeaderErr, fmt.Errorf("wrong Header.ValidatorsHash: %v", err))
	}
	if err := tdmtypes.ValidateHash(h.NextValidatorsHash); err != nil {

	}
	if err := tdmtypes.ValidateHash(h.ConsensusHash); err != nil {
		return NewBNBRelayingError(InvalidBasicHeaderErr, fmt.Errorf("wrong Header.ConsensusHash: %v", err))
	}
	// NOTE: AppHash is arbitrary length
	if err := tdmtypes.ValidateHash(h.LastResultsHash); err != nil {
		return NewBNBRelayingError(InvalidBasicHeaderErr, fmt.Errorf("wrong Header.LastResultsHash: %v", err))
	}

	// Validate evidence and its hash.
	if err := tdmtypes.ValidateHash(h.EvidenceHash); err != nil {
		return NewBNBRelayingError(InvalidBasicHeaderErr, fmt.Errorf("wrong Header.EvidenceHash: %v", err))
	}

	if len(h.ProposerAddress) != crypto.AddressSize {
		return NewBNBRelayingError(InvalidBasicHeaderErr,
			fmt.Errorf("expected len(Header.ProposerAddress) to be %d, got %d",
			crypto.AddressSize, len(h.ProposerAddress)))
	}

	return nil
}

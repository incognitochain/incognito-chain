package ppos

import "errors"

var (
	errBlockSizeExceed       = errors.New("block size is too big")
	errNotInCommittee        = errors.New("user not in committee")
	errSigWrongOrNotExits    = errors.New("signature is wrong or not existed in block header")
	errChainNotFullySynced   = errors.New("chains are not fully synced")
	errTxIsWrong             = errors.New("transaction is wrong")
	errNotEnoughChainData    = errors.New("not enough chain data")
	errCantFinalizeBlock     = errors.New("can't finalize block")
	errMerkleRootCommitments = errors.New("MerkleRootCommitments is wrong")
	errNotEnoughSigs         = errors.New("not enough signatures")
)

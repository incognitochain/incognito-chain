//nolint:revive // skip linter for this legacy package name
package agg_interface

import "github.com/incognitochain/incognito-chain/privacy/operation"

type AggregatedRangeProof interface {
	Init()
	IsNil() bool
	Bytes() []byte
	SetBytes([]byte) error
	Verify() (bool, error)
	GetCommitments() []*operation.Point
	SetCommitments([]*operation.Point)
	ValidateSanity() bool
}

// type AggregatedRangeProofV1 = aggregatedrange.AggregatedRangeProof
// type AggregatedRangeProofV2 = bulletproofs.AggregatedRangeProof


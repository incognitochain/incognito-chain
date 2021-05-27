package metadata

import "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

type PortalUnshieldRequestBatchContent struct {
	BatchID       string // Hash(beaconHeight || UnshieldIDs)
	RawExternalTx string
	TokenID       string
	UnshieldIDs   []string
	UTXOs         []*statedb.UTXO
	NetworkFee    uint
	BeaconHeight  uint64
}

type ExternalFeeInfo struct {
	NetworkFee    uint
	RBFReqIncTxID string
}

type PortalUnshieldRequestBatchStatus struct {
	BatchID       string // Hash(beaconHeight || UnshieldIDs)
	TokenID       string
	UnshieldIDs   []string
	UTXOs         []*statedb.UTXO
	RawExternalTx string
	NetworkFees   map[uint64]ExternalFeeInfo
	BeaconHeight  uint64
	Status        byte
}

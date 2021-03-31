package metadata

import "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

type PortalUnshieldRequestBatchContent struct {
	BatchID       string // Hash(beaconHeight || UnshieldIDs)
	RawExternalTx string
	TokenID       string
	UnshieldIDs   []string
	UTXOs         []*statedb.UTXO
	NetworkFee    map[uint64]uint
	BeaconHeight  uint64
}

package transaction

// What needs to know when Tx Bridge with Privacy package?
// Only 2 main things:
// "Prove" and "Verify" these rules:
// - For each input conceal our real input by getting random inputs (ring signature).
// - Ensure sum input = sum output (pedersen commitment)
// - Ensure all output is non-negative (bulletproofs, aggregatedrangeproof)

// Ver 1:
// Prove:
// - Prove the input is oneofmany with other random inputs (with sum input = output by Pedersen)
// - Prove the non-negative with bulletproofs (aggregatedrangeproof)
// - Sign the above proofs

// Ver 2:
// Prove:
// - Prove the non-negative with bulletproofs (aggregatedrangeproof)
// - Prove the input is one of many with other random inputs plus sum input = output using MLSAG. (it also provides signature).

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
)

type TxVersionSwitcher interface {
	// It should store to tx the tx.sig and tx.proof
	Prove(tx *Tx, params *TxPrivacyInitParams) error

	// It should verify based on
	Verify(tx *Tx, hasPrivacy bool, db database.DatabaseInterface, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error)
}

// Used in Tx.Init
// For Tx to be formed correctly by using privacy package
func proveAndSignVersionSwitcher(tx *Tx, params *TxPrivacyInitParams) error {
	// Init interface
	var versionSwitcher TxVersionSwitcher
	if tx.Version == 1 {
		versionSwitcher = new(TxVersion1)
	} else if tx.Version == 2 {
		versionSwitcher = new(TxVersion2)
	}
	// Start proving and verifying
	return versionSwitcher.Prove(tx, params)
}

func verifierVersionSwitcher(tx *Tx, hasPrivacy bool, db database.DatabaseInterface, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	// Init interface
	var versionSwitcher TxVersionSwitcher
	if tx.Version == 1 {
		versionSwitcher = new(TxVersion1)
	} else if tx.Version == 2 {
		versionSwitcher = new(TxVersion2)
	}

	// Start proving and verifying
	return versionSwitcher.Verify(tx, hasPrivacy, db, shardID, tokenID, isBatch, isNewTransaction)
}

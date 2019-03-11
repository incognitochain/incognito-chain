package transaction

// import (
// 	"bytes"

// 	"github.com/big0t/constant-chain/common"
// 	"github.com/big0t/constant-chain/common/base58"
// 	"github.com/big0t/constant-chain/database"
// 	"github.com/big0t/constant-chain/privacy"
// )

// // count in miliconstant
// // 100 *10^3 mili constant
// const stakeShardAmount = 100000

// // count in miliconstant
// // 10000 *10^3 mili constant
// const stakeBeaconAmount = 10000000

// // Burning address
// // Using as receiver of staking transaction
// // All bytes are zero
// var publicKey = make([]byte, 33)
// var transmissionKey = make([]byte, 33)
// var stakeShardAddress = privacy.PaymentInfo{
// 	PaymentAddress: privacy.PaymentAddress{
// 		Pk: publicKey,
// 		Tk: transmissionKey,
// 	},
// 	Amount: stakeShardAmount,
// }

// var stakeBeaconAddress = privacy.PaymentInfo{
// 	PaymentAddress: privacy.PaymentAddress{
// 		Pk: publicKey,
// 		Tk: transmissionKey,
// 	},
// 	Amount: stakeBeaconAmount,
// }

// //TODO: check again stake transaction
// func (tx *Tx) validateTxStake(db database.DatabaseInterface, shardID byte) bool {
// 	constantTokenID := &common.Hash{}
// 	constantTokenID.SetBytes(common.ConstantID[:])
// 	valid := tx.ValidateTransaction(false, db, shardID, constantTokenID)
// 	if valid == false {
// 		return valid
// 	}
// 	// Check staking info:
// 	// - Check outputcoin
// 	// - Check staking address
// 	// Only one output at outputCoin
// 	if len(tx.Proof.OutputCoins) != 1 {
// 		return false
// 	}
// 	// No privacy
// 	if tx.Proof.OutputCoins[0].CoinDetailsEncrypted.IsNil() == false {
// 		return false
// 	}
// 	// Burning address (publickey are all zero)
// 	if bytes.Compare(tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress(), publicKey) != 0 {
// 		return false
// 	}
// 	return true
// }

// func (tx *Tx) ValidateTxStakeShard(db database.DatabaseInterface, shardID byte) bool {
// 	if tx.validateTxStake(db, shardID) == false {
// 		return false
// 	}
// 	// validate staking amount
// 	if tx.Proof.OutputCoins[0].CoinDetails.Value != stakeShardAmount {
// 		return false
// 	}
// 	return true
// }

// func (tx *Tx) ValidateTxStakeBeacon(db database.DatabaseInterface, shardID byte) bool {
// 	if tx.validateTxStake(db, shardID) == false {
// 		return false
// 	}
// 	// validate staking amount
// 	if tx.Proof.OutputCoins[0].CoinDetails.Value != stakeBeaconAmount {
// 		return false
// 	}
// 	return true
// }

// // return param:
// // #param1: state shard Address
// // #param2: state beacon Address
// // #param3: has staker or not?

// //using b, _, err := base58.Base58Check{}.Decode(data) for decode base58 string
// func (tx *Tx) ProcessTxStake(db database.DatabaseInterface, shardID byte) (string, string, bool) {
// 	if tx.ValidateTxStakeBeacon(db, shardID) == true {
// 		// skip comparing all address in input coin
// 		// ASSUME that all address are the same
// 		res := base58.Base58Check{}.Encode(tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress(), byte(0x00))
// 		return common.EmptyString, res, true
// 	}

// 	if tx.ValidateTxStakeShard(db, shardID) == true {
// 		// skip comparing all address in input coin
// 		// ASSUME that all address are the same
// 		res := base58.Base58Check{}.Encode(tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress(), byte(0x00))
// 		return res, common.EmptyString, true
// 	}
// 	return common.EmptyString, common.EmptyString, false
// }

// /*
// 	This function only extract staker from a valid transaction
// 	return param:
// 		#param1: state shard Address
// 		#param2: state beacon Address
// 		#param3: has staker or not?
// */
// func (tx *Tx) GetStakerFromTransaction() (string, string, bool) {
// 	if len(tx.Proof.OutputCoins) != 1 {
// 		return common.EmptyString, common.EmptyString, false
// 	}
// 	// No privacy
// 	if tx.Proof.OutputCoins[0].CoinDetailsEncrypted.IsNil() == false {
// 		return common.EmptyString, common.EmptyString, false
// 	}
// 	// Burning address (publickey are all zero)
// 	if bytes.Compare(tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress(), publicKey) != 0 {
// 		return common.EmptyString, common.EmptyString, false
// 	}
// 	if tx.Proof.OutputCoins[0].CoinDetails.Value == stakeBeaconAmount {
// 		res := base58.Base58Check{}.Encode(tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress(), byte(0x00))
// 		return common.EmptyString, res, true
// 	}

// 	if tx.Proof.OutputCoins[0].CoinDetails.Value == stakeShardAmount {
// 		res := base58.Base58Check{}.Encode(tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress(), byte(0x00))
// 		return res, common.EmptyString, true
// 	}
// 	return common.EmptyString, common.EmptyString, false
// }

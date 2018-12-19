package transaction

import (
	"bytes"

	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy-protocol"
)

// count in miliconstant
// 100 *10^3 mili constant
const stakeShardAmount = 100000

// count in miliconstant
// 10000 *10^3 mili constant
const stakeBeaconAmount = 10000000

// Burning address
// Using as receiver of staking transaction
// All bytes are zero
var publicKey = make([]byte, 33)
var transmissionKey = make([]byte, 33)
var stakeShardAddress = privacy.PaymentInfo{
	PaymentAddress: privacy.PaymentAddress{
		Pk: publicKey,
		Tk: transmissionKey,
	},
	Amount: stakeShardAmount,
}

var stakeBeaconAddress = privacy.PaymentInfo{
	PaymentAddress: privacy.PaymentAddress{
		Pk: publicKey,
		Tk: transmissionKey,
	},
	Amount: stakeBeaconAmount,
}

// func (tx *Tx) CreateStakeTx(
// 	senderSK *privacy.SpendingKey,
// 	usableTx []*Tx,
// 	fee uint64,
// 	db database.DatabaseInterface,
// ) error {
// 	err := tx.CreateTx(senderSK, stakingInfo, usableTx, fee, false, db)
// 	tx.Metadata = stakeTx{flag: "stake"}
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (tx *Tx) ValidateTxStake(db database.DatabaseInterface, shardID byte) bool {
// 	valid := tx.ValidateTransaction(false, db, shardID)
// 	if valid == false {
// 		fmt.Printf("Error validate transaction")
// 		return false
// 	}
// 	metaData := tx.Metadata.(stakeTx)
// 	if metaData.flag != "stake" {
// 		fmt.Printf("Not stake transaction")
// 		return false
// 	}
// 	return true
// }

	return true
}

func (tx *Tx) ValidateTxStakeShard(db database.DatabaseInterface, chainID byte) bool {
	if tx.validateTxStake(db, chainID) == false {
		return false
	}
	// validate staking amount
	if tx.Proof.OutputCoins[0].CoinDetails.Value != stakeShardAmount {
		return false
	}
	return true
}

func (tx *Tx) ValidateTxStakeBeacon(db database.DatabaseInterface, chainID byte) bool {
	if tx.validateTxStake(db, chainID) == false {
		return false
	}
	// validate staking amount
	if tx.Proof.OutputCoins[0].CoinDetails.Value != stakeBeaconAmount {
		return false
	}
	return true
}

// return param:
// #param1: state shard Address
// #param2: state beacon Address
// #param3: has staker or not?

//using b, _, err := base58.Base58Check{}.Decode(data) for decode base58 string
func (tx *Tx) ProcessTxStake(db database.DatabaseInterface, chainID byte) (string, string, bool) {
	if tx.ValidateTxStakeBeacon(db, chainID) == true {
		// skip comparing all address in input coin
		// ASSUME that all address are the same
		res := base58.Base58Check{}.Encode(tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress(), byte(0x00))
		return "", res, true
	}

	if tx.ValidateTxStakeShard(db, chainID) == true {
		// skip comparing all address in input coin
		// ASSUME that all address are the same
		res := base58.Base58Check{}.Encode(tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress(), byte(0x00))
		return res, "", true
	}
	return "", "", false
}

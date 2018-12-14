package transaction

// import (
// 	"fmt"

// 	"github.com/ninjadotorg/constant/database"
// 	"github.com/ninjadotorg/constant/privacy-protocol"
// )

// // count in miliconstant
// // 100 *10^3 mili constant
// const stake = 100000

// // Burning address
// // Using as receiver of staking transaction
// // All bytes are zero
// var publicKey = make([]byte, 33)
// var transmissionKey = make([]byte, 33)
// var stakingAddress = privacy.PaymentInfo{
// 	PaymentAddress: privacy.PaymentAddress{
// 		Pk: publicKey,
// 		Tk: transmissionKey,
// 	},
// 	Amount: stake,
// }

// // staking info contain only one address 0x0000000
// // staking amount defined in stake variable
// var stakingInfo = []*privacy.PaymentInfo{&stakingAddress}

// type stakeTx struct {
// 	flag string
// }

// func (tx *Tx) CreateStakeTx(
// 	senderSK *privacy.SpendingKey,
// 	usableTx []*Tx,
// 	fee uint64,
// 	commitmentsDB [][]byte,
// 	db database.DatabaseInterface,
// ) error {
// 	/*err := tx.CreateTx(senderSK, stakingInfo, usableTx, fee, commitmentsDB, false, db)
// 	tx.Metadata = stakeTx{flag: "stake"}
// 	if err != nil {
// 		return err
// 	}*/
// 	return nil
// }

// func (tx *Tx) ValidateTxStake(db database.DatabaseInterface, chainID byte) bool {
// 	valid := tx.ValidateTransaction(false, db, chainID)
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

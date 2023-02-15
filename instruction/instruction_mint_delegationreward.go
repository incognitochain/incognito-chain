package instruction

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type MintDelegationRewardInstruction struct {
	PaymentAddresses       []string
	PaymentAddressesStruct []key.PaymentAddress
	RewardAmount           []uint64
	TxRequestIDs           []string
	TxRequestIdHashes      []common.Hash
}

func NewMintDelegationRewardInsWithValue(
	txsRequest []string,
	paymentAddresses []string,
	rewardAmount []uint64,
) *MintDelegationRewardInstruction {
	mI := &MintDelegationRewardInstruction{}
	mI, _ = mI.SetPaymentAddresses(paymentAddresses)
	mI, _ = mI.SetRewardAmount(rewardAmount)
	mI, _ = mI.SetRequestTXIDs(txsRequest)
	return mI
}

func NewMintDelegationRewardIns() *MintDelegationRewardInstruction {
	return &MintDelegationRewardInstruction{}
}

func (mI *MintDelegationRewardInstruction) IsEmpty() bool {
	return reflect.DeepEqual(mI, NewReturnStakeIns()) ||
		len(mI.PaymentAddresses) == 0 && len(mI.RewardAmount) == 0
}

func (mI *MintDelegationRewardInstruction) SetPaymentAddresses(payments []string) (*MintDelegationRewardInstruction, error) {
	if payments == nil {
		return nil, errors.New("PaymentAddresses Are Null")
	}
	mI.PaymentAddresses = payments
	for _, payment := range payments {
		keyWallet, err := wallet.Base58CheckDeserialize(payment)
		if err != nil {
			Logger.Log.Errorf("ERROR: an error occured while deserializing reward receiver address string: %+v", err)
			return nil, err
		}
		receiverAddr := keyWallet.KeySet.PaymentAddress
		mI.PaymentAddressesStruct = append(mI.PaymentAddressesStruct, receiverAddr)
	}

	return mI, nil
}

func (mI *MintDelegationRewardInstruction) SetRewardAmount(amount []uint64) (*MintDelegationRewardInstruction, error) {
	if amount == nil {
		return nil, errors.New("List reward amount is ")
	}
	mI.RewardAmount = amount
	return mI, nil
}
func (mI *MintDelegationRewardInstruction) SetRequestTXIDs(txIDs []string) (*MintDelegationRewardInstruction, error) {
	if txIDs == nil {
		return nil, errors.New("Tx Hashes Are Null")
	}
	mI.TxRequestIDs = txIDs
	mI.TxRequestIdHashes = make([]common.Hash, len(txIDs))
	for i, v := range mI.TxRequestIDs {
		temp, err := common.Hash{}.NewHashFromStr(v)
		if err != nil {
			return mI, err
		}
		mI.TxRequestIdHashes[i] = *temp
	}
	return mI, nil
}
func (mI *MintDelegationRewardInstruction) GetType() string {
	return MINT_DREWARD_ACTION
}

func (mI *MintDelegationRewardInstruction) GetPaymentAddresses() []string {
	return mI.PaymentAddresses
}

func (mI *MintDelegationRewardInstruction) GetPaymentAddressesStruct() []key.PaymentAddress {
	return mI.PaymentAddressesStruct
}

func (mI *MintDelegationRewardInstruction) GetRequestIDs() []string {
	return mI.TxRequestIDs
}

func (mI *MintDelegationRewardInstruction) GetRequestIDHashes() []common.Hash {
	return mI.TxRequestIdHashes
}

func (mI *MintDelegationRewardInstruction) GetRewardAmount() []uint64 {
	return mI.RewardAmount
}

func (mI *MintDelegationRewardInstruction) ToString() []string {
	mintDRewardInsStr := []string{MINT_DREWARD_ACTION}
	mintDRewardInsStr = append(mintDRewardInsStr, strings.Join(mI.PaymentAddresses, SPLITTER))
	rewardAmountStr := make([]string, len(mI.RewardAmount))
	for i, v := range mI.RewardAmount {
		rewardAmountStr[i] = strconv.FormatUint(v, 10)
	}
	mintDRewardInsStr = append(mintDRewardInsStr, strings.Join(rewardAmountStr, SPLITTER))
	mintDRewardInsStr = append(mintDRewardInsStr, strings.Join(mI.TxRequestIDs, SPLITTER))
	return mintDRewardInsStr
}

func (mI *MintDelegationRewardInstruction) AddNewMintInfo(paymentAddress string, amount uint64, requestID string) {
	keyWallet, err := wallet.Base58CheckDeserialize(paymentAddress)
	if err != nil {
		Logger.Log.Errorf("ERROR: an error occured while deserializing reward receiver address %v string: %+v", paymentAddress, err)
		return
	}
	requestIDHash, err := common.Hash{}.NewHashFromStr(requestID)
	if err != nil {
		Logger.Log.Errorf("ERROR: an error occured while deserializing tx hash %v string: %+v", requestID, err)
		return
	}
	mI.PaymentAddresses = append(mI.PaymentAddresses, paymentAddress)
	receiverAddr := keyWallet.KeySet.PaymentAddress
	mI.PaymentAddressesStruct = append(mI.PaymentAddressesStruct, receiverAddr)
	mI.RewardAmount = append(mI.RewardAmount, amount)
	mI.TxRequestIDs = append(mI.TxRequestIDs, requestID)
	mI.TxRequestIdHashes = append(mI.TxRequestIdHashes, *requestIDHash)

}

func ValidateAndImportMintDelegationRewardInstructionFromString(instruction []string) (*MintDelegationRewardInstruction, error) {
	if err := ValidateReturnStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportMintDelegationRewardInstructionFromString(instruction)
}

func ImportMintDelegationRewardInstructionFromString(instruction []string) (*MintDelegationRewardInstruction, error) {
	mintDelegationRewardIns := NewMintDelegationRewardIns()
	var err error
	mintDelegationRewardIns, err = mintDelegationRewardIns.SetPaymentAddresses(strings.Split(instruction[1], SPLITTER))
	if err != nil {
		return nil, err
	}

	mintDelegationRewardIns, err = mintDelegationRewardIns.SetRequestTXIDs(strings.Split(instruction[3], SPLITTER))
	if err != nil {
		return nil, err
	}

	listRewardAmountStr := strings.Split(instruction[2], SPLITTER)
	listRewardAmount := make([]uint64, len(listRewardAmountStr))
	for i, v := range listRewardAmountStr {
		rewardAmount, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, err
		}
		listRewardAmount[i] = rewardAmount
	}
	mintDelegationRewardIns.SetRewardAmount(listRewardAmount)
	return mintDelegationRewardIns, err
}

func ValidateMintDelegationRewardInstructionSanity(instruction []string) error {
	if len(instruction) != 4 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != MINT_DREWARD_ACTION {
		return fmt.Errorf("invalid return staking action, %+v", instruction)
	}
	payments := strings.Split(instruction[1], SPLITTER)
	for _, payment := range payments {
		_, err := wallet.Base58CheckDeserialize(payment)
		if err != nil {
			Logger.Log.Errorf("ERROR: an error occured while deserializing reward receiver address string: %+v", err)
			return err
		}
	}

	listRewardAmountStr := strings.Split(instruction[2], SPLITTER)
	listRewardAmount := make([]uint64, len(listRewardAmountStr))
	for i, v := range listRewardAmountStr {
		rewardAmount, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid percent return %+v", err)
		}
		listRewardAmount[i] = rewardAmount
	}
	if len(listRewardAmount) != len(payments) {
		return fmt.Errorf("invalid reward percentReturns & payments length, %+v", instruction)
	}
	txRequests := strings.Split(instruction[3], SPLITTER)
	for _, txRequest := range txRequests {
		_, err := common.Hash{}.NewHashFromStr(txRequest)
		if err != nil {
			log.Println("err:", err)
			return fmt.Errorf("invalid tx request delegation reward %+v %+v", txRequests, err)
		}
	}
	if len(payments) != len(txRequests) {
		return fmt.Errorf("invalid tx requests & public Keys length, %+v", instruction)
	}
	return nil
}

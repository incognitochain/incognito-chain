package instruction

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

var (
	ErrRequestDelegationRewardInstruction = errors.New("Request delegation reward instruction error")
)

type RequestDelegationRewardInstruction struct {
	IncPaymentAddrs       []string
	IncPaymentAddrStructs []key.PaymentAddress
	TxRequestIDs          []string
	// IncPaymentAddrStructs []key.PaymentAddress
}

func NewRequestDelegationRewardInstructionWithValue(paymentAddrs []string, requestID []string) *RequestDelegationRewardInstruction {
	reqDRewardInstruction := &RequestDelegationRewardInstruction{}
	// finishSyncInstruction.SetChainID(chainID)
	reqDRewardInstruction.SetPaymentAddrs(paymentAddrs)
	reqDRewardInstruction.TxRequestIDs = requestID
	return reqDRewardInstruction
}

func NewRequestDelegationRewardInstruction() *RequestDelegationRewardInstruction {
	return &RequestDelegationRewardInstruction{}
}

func (r *RequestDelegationRewardInstruction) GetType() string {
	return REQ_DREWARD_ACTION
}

func (r *RequestDelegationRewardInstruction) IsEmpty() bool {
	return reflect.DeepEqual(r, NewRequestDelegationRewardInstruction()) ||
		len(r.IncPaymentAddrs) == 0
}

func (r *RequestDelegationRewardInstruction) ToString() []string {
	reqDRewardInstructionStr := []string{REQ_DREWARD_ACTION}
	reqDRewardInstructionStr = append(reqDRewardInstructionStr, strings.Join(r.IncPaymentAddrs, SPLITTER))
	reqDRewardInstructionStr = append(reqDRewardInstructionStr, strings.Join(r.TxRequestIDs, SPLITTER))
	return reqDRewardInstructionStr
}

func (r *RequestDelegationRewardInstruction) SetPaymentAddrs(paymentAddrs []string) (*RequestDelegationRewardInstruction, error) {
	r.IncPaymentAddrs = paymentAddrs
	for _, payment := range paymentAddrs {
		keyWallet, err := wallet.Base58CheckDeserialize(payment)
		if err != nil {
			Logger.Log.Errorf("ERROR: an error occured while deserializing reward receiver address string: %+v", err)
			return nil, err
		}
		receiverAddr := keyWallet.KeySet.PaymentAddress
		r.IncPaymentAddrStructs = append(r.IncPaymentAddrStructs, receiverAddr)
	}

	return r, nil
}

func (mI *RequestDelegationRewardInstruction) SetRequestTXIDs(txIDs []string) (*RequestDelegationRewardInstruction, error) {
	if txIDs == nil {
		return nil, errors.New("Tx Hashes Are Null")
	}
	mI.TxRequestIDs = txIDs
	for _, v := range mI.TxRequestIDs {
		_, err := common.Hash{}.NewHashFromStr(v)
		if err != nil {
			return mI, err
		}
	}
	return mI, nil
}

func ValidateAndImportRequestDelegationRewardInstructionFromString(instruction []string) (*RequestDelegationRewardInstruction, error) {
	if err := ValidateRequestDelegationRewardInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportRequestDelegationRewardInstructionFromString(instruction)
}

// ImportRequestDelegationRewardInstructionFromString is unsafe method
func ImportRequestDelegationRewardInstructionFromString(instruction []string) (*RequestDelegationRewardInstruction, error) {
	var err error
	reqDRewardInstruction := NewRequestDelegationRewardInstruction()
	reqDRewardInstruction, err = reqDRewardInstruction.SetPaymentAddrs(strings.Split(instruction[1], SPLITTER))
	if err != nil {
		return nil, err
	}

	reqDRewardInstruction, err = reqDRewardInstruction.SetRequestTXIDs(strings.Split(instruction[2], SPLITTER))
	if err != nil {
		return nil, err
	}

	return reqDRewardInstruction, nil
}

// ValidateRequestDelegationRewardInstructionSanity ...
func ValidateRequestDelegationRewardInstructionSanity(instruction []string) error {
	if len(instruction) != 3 {
		return errors.Errorf("%+v: invalid length, %+v", ErrRequestDelegationRewardInstruction, instruction)
	}
	if instruction[0] != REQ_DREWARD_ACTION {
		return errors.Errorf("%+v: invalid finish sync action, %+v", ErrRequestDelegationRewardInstruction, instruction)
	}
	payments := strings.Split(instruction[1], SPLITTER)
	for _, payment := range payments {
		_, err := wallet.Base58CheckDeserialize(payment)
		if err != nil {
			Logger.Log.Errorf("ERROR: an error occured while deserializing reward receiver address string: %+v", err)
			return err
		}
	}

	txRequests := strings.Split(instruction[2], SPLITTER)
	if len(payments) != len(payments) {
		return fmt.Errorf("invalid request tx ID & payments length, %+v", instruction)
	}
	for _, txRequest := range txRequests {
		_, err := common.Hash{}.NewHashFromStr(txRequest)
		if err != nil {
			log.Println("err:", err)
			return fmt.Errorf("invalid tx request delegation reward %+v %+v", txRequests, err)
		}
	}
	return nil
}

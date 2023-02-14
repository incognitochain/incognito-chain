package instruction

import (
	"reflect"
	"strings"

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
}

func NewRequestDelegationRewardInstructionWithValue(paymentAddrs []string) *RequestDelegationRewardInstruction {
	reqDRewardInstruction := &RequestDelegationRewardInstruction{}
	// finishSyncInstruction.SetChainID(chainID)
	reqDRewardInstruction.SetPaymentAddrs(paymentAddrs)
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
	return reqDRewardInstructionStr
}

func (r *RequestDelegationRewardInstruction) SetPaymentAddrs(paymentAddrs []string) *RequestDelegationRewardInstruction {
	r.IncPaymentAddrs = paymentAddrs
	for _, payment := range paymentAddrs {
		keyWallet, err := wallet.Base58CheckDeserialize(payment)
		if err != nil {
			Logger.Log.Errorf("ERROR: an error occured while deserializing reward receiver address string: %+v", err)
			return nil
		}
		receiverAddr := keyWallet.KeySet.PaymentAddress
		r.IncPaymentAddrStructs = append(r.IncPaymentAddrStructs, receiverAddr)
	}

	return r
}

func ValidateAndImportRequestDelegationRewardInstructionFromString(instruction []string) (*RequestDelegationRewardInstruction, error) {
	if err := ValidateRequestDelegationRewardInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportRequestDelegationRewardInstructionFromString(instruction)
}

// ImportRequestDelegationRewardInstructionFromString is unsafe method
func ImportRequestDelegationRewardInstructionFromString(instruction []string) (*RequestDelegationRewardInstruction, error) {
	reqDRewardInstruction := NewRequestDelegationRewardInstruction()
	reqDRewardInstruction.SetPaymentAddrs(strings.Split(instruction[1], SPLITTER))

	return reqDRewardInstruction, nil
}

// ValidateRequestDelegationRewardInstructionSanity ...
func ValidateRequestDelegationRewardInstructionSanity(instruction []string) error {
	if len(instruction) != 2 {
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
	return nil
}

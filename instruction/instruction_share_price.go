package instruction

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"reflect"
	"strings"
)

var (
	ErrSharePriceInstruction = errors.New("share price instruction error")
)

type SharePriceInstruction struct {
	committeeHash []string
	prices        []uint64
}

func NewSharePriceInstruction() *SharePriceInstruction {
	return &SharePriceInstruction{}
}

func (f *SharePriceInstruction) GetType() string {
	return SHARE_PRICE
}

func (f *SharePriceInstruction) IsEmpty() bool {
	return reflect.DeepEqual(f, NewSharePriceInstruction()) ||
		len(f.committeeHash) == 0 && len(f.prices) == 0
}

func (f *SharePriceInstruction) ToString() []string {
	sharePriceInstructionStr := []string{SHARE_PRICE}
	sharePriceInstructionStr = append(sharePriceInstructionStr, strings.Join(f.committeeHash, SPLITTER))
	prices := ""
	for i, v := range f.prices {
		if i == len(f.prices)-1 {
			prices += fmt.Sprintf("%v", v)
		} else {
			prices += fmt.Sprintf("%v,", v)
		}
	}
	sharePriceInstructionStr = append(sharePriceInstructionStr, prices)
	return sharePriceInstructionStr
}
func (f *SharePriceInstruction) GetValue() map[string]uint64 {
	res := map[string]uint64{}
	for i, v := range f.committeeHash {
		res[v] = f.prices[i]
	}
	return res
}

func (f *SharePriceInstruction) AddPrice(id string, price uint64) *SharePriceInstruction {
	f.committeeHash = append(f.committeeHash, id)
	f.prices = append(f.prices, price)
	return f
}

func ValidateAndImportSharePriceInstructionFromString(instruction []string) (*SharePriceInstruction, error) {
	if err := ValidateSharePriceInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return importSharePriceInstructionFromString(instruction)
}

func importSharePriceInstructionFromString(instruction []string) (*SharePriceInstruction, error) {
	sharePriceInstruction := NewSharePriceInstruction()
	tempIDs := strings.Split(instruction[1], SPLITTER)
	tempPrices := strings.Split(instruction[2], SPLITTER)
	if len(tempIDs) != len(tempPrices) {
		return nil, errors.New("Element not match!")
	}
	for i, v := range tempIDs {
		p, err := math.ParseUint64(tempPrices[i])
		if !err {
			return nil, errors.New("Parse price error")
		}
		sharePriceInstruction.AddPrice(v, p)
	}
	return sharePriceInstruction, nil
}

//ValidateSharePriceInstructionSanity ...
func ValidateSharePriceInstructionSanity(instruction []string) error {
	if len(instruction) != 3 {
		return fmt.Errorf("%+v: invalid length, %+v", ErrSharePriceInstruction, instruction)
	}
	tempIDs := strings.Split(instruction[1], SPLITTER)
	tempPrices := strings.Split(instruction[2], SPLITTER)
	if len(tempIDs) != len(tempPrices) {
		return errors.New("Element not match!")
	}
	return nil
}

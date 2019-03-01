package frombeaconins

import (
	"encoding/json"
	"strconv"

	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
)

type InstructionFromBeacon interface {
	GetStringFormat() ([]string, error)
}

type AcceptDCBBoardIns struct {
	BoardPaymentAddress []privacy.PaymentAddress
	StartAmountToken    uint64
}

func (acceptDCBBoardIns *AcceptDCBBoardIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(acceptDCBBoardIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(metadata.AcceptDCBBoardIns),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

type AcceptGOVBoardIns struct {
	BoardPaymentAddress []privacy.PaymentAddress
	StartAmountToken    uint64
}

func (acceptGOVBoardIns *AcceptGOVBoardIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(acceptGOVBoardIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(metadata.AcceptGOVBoardIns),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

//used in 2 cases:
//1. In Beacon chain
//2. In shard
func NewAcceptBoardIns(
	boardType metadata.BoardType,
	boardPaymentAddress []privacy.PaymentAddress,
	startAmountToken uint64,
) InstructionFromBeacon {
	if boardType == metadata.DCBBoard {
		return NewAcceptDCBBoardIns(
			boardPaymentAddress,
			startAmountToken,
		)
	} else {
		return NewAcceptGOVBoardIns(
			boardPaymentAddress,
			startAmountToken,
		)
	}
}

func NewAcceptDCBBoardIns(
	boardPaymentAddress []privacy.PaymentAddress,
	startAmountToken uint64,
) *AcceptDCBBoardIns {
	return &AcceptDCBBoardIns{
		BoardPaymentAddress: boardPaymentAddress,
		StartAmountToken:    startAmountToken,
	}
}

func NewAcceptGOVBoardIns(
	boardPaymentAddress []privacy.PaymentAddress,
	startAmountToken uint64,
) *AcceptGOVBoardIns {
	return &AcceptGOVBoardIns{
		BoardPaymentAddress: boardPaymentAddress,
		StartAmountToken:    startAmountToken,
	}
}

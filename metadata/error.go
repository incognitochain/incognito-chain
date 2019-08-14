package metadata

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota

	IssuingEthRequestDecodeInstructionError
	IssuingEthRequestUnmarshalJsonError
	IssuingEthRequestNewIssuingETHRequestFromMapEror
	IssuingEthRequestValidateTxWithBlockChainError
	IssuingEthRequestValidateSanityDataError
	IssuingEthRequestBuildReqActionsError
	IssuingEthRequestVerifyProofAndParseReceipt

	IssuingRequestDecodeInstructionError
	IssuingRequestUnmarshalJsonError
	IssuingRequestNewIssuingRequestFromMapEror
	IssuingRequestValidateTxWithBlockChainError
	IssuingRequestValidateSanityDataError
	IssuingRequestBuildReqActionsError
	IssuingRequestVerifyProofAndParseReceipt
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedError: {-1, "Unexpected error"},

	// -1xxx issuing eth request
	IssuingEthRequestDecodeInstructionError:          {-1001, "Can not decode instruction"},
	IssuingEthRequestUnmarshalJsonError:              {-1002, "Can not unmarshall json"},
	IssuingEthRequestNewIssuingETHRequestFromMapEror: {-1003, "Can no new issuing eth request from map"},
	IssuingEthRequestValidateTxWithBlockChainError:   {-1004, "Validate tx with block chain error"},
	IssuingEthRequestValidateSanityDataError:         {-1005, "Validate sanity data error"},
	IssuingEthRequestBuildReqActionsError:            {-1006, "Build request action error"},
	IssuingEthRequestVerifyProofAndParseReceipt:      {-1007, "Verify proof and parse receipt"},

	// -2xxx issuing eth request
	IssuingRequestDecodeInstructionError:        {-2001, "Can not decode instruction"},
	IssuingRequestUnmarshalJsonError:            {-2002, "Can not unmarshall json"},
	IssuingRequestNewIssuingRequestFromMapEror:  {-2003, "Can no new issuing eth request from map"},
	IssuingRequestValidateTxWithBlockChainError: {-2004, "Validate tx with block chain error"},
	IssuingRequestValidateSanityDataError:       {-2005, "Validate sanity data error"},
	IssuingRequestBuildReqActionsError:          {-2006, "Build request action error"},
	IssuingRequestVerifyProofAndParseReceipt:    {-2007, "Verify proof and parse receipt"},
}

type MetadataTxError struct {
	Code    int    // The code to send with reject messages
	Message string // Human readable message of the issue
	Err     error
}

// Error satisfies the error interface and prints human-readable errors.
func (e MetadataTxError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.Err)
}

func NewMetadataTxError(key int, err error, params ...interface{}) *MetadataTxError {
	return &MetadataTxError{
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].Message, params),
		Err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}

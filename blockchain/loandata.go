package blockchain

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/metadata"
	"github.com/pkg/errors"
)

const (
	dataSep  = "-"
	valueSep = "_"
)

var (
	loanIDKeyPrefix   = []byte("loanID-")
	loanRespKeyPrefix = []byte("loanResp-")
)

func getLoanRequestKeyBeacon(loanID []byte) string {
	return string(loanIDKeyPrefix) + string(loanID)
}

type LoanRespData struct {
	SenderPubkey []byte
	Response     metadata.ValidLoanResponse
}

func (lrd *LoanRespData) String() string {
	return strings.Join([]string{base64.StdEncoding.EncodeToString(lrd.SenderPubkey), string(lrd.Response)}, dataSep)
}

func parseLoanRespData(data string) (*LoanRespData, error) {
	s := strings.Split(data, dataSep)
	if len(s) != 2 {
		return nil, errors.Errorf("Error parsing loan response data")
	}
	errSaver := &metadata.ErrorSaver{}
	sender, errSender := base64.StdEncoding.DecodeString(s[0])
	response, errResp := strconv.Atoi(s[1])
	if errSaver.Save(errSender, errResp) != nil {
		return nil, errSaver.Get()
	}
	lrd := &LoanRespData{
		SenderPubkey: sender,
		Response:     metadata.ValidLoanResponse(response),
	}
	return lrd, nil
}

func getLoanResponseKeyBeacon(loanID []byte) string {
	return string(loanRespKeyPrefix) + string(loanID)
}

func getLoanResponseValueBeacon(data []*LoanRespData) string {
	s := []string{}
	for _, lrd := range data {
		s = append(s, lrd.String())
	}
	return strings.Join(s, valueSep)
}

func parseLoanResponseValueBeacon(data string) ([]*LoanRespData, error) {
	splits := strings.Split(data, valueSep)
	lrds := []*LoanRespData{}
	for _, s := range splits {
		lrd, err := parseLoanRespData(s)
		if err != nil {
			return nil, err
		}
		lrds = append(lrds, lrd)
	}
	return lrds, nil
}

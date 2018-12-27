package metadata

import (
	"encoding/json"
	"errors"
)

func ParseMetadata(meta interface{}) (Metadata, error) {
	if meta == nil {
		return nil, nil
	}

	mtTemp := map[string]interface{}{}
	metaInBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metaInBytes, &mtTemp)
	if err != nil {
		return nil, err
	}
	var md Metadata
	switch int(mtTemp["Type"].(float64)) {
	case BuyFromGOVRequestMeta:
		md = &BuySellRequest{}

	case BuyBackRequestMeta:
		md = &BuyBackRequest{}

	case BuyFromGOVResponseMeta:
		md = &BuySellResponse{}

	case BuyBackResponseMeta:
		md = &BuyBackResponse{}

	case LoanRequestMeta:
		md = &LoanRequest{}

	case LoanResponseMeta:
		md = &LoanResponse{}

	default:
		return nil, errors.New("Could not parse metadata with known types.")
	}

	err = json.Unmarshal(metaInBytes, &md)
	if err != nil {
		return nil, err
	}
	return md, nil
}

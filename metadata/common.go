package metadata

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
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
	case ResponseBaseMeta:
		md = &ResponseBase{}

	case IssuingRequestMeta:
		md = &IssuingRequest{}

	case IssuingResponseMeta:
		md = &IssuingResponse{}

	case ContractingRequestMeta:
		md = &ContractingRequest{}

	case ContractingResponseMeta:
		md = &ResponseBase{}

	case ShardBlockSalaryResponseMeta:
		md = &ShardBlockSalaryRes{}
	case BeaconSalaryResponseMeta:
		md = &BeaconBlockSalaryRes{}

	case ShardStakingMeta:
		md = &StakingMetadata{}
	case BeaconStakingMeta:
		md = &StakingMetadata{}
	case ReturnStakingMeta:
		md = &ReturnStakingMetadata{}

	default:
		fmt.Printf("[db] parse meta err: %+v\n", meta)
		return nil, errors.Errorf("Could not parse metadata with type: %d", int(mtTemp["Type"].(float64)))
	}

	err = json.Unmarshal(metaInBytes, &md)
	if err != nil {
		return nil, err
	}
	return md, nil
}

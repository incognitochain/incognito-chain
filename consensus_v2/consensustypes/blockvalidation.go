package consensustypes

import (
	"encoding/base64"
	"encoding/json"

	ggproto "github.com/golang/protobuf/proto"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
	"github.com/incognitochain/incognito-chain/proto"
)

type ValidationData struct {
	ProducerBLSSig []byte
	ProducerBriSig []byte
	ValidatiorsIdx []int
	AggSig         []byte
	BridgeSig      [][]byte
	PortalSig      []*portalprocessv4.PortalSig
}

func DecodeValidationData(data string) (*ValidationData, error) {
	var valData ValidationData
	err := json.Unmarshal([]byte(data), &valData)
	if err != nil {
		return nil, err
	}
	return &valData, nil
}

func (v ValidationData) ToBytes() ([]byte, error) {
	vBytes := proto.ValidationDataBytes{}
	vBytes.BLSSig = v.ProducerBLSSig
	vBytes.BriSig = v.ProducerBriSig
	vBytes.AllBriSig = v.BridgeSig
	vBytes.AggSig = v.AggSig
	for _, idx := range v.ValidatiorsIdx {
		vBytes.ValIdx = append(vBytes.ValIdx, int32(idx))
	}
	for _, sig := range v.PortalSig {
		vBytes.PortalSig = append(vBytes.PortalSig, sig.ToPortalSigBytes())
	}
	vB, err := ggproto.Marshal(&vBytes)
	if err != nil {
		return nil, err
	}
	return vB, nil
}

func (v *ValidationData) ToBase64() (string, error) {
	vBytes, err := v.ToBytes()
	if err != nil {
		return "", err
	}
	vStr := base64.StdEncoding.EncodeToString(vBytes)
	return vStr, nil
}

func (v *ValidationData) FromBytes(vDataBytes []byte) error {
	protoVData := proto.ValidationDataBytes{}
	err := ggproto.Unmarshal(vDataBytes, &protoVData)
	if err != nil {
		return err
	}
	v.AggSig = protoVData.AggSig
	v.BridgeSig = protoVData.AllBriSig
	if len(protoVData.PortalSig) > 0 {
		v.PortalSig = []*portalprocessv4.PortalSig{}
		for _, sigBytes := range protoVData.PortalSig {
			sig := portalprocessv4.PortalSig{}
			err = sig.FromPortalSigBytes(sigBytes)
			if err != nil {
				return err
			}
			v.PortalSig = append(v.PortalSig, &sig)
		}
	}
	v.ProducerBLSSig = protoVData.BLSSig
	v.ProducerBriSig = protoVData.BriSig
	for _, idx := range protoVData.ValIdx {
		v.ValidatiorsIdx = append(v.ValidatiorsIdx, int(idx))
	}
	return nil
}

func (v *ValidationData) FromBase64(vString string) error {
	vBytes, err := base64.StdEncoding.DecodeString(vString)
	if err != nil {
		return err
	}
	return v.FromBytes(vBytes)
}

func EncodeValidationData(validationData ValidationData) (string, error) {
	result, err := json.Marshal(validationData)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

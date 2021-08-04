package pdexv3

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// Check if the given OTA address string is a valid address and has the expected shard ID
func isValidReceiverAddressStr(addressStr string, expectedShardID byte) (privacy.OTAReceiver, error) {
	receiverAddress := privacy.OTAReceiver{}
	err := receiverAddress.FromString(addressStr)
	if err != nil {
		return receiverAddress, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if !receiverAddress.IsValid() {
		return receiverAddress, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("ReceiverAddress is invalid"))
	}

	pkb := receiverAddress.PublicKey.ToBytesS()
	currentShardID := common.GetShardIDFromLastByte(pkb[len(pkb)-1])
	if currentShardID != expectedShardID {
		return receiverAddress, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("ReceiverAddress shard ID is wrong"))
	}

	return receiverAddress, nil
}

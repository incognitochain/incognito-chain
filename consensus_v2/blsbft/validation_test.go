package blsbft

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"testing"
)

func TestDecodeAndValidateSig(t *testing.T) {
	validationData := "{\"ProducerBLSSig\":\"bpnA8vEbSSntTxFcSiD+RHSoGDsCX8qCXDAuc2l41t4ypfQ9oNypU8MdXF/DUILW0wfzI+6BPArEHnw7ZoGFKQE=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[1,2,3],\"AggSig\":\"kjVndVesMYMUz+Us7z98B1RzKWh4qSgvpxavlZ1yrRg=\",\"BridgeSig\":[\"\",\"\",\"\"]}"
	committeeKey := "121VhftSAygpEJZ6i9jGkEKLMQTKTiiHzeUfeuhpQCcLZtys8FazpWwytpHebkAwgCxvqgUUF13fcSMtp5dgV1YkbRMj3z42TW2EebzAaiGg2DkGPodckN2UsbqhVDibpMgJUHVkLXardemfLdgUqWGtymdxaaRyPM38BAZcLpo2pAjxKv5vG5Uh9zHMkn7ZHtdNHmBmhG8B46UeiGBXYTwhyMe9KGS83jCMPAoUwHhTEXj5qQh6586dHjVxwEkRzp7SKn9iG1FFWdJ97xEkP2ezAapNQ46quVrMggcHFvoZofs1xdd4o5vAmPKnPTZtGTKunFiTWGnpSG9L6r5QpcmapqvRrK5SiuFhNM5DqgzUeHBb7fTfoiWd2N29jkbTGSq8CPUSjx3zdLR9sZguvPdnAA8g25cFPGSZt8aEnFJoPRzM"

	valData, err := consensustypes.DecodeValidationData(validationData)
	if err != nil {
		t.Fatal(err)
	}

	producerKey := incognitokey.CommitteePublicKey{}
	err = producerKey.FromBase58(committeeKey)
	if err != nil {
		t.Fatal(err)
	}
	hash, _ := common.Hash{}.NewHashFromStr("030bcfa39a9da5dc525c2424c44bf7a58480c5bb9e82f50c81a0f9a5093fc69c")
	if err := validateSingleBriSig(hash, valData.ProducerBLSSig, producerKey.MiningPubKey[common.BridgeConsensus]); err != nil {
		t.Fatal(err)
	}
}

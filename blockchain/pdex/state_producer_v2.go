package pdex

import "github.com/incognitochain/incognito-chain/metadata"

type stateProducerV2 struct {
	stateProducerBase
}

func (sp *stateProducerV2) addLiquidity(
	metaData []metadata.PDEV3AddLiquidity,
	beaconHeight uint64,
) ([][]string, error) {
	return [][]string{}, nil
}

func (sp *stateProducerV2) modifyParams(
	actions [][]string,
	beaconHeight uint64,
	params Params,
) ([][]string, error) {
	return [][]string{}, nil
}

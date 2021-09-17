package main

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

const (
	privateKey = "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
)

type empty struct{}

var (
	nftID         common.Hash
	customTokenID common.Hash
	poolPairID    string
)

func submitKey(url string) error {
	var params []interface{}
	params = append(params, "14yJXBcq3EZ8dGh2DbL3a78bUUhWHDN579fMFx6zGVBLhWGzr2V4ZfUgjGHXkPnbpcvpepdzqAJEKJ6m8Cfq4kYiqaeSRGu37ns87ss")
	params = append(params, "0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11")
	params = append(params, 0)
	params = append(params, false)

	_, err := sendHttpRequest(url, "authorizedsubmitkey", params, true)
	return err
}

func convertCoin(url string) error {
	var params []interface{}
	params = append(params, privateKey)
	params = append(params, 1)

	_, err := sendHttpRequest(url, "createconvertcoinver1tover2transaction", params, true)
	return err
}

func initToken(url string) error {

	type Param struct {
		PrivateKey  string `json:"PrivateKey"`
		TokenName   string `json:"TokenName"`
		TokenSymbol string `json:"TokenSymbol"`
		Amount      uint64 `json:"Amount"`
	}

	param := Param{
		PrivateKey:  privateKey,
		TokenName:   "pETH",
		TokenSymbol: "pETH",
		Amount:      100000000000000,
	}

	var params []interface{}
	params = append(params, param)
	data, err := sendHttpRequest(url, "createandsendtokeninittransaction", params, true)
	type Temp struct {
		TokenID string `json:"TokenID"`
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var temp Temp
	err = json.Unmarshal(dataBytes, &temp)
	if err != nil {
		return err
	}
	tokenHash, err := common.Hash{}.NewHashFromStr(temp.TokenID)
	customTokenID = *tokenHash
	return err
}

func mintNft(url string) error {
	var params []interface{}
	params = append(params, privateKey)
	params = append(params, empty{})
	params = append(params, -1)
	params = append(params, 1)
	params = append(params, empty{})
	_, err := sendHttpRequest(url, "pdexv3_txMintNft", params, true)
	return err
}

func getBeaconBestState(url string) (*jsonresult.GetBeaconBestState, error) {
	data, err := sendHttpRequest(url, "getbeaconbeststate", nil, false)
	if err != nil {
		return nil, err
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	beaconBestState := &jsonresult.GetBeaconBestState{}
	err = json.Unmarshal(dataBytes, &beaconBestState)
	if err != nil {
		return nil, err
	}
	return beaconBestState, err
}

func getPdexBestState(url string) (*jsonresult.Pdexv3State, error) {
	type Temp struct {
		BeaconHeight uint64 `json:"BeaconHeight"`
	}
	beaconBestState, err := getBeaconBestState(url)
	if err != nil {
		return nil, err
	}
	temp := Temp{BeaconHeight: beaconBestState.BeaconHeight}
	var params []interface{}
	params = append(params, temp)
	data, err := sendHttpRequest(url, "pdexv3_getState", params, false)
	if err != nil {
		return nil, err
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	pdexv3State := &jsonresult.Pdexv3State{}
	err = json.Unmarshal(dataBytes, &pdexv3State)
	if err != nil {
		return nil, err
	}
	return pdexv3State, nil
}

func readNftID(url string) (common.Hash, error) {
	var res common.Hash
	pdexv3State, err := getPdexBestState(url)
	if err != nil {
		return res, err
	}
	for k := range pdexv3State.NftIDs {
		nftHash, err := common.Hash{}.NewHashFromStr(k)
		if err != nil {
			return res, err
		}
		res = *nftHash
	}
	return res, nil
}

func addLiquidity(url string, isFirstTx bool) error {
	var tokenID common.Hash
	var amount string
	tokenID = common.PRVCoinID
	amount = "100000"
	if !isFirstTx {
		tokenID = customTokenID
		amount = "400000"
	}
	var params []interface{}
	type Temp struct {
		NftID             string `json:"NftID"`
		TokenID           string `json:"TokenID"`
		PoolPairID        string `json:"PoolPairID"`
		PairHash          string `json:"PairHash"`
		ContributedAmount string `json:"ContributedAmount"`
		Amplifier         string `json:"Amplifier"`
	}
	temp := Temp{
		NftID:             nftID.String(),
		TokenID:           tokenID.String(),
		PairHash:          "pair_hash",
		ContributedAmount: amount,
		Amplifier:         "20000",
	}

	params = append(params, privateKey)
	params = append(params, empty{})
	params = append(params, -1)
	params = append(params, 1)
	params = append(params, temp)
	_, err := sendHttpRequest(url, "pdexv3_txAddLiquidity", params, true)
	return err
}

func addStakingPoolLiquidity(url string, stakingPoolID common.Hash) error {
	var params []interface{}
	type Temp struct {
		NftID         string `json:"NftID"`
		StakingPoolID string `json:"StakingPoolID"`
		Amount        string `json:"Amount"`
	}
	temp := Temp{
		NftID:         nftID.String(),
		Amount:        "2000",
		StakingPoolID: stakingPoolID.String(),
	}

	params = append(params, privateKey)
	params = append(params, empty{})
	params = append(params, -1)
	params = append(params, 1)
	params = append(params, temp)
	_, err := sendHttpRequest(url, "pdexv3_txStake", params, true)
	return err
}

func modifyParam(url string) error {
	var params []interface{}
	type Temp struct {
		NewParams pdex.Params `json:"NewParams"`
	}
	temp := Temp{
		NewParams: pdex.Params{},
	}

	params = append(params, privateKey)
	params = append(params, empty{})
	params = append(params, -1)
	params = append(params, 1)
	params = append(params, temp)
	_, err := sendHttpRequest(url, "pdexv3_txModifyParams", params, true)
	return err
}

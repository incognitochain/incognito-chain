package main

const (
	privateKey = "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
)

type empty struct{}

func submitKey(url string) error {
	var params []interface{}
	params = append(params, "14yJXBcq3EZ8dGh2DbL3a78bUUhWHDN579fMFx6zGVBLhWGzr2V4ZfUgjGHXkPnbpcvpepdzqAJEKJ6m8Cfq4kYiqaeSRGu37ns87ss")
	params = append(params, "0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11")
	params = append(params, 0)
	params = append(params, false)

	return sendHttpRequest(url, "authorizedsubmitkey", params)
}

func convertCoin(url string) error {
	var params []interface{}
	params = append(params, privateKey)
	params = append(params, 1)

	return sendHttpRequest(url, "createconvertcoinver1tover2transaction", params)
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
	return sendHttpRequest(url, "createandsendtokeninittransaction", params)
}

func mintNft(url string) error {
	var params []interface{}
	params = append(params, privateKey)
	params = append(params, empty{})
	params = append(params, -1)
	params = append(params, 1)
	params = append(params, empty{})
	return sendHttpRequest(url, "pdexv3_txMintNft", params)
}

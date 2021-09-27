package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func setNoVoteRule(url string) {

	method := "POST"

	payload := strings.NewReader(`{
	"jsonrpc": "1.0",
    "method": "setnovoteruleflag",
    "params": [true],
    "id": 1
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", "MwUBtRBWcH8kDr9m40Y027Rt6GyqjOpC73iioXTf")
	req.Header.Add("X-Amz-Content-Sha256", "beaead3198f7da1e70d03ab969765e0821b24fc913697e929e726aeaebf0eba3")
	req.Header.Add("X-Amz-Date", "20210927T012219Z")
	req.Header.Add("Authorization", "AWS4-HMAC-SHA256 Credential=AKIAIFNQPBGPLTLC2AKA/20210927/us-east-2c/execute-api/aws4_request, SignedHeaders=content-type;host;x-amz-content-sha256;x-amz-date;x-api-key, Signature=4c14129241431974c4fdbb3c34e838da99228dde2ef16da91e919e4cef525fc1")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}

func setVoteRule(url string) {

	method := "POST"

	payload := strings.NewReader(`{
	"jsonrpc": "1.0",
    "method": "setnovoteruleflag",
    "params": [false],
    "id": 1
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", "MwUBtRBWcH8kDr9m40Y027Rt6GyqjOpC73iioXTf")
	req.Header.Add("X-Amz-Content-Sha256", "beaead3198f7da1e70d03ab969765e0821b24fc913697e929e726aeaebf0eba3")
	req.Header.Add("X-Amz-Date", "20210927T012219Z")
	req.Header.Add("Authorization", "AWS4-HMAC-SHA256 Credential=AKIAIFNQPBGPLTLC2AKA/20210927/us-east-2c/execute-api/aws4_request, SignedHeaders=content-type;host;x-amz-content-sha256;x-amz-date;x-api-key, Signature=4c14129241431974c4fdbb3c34e838da99228dde2ef16da91e919e4cef525fc1")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}

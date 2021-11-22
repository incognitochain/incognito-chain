package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func setNoVoteRule(url string) {

	method := "POST"

	payload := strings.NewReader(`{
	"jsonrpc": "1.0",
    "method": "setconsensusrule",
    "params": [{
        "vote_rule": "no-vote",
        "create_rule": "create-repropose",
		"handle_vote_rule": "collect-vote",
		"handle_propose_rule": "handle-propose-message",
		"insert_rule": "insert-and-broadcast"
		"validator_rule": "validator-lemma2"
    }],
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

func setTestByzantineDetector(url string) {

	method := "POST"

	payload := strings.NewReader(`{
	"jsonrpc": "1.0",
    "method": "setconsensusrule",
    "params": [{
        "vote_rule": "vote",
        "create_rule": "create-only",
		"handle_vote_rule": "no-collect-vote",
		"handle_propose_rule": "handle-propose-message",
		"insert_rule": "insert-and-broadcast",
		"validator_rule": "validator-no-validate"
    }],
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
    "method": "setconsensusrule",
    "params": [{
        "vote_rule": "vote",
        "create_rule": "create-repropose",
		"handle_vote_rule": "collect-vote"
    }],
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

func setFixedNodeNoHandleProposeMsg(url string) {

	method := "POST"

	payload := strings.NewReader(`{
	"jsonrpc": "1.0",
    "method": "setconsensusrule",
    "params": [{
 		"vote_rule": "vote", 
        "create_rule": "create-repropose",
        "handle_vote_rule": "collect-vote",
        "handle_propose_rule": "no-handle-propose-message",
        "insert_rule": "insert-and-broadcast",
 		"validator_rule": "validator-lemma2"
    }],
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

func setCreateOnly(url string) {

	method := "POST"

	payload := strings.NewReader(`{
	"jsonrpc": "1.0",
    "method": "setconsensusrule",
    "params": [{
 		"vote_rule": "vote", 
        "create_rule": "create-only",
        "handle_vote_rule": "collect-vote",
        "handle_propose_rule": "handle-propose-message",
        "insert_rule": "insert-and-broadcast",
 		"validator_rule": "validator-lemma2"
    }],
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

func setLemma2NormalRule(url string) {

	method := "POST"

	payload := strings.NewReader(`{
	"jsonrpc": "1.0",
    "method": "setconsensusrule",
    "params": [{
 		"vote_rule": "vote", 
        "create_rule": "create-repropose",
        "handle_vote_rule": "collect-vote",
        "handle_propose_rule": "handle-propose-message",
        "insert_rule": "insert-and-broadcast",
 		"validator_rule": "validator-lemma2"
    }],
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

func setNoHandleVote(url string) {

	method := "POST"

	payload := strings.NewReader(`{
	"jsonrpc": "1.0",
    "method": "setconsensusrule",
    "params": [{
        "vote_rule": "vote",
        "create_rule": "create-repropose",
		"handle_vote_rule": "no-collect-vote",
		"handle_propose_rule": "handle-propose-message",
        "insert_rule": "insert-and-broadcast"
    }],
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

func setHandleVote(url string) {

	method := "POST"

	payload := strings.NewReader(`{
	"jsonrpc": "1.0",
    "method": "setconsensusrule",
    "params": [{
        "vote_rule": "vote",
        "create_rule": "create-repropose",
		"handle_vote_rule": "collect-vote"
    }],
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

func setNoHandleVoteNoVote(url string) {

	method := "POST"

	payload := strings.NewReader(`{
	"jsonrpc": "1.0",
    "method": "setconsensusrule",
    "params": [{
        "vote_rule": "no-vote",
        "create_rule": "create-repropose",
		"handle_vote_rule": "no-collect-vote"
    }],
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

func setNoVoteRuleShard0() {
	for i, url := range shard0UrlList {
		if i == 0 {
			continue
		}
		setNoVoteRule(url)
	}
}

func setByzantineDetectorShard0() {
	for _, url := range shard0UrlList {
		setTestByzantineDetector(url)
	}
}

func setVoteRuleShard0() {
	for i, url := range shard0UrlList {
		if i == 0 {
			continue
		}
		setVoteRule(url)
	}
}

func setFixedNodeNoHandleProposeMsgShard0() {
	for i, url := range shard0UrlList {
		if i == 0 {
			continue
		}
		setFixedNodeNoHandleProposeMsg(url)
	}
}

func setNode12CreateOnly() {
	for i, url := range shard0UrlList {
		if i == 1 || i == 2 {
			setCreateOnly(url)
		}
	}
}

func setLemma2NormalRuleShard0() {
	for _, url := range shard0UrlList {
		setLemma2NormalRule(url)
	}
}

func setNoHandleVoteShard0() {
	for i, url := range shard0UrlList {
		if i == 0 {
			continue
		}
		setNoHandleVote(url)
	}
}

func setHandleVoteShard0() {
	for i, url := range shard0UrlList {
		if i == 0 {
			continue
		}
		setHandleVote(url)
	}
}

func setNoHandleVoteNoVoteShard0() {
	for i, url := range shard0UrlList {
		if i == 0 {
			continue
		}
		setNoHandleVoteNoVote(url)
	}
}

func setIntervalVoteRule() {

	common.TIMESLOT = 10
	flag := false
	ticker := time.Tick(1 * time.Second)
	interval := int64(32)
	for _ = range ticker {

		currentTimeSlot := common.CalculateTimeSlot(time.Now().Unix())
		if currentTimeSlot%interval == 0 {
			if flag {
				fmt.Println(currentTimeSlot, flag, "Set No Vote")
				setNoVoteRuleShard0()
				flag = !flag
			}
			if !flag {
				fmt.Println(currentTimeSlot, flag, "Set Allow Vote")
				setVoteRuleShard0()
				flag = !flag

			}
		}
	}
}

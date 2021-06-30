package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func submitKeyShard0_0() {

	for _, url := range shard0UrlListWithBeacon {
		method := "POST"

		payload := strings.NewReader(`{
        "Jsonrpc": "1.0",
        "Method": "submitkey",
        "Params": [
                "14yHDAJRfR9fgx8udhNjYsgZ5jPLxhN2TER53L896kMg7JE5u1AbcD63Rf5yFk2rM2JL2qUvmDg2x5GkY2hPVmCAhwj76DptYtKSwSN"
        ],
        "Id": 1
}`)
		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Add("Content-Type", "application/json")

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
}
func submitKeyShard0_1() {
	for _, url := range shard0UrlListWithBeacon {

		method := "POST"

		payload := strings.NewReader(`{
        "Jsonrpc": "1.0",
        "Method": "submitkey",
        "Params": [
                "14yJXBcq3EZ8dGh2DbL3a78bUUhWHDN579fMFx6zGVBLhWGzr2V4ZfUgjGHXkPnbpcvpepdzqAJEKJ6m8Cfq4kYiqaeSRGu37ns87ss"
        ],
        "Id": 1
}`)

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Add("Content-Type", "application/json")

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
}
func submitKeyShard0_2() {
	for _, url := range shard0UrlListWithBeacon {

		method := "POST"

		payload := strings.NewReader(`{
        "Jsonrpc": "1.0",
        "Method": "submitkey",
        "Params": [
                "14y5uoGZUnXBZEX9x1AxAivmUK7tey8NBgNctzp2oLFJVNuHQ2PiKnXbk3T5SYKJiNSZb67hMBkUemVx2L8tpHZTm2wNtasGVbaYLuB"
        ],
        "Id": 1
}`)

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Add("Content-Type", "application/json")

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
}
func submitKeyShard0_3() {
	for _, url := range shard0UrlListWithBeacon {

		method := "POST"

		payload := strings.NewReader(`{
        "Jsonrpc": "1.0",
        "Method": "submitkey",
        "Params": [
                "14yChdmPo43znzjo3UsjXXPbbKPRzjejer2KAEnLorbxHziNqAWHxLefRWDxvNjmxY59CDzfyW6MnyP34NtGVkfgM4J5XRo5qhAXEMM"
        ],
        "Id": 1
}`)

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Add("Content-Type", "application/json")

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
}
func submitKeyShard1_1() {
	for _, url := range shard1UrlListWithBeacon {

		method := "POST"

		payload := strings.NewReader(`{
        "Jsonrpc": "1.0",
        "Method": "submitkey",
        "Params": [
                "14y6v1dqm3T7hk1KdrmQJXMDchUs9nLwo7EpuZ4DrecF2q2k57bth5ai2gjRyG2zC3coYUt8Zm6Ny9L4PdHXm9EDct5biefeTHSncXm"
        ],
        "Id": 1
}`)

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Add("Content-Type", "application/json")

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
}
func submitKeyShard1_2() {
	for _, url := range shard1UrlListWithBeacon {

		method := "POST"

		payload := strings.NewReader(`{
        "Jsonrpc": "1.0",
        "Method": "submitkey",
        "Params": [
                "14yCLZLJVf22sc9rFoNm4cwVnFyLdpB7tiD8jvQds1YtkZKssSs6YQhpfRh1hc6ZbMp3yoEpWEJJfXZxa6UFnPCzd5Qf41LawDbxoj3"
        ],
        "Id": 1
}`)

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Add("Content-Type", "application/json")

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
}
func submitKeyShard1_3() {
	for _, url := range shard1UrlListWithBeacon {

		method := "POST"

		payload := strings.NewReader(`{
        "Jsonrpc": "1.0",
        "Method": "submitkey",
        "Params": [
                "14yHcWG3Nc9gSqrMFo8Asswp3973hyFwm6qcbg5yJ1PAevJ4EcEdcLAjMJ6MDrPT3u2scW6RQc8bVQ7MrFfhePa7JzUCCownHsCKxnn"
        ],
        "Id": 1
}`)

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Add("Content-Type", "application/json")

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
}

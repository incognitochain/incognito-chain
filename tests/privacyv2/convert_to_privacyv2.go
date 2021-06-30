package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func convertShard0_0() {

	for _, url := range shard0UrlList[:1] {

		method := "POST"

		payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createconvertcoinver1tover2transaction",
   "params":[
      "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or",
      1
   ],
   "id":1
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
func convertShard0_1() {

	for _, url := range shard0UrlList[:1] {

		method := "POST"

		payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createconvertcoinver1tover2transaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      1
   ],
   "id":1
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
func convertShard0_2() {

	for _, url := range shard0UrlList[:1] {

		method := "POST"

		payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createconvertcoinver1tover2transaction",
   "params":[
      "112t8rrEW3NPNgU8xzbeqE7cr4WTT8JvyaQqSZyczA5hBJVvpQMTBVqNfcCdzhvquWCHH11jHihZtgyJqbdWPhWYbmmsw5aV29WSXBEsgbVX",
      1
   ],
   "id":1
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
func convertShard0_3() {

	for _, url := range shard0UrlList[:1] {

		method := "POST"

		payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createconvertcoinver1tover2transaction",
   "params":[
      "112t8rnY15HgjpbJn1es84ysseB6q9UQ5SwB6Eb82yejEQ3yzhd1dm5ShEiezdfMoEzBgvkuKcFdP5TY3SuWNHXKa1Krprsfxnk5wy7wZ6Dg",
      1
   ],
   "id":1
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
func convertShard1_1() {

	for _, url := range shard1UrlList[:1] {

		method := "POST"

		payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createconvertcoinver1tover2transaction",
   "params":[
      "112t8roHikeAFyuBpdCU76kXurEqrC9VYWyRyfFb6PwX6nip9KGYbwpXL78H92mUoWK2GWkA2WysgXbHqwSxnC6XCkmtxBVb3zJeCXgfcYyL",
      1
   ],
   "id":1
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
func convertShard1_2() {

	for _, url := range shard1UrlList[:1] {

		method := "POST"

		payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createconvertcoinver1tover2transaction",
   "params":[
      "112t8rr4sE2L8WzsVNEN9WsiGcMTDCmEH9TC1ZK8517cxURRFNoWoStYQTgqXpiAMU4gzmkmnWahHdGvQqFaY1JTVsn3nHfD5Ppgz8hQDiVC",
      1
   ],
   "id":1
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
func convertShard1_3() {

	for _, url := range shard1UrlList[:1] {

		method := "POST"

		payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createconvertcoinver1tover2transaction",
   "params":[
      "112t8rtt9Kd5LUcfXNmd7aMnQehCnKabArVB3BUk2RHVjeh88x5MJnJY4okB8JdFm4JNm4A2WjSe58qWNVkJPEFjpLHNYfKHpWfRdqyfDD9f",
      1
   ],
   "id":1
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

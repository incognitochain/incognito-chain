package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func main() {
	sendTransactionToShard1()
	ticker := time.Tick(500 * time.Millisecond)
	for _ = range ticker {
		sendTransactionFromTestnetGenesisKeyFromShard0()
		sendTransactionFromTestnetGenesisKeyFromShard1()
	}
}

func sendTransactionFromTestnetGenesisKeyFromShard0() {
	url := "http://localhost:9334"
	method := "POST"

	payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendtransaction",
   "params": [
        "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or",
        {
          "12RsFovpqaVaPggx7kFCNwczn5QXkUuuNWV6JLSDmz9Ec2xb5NmzSGdD9MW7Xzu5Fzu5kTQEay6nnHSdi2tp6znNrvWkbY6dhnobQ8D": 1,
          "12RtRb3B38MRcpMsQA93ggaj8U4XvmppKfjkcD3RiP6ieNVorasrMtu4GAB3gVivsDtgjACWLNj7bxE25NDzBm44TMkk6J7GBeqjVQM": 1,
          "12RpJWwrAqmw5uBCk1evWboTAukHUPgfG2DMwXQXuSJZPS9jA5vhj7s8uzPDUqSXaFPjnWbUFonSckBHKJ6fmEZtc7QAnjG2WsqsGG6": 1,
          "12S6ewfMcxxbppiPNPhT9qdZDcztUb5RbeqktuGfXQ3Xy1BpgTnJp2fGLPgQJBXScMXCB6tFT8rgTXwbUruUhMxfQ7EsBRAQe5eMEWS": 1,
          "12RpJxyS2qCPrtwmfWwzJF43Y7aiqivc7MBB7jpLtQxmtPsbq33u3s2Pk4azDdeQcvJvW96HcpQZ7f3FRKdduHjV5nv1P3NBsXeB7q9": 1,
          "12RpKTpuAFMbuaAFp7fXX5Q8uL6B4Ts9zxt4k5p8saTE6HniBvEJEPwa73AxVHUBt5MUXW6uASSfbiK8VCD7AmzbMYivGbmbFxo5EEN": 1,
          "12RpL484jktHW15z3cu6SSpqNYqVh5jMYB2NZBZh5xBFFLy86nfvoqd28QyYcpdwrvN9C9EmLcrA3ak4SdXAr8URpWogiMPNpEdhEmc": 1,
          "12RpHjMmRjFQJbNeDeBkiSavFkevfmmYkRegRdzfCK9PEwet5VdCKgQqtnkodP1nRBcuCCa3pYDrqFWxH1saKMLbnt1ZsfHWZ5DtXJX": 1,
          "12RyfHdhz4t6aQwhg6hVQrvHJPHXMtUhSbKF2RMtrhFSbHuU1eng2fk3jaowdAfLnodjaUExsq46Wprg7UNGJMYmPY5DBYFqPip87sg": 1,
          "12Ruy2g98j4Z8ytqQG9AxNvcAyxrsZKdC2bfCy5RrHRYMJpfPiYjTHRAvTDFbQa34g9X48eMVAtAhjjci5qHT7mN835evVyBkbfgV3P": 1,
          "12RpMCmQ8eoyJVe6vRHcRzhsyMEguKGvNYJJdPmaZJDXFeAmJKCmjBCXsALv5jdiyTRspWK1Lh3bXEd95x3MGMB89E9x4MWeoPwX4TR": 1,
          "12S1m3P4Sqqpj1WJistvEPXWu6ixQTbEGNYRunGz2m7hxFKxAUK1oRHpw37fMy243qtByRVVbwueLFssCB31wJR6NS5LpmpKJ1kWpxw": 1,
          "12Rsk58B8sNs5oUfjnFPC6neQUjNDx6NXvZ8CGWmMKcnUKhZMfNXsDg68ayAei9cw6mphRYB84PgTJK4r3eN9fMQTcAnx7Ee3S6oMAF": 1,
          "12RxtqEmTDf9DdQKM3yqC3XzMx4x4mtQHdT9KDmCJ5c8Br9KuC6Bjehsb9vxF5YEnH4XRrv8HgvrfyiGZdVDYZ1HwTVvtTMf3bvkLC6": 1,
          "12Rxu4juXSmAB39BgSuPavxWz2orXZMQFhGrvUY9e62rw44W5yCLBtcN11FrthNLquNk9CV2Pw9fs9Q6hSTbGZZdWTnv3pqKEkhWX6e": 1,
          "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci": 100000
        },
        10,
        0
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

func sendTransactionFromTestnetGenesisKeyFromShard1() {
	url := "http://localhost:9338"
	method := "POST"

	payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendtransaction",
   "params": [
        "112t8rr4sE2L8WzsVNEN9WsiGcMTDCmEH9TC1ZK8517cxURRFNoWoStYQTgqXpiAMU4gzmkmnWahHdGvQqFaY1JTVsn3nHfD5Ppgz8hQDiVC",
        {
          "12RsFovpqaVaPggx7kFCNwczn5QXkUuuNWV6JLSDmz9Ec2xb5NmzSGdD9MW7Xzu5Fzu5kTQEay6nnHSdi2tp6znNrvWkbY6dhnobQ8D": 1,
          "12RtRb3B38MRcpMsQA93ggaj8U4XvmppKfjkcD3RiP6ieNVorasrMtu4GAB3gVivsDtgjACWLNj7bxE25NDzBm44TMkk6J7GBeqjVQM": 1,
          "12RpJWwrAqmw5uBCk1evWboTAukHUPgfG2DMwXQXuSJZPS9jA5vhj7s8uzPDUqSXaFPjnWbUFonSckBHKJ6fmEZtc7QAnjG2WsqsGG6": 1,
          "12S6ewfMcxxbppiPNPhT9qdZDcztUb5RbeqktuGfXQ3Xy1BpgTnJp2fGLPgQJBXScMXCB6tFT8rgTXwbUruUhMxfQ7EsBRAQe5eMEWS": 1,
          "12RpJxyS2qCPrtwmfWwzJF43Y7aiqivc7MBB7jpLtQxmtPsbq33u3s2Pk4azDdeQcvJvW96HcpQZ7f3FRKdduHjV5nv1P3NBsXeB7q9": 1,
          "12RpKTpuAFMbuaAFp7fXX5Q8uL6B4Ts9zxt4k5p8saTE6HniBvEJEPwa73AxVHUBt5MUXW6uASSfbiK8VCD7AmzbMYivGbmbFxo5EEN": 1,
          "12RpL484jktHW15z3cu6SSpqNYqVh5jMYB2NZBZh5xBFFLy86nfvoqd28QyYcpdwrvN9C9EmLcrA3ak4SdXAr8URpWogiMPNpEdhEmc": 1,
          "12RpHjMmRjFQJbNeDeBkiSavFkevfmmYkRegRdzfCK9PEwet5VdCKgQqtnkodP1nRBcuCCa3pYDrqFWxH1saKMLbnt1ZsfHWZ5DtXJX": 1,
          "12RyfHdhz4t6aQwhg6hVQrvHJPHXMtUhSbKF2RMtrhFSbHuU1eng2fk3jaowdAfLnodjaUExsq46Wprg7UNGJMYmPY5DBYFqPip87sg": 1,
          "12Ruy2g98j4Z8ytqQG9AxNvcAyxrsZKdC2bfCy5RrHRYMJpfPiYjTHRAvTDFbQa34g9X48eMVAtAhjjci5qHT7mN835evVyBkbfgV3P": 1,
          "12RpMCmQ8eoyJVe6vRHcRzhsyMEguKGvNYJJdPmaZJDXFeAmJKCmjBCXsALv5jdiyTRspWK1Lh3bXEd95x3MGMB89E9x4MWeoPwX4TR": 1,
          "12S1m3P4Sqqpj1WJistvEPXWu6ixQTbEGNYRunGz2m7hxFKxAUK1oRHpw37fMy243qtByRVVbwueLFssCB31wJR6NS5LpmpKJ1kWpxw": 1,
          "12Rsk58B8sNs5oUfjnFPC6neQUjNDx6NXvZ8CGWmMKcnUKhZMfNXsDg68ayAei9cw6mphRYB84PgTJK4r3eN9fMQTcAnx7Ee3S6oMAF": 1,
          "12RxtqEmTDf9DdQKM3yqC3XzMx4x4mtQHdT9KDmCJ5c8Br9KuC6Bjehsb9vxF5YEnH4XRrv8HgvrfyiGZdVDYZ1HwTVvtTMf3bvkLC6": 1,
          "12Rxu4juXSmAB39BgSuPavxWz2orXZMQFhGrvUY9e62rw44W5yCLBtcN11FrthNLquNk9CV2Pw9fs9Q6hSTbGZZdWTnv3pqKEkhWX6e": 1,
          "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci": 100000
        },
        10,
        0
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

func sendTransactionToShard1() {
	url := "http://localhost:9334"
	method := "POST"

	payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendtransaction",
   "params": [
        "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or",
        {
          "12RyAEaUz4sErApu1f23PEydvotxDnC5gHoWDy5Th7JQuoT57oUowk8eSQN44ojPj3wZ5sEYFcLeFU5R8zgiXkSbAuY367Tek31gM1z": 10000,
        },
        10,
        0
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

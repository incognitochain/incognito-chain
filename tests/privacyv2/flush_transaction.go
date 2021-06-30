package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func sendTransactionFromTestnetGenesisKeyFromShard0_0() {
	url := "http://localhost:9334"
	method := "POST"

	payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendtransaction",
   "params": [
        "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
        {
          "12stx51rFNsgey1yBAAtJguaahKFbZPQnQQ7UdehnMNjuuYKWbNVc7V8VSYyjv5DhD3xpciXPcRNbPiMbrt5oKqHMaHWW6o5sGPjga6JTUab1YVtayAk6S13QEUSmroWDh8Bdyc85YUNL1QHASyu": 1,
          "12se8n5FxnT2cgWCkznBtDsuSnAjGzxNaxT2x6inPT5p2DbJDoQpt8YCwqxU8YeevAsJKaUuCSy4UA5tMfTfnkHAmZre8nsaJg68yVdPCbj5YYaNoNtbADFNmKjXgt7P841k8nnNFGgXgwXDkPdm": 1,
          "12so3AUr2sHyfMvU1PtmboSYJQDgW88n7QcDmS6qH5zn7fVmd5zAM5TXgE3pAknSgtUoHECnQaKqi5GcAUU85bkd31an17j7owc7P7T3EuY7xh2da7aeYYxzbgPhm5hfSMbwLBJT2dLKwEMkJPq3": 1,
          "12sfT74pXGVwjDdhLaP4J5ioN4y9bXMaoaGeGDNanyJuf4bA7783zvodrr9omz8EF5mywqfMbt77B1a43L1s1t953dVoe935YJoRZmss48o8YoosjWESwiAqa7kJpXM71JdymynT7TPZCWnFhy9c": 1,
          "12suUgHZQfPMB6vs7PttqwkPpbKBWdWXcXnuKtiewb6fkrb7tbewuNoapZ6dbG9JWwpAHbHmHPvfjPhPqPuguw5ooY6SnQy9ndzfafjvMitckEbZPfFuknnuDqKjFb7fKokqGN7gykTpuTciy5kd": 1,
          "12shkVKKfMA812gcxsg4yAFHPtvehYzVrATLQ535ZA9Gkg6BcgBFRXg593JSzexPySrNBFNCCHQ8ss1TuGmjHfe8TE82Mk19EuveHiJyh7jweuPBtVfpoARUhRYDBy3odhcmS1gqbvajGkqW6t3w": 1
        },
        10,
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
func sendTransactionFromTestnetGenesisKeyFromShard0_1() {
	url := "http://localhost:9334"
	method := "POST"

	payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendtransaction",
   "params": [
        "112t8rrEW3NPNgU8xzbeqE7cr4WTT8JvyaQqSZyczA5hBJVvpQMTBVqNfcCdzhvquWCHH11jHihZtgyJqbdWPhWYbmmsw5aV29WSXBEsgbVX",
        {
		  "12stx51rFNsgey1yBAAtJguaahKFbZPQnQQ7UdehnMNjuuYKWbNVc7V8VSYyjv5DhD3xpciXPcRNbPiMbrt5oKqHMaHWW6o5sGPjga6JTUab1YVtayAk6S13QEUSmroWDh8Bdyc85YUNL1QHASyu": 1,
          "12se8n5FxnT2cgWCkznBtDsuSnAjGzxNaxT2x6inPT5p2DbJDoQpt8YCwqxU8YeevAsJKaUuCSy4UA5tMfTfnkHAmZre8nsaJg68yVdPCbj5YYaNoNtbADFNmKjXgt7P841k8nnNFGgXgwXDkPdm": 1,
          "12so3AUr2sHyfMvU1PtmboSYJQDgW88n7QcDmS6qH5zn7fVmd5zAM5TXgE3pAknSgtUoHECnQaKqi5GcAUU85bkd31an17j7owc7P7T3EuY7xh2da7aeYYxzbgPhm5hfSMbwLBJT2dLKwEMkJPq3": 1,
          "12sfT74pXGVwjDdhLaP4J5ioN4y9bXMaoaGeGDNanyJuf4bA7783zvodrr9omz8EF5mywqfMbt77B1a43L1s1t953dVoe935YJoRZmss48o8YoosjWESwiAqa7kJpXM71JdymynT7TPZCWnFhy9c": 1,
          "12suUgHZQfPMB6vs7PttqwkPpbKBWdWXcXnuKtiewb6fkrb7tbewuNoapZ6dbG9JWwpAHbHmHPvfjPhPqPuguw5ooY6SnQy9ndzfafjvMitckEbZPfFuknnuDqKjFb7fKokqGN7gykTpuTciy5kd": 1,
          "12shkVKKfMA812gcxsg4yAFHPtvehYzVrATLQ535ZA9Gkg6BcgBFRXg593JSzexPySrNBFNCCHQ8ss1TuGmjHfe8TE82Mk19EuveHiJyh7jweuPBtVfpoARUhRYDBy3odhcmS1gqbvajGkqW6t3w": 1
        },
        10,
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
func sendTransactionFromTestnetGenesisKeyFromShard1_0() {
	url := "http://localhost:9338"
	method := "POST"

	payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendtransaction",
   "params": [
        "112t8roHikeAFyuBpdCU76kXurEqrC9VYWyRyfFb6PwX6nip9KGYbwpXL78H92mUoWK2GWkA2WysgXbHqwSxnC6XCkmtxBVb3zJeCXgfcYyL",
        {
		  "12stx51rFNsgey1yBAAtJguaahKFbZPQnQQ7UdehnMNjuuYKWbNVc7V8VSYyjv5DhD3xpciXPcRNbPiMbrt5oKqHMaHWW6o5sGPjga6JTUab1YVtayAk6S13QEUSmroWDh8Bdyc85YUNL1QHASyu": 1,
          "12se8n5FxnT2cgWCkznBtDsuSnAjGzxNaxT2x6inPT5p2DbJDoQpt8YCwqxU8YeevAsJKaUuCSy4UA5tMfTfnkHAmZre8nsaJg68yVdPCbj5YYaNoNtbADFNmKjXgt7P841k8nnNFGgXgwXDkPdm": 1,
          "12so3AUr2sHyfMvU1PtmboSYJQDgW88n7QcDmS6qH5zn7fVmd5zAM5TXgE3pAknSgtUoHECnQaKqi5GcAUU85bkd31an17j7owc7P7T3EuY7xh2da7aeYYxzbgPhm5hfSMbwLBJT2dLKwEMkJPq3": 1,
          "12sfT74pXGVwjDdhLaP4J5ioN4y9bXMaoaGeGDNanyJuf4bA7783zvodrr9omz8EF5mywqfMbt77B1a43L1s1t953dVoe935YJoRZmss48o8YoosjWESwiAqa7kJpXM71JdymynT7TPZCWnFhy9c": 1,
          "12suUgHZQfPMB6vs7PttqwkPpbKBWdWXcXnuKtiewb6fkrb7tbewuNoapZ6dbG9JWwpAHbHmHPvfjPhPqPuguw5ooY6SnQy9ndzfafjvMitckEbZPfFuknnuDqKjFb7fKokqGN7gykTpuTciy5kd": 1,
          "12shkVKKfMA812gcxsg4yAFHPtvehYzVrATLQ535ZA9Gkg6BcgBFRXg593JSzexPySrNBFNCCHQ8ss1TuGmjHfe8TE82Mk19EuveHiJyh7jweuPBtVfpoARUhRYDBy3odhcmS1gqbvajGkqW6t3w": 1,
          "12so3AUr2sHyfMvU1PtmboSYJQDgW88n7QcDmS6qH5zn7fVmd5zAM5TXgE3pAknSgtUoHECnQaKqi5GcAUU85bkd31an17j7owc7P7T3EuY7xh2da7aeYYxzbgPhm5hfSMbwLBJT2dLKwEMkJPq3": 1
        },
        10,
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
func sendTransactionFromTestnetGenesisKeyFromShard1_1() {
	url := "http://localhost:9338"
	method := "POST"

	payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendtransaction",
   "params": [
        "112t8rr4sE2L8WzsVNEN9WsiGcMTDCmEH9TC1ZK8517cxURRFNoWoStYQTgqXpiAMU4gzmkmnWahHdGvQqFaY1JTVsn3nHfD5Ppgz8hQDiVC",
        {
		  "12stx51rFNsgey1yBAAtJguaahKFbZPQnQQ7UdehnMNjuuYKWbNVc7V8VSYyjv5DhD3xpciXPcRNbPiMbrt5oKqHMaHWW6o5sGPjga6JTUab1YVtayAk6S13QEUSmroWDh8Bdyc85YUNL1QHASyu": 1,
          "12se8n5FxnT2cgWCkznBtDsuSnAjGzxNaxT2x6inPT5p2DbJDoQpt8YCwqxU8YeevAsJKaUuCSy4UA5tMfTfnkHAmZre8nsaJg68yVdPCbj5YYaNoNtbADFNmKjXgt7P841k8nnNFGgXgwXDkPdm": 1,
          "12so3AUr2sHyfMvU1PtmboSYJQDgW88n7QcDmS6qH5zn7fVmd5zAM5TXgE3pAknSgtUoHECnQaKqi5GcAUU85bkd31an17j7owc7P7T3EuY7xh2da7aeYYxzbgPhm5hfSMbwLBJT2dLKwEMkJPq3": 1,
          "12sfT74pXGVwjDdhLaP4J5ioN4y9bXMaoaGeGDNanyJuf4bA7783zvodrr9omz8EF5mywqfMbt77B1a43L1s1t953dVoe935YJoRZmss48o8YoosjWESwiAqa7kJpXM71JdymynT7TPZCWnFhy9c": 1,
          "12suUgHZQfPMB6vs7PttqwkPpbKBWdWXcXnuKtiewb6fkrb7tbewuNoapZ6dbG9JWwpAHbHmHPvfjPhPqPuguw5ooY6SnQy9ndzfafjvMitckEbZPfFuknnuDqKjFb7fKokqGN7gykTpuTciy5kd": 1,
          "12shkVKKfMA812gcxsg4yAFHPtvehYzVrATLQ535ZA9Gkg6BcgBFRXg593JSzexPySrNBFNCCHQ8ss1TuGmjHfe8TE82Mk19EuveHiJyh7jweuPBtVfpoARUhRYDBy3odhcmS1gqbvajGkqW6t3w": 1
        },
        10,
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
func sendTransactionFromTestnetGenesisKeyFromShard1_2() {
	url := "http://localhost:9338"
	method := "POST"

	payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendtransaction",
   "params": [
        "112t8rtt9Kd5LUcfXNmd7aMnQehCnKabArVB3BUk2RHVjeh88x5MJnJY4okB8JdFm4JNm4A2WjSe58qWNVkJPEFjpLHNYfKHpWfRdqyfDD9f",
        {
		  "12stx51rFNsgey1yBAAtJguaahKFbZPQnQQ7UdehnMNjuuYKWbNVc7V8VSYyjv5DhD3xpciXPcRNbPiMbrt5oKqHMaHWW6o5sGPjga6JTUab1YVtayAk6S13QEUSmroWDh8Bdyc85YUNL1QHASyu": 1,
          "12se8n5FxnT2cgWCkznBtDsuSnAjGzxNaxT2x6inPT5p2DbJDoQpt8YCwqxU8YeevAsJKaUuCSy4UA5tMfTfnkHAmZre8nsaJg68yVdPCbj5YYaNoNtbADFNmKjXgt7P841k8nnNFGgXgwXDkPdm": 1,
          "12so3AUr2sHyfMvU1PtmboSYJQDgW88n7QcDmS6qH5zn7fVmd5zAM5TXgE3pAknSgtUoHECnQaKqi5GcAUU85bkd31an17j7owc7P7T3EuY7xh2da7aeYYxzbgPhm5hfSMbwLBJT2dLKwEMkJPq3": 1,
          "12sfT74pXGVwjDdhLaP4J5ioN4y9bXMaoaGeGDNanyJuf4bA7783zvodrr9omz8EF5mywqfMbt77B1a43L1s1t953dVoe935YJoRZmss48o8YoosjWESwiAqa7kJpXM71JdymynT7TPZCWnFhy9c": 1,
          "12suUgHZQfPMB6vs7PttqwkPpbKBWdWXcXnuKtiewb6fkrb7tbewuNoapZ6dbG9JWwpAHbHmHPvfjPhPqPuguw5ooY6SnQy9ndzfafjvMitckEbZPfFuknnuDqKjFb7fKokqGN7gykTpuTciy5kd": 1,
          "12shkVKKfMA812gcxsg4yAFHPtvehYzVrATLQ535ZA9Gkg6BcgBFRXg593JSzexPySrNBFNCCHQ8ss1TuGmjHfe8TE82Mk19EuveHiJyh7jweuPBtVfpoARUhRYDBy3odhcmS1gqbvajGkqW6t3w": 1,
          "12snZXfS1ZLfsSRSzcruFd1qdknY1NkpNZCmaYn8efj2Wo7xjBgrZAuuuhshUeBiZMZysRBR8KfWrSoRZ9hTA6y8612nTfgyhYrG225P1AhvttcunNoFUFsnG5AQQvxHv4DFwuHbokQqzFReB2iA": 1
        },
        10,
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

func sendTransactionToShard1() {
	url := "http://localhost:9334"
	method := "POST"

	payload := strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendtransaction",
   "params": [
        "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or",
        {
		  "12stx51rFNsgey1yBAAtJguaahKFbZPQnQQ7UdehnMNjuuYKWbNVc7V8VSYyjv5DhD3xpciXPcRNbPiMbrt5oKqHMaHWW6o5sGPjga6JTUab1YVtayAk6S13QEUSmroWDh8Bdyc85YUNL1QHASyu": 1,
          "12se8n5FxnT2cgWCkznBtDsuSnAjGzxNaxT2x6inPT5p2DbJDoQpt8YCwqxU8YeevAsJKaUuCSy4UA5tMfTfnkHAmZre8nsaJg68yVdPCbj5YYaNoNtbADFNmKjXgt7P841k8nnNFGgXgwXDkPdm": 1,
          "12so3AUr2sHyfMvU1PtmboSYJQDgW88n7QcDmS6qH5zn7fVmd5zAM5TXgE3pAknSgtUoHECnQaKqi5GcAUU85bkd31an17j7owc7P7T3EuY7xh2da7aeYYxzbgPhm5hfSMbwLBJT2dLKwEMkJPq3": 1,
          "12sfT74pXGVwjDdhLaP4J5ioN4y9bXMaoaGeGDNanyJuf4bA7783zvodrr9omz8EF5mywqfMbt77B1a43L1s1t953dVoe935YJoRZmss48o8YoosjWESwiAqa7kJpXM71JdymynT7TPZCWnFhy9c": 1,
          "12suUgHZQfPMB6vs7PttqwkPpbKBWdWXcXnuKtiewb6fkrb7tbewuNoapZ6dbG9JWwpAHbHmHPvfjPhPqPuguw5ooY6SnQy9ndzfafjvMitckEbZPfFuknnuDqKjFb7fKokqGN7gykTpuTciy5kd": 1,
          "12shkVKKfMA812gcxsg4yAFHPtvehYzVrATLQ535ZA9Gkg6BcgBFRXg593JSzexPySrNBFNCCHQ8ss1TuGmjHfe8TE82Mk19EuveHiJyh7jweuPBtVfpoARUhRYDBy3odhcmS1gqbvajGkqW6t3w": 1
        },
        10,
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

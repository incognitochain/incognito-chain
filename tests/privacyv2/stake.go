package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	stakePayloads = []*strings.Reader{
		strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendstakingtransaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      {
         "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA":1750000000000
      },
     10,
     0,
     {
	    "StakingType": 63,
	    "CandidatePaymentAddress":
    "12stx51rFNsgey1yBAAtJguaahKFbZPQnQQ7UdehnMNjuuYKWbNVc7V8VSYyjv5DhD3xpciXPcRNbPiMbrt5oKqHMaHWW6o5sGPjga6JTUab1YVtayAk6S13QEUSmroWDh8Bdyc85YUNL1QHASyu",
    "PrivateSeed":"129pZpqYqYAA8wTAeDKuVwRthoBjNLUFm8FnLwUTkXddUqwShN9",
    "RewardReceiverPaymentAddress":"12stx51rFNsgey1yBAAtJguaahKFbZPQnQQ7UdehnMNjuuYKWbNVc7V8VSYyjv5DhD3xpciXPcRNbPiMbrt5oKqHMaHWW6o5sGPjga6JTUab1YVtayAk6S13QEUSmroWDh8Bdyc85YUNL1QHASyu",
    "AutoReStaking": true
    }
   ],
   "id":1
}`),
		strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendstakingtransaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      {
         "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA":1750000000000
      },
     10,
     0,
     {
    "StakingType": 63,
    "CandidatePaymentAddress":"12se8n5FxnT2cgWCkznBtDsuSnAjGzxNaxT2x6inPT5p2DbJDoQpt8YCwqxU8YeevAsJKaUuCSy4UA5tMfTfnkHAmZre8nsaJg68yVdPCbj5YYaNoNtbADFNmKjXgt7P841k8nnNFGgXgwXDkPdm",
    "PrivateSeed": "12JqKehM24bfSkfv3FKGtzFw4seoJSJbbgAqaYtX3w6DjVuH8mb",
    "RewardReceiverPaymentAddress":"12se8n5FxnT2cgWCkznBtDsuSnAjGzxNaxT2x6inPT5p2DbJDoQpt8YCwqxU8YeevAsJKaUuCSy4UA5tMfTfnkHAmZre8nsaJg68yVdPCbj5YYaNoNtbADFNmKjXgt7P841k8nnNFGgXgwXDkPdm",
    "AutoReStaking": true
     }
   ],
   "id":1
}`),
		strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendstakingtransaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      {
         "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA":1750000000000
      },
     10,
     0,
     {
    "StakingType": 63,
    "CandidatePaymentAddress": "12so3AUr2sHyfMvU1PtmboSYJQDgW88n7QcDmS6qH5zn7fVmd5zAM5TXgE3pAknSgtUoHECnQaKqi5GcAUU85bkd31an17j7owc7P7T3EuY7xh2da7aeYYxzbgPhm5hfSMbwLBJT2dLKwEMkJPq3",
    "PrivateSeed":"1tLA3U2fEBHF2b7U1Xry4K4CUYddxAs9Hi8e2KjQ7nHRd8nVbK",
    "RewardReceiverPaymentAddress":"12so3AUr2sHyfMvU1PtmboSYJQDgW88n7QcDmS6qH5zn7fVmd5zAM5TXgE3pAknSgtUoHECnQaKqi5GcAUU85bkd31an17j7owc7P7T3EuY7xh2da7aeYYxzbgPhm5hfSMbwLBJT2dLKwEMkJPq3",
    "AutoReStaking":true
     }
   ],
   "id":1
}`),
		strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendstakingtransaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      {
         "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA":1750000000000
      },
     10,
     0,
     {
    "StakingType": 63,
    "CandidatePaymentAddress": "12sfT74pXGVwjDdhLaP4J5ioN4y9bXMaoaGeGDNanyJuf4bA7783zvodrr9omz8EF5mywqfMbt77B1a43L1s1t953dVoe935YJoRZmss48o8YoosjWESwiAqa7kJpXM71JdymynT7TPZCWnFhy9c",
    "PrivateSeed": "12VkNWyH1DZqLECcLKABojwMAL5ygmPR6haioYHQwW57YKCmRai",
    "RewardReceiverPaymentAddress": "12sfT74pXGVwjDdhLaP4J5ioN4y9bXMaoaGeGDNanyJuf4bA7783zvodrr9omz8EF5mywqfMbt77B1a43L1s1t953dVoe935YJoRZmss48o8YoosjWESwiAqa7kJpXM71JdymynT7TPZCWnFhy9c",
    "AutoReStaking":true
     }
   ],
   "id":1
}`),
		strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendstakingtransaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      {
         "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA":1750000000000
      },
     10,
     0,
    {
    "StakingType": 63,
    "CandidatePaymentAddress": "12snZXfS1ZLfsSRSzcruFd1qdknY1NkpNZCmaYn8efj2Wo7xjBgrZAuuuhshUeBiZMZysRBR8KfWrSoRZ9hTA6y8612nTfgyhYrG225P1AhvttcunNoFUFsnG5AQQvxHv4DFwuHbokQqzFReB2iA",
    "PrivateSeed": "1g19rDzcdsMGz6xrVD2WeUmwBYRRDFhN4zMJLNFGt1VLzqtL8S",
    "RewardReceiverPaymentAddress": "12snZXfS1ZLfsSRSzcruFd1qdknY1NkpNZCmaYn8efj2Wo7xjBgrZAuuuhshUeBiZMZysRBR8KfWrSoRZ9hTA6y8612nTfgyhYrG225P1AhvttcunNoFUFsnG5AQQvxHv4DFwuHbokQqzFReB2iA",
    "AutoReStaking":true
     }
   ],
   "id":1
}`),
		strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendstakingtransaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      {
         "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA":1750000000000
      },
     10,
     0,
    {
    "StakingType": 63,
    "CandidatePaymentAddress": "12suUgHZQfPMB6vs7PttqwkPpbKBWdWXcXnuKtiewb6fkrb7tbewuNoapZ6dbG9JWwpAHbHmHPvfjPhPqPuguw5ooY6SnQy9ndzfafjvMitckEbZPfFuknnuDqKjFb7fKokqGN7gykTpuTciy5kd",
    "PrivateSeed": "12WKfbe8tjhDXNpVoK6zBaxMyLCc11froj3mdDMHwMBttuMh6n7",
    "RewardReceiverPaymentAddress": "12suUgHZQfPMB6vs7PttqwkPpbKBWdWXcXnuKtiewb6fkrb7tbewuNoapZ6dbG9JWwpAHbHmHPvfjPhPqPuguw5ooY6SnQy9ndzfafjvMitckEbZPfFuknnuDqKjFb7fKokqGN7gykTpuTciy5kd",
    "AutoReStaking":true
     }
   ],
   "id":1
}`),
		strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendstakingtransaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      {
         "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA":1750000000000
      },
     10,
     0,
    {
    "StakingType": 63,
    "CandidatePaymentAddress": "12shkVKKfMA812gcxsg4yAFHPtvehYzVrATLQ535ZA9Gkg6BcgBFRXg593JSzexPySrNBFNCCHQ8ss1TuGmjHfe8TE82Mk19EuveHiJyh7jweuPBtVfpoARUhRYDBy3odhcmS1gqbvajGkqW6t3w",
    "PrivateSeed": "14J5qkQ6SaZxRxujVExtvm6uRbcEZxGU64bMS91L59fbTyN9cj",
    "RewardReceiverPaymentAddress": "12shkVKKfMA812gcxsg4yAFHPtvehYzVrATLQ535ZA9Gkg6BcgBFRXg593JSzexPySrNBFNCCHQ8ss1TuGmjHfe8TE82Mk19EuveHiJyh7jweuPBtVfpoARUhRYDBy3odhcmS1gqbvajGkqW6t3w",
    "AutoReStaking":true
     }
   ],
   "id":1
}`),
		strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendstakingtransaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      {
         "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA":1750000000000
      },
     10,
     0,
    {
    "StakingType": 63,
    "CandidatePaymentAddress": "12sw7BWe13ApYguEyBZCBLH6d5XWPzGsXy7EY3vtQ9ZsXrJ3yTJejiiC2PJW2DTqTKGTcgvK1S67TX2TwicjLVqcuD6ZWuDEkEmKSoaJfeiFAtFVzSVVJ6rFpbG8Vz9D6hmQSJ51JgayCz8ghhW9",
    "PrivateSeed": "1kbmGmRfAPoqbem2mTsMwmjr72mJR1U9CTFNJ2iDmXHk5rh4tv",
    "RewardReceiverPaymentAddress": "12sw7BWe13ApYguEyBZCBLH6d5XWPzGsXy7EY3vtQ9ZsXrJ3yTJejiiC2PJW2DTqTKGTcgvK1S67TX2TwicjLVqcuD6ZWuDEkEmKSoaJfeiFAtFVzSVVJ6rFpbG8Vz9D6hmQSJ51JgayCz8ghhW9",
    "AutoReStaking":true
     }
   ],
   "id":1
}`),
		strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendstakingtransaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      {
         "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA":1750000000000
      },
     10,
     0,
    {
    "StakingType": 63,
    "CandidatePaymentAddress": "12smfcjJ12WXVXah6K181U3nr3PFYMinXiMRnCTYEJUZLS1DdMkP8VqjRgBhJyKt9oRaiWU2it7Qtfi6VfQxpKUMta2Sx49F78kfgVVLQRYE8JjDHPfCDZwHi5MpGnakBcMLhX5V6JfFoxCrFvFA",
    "PrivateSeed": "19Md7MynbrZaGUKk1YnD1BKcrQybF2ZSu1pkxMGnksJHSwPLTv",
    "RewardReceiverPaymentAddress": "12smfcjJ12WXVXah6K181U3nr3PFYMinXiMRnCTYEJUZLS1DdMkP8VqjRgBhJyKt9oRaiWU2it7Qtfi6VfQxpKUMta2Sx49F78kfgVVLQRYE8JjDHPfCDZwHi5MpGnakBcMLhX5V6JfFoxCrFvFA",
    "AutoReStaking":true
     }
   ],
   "id":1
}`),
		strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendstakingtransaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      {
         "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA":1750000000000
      },
     10,
     0,
    {
    "StakingType": 63,
    "CandidatePaymentAddress": "12smuyxf6jP3NFdAzEEaHALjA6r8rqFEVURKdqt3rUANZkgdBAH1Cp5cTwEnvU3gQUgYPEVbVuWmBoFrqT1ARLAkPde9ykHY2iVeTVi6QeJQWp3mJhLCsFBJPBUNrAQ8YytyKdCi7YwxJVQaKU98",
    "PrivateSeed": "1mDAyCjk6byv3mad6KsVSJg9TQ4G6QtT5xEFWMof4fKaUG3hH3",
    "RewardReceiverPaymentAddress": "12smuyxf6jP3NFdAzEEaHALjA6r8rqFEVURKdqt3rUANZkgdBAH1Cp5cTwEnvU3gQUgYPEVbVuWmBoFrqT1ARLAkPde9ykHY2iVeTVi6QeJQWp3mJhLCsFBJPBUNrAQ8YytyKdCi7YwxJVQaKU98",
    "AutoReStaking":true
     }
   ],
   "id":1
}`),
		strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendstakingtransaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      {
         "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA":1750000000000
      },
     10,
     0,
    {
    "StakingType": 63,
    "CandidatePaymentAddress": "12svitvWuKywRUkzAWwU1omouuF3gAEgvrXnf4b84mMMQ4ep5YMdxrZpwxRPrGQv2JBxdAWQb4a8faqndeSvn7hLTQCTYdWyfvrjjKbVYND1ydBMKx2s5uWJAxFMZgT1yDwGHfCSdN8PQnG2KAuD",
    "PrivateSeed": "12hvZFa6nCEqQFqZuQjMc6Pdvyo153kTC8Xg8FyefBDpXhg8LX7",
    "RewardReceiverPaymentAddress": "12svitvWuKywRUkzAWwU1omouuF3gAEgvrXnf4b84mMMQ4ep5YMdxrZpwxRPrGQv2JBxdAWQb4a8faqndeSvn7hLTQCTYdWyfvrjjKbVYND1ydBMKx2s5uWJAxFMZgT1yDwGHfCSdN8PQnG2KAuD",
    "AutoReStaking":true
     }
   ],
   "id":1
}`),
		strings.NewReader(`{
   "jsonrpc":"1.0",
   "method":"createandsendstakingtransaction",
   "params":[
      "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
      {
         "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA":1750000000000
      },
     10,
     0,
    {
    "StakingType": 63,
	"CandidatePaymentAddress": "12snQLH2TcATCiZhre6fTKphX2Di86mAEsS1pKzfq58ZVk3fqqhMt4K6gw8y9BdHd11Z9efe56bRQoyYmDiMJHM6d8sc6vqXetdFQzLyKBRv9ntw8fpyho78UqRvcbvrdRrL3c4k6ypii6LLQwNb",
    "PrivateSeed": "12rtQu37z9bhDRL3ERZFm383jjrZ3cZTVB1rpgzhfog3o3KjviV",
    "RewardReceiverPaymentAddress": "12snQLH2TcATCiZhre6fTKphX2Di86mAEsS1pKzfq58ZVk3fqqhMt4K6gw8y9BdHd11Z9efe56bRQoyYmDiMJHM6d8sc6vqXetdFQzLyKBRv9ntw8fpyho78UqRvcbvrdRrL3c4k6ypii6LLQwNb",
    "AutoReStaking":true
     }
   ],
   "id":1
}`),
	}
)

func stake(numberOfKey int) {
	i := 0
	for i < numberOfKey {
		time.Sleep(5 * time.Second)
		fmt.Println("Try stake key ", i)
		payload := stakePayloads[i]
		url := "http://localhost:9334"
		method := "POST"

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
		if len(body) == 0 {
			fmt.Println(i, "Return empty body")
			i++
			continue
		}
		rpcResponseRaw := make(map[string]interface{})
		json.Unmarshal(body, &rpcResponseRaw)
		if rpcResponseRaw["Error"] == nil {
			fmt.Println("No Error", i, rpcResponseRaw)
			i++
			continue
		} else {
			rpcResponseRawError := rpcResponseRaw["Error"]
			rpcResponseError, _ := rpcResponseRawError.(map[string]interface{})
			stackTrace := rpcResponseError["StackTrace"].(string)
			if strings.Contains(stackTrace, "This pubkey may staked already") {
				fmt.Println(i, "This pubkey may staked already")
				i++
				continue
			} else {
				continue
			}
		}
	}
}

This is the setup to run the very basic incognito blockchain on your local server. Then blockchain will contain 2 Highway-Proxies, 4 nodes in Beacon, 4 nodes in Shard-0, 4 nodes in Shard-1. In short, you will setup 2 Highway and 12 nodes in only one server.

**Prerequisites:** 
 + Ubuntu 18.04 or higher
 + git
 + golang 1.12.9
 + tmux
 + curl
 
**Minimum hardware requirement:**
 + 8 CPU ~ 2.4Ghz
 + 12 GB of RAM
 
**recommended hardware requirement:**
 + 16 CPU ~ 2.4Ghz
 + 24 GB of RAM

#SECTION I: DEPLOY THE INCOGNITO-HIGHWAY
--------------------
1. Clone the repo:

`git clone https://github.com/incognitochain/incognito-highway.git`

2. Checkout the branch v1

`git checkout v1`

3. Create tmux sessions:
```
tmux new -s highway1
Ctrl B D (detach)
tmux new -s highway2
Ctrl B D (detach)
```

4. Build the highway binary
```
go build -o highway
```

5. Run the highway
```
tmux send-keys -t highway1 C-C ENTER ./run.sh lc1 ENTER
tmux send-keys -t highway2 C-C ENTER ./run.sh lc2 ENTER
```

#SECTION II: DEPLOY THE CHAIN
--------------------
1. Clone the repo:

`git clone https://github.com/incognitochain/incognito-chain.git`

2. Checkout the tag 20200114_1

`git checkout 20200114_1`

3. Modify the following params in *./incognito-chain/common/constants.go*
```
MaxShardNumber = 2
```

4. Generate 4 (or more) key set for user: 
- Edit the *./incognito-chain/utility/genkeywithpassword.go*
```
numberOfKey := 2                              		// Number of keyset that you want to be generated; 2 address x 2 shard = 4 keyset
randomString := []byte("YourRandomStringOne")    	// A random string used to create keyset. The same string create the same keyset
```
- Execute the script to generate keyset:

`go run utility/genkeywithpassword.go `

- Sample output:
```
 ***** Shard 0 **** 
0
Payment Address  : 12RqRRt4q6nVis3bfVVf7L4fGquHU8KReA4ZWjjs43kH3VfWRFfGHwZjexHBgSkWMrAQrmq5CeWPkjD1hYt4KousUY7GUWn3Cg3iyJk
Private Key      : 112t8rncBDbGaFrAE7MZz14d2NPVWprXQuHHXCD2TgSV8USaDFZY3MihVWSqKjwy47sTQ6XvBgNYgdKH2iDVZruKQpRSB5JqxDAX6sjMoUT6
Public key       : 18fTFPj6JFBAGoAeyq6JqfW15vKhxKNHHh3rdvoFMyt3ThdWkR
ReadOnly key     : 13hT1bFyatrtjai5HZg1x23bCZwUg9JwkNGNmiQT8zsWxsbueV6pMBpNVUbcfvQrCjWG1hNa6Z8gJjDFstRZ4cqWgN7zzRSCsci8b3R
Validator key    : 1NS8Qs94HRza6C854Y9q7NSdHRJnkjf9nPx2sb7VYDp66kFDjL
Committee key set: 121VhftSAygpEJZ6i9jGkAnFWAnsZj2BjYRUaqkvpix4uZGprCKuLB9VCVqcSUzhXB5YZbq56oAd3HF3oiPhS1jBLko6B4L9EpTarT982GYPtVH7RJCJoMmbNGi8fkjbgy2QXRKwLwGCTXZwvW1g93AgqtMcYWBf8krS1HKY1JTy86j3uWGifi9gDKMDwqqdgLvZfSP2rGBHK7XEw9RnJ7GA43Eqeu8hWDAogt2aU3A7K8KV4iZMRJ2n7Noxmb3DtiboBoobjZcTCM7wPJGa9AxmfCBBtt1UqfCf4Qt8ZHhFvP7KJTYM5ToxwukuURDD3LvZF7ZVydrx7MVXGuYas3gdp8VB4jbaq1SDRFJ36eBEHLe5yATSh1r7HUTosD7TZMtp3wR9AX4vKpGZSJ7jeWFijNMNWdKXrVXmDpBonBECESvK
------------------------------------------------------------
1
Payment Address  : 12S1FVsi5eBSFeGEr6YVBYDdF54dyJZvE59cuZ9dFK9RSkcjznSnDLqaWtsCCFTxPzBWMSA5MfepkdTma8ZAZjJqqadthN2C3QCNh5N
Private Key      : 112t8rnfXYskvWnHAXKs8dXLtactxRqpPTYJ6PzwkVHnF1begkenMviATTJVM6gVAgSdXsN5DEpTkLFPHtFVnS5RePi6aqTSth6dP4frcJUT
Public key       : 12H791Xr9R7EwAZeGG38Lgg6iHZngogddR9NxtMw2gdYkqik3NP
ReadOnly key     : 13hcqfFcqSFqGMviUAir2ECZAo6qBKZSLHMSAXpDLGGfN8Z9FEEjcyPj7AADH1BCd3tgJ6ukpYzrTSyMweWMofu2NgroNJSqSbUtzWa
Validator key    : 1mv64aA7VRFN4RTKYNcqoAccghojsWk9XZXtQaSF6J48pEqF4q
Committee key set: 121VhftSAygpEJZ6i9jGkPeBLVhCpkkpi7v2wN5soqvfHjD6GtQhBXoEXXqzra5BVQWdDy7oe2Yttoju6bMm2amNHgJBzjoPj78iV16t3s9jQRtNqAqLpB68k4HLyR92vp2GWSwhwVeVcWfuEoQUnwFSed5qiJ2BfP8dJTWwUm6g2yrBx2uMB7Uhbg9EbcAw9Sese6jN6T3yEaS4YKeJN5ijzgqJX496FT23vdiy24bJdggdEb9b5p1b7ir7JQzoxy1AKm9mzBswgc3eByRkNZTkgwHRwxbHTDaB8WyDBQCcNu5bjAM5JfxewNdRRmAdGbNeLngP6AFL3CmPYbeiecbgn86GDc9nfQfWJcu4mZJnAzXFPfVgCQ33qPK4XVxbqhWj59WbmTLADtsj4MSq55Vfj5fgJgwewfbYvmNpwWLjvauC
------------------------------------------------------------
 ***** Shard 1 **** 
0
Payment Address  : 12RquWY3vpaSPMtAQEozAB1pgbJJnnphzhJTux2VGaX5eHBxYGKcUTYEqJqQdAUsjzr8cpNQRnSTygnduxBpBvrqH1XthdrJMxCQyaC
Private Key      : 112t8rnY3WLfkE9MsKyW9s3Z5qGnPgCkeutTXJzcT5KJgAMS3vgTL9YbaJ7wyc52CzMnrj8QtwHuCpDzo47PV1qCnrui2dfJzKpuYJ3H6fa9
Public key       : 1BwH1jG6Ei23L54RtFdr63d88Ua4UkD1wMmAN9aPRBcMUJCD7i
ReadOnly key     : 13hTVfuxgceqQ5Ye2JzLzrzkcKLVzopE6uWHAvh5MXeKZf8MwuPmBWYVnWao4mYNNsz6EBLVzcek4KRgYhaM3jB4TQ4PVhtf7KcjdjC
Validator key    : 12wBQta4G1L29hVWKE7fPEm1YURjtidA4pc2Q5tmGnA1PxHmsyH
Committee key set: 121VhftSAygpEJZ6i9jGkBMcUpQm2hozNNwSpTz9bmVN8FTbQj167B4AmetT8VUFVyVMs6AcBrpQKaB9iqTCVycbnsmRkfuYHDEFGfo85o8c6vsF3AcvCcMRcVTAt1C8oAy5nSTg2T7K7jrCmcqdgXWdHRRFHpceCzNyBm3CoWWPxknFif2TD6PVRRfGVwyJEKRufhfQN7Mcq4rCUCkJzjz6vYN2NTssxspegwTiyvXGB7q6fcRHDh7JAVHkaH7oLQpwD9RsZc1qVTvsBPoywaznSqhhCmqkxsoSwGvifVqMp2h3W2ZndLs3JRBvRGdYqewPkMD9t8Jp7xWhbP7XsBXQP6DkRfFGrY9ZjHMbdwADUUCDbU5216ugyjKtQk1bN7747eTCCmucGPSLRjPgQq9PZb8EK96vXDffTmpKS6kuXisX
------------------------------------------------------------
1
Payment Address  : 12S1X46G28CSfcAnxz1bT6LymKuLodW9RDT9hLckBVzGozAiCGoZ1xP9yA6DpyUQYuXCQXvW1fkUeNSJRryQwHtvFUh5WnFycngdzu5
Private Key      : 112t8rnYRAAQ9BqLA9CF7ESWQzAAUBL1EZQwVPx4z5gPstyNpLk9abFp7iXQFu1rQ5xKukKtvorrxyetpP6Crs7Hj7GeVaVPDL5oW12zx6sQ
Public key       : 12JvJ2NjFSenbk7GKmnjRpznPzBVVZkpHvaejVW44hUAGHmi8Em
ReadOnly key     : 13hd7DUAmvGqgKqGb4BxHnKuh3wY1eVfXRexxKHLGT7WjN77Uv7KR41rJuPCKMXr5r3MT4XNdB51EE2pfvA2tA8VAw3G1oup43dTdpY
Validator key    : 1toGe3bMs61zQJyofSYGG5Rca1TL6dnf4dtFUXhB55xRt4Aypa
Committee key set: 121VhftSAygpEJZ6i9jGkPwHkhs51e63kTqfpfs7pSZTqfozwtHS8kdNK5BjciRQounVVirKQACZkXqh1dCFZyTiCgi2kwxKm8pi7EACx7S8WCe3Pcj4r1KitETVYxZfmiodKyaa9pHDdey9AFFjWzRGuR7roQGwYsATqYhZ1cEtsLvUHTiQY9p876nKis3qDUUYwDdXUWN12ZK6fgBoShayAfNgtnqd42WRK5oTMPXbWbHcmsDqXxY2wuEdpnvJV6jTQEd7TUwToMrJ5sg3yzT5VjDZoZXe9hDDvjocZVQaDUNNgDdVhBwVbszyoVrymry1vZpJFBMCJQPpFvBcEcor7YFpAbDiyJXaP4SBGtytMBiJudoqrSPd4PdTpZ4Vkmv9JCnkQvryCPe64j2fKdTqQJG3rH1amCAUEff6Ek8joG89
------------------------------------------------------------
```

5. Generate genesis block data. Use the address of shard-0 that we created at step4, as the address to received the initial PRV. Edit the file *utility/transaction/main.go*

- Comment the line `initThankTx`
```
// initThankTx(db)
```

- Replace the privatekey and initial amount as you want
```
testUserkeyList := map[string]uint64{
		"112t8rncBDbGaFrAE7MZz14d2NPVWprXQuHHXCD2TgSV8USaDFZY3MihVWSqKjwy47sTQ6XvBgNYgdKH2iDVZruKQpRSB5JqxDAX6sjMoUT6": uint64(5000000000000000),
	}
```

- Run the script. For eg.: (5mil PRV)

`go run utility/transaction/main.go` 

```
[{
  "Version": 1,
  "Type": "s",
  "LockTime": 1571901589,
  "Fee": 0,
  "Info": null,
  "SigPubKey": "au7tQYOeTVqLKkxlRnGo1zGJelpAr0aE0mUqhVtGvTg=",
  "Sig": "UcnJxAMxc/iioW+Q0IxAMifM1t+PPDMlXFtKB+dTtAl7Gy0n7sEWJmIdPnhyzAki47H5Zoerwiznm4iSJ0+JBg==",
  "Proof": "AAAAAAAAAbAAriBq7u1Bg55NWosqTGVGcajXMYl6WkCvRoTSZSqFW0a9OCCgp32coDuD8DhY25He3dSYXxtQoFaY2FA0DFTOABUrPCDIj9F7UIutBUz21WAFteCwAHOobhMIH3CEKRQsVKdJAiABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAJmX+VPbL7lG9SB1LZYW9OerGo48wVgODS29I+f8JdBgcRw3k34IAAAAAAAAAAAAA=",
  "PubKeyLastByteSender": 56,
  "Metadata": null
}]
```

6. Copy genesis block data to *./incognito-chain/blockchain/constantstx.go*, 
replace the output at step5 to the section TestnetInitPRV. 
```
var TestnetInitPRV = []string{
	`{
		"Version": 1,
		"Type": "s",
		"LockTime": 1571901589,
		"Fee": 0,
		"Info": null,
		"SigPubKey": "au7tQYOeTVqLKkxlRnGo1zGJelpAr0aE0mUqhVtGvTg=",
		"Sig": "UcnJxAMxc/iioW+Q0IxAMifM1t+PPDMlXFtKB+dTtAl7Gy0n7sEWJmIdPnhyzAki47H5Zoerwiznm4iSJ0+JBg==",
		"Proof": "AAAAAAAAAbAAriBq7u1Bg55NWosqTGVGcajXMYl6WkCvRoTSZSqFW0a9OCCgp32coDuD8DhY25He3dSYXxtQoFaY2FA0DFTOABUrPCDIj9F7UIutBUz21WAFteCwAHOobhMIH3CEKRQsVKdJAiABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAJmX+VPbL7lG9SB1LZYW9OerGo48wVgODS29I+f8JdBgcRw3k34IAAAAAAAAAAAAA=",
		"PubKeyLastByteSender": 56,
		"Metadata": null
	}`,
}
```


7. Modify the following params in *./incognito-chain/blockchain/constants.go*
```
TestRandom              = true			// System will auto generate the random number
TestnetEpoch            = 20			// An epoch = 20 beacon blocks
TestnetRandomTime       = 10			// At beacon blocks 10th, the random number will be generated

TestNetShardCommitteeSize     = 6		// Maximum number of committee in a shard
TestNetMinShardCommitteeSize  = 4		// Minimum number of committee in a shard
TestNetBeaconCommitteeSize    = 4		// Beacon committee size
TestNetMinBeaconCommitteeSize = 4		// Minimum number of committee in beacon
TestNetActiveShards           = 2		// Number of Shard in Incognito Blockchain
```
8. Modify the following params in *./incognito-chain/blockchain/params.go*: (Under ChainTestParam section)
```
CheckForce:   false, 					          // Avoid system update when received signal from Master Server
BeaconHeightBreakPointBurnAddr: 1,      // Apply newest burning address started from block beacon no. 1
```

9. Generate 12 keyset for committee node:
- edit the *./incognito-chain/utility/genkeywithpassword.go*
```
numberOfKey := 6                              		// Number of keyset that you want to be generated. 6 address in 2 shard: 6x2=12 (keyset)
randomString := []byte("YourRandomStringTwo")    	// A random string used to create keyset. The same string create the same keyset
```
- Execute the script to generate keyset:

`go run utility/genkeywithpassword.go `

- Sample Output:

`go run utility/genkeywithpassword.go `
```
 ***** Shard 0 **** 
0
<keyset_0>
------------------------------------------------------------
1
<keyset_1>
------------------------------------------------------------
2
<keyset_2>
------------------------------------------------------------
3
<keyset_3>
------------------------------------------------------------
4
<keyset_4>
------------------------------------------------------------
5
<keyset_5>
------------------------------------------------------------

 ***** Shard 1 **** 
0
<keyset_0>
------------------------------------------------------------
1
<keyset_1>
------------------------------------------------------------
2
<keyset_2>
------------------------------------------------------------
3
<keyset_3>
------------------------------------------------------------
4
<keyset_4>
------------------------------------------------------------
5
<keyset_5>
------------------------------------------------------------
```

10. Edit the *./incognito-chain/keylist.json*, replace the PaymentAddress and CommitteePublicKey that we generated at step9
 - use 2 shard-0 keyset for `beacon-0` & `beacon-1`
 - use 2 shard-1 keyset for `beacon-2` & `beacon-3`
 - use 4 shard-0 keyset for `shard-00 01 02 03`
 - use 4 shard-1 keyset for `shard-10 11 12 13 `
 - ignore the keylist in `shard-2 3 4 5 6 7`

11. Edit the *./incognito-chain/start_node.sh*, replace the PrivateKey that we generated at step9


12. Build Incognito binary file:
```
cd ./incognito-chain/
go build -o incognito
```

13. Create Tmux session

`bash create_tmux.sh`

14. Start the chain

`bash start_chain.sh`

15. Verify that the chain is running: go to each tmux session, you would see the running log on screen.
eg:

`tmux a -t fullnode`

#SECTION III: TEST THE CHAIN
--------------------
Incognito Blockchain can be tested by making RPC request.

**Get block chain info:**
```
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"jsonrpc":"1.0","method":"getblockchaininfo","params":[],"id":1}' \
  http://192.168.0.1:9354
  ```
  
**Get balance by private key:**
```
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"jsonrpc":"1.0","method":"getbalancebyprivatekey","params":["<private_key>"],"id":1}' \
  http://192.168.0.1:9334
```

**Send PRV:**
```
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"jsonrpc": "1.0","method": "createandsendtransaction","params": ["<private_key>",{"<payment_address>": <ammount_in_nanoPRV>}, -1, 0],"id": 1}' \
  http://192.168.0.1:9334
```

**Eg:**
- Get balance address 0
```
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"jsonrpc":"1.0","method":"getbalancebyprivatekey","params":["112t8rnX5E2Mkqywuid4r4Nb2XTeLu3NJda43cuUM1ck2brpHrufi4Vi42EGybFhzfmouNbej81YJVoWewJqbR4rPhq2H945BXCLS2aDLBTA"],"id":1}' \
  http://vps162:9334
```
- send from address 0 to address 1 (same shard)
```
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"jsonrpc": "1.0","method": "createandsendtransaction","params": ["112t8rnX5E2Mkqywuid4r4Nb2XTeLu3NJda43cuUM1ck2brpHrufi4Vi42EGybFhzfmouNbej81YJVoWewJqbR4rPhq2H945BXCLS2aDLBTA",{"12RyJTSL2G8KvjN7SUFuiS9Ek4pvFFze3EMMic31fmXVw8McwYzpKPpxeW6TLsNo1UoPhCHKV3GDRLQwdLF41PED3LQNCLsGNKzmCE5": 99000000000}, -1, 0],"id": 1}' \
  http://vps162:9334
```
- send from address 0 to address 2 (cross shard)
```
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"jsonrpc": "1.0","method": "createandsendtransaction","params": ["112t8rnX5E2Mkqywuid4r4Nb2XTeLu3NJda43cuUM1ck2brpHrufi4Vi42EGybFhzfmouNbej81YJVoWewJqbR4rPhq2H945BXCLS2aDLBTA",{"12RtmpqwyzghGQJHXGRhXnNqs7SDhx1wXemgAZNC2xePj9DNpxcTZfpwCeNoBvvyxNU8n2ChVijPhSsNhGCDmFmiwXSjQEMSef4cMFG": 69000000000}, -1, 0],"id": 1}' \
  http://vps162:9334
```
- Get balance address 0
- Get balance address 1
- Get balance address 2

#SECTION IV: WRITING GOLANG TEST SCRIPT:
-----------------------------------
- Check the file `./incognito-chain/testexample/test-chain.go` for sample test case template
- Execute the test script:

`go run test-chain.go`

- Modify it as you want

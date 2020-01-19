# Burn Address


* Burn address: is a special address on Incognito chain, which is not accessible by anyone, since no one owns the private keys to that address.

* Burned assets (coins or tokens): mean eliminated assets by sending them to a burn address.

* Minted assets (coins or tokens): mean created new assets then distributed them to some address on Incognito chain.

In Incognito chain, we use "burn and mint mechanism" to archive consensus in some cases such as stake/unstake, deposit/withdraw token in the bridge, and pDEX.
  
#### How to generate Burn Address
* The formula to generate burn address is as follows:

                      Hash(G, "burningaddress", index)

    where 
    * ```Keccak256``` is used as the hash function,
    * ```G``` is the base point of Edwards25519,
    * ```"burningaddress" ``` is a hard code string, 
    * ```index``` is the minimum positive integer number such that the last byte of output is zero (it will make sure that burn address belongs in Shard 0).
   
* The burn address at ```index = 68``` in ```base58encode``` is:
```15pGE5XZc7gsHSLgQxgZRVt7UsXBodiyTYe6xtvXipTFQPMpdQqtXhkzriL```

### Why no one can recompute the private key

Based on two assumptions:
* Function ```Hash```  is one-way and unpredictable, the output of ```Hash```  is an unpredictable point ```H=xG```
* To recover the private key ```x```, we need to solve the discrete log problem. The probabilistic to solve the discrete log problem on Edward25519 is negligible. 

Thus, no one can recompute the private key.  
 
#### Reference
[1] https://github.com/incognitochain/incognito-chain/blob/dev/master/utility/generateburnaddress.go
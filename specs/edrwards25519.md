# New Implementation based on Edward25519 

## Introduction
Elliptic Curve Cryptography (ECC) is a form of public-key cryptography from which the security relies on the Elliptic Curve Discrete Logarithm Problem.  Recently, ECC has become the predominant cryptosystem found in most blockchain projects (except some projects based on post-quantum cryptosystem).

At the first stage of the Incognito project, we use the P-256 ECC library built-in of Golang to implement cryptography protocols. However, the performance is not as we expected. Finally, we decide to use the Edwards25519 library from Dero Project [1], this curve is also used in Monero Project. 

## What is Edwards25519
Edwards25519 is a variant of Curve25519 built by Bernstein in 2006 that is very useful for digital signatures scheme. This is the standard Edwards25519 curve definition from [2]. 

#### Curve equation: 
Edwards25519 is defined over the prime field ![equation](https://latex.codecogs.com/gif.latex?F_%7B2%5E%7B255%7D%20-19%7D) as follows:

![equation](https://latex.codecogs.com/gif.latex?-x%5E2%20&plus;%20y%5E2%20%3D%201%20-%20%5Cfrac%7B121665%7D%7B121666%7D%20x%5E2y%5E2)

#### Base Point:
G = (x, 4/5) 

This is a specific point on the curve. It is used as a basis for all calculations on this curve. 

#### Order of the base point: 
![equation](https://latex.codecogs.com/gif.latex?l%20%3D%202%5E%7B252%7D%20&plus;%2027742317777372353535851937790883648493)

This value is a prime number specified by the curve authors. The ![equation](https://latex.codecogs.com/gif.latex?l) defines the maximum scalar we can use.  

#### Total number of points on the curve:
 
![equation](https://latex.codecogs.com/gif.latex?q%20%3D%202%5E%7B255%20%7D-%2019). 

This is a prime number and gives the name of the curve.

## Why is it better than our previous scheme

**Security**: In 2013 Bernstein and Lange started a website called SafeCurves that lists criteria for a structured analysis of the security of different elliptic curves. It can be observed that Curve25519 passes all the criteria. This result is applicable to Edwards25519 as well because the used curve is equivalent to Curve25519. P-256, however, does not pass all criteria. In particular, the following four criteria: Rigidity, Completeness, Infistinguishability, Ladders. See more detail in [3].

**Performance**: The performance of Edwards25519 and P-256 were similar when comparing the fast assembly language implementation of P-256 and the reference implementation of Edwards25519 that were available in OpenSSL as of June 2017. However, our benchmarks show that the performance of cryptography protocols based on Edwards25519 is more times faster than the implementation based on P-256.

##Benchmarks

These results show that the implementation based on Edwards25519 at least two times faster the implementation based on P-256. This improvement will help the Incognito chain can produce more transactions in each block. 

| Number of Outputs | 2       | 4        | 8        | 16       |
|:-----------------:|:-------:|:--------:|:--------:|:--------:|
| Prove             | 34.06ms | 121.47ms | 242.29ms | 473.66ms |
| Verify            | 17.62ms | 66.26ms  | 130.95ms | 263.93ms |

Table 1: Performance of Prove and Verify functions based on Edwards25519


| Number of Outputs | 2        | 4        | 8        | 16       |
|:-----------------:|:--------:|:--------:|:--------:|:--------:|
| Prove             | 106.10ms | 398.58ms | 788.21ms | 1566.16ms |
| Verify            | 44.74ms  | 176.24ms  | 352.49ms | 709.98ms |

Table2: Performance of Prove and Verify functions based on P-256 


## References
[1] https://github.com/deroproject/derosuite/tree/master/crypto

[2] Bernstein, Daniel J.; Lange, Tanja (2007). Kurosawa, Kaoru (ed.). Faster addition and doubling on elliptic curves. Advances in Cryptology—ASIACRYPT.
 
[3] https://safecurves.cr.yp.to/index.html


#!/bin/bash
cd /go/src/github.com/ninjadotorg/constant
git pull
/usr/local/go/bin/go build
./cash-prototype --norpc --discoverpeers --generate --producerkeyset "13horJt6gBxUDDcx2teNjSg4kTMoTcw6Zt1NGKVYJjzrAoPQ2gZqWuFzZuLWGWEcH9rZhineEEkuXxtTvEoHezrJAT1L38BrBSCoaC98YhrWdcf396hBZ1CKH9vw3C7adt3QcjEwka192qnwAbyc3FaDZz7mLH8aqbqwbArQBWytgC4p4BWNEpwLSu8wBhEwbgYA4ySLdF1geeG5qudNVSRwwEyQMxQ4qw2aYMDf3AaXGiq8QT6bG15kddQGAoUt8XLRLTptKs4T19N4771ZHXHRwiWhf58hSMpQJyuzN9uz4z171Dbd6Hzu3Nvae2VQ8NKPCcMVYSwzJmqfogMEeWwqghgWnFK6nsKmWCNXFQHzsp2zE5Ggo8uX3aAXuv8PrNX2B8YuhtVS87WhNKBACvptwGsq7eYzc3PZMA4Gij7oqR3cUdtxZ5X4GdisBXjEp3rHvoy41r8qJujzgKcoqWspeQKfVG7w7wAaANUZytizYJXxyzxRqnhNfQskcgscSgbiAcngh2zqqYXzHGuNSKqxzmwdvjXCygSpzN2EXSHn5fvBDAc68NwqzWhjS626gmmRBcbsf7uk4A8BTMomJq2ZyGWU58tqqo2EedBH386S3G499beAVydMP7HJt8e9T3nSJZMWhKa26gjo6BGitLXSbtMrouhkBt4TwwUJ3WsEZLvSFJpWAsKmR1pZKZsDzLXrzqLvUpq8AHE727yxGKWg7ocKXNPJGU2azBzPqtXAhFeYnK69pkqQGtVXmhYjjTNhPLeiCfJN"
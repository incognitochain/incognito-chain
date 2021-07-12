#!/usr/bin/env bash
#GrafanaURL=http://128.199.96.206:8086/write?db=mydb
###### MULTI_MEMBERS
# Shard 0
if [ "$1" == "shard0-0" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnkwspf7rutf9tYXhb77N3LBTxfmU6oyE7eMQ4wyy8iH7gEZzeePsXX8fx9ZjY8ozeWySuYnWZzwACoDog4GZzViUe5kH46hGxBJzrf"  --datadir "data/shard0-0" --listen "0.0.0.0:9434" --externaladdress "0.0.0.0:9434" --norpcauth --rpclisten "0.0.0.0:9334" --rpcwslisten "0.0.0.0:19334"
fi
if [ "$1" == "shard0-1" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnr1HRcYozjwt1Y4t1o7kAvKM876kQeN7s9jCUodhFbsrPZ1UU6vDcSSFAQMAMLz6i576gpVmyy3K4py9ao3aF7hqozCqSw6amKnUjw"  --datadir "data/shard0-1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9335" --rpcwslisten "0.0.0.0:19335"
fi
if [ "$1" == "shard0-2" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rpiamZoiyGXTkJ8mMsLxJwgnyUUTTrfpdnpm3KckC3VjaZRzpU24jQTfmfztZUKYqdJUKUt7BrgMKMqKLjqczwffmCx9QM5D1PJShPX"  --datadir "data/shard0-2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336"
fi
if [ "$1" == "shard0-3" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rqDSA3EThnGp2doR4zuRdgGRiGLdyYi6Deo1QTwW41nttvmo5dVEUjbE3PTW2wtWc9cthKA4ZmJjwQyRShqqXL5NQGRUhoVYjRCR5Qo"  --datadir "data/shard0-3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337"
fi
if [ "$1" == "shard0-new-0" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rospAEaouNQgnK8vAAJGzH6ysLAGeZGqmZ5RJTT7CrF1zK8zwqVqx4DEdoD6MDTNiSK9W1vbZXtVe7vqvfEuf6LpuBbUiHvvkfF9L3X"  --datadir "data/shard0-0" --listen "0.0.0.0:9434" --externaladdress "0.0.0.0:9434" --norpcauth --rpclisten "0.0.0.0:9334" --rpcwslisten "0.0.0.0:19334"
fi
if [ "$1" == "shard0-new-1" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rs4QMxRdQLLyG7b5ifeBwH39TuZU92UuTXHynDKVgtE366Jd6qs99gj6gKtz46ad5NKXbaJ2UyXfxbjCouhtN6Es8ve5yyQVyEtgXod"  --datadir "data/shard0-1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9335" --rpcwslisten "0.0.0.0:19335"
fi
if [ "$1" == "shard0-new-2" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rtB7mXUvqSqgGJZTa84JwUJyruLXNYwrypD98t8UkAGxhBhhs6P696Z7iZ1WxdhWEFKeDbEkR5PdXNf4V8CRwmszhjUAU6AqyXQ6ME7"  --datadir "data/shard0-2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336"
fi
if [ "$1" == "shard0-new-3" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rwXq7wh46QCmZUuqUTaQwbrXi7mvZnPayhjaUqa2MCnGbXRxHQhxyShfRLpvQzUzqqmDGMqfmLR6R2WayyyBhXU1b6Hhz8CnZVPTBG7"  --datadir "data/shard0-3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337"
fi

if [ "$1" == "staker-0" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "12MZ4QiFoETNbdLKgRQWPMQMqsceWPKo71Jma9NzwvLTabpcDhn"  --datadir "data/shard2-0" --listen "0.0.0.0:9442" --externaladdress "0.0.0.0:9442" --norpcauth --rpclisten "0.0.0.0:9342"
fi
if [ "$1" == "staker-1" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "158ZGK5EHmoyrHEd8aA2HCaqbNQ4r45ZsnwL4Zh8mH8dueWHWs"  --datadir "data/shard2-1" --listen "0.0.0.0:9443" --externaladdress "0.0.0.0:9443" --norpcauth --rpclisten "0.0.0.0:9343"
fi
if [ "$1" == "staker-2" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "1G5Q9uGSxekPSgC1w1ZFaDJ8RxeYrekk2FtFLF33QCKNbg2V88"  --datadir "data/shard2-2" --listen "0.0.0.0:9444" --externaladdress "0.0.0.0:9444" --norpcauth --rpclisten "0.0.0.0:9344"
fi
if [ "$1" == "staker-3" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "12MZ4QiFoETNbdLKgRQWPMQMqsceWPKo71Jma9NzwvLTabpcDhn"  --datadir "data/shard2-3" --listen "0.0.0.0:9445" --externaladdress "0.0.0.0:9445" --norpcauth --rpclisten "0.0.0.0:9345"
fi
if [ "$1" == "staker-4" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "1cQCTV1m33LxBKpNW2SisbuJfp5VcBSEau7PE5aD16gGLAN7eq"  --datadir "data/shard3-0" --listen "0.0.0.0:9446" --externaladdress "0.0.0.0:9446" --norpcauth --rpclisten "0.0.0.0:9346"
fi

# Shard 1
if [ "$1" == "shard1-0" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rncXBFzLgSGvci44XrLfzjnXELeBn6sifqbAictF4cyiKSdssN5ud43zuoa5zTaKNiqdv5LgPPKmN4Zmu4CmGxmzbZeELxprhrfAYcu"  --datadir "data/shard1-0" --listen "0.0.0.0:9438" --externaladdress "0.0.0.0:9438" --norpcauth --rpclisten "0.0.0.0:9338" --rpcwslisten "127.0.0.1:19338"
fi
if [ "$1" == "shard1-1" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8roiAj2dbQtiDRQnLug5uUBfw8mm2FcqfVb5xUwNs18GAvot9xScRaigH1RQDoM1ghWbCN3ZLtBVUtpSe7mTj2HKYAznkfPd1ZZQEb6o"  --datadir "data/shard1-1" --listen "0.0.0.0:9439" --externaladdress "0.0.0.0:9439" --norpcauth --rpclisten "0.0.0.0:9339" --rpcwslisten "127.0.0.1:19339"
fi
if [ "$1" == "shard1-2" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rrBVDoKsTXFUicsptwgq5NMHupjDp63q4tmuaAmetWWhFhRCjnQMtpk1fEatrmGZSrYyqXrgppD7d2jKSQgRQGju5b23SSDxGab1esB"  --datadir "data/shard1-2" --listen "0.0.0.0:9440" --externaladdress "0.0.0.0:9440" --norpcauth --rpclisten "0.0.0.0:9340"
fi
if [ "$1" == "shard1-3" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rrt9rfVxZs7cGfxnUswThPhByn8UPef31KVkHeKCGtdT36hgNtMuXKNgFj9W8FSN6nafQSZsiLqCgpyJWJRe3cEA75SmKw8Gbov77P6"  --datadir "data/shard1-3" --listen "0.0.0.0:9441" --externaladdress "0.0.0.0:9441" --norpcauth --rpclisten "0.0.0.0:9341"
fi
# Shard 2
if [ "$1" == "shard2-0" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnX2Fngsy1KuJv5GLLNUeB3gZwoHMss2cqRr1ECa2ibR7FQNUyE7kMFvq7rGtqVJULo8XfAxThuLxUwd8vv76MojbL3wPhxmTvbcd2S"  --datadir "data/shard2-0" --listen "0.0.0.0:9442" --externaladdress "0.0.0.0:9442" --norpcauth --rpclisten "0.0.0.0:9342"
fi
if [ "$1" == "shard2-1" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXRrZ1gC7MYNFVUA1paZrE3iSiAb9AR7Z5quNBgR2ovrcfj8p4kTb3ynx6ddjnoPey3qA2vRiP17tCvpCHU9xBDwMq8D1Mg2GBM9eC"  --datadir "data/shard2-1" --listen "0.0.0.0:9443" --externaladdress "0.0.0.0:9443" --norpcauth --rpclisten "0.0.0.0:9343"
fi
if [ "$1" == "shard2-2" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXbsJF4f5xtzM2jPW6dCYHChNksth7X64iUkLR4bGi2JBVjgJQRLeKRsdbiFaYMsxzrfbfKAp4TELGre45QkxHWCnwVXPGnnZjJKVL"  --datadir "data/shard2-2" --listen "0.0.0.0:9444" --externaladdress "0.0.0.0:9444" --norpcauth --rpclisten "0.0.0.0:9344"
fi
if [ "$1" == "shard2-3" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnY4DeSGZYb8r8sSN5WJr6ZL3NCafYAQ7f7Am9KXQDGSc3Qddpn7BfHW1i6CoVVk8vKEzJ25vA9uc9EdhoLU98eoUw7fMrPPrBdNB7Q"  --datadir "data/shard2-3" --listen "0.0.0.0:9445" --externaladdress "0.0.0.0:9445" --norpcauth --rpclisten "0.0.0.0:9345"
fi
# Shard 3
if [ "$1" == "shard3-0" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXBg35mhRzZDw7nNQCiRcM9QLg2DTHewY7kRZ3XCmgkdQa4iVcMUwMd5DTvvjEcCvv3SnCo5zSQpS93zskAxG6tdvR1QPBxtCmaBCK"  --datadir "data/shard3-0" --listen "0.0.0.0:9446" --externaladdress "0.0.0.0:9446" --norpcauth --rpclisten "0.0.0.0:9346"
fi
if [ "$1" == "shard3-1" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXa3USo2wMdyuZTfurZrYe1ZJhqibnp9GHx8Jkf9dh5cU39sjqBTKoWPtNHvVZn2eqGj6V26PmELez85bUUBMBKG6tqFQrer2GkuJ4"  --datadir "data/shard3-1" --listen "0.0.0.0:9447" --externaladdress "0.0.0.0:9447" --norpcauth --rpclisten "0.0.0.0:9347"
fi
if [ "$1" == "shard3-2" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXc38wPUsDdRasThfk58ME3KoSdeyXJS3tsyRuxPrSkvuc8MBT1gxiQjSCMBbifqibxEVAHimGcfnLbjVEEett8FtwFuBQ8zAUHTet"  --datadir "data/shard3-2" --listen "0.0.0.0:9448" --externaladdress "0.0.0.0:9448" --norpcauth --rpclisten "0.0.0.0:9348"
fi
if [ "$1" == "shard3-3" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnY6BgeU6bk7Nui1TwB2w2MsG3UeSXrgJ1n2tGVs66Qk9YT6E15PbXR4ai7eW1qyPrW7a2AUtN2otuXtBAZKtm2DDiUxh3ngZ6JPYHv"  --datadir "data/shard3-3" --listen "0.0.0.0:9449" --externaladdress "0.0.0.0:9449" --norpcauth --rpclisten "0.0.0.0:9349"
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "117FYq9sjG87ny4nFq9cqdtFonzJ5BFpdidLJ7FCXe2dmxEEBu"  --datadir "data/beacon-0" --listen "0.0.0.0:9450" --externaladdress "0.0.0.0:9450" --norpcauth --rpclisten "0.0.0.0:9350"
fi
if [ "$1" == "beacon-1" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "12CeFkSTYxzTG8eQyEeGaof4VWGyqav6qMrXUkwtHqKU1yA2cRU"  --datadir "data/beacon-1" --listen "0.0.0.0:9451" --externaladdress "0.0.0.0:9451" --norpcauth --rpclisten "0.0.0.0:9351"
fi
if [ "$1" == "beacon-2" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "12njcQBofHdiLVA6qZAC9apvBnS2SSfTHon2AsBixKBfahz7eQR"  --datadir "data/beacon-2" --listen "0.0.0.0:9452" --externaladdress "0.0.0.0:9452" --norpcauth --rpclisten "0.0.0.0:9352"
fi
if [ "$1" == "beacon-3" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "1TmwTeXvXQMfb82Hb7fPuGGuySwTqQrXERMEzmZgUMcJq2ybYX"  --datadir "data/beacon-3" --listen "0.0.0.0:9453" --externaladdress "0.0.0.0:9453" --norpcauth --rpclisten "0.0.0.0:9353"
fi
# Beacon
if [ "$1" == "beacon-new-0" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8sSijugr8azxAiHMWS9rA22grKDv5o7AEXQ9datpT1V7N5FLHiJMjvfVnXcitL3fpj35Xt5DNnBq8iFq618X31nCgn2RjrYx5tZZWCtj"  --datadir "data/beacon-0" --listen "0.0.0.0:9460" --externaladdress "0.0.0.0:9460" --norpcauth --rpclisten "0.0.0.0:9360"
fi
if [ "$1" == "beacon-new-1" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8sSjSEck5J5RKWGWurVfigruYDxEzjjVPqHaTRJ57YFNo7gXBH8onUQxtdpoyFnBZrLhfGWQ4k4MNadwa6F7qYwcuFLW9R1VxTfN7q4d"  --datadir "data/beacon-1" --listen "0.0.0.0:9461" --externaladdress "0.0.0.0:9461" --norpcauth --rpclisten "0.0.0.0:9361"
fi
if [ "$1" == "beacon-new-2" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8sSjkAqVJi4KkbCS75GYrsag7QZYP7FTPRfZ63D1AJgzfmdHnE9sbpdJV4Kx5tN9MgbqbRYDgzER2xpgsxrHvWxNgTHHrghYwLJLfe2R"  --datadir "data/beacon-2" --listen "0.0.0.0:9462" --externaladdress "0.0.0.0:9462" --norpcauth --rpclisten "0.0.0.0:9362"
fi
if [ "$1" == "beacon-new-3" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8sSj637mhpaJUboUEjkXsEUQm8q82T6kND3mWtNwig71qX2aFeZegWYsLVtyxBWdiZMBoNkdJ1MZYAcWetUP8DjYFnUac4vW7kzHfYsc"  --datadir "data/beacon-3" --listen "0.0.0.0:9463" --externaladdress "0.0.0.0:9463" --norpcauth --rpclisten "0.0.0.0:9363"
fi
# FullNode testnet
if [ "$1" == "fullnode" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=100 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" GETH_NAME="http://127.0.0.1:8545" GETH_PORT="" GETH_PROTOCOL="" --relayshards "all" --datadir "data" --listen "0.0.0.0:9433" --externaladdress "0.0.0.0:9433" --norpcauth --rpclisten "0.0.0.0:8334" --rpcwslisten 0.0.0.0:18334 --txpoolmaxtx 100000
fi
if [ "$1" == "fullnode-devnet" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=100 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "139.162.55.124:9330" GETH_NAME="https://kovan.infura.io/v3/1138a1e99b154b10bae5c382ad894361" GETH_PORT="" GETH_PROTOCOL="" --relayshards "all" --datadir "/data/devnet" --listen "0.0.0.0:9433" --externaladdress "0.0.0.0:9433" --norpcauth --rpclisten "0.0.0.0:8334" --rpcwslisten 0.0.0.0:18334 --txpoolmaxtx 100000
fi
if [ "$1" == "fullnode2-devnet" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=100 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "139.162.55.124:9330" GETH_NAME="https://kovan.infura.io/v3/1138a1e99b154b10bae5c382ad894361" GETH_PORT="" GETH_PROTOCOL="" --relayshards "all" --datadir "/data/devnet2" --listen "0.0.0.0:39433" --externaladdress "0.0.0.0:39433" --norpcauth --rpclisten "0.0.0.0:38334" --rpcwslisten 0.0.0.0:48334 --txpoolmaxtx 100000
fi
if [ "$1" == "fullnode-testnet-b" ]; then
GO111MODULE--usecoindata --coindatapre="__coins__" =on GETH_NAME=kovan.infura.io/v3/1138a1e99b154b10bae5c382ad894361 GETH_PROTOCOL=https GETH_PORT="" INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --testnet true --nodemode "relay" --relayshards "[0]" --externaladdress "127.0.0.1:9433" --enablewallet --wallet "wallet" --walletpassphrase "12345678" --walletautoinit --norpcauth --datadir "../testnet/fullnode" --discoverpeersaddress "testnet-bootnode.incognito.org:9330" --norpcauth --rpclisten "0.0.0.0:8334" --rpcwslisten "127.0.0.1:18338" > ../testnet/log.txt 2> ../testnet/error_log.txt &
fi
if [ "$1" == "fullnode-mainnet" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --testnet true --nodemode "relay" --externaladdress "127.0.0.1:9433" --enablewallet --wallet "wallet" --walletpassphrase "12345678" --walletautoinit --norpcauth --datadir "../mainnet/fullnode" --discoverpeersaddress "mainnet-bootnode.incognito.org:9330"
fi
######
if [ "$1" == "shard-candidate0-1" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f"  --datadir "data/shard-stake" --listen "127.0.0.1:9455" --externaladdress "127.0.0.1:9455" --norpcauth --rpclisten "127.0.0.1:9355"
fi
if [ "$1" == "shard-candidate0-2" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rrEW3NPNgU8xzbeqE7cr4WTT8JvyaQqSZyczA5hBJVvpQMTBVqNfcCdzhvquWCHH11jHihZtgyJqbdWPhWYbmmsw5aV29WSXBEsgbVX"  --datadir "data/shard-stake-2" --listen "127.0.0.1:9456" --externaladdress "127.0.0.1:9456" --norpcauth --rpclisten "127.0.0.1:9356"
fi
if [ "$1" == "shard-candidate0-3" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnY15HgjpbJn1es84ysseB6q9UQ5SwB6Eb82yejEQ3yzhd1dm5ShEiezdfMoEzBgvkuKcFdP5TY3SuWNHXKa1Krprsfxnk5wy7wZ6Dg"  --datadir "data/shard-stake-6" --listen "0.0.0.0:9460" --externaladdress "0.0.0.0:9460" --norpcauth --rpclisten "0.0.0.0:9360"
fi
if [ "$1" == "shard-candidate1-1" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8roHikeAFyuBpdCU76kXurEqrC9VYWyRyfFb6PwX6nip9KGYbwpXL78H92mUoWK2GWkA2WysgXbHqwSxnC6XCkmtxBVb3zJeCXgfcYyL"  --datadir "data/shard-stake-3" --listen "0.0.0.0:9457" --externaladdress "0.0.0.0:9457" --norpcauth --rpclisten "0.0.0.0:9357"
fi
if [ "$1" == "shard-candidate1-2" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rr4sE2L8WzsVNEN9WsiGcMTDCmEH9TC1ZK8517cxURRFNoWoStYQTgqXpiAMU4gzmkmnWahHdGvQqFaY1JTVsn3nHfD5Ppgz8hQDiVC"  --datadir "data/shard-stake-4" --listen "0.0.0.0:9458" --externaladdress "0.0.0.0:9458" --norpcauth --rpclisten "0.0.0.0:9358"
fi
if [ "$1" == "shard-candidate1-3" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rtt9Kd5LUcfXNmd7aMnQehCnKabArVB3BUk2RHVjeh88x5MJnJY4okB8JdFm4JNm4A2WjSe58qWNVkJPEFjpLHNYfKHpWfRdqyfDD9f"  --datadir "data/shard-stake-5" --listen "0.0.0.0:9459" --externaladdress "0.0.0.0:9459" --norpcauth --rpclisten "0.0.0.0:9359"
fi
if [ "$1" == "shard-candidate1-4" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnY15HgjpbJn1es84ysseB6q9UQ5SwB6Eb82yejEQ3yzhd1dm5ShEiezdfMoEzBgvkuKcFdP5TY3SuWNHXKa1Krprsfxnk5wy7wZ6Dg"  --datadir "data/shard-stake-7" --listen "0.0.0.0:9461" --externaladdress "0.0.0.0:9461" --norpcauth --rpclisten "0.0.0.0:9361"
fi
if [ "$1" == "shard-candidate1-5" ]; then
INCOGNITO_CONFIG_MODE_KEY=flag INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rna7M8BYBfNjNHmw3Tie6Yir9mQgp5rSRgUngTqn6A6iSRvAPex4sXsmGxVzXcpUUDfnRfRys3QrPnTHauiipdUNtj7Ef6t3mHUwiC3"  --datadir "data/shard-stake-8" --listen "0.0.0.0:9462" --externaladdress "0.0.0.0:9462" --norpcauth --rpclisten "0.0.0.0:9362"
fi

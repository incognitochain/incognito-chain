#!/usr/bin/env bash
#GrafanaURL=http://128.199.96.206:8086/write?db=mydb
###### MULTI_MEMBERS
# Shard 0
if [ "$1" == "shard0-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXA4JgwFj3xVhHUKj6amVieT5meqgiTQDfubW1atB9fXxnUoLUd2yVxb175k9gikB6sRWinLLUzBq8RRFSyq41SZVz2KCJBhVbo4HR"  --datadir "data/shard0-0" --listen "0.0.0.0:9434" --externaladdress "0.0.0.0:9434" --norpcauth --rpclisten "0.0.0.0:9334" --rpcwslisten "0.0.0.0:19334"
fi
if [ "$1" == "shard0-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXLTy4FNLiKbzpnusifvCD6Mjz1s9BpNqfTM4KVLwP8TGTo8YY4hcBPtfsRihiYMATCMMUHggDWcgQGA7fsLcp5Cwui8UTF3xPsv7L"  --datadir "data/shard0-1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9335" --rpcwslisten "0.0.0.0:19335"
fi
if [ "$1" == "shard0-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXgVNaAxS1ip36w8ZhEQsahdsNeay6hFbdAm7bqGUL9SBpsekQaJUSihw8yoyPyPzu9bLCXXduPHbqBfv4833dNdBBTy4ML6YsSgJF"  --datadir "data/shard0-2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336"
fi
if [ "$1" == "shard0-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnYAvADMBERP5BAYPM9xwWDnDRQMVnHjf7ZHiT7LdW6BkrVz6EhcZHV2cDTvtAm1q4RvaBu7xZ4iE3Etrw7c8ktm48j7qgoctxQZnMn"  --datadir "data/shard0-3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337"
fi
if [ "$1" == "shard0-new-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rospAEaouNQgnK8vAAJGzH6ysLAGeZGqmZ5RJTT7CrF1zK8zwqVqx4DEdoD6MDTNiSK9W1vbZXtVe7vqvfEuf6LpuBbUiHvvkfF9L3X"  --datadir "data/shard0-0" --listen "0.0.0.0:9434" --externaladdress "0.0.0.0:9434" --norpcauth --rpclisten "0.0.0.0:9334" --rpcwslisten "0.0.0.0:19334"
fi
if [ "$1" == "shard0-new-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rs4QMxRdQLLyG7b5ifeBwH39TuZU92UuTXHynDKVgtE366Jd6qs99gj6gKtz46ad5NKXbaJ2UyXfxbjCouhtN6Es8ve5yyQVyEtgXod"  --datadir "data/shard0-1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9335" --rpcwslisten "0.0.0.0:19335"
fi
if [ "$1" == "shard0-new-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rtB7mXUvqSqgGJZTa84JwUJyruLXNYwrypD98t8UkAGxhBhhs6P696Z7iZ1WxdhWEFKeDbEkR5PdXNf4V8CRwmszhjUAU6AqyXQ6ME7"  --datadir "data/shard0-2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336"
fi
if [ "$1" == "shard0-new-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rwXq7wh46QCmZUuqUTaQwbrXi7mvZnPayhjaUqa2MCnGbXRxHQhxyShfRLpvQzUzqqmDGMqfmLR6R2WayyyBhXU1b6Hhz8CnZVPTBG7"  --datadir "data/shard0-3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337"
fi

if [ "$1" == "staker-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "12MZ4QiFoETNbdLKgRQWPMQMqsceWPKo71Jma9NzwvLTabpcDhn"  --datadir "data/shard2-0" --listen "0.0.0.0:9442" --externaladdress "0.0.0.0:9442" --norpcauth --rpclisten "0.0.0.0:9342"
fi
if [ "$1" == "staker-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "158ZGK5EHmoyrHEd8aA2HCaqbNQ4r45ZsnwL4Zh8mH8dueWHWs"  --datadir "data/shard2-1" --listen "0.0.0.0:9443" --externaladdress "0.0.0.0:9443" --norpcauth --rpclisten "0.0.0.0:9343"
fi
if [ "$1" == "staker-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "1G5Q9uGSxekPSgC1w1ZFaDJ8RxeYrekk2FtFLF33QCKNbg2V88"  --datadir "data/shard2-2" --listen "0.0.0.0:9444" --externaladdress "0.0.0.0:9444" --norpcauth --rpclisten "0.0.0.0:9344"
fi
if [ "$1" == "staker-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "12MZ4QiFoETNbdLKgRQWPMQMqsceWPKo71Jma9NzwvLTabpcDhn"  --datadir "data/shard2-3" --listen "0.0.0.0:9445" --externaladdress "0.0.0.0:9445" --norpcauth --rpclisten "0.0.0.0:9345"
fi
if [ "$1" == "staker-4" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --miningkeys "1cQCTV1m33LxBKpNW2SisbuJfp5VcBSEau7PE5aD16gGLAN7eq"  --datadir "data/shard3-0" --listen "0.0.0.0:9446" --externaladdress "0.0.0.0:9446" --norpcauth --rpclisten "0.0.0.0:9346"
fi

# Shard 1
if [ "$1" == "shard1-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXHTza2DB23sR5UWpKSj9BV254U5Se8FpTrvUZyXDUP1K3vuZdeukuBQKQqTGH3Nr2UGt1i84cqbYpTZx3JKYgSUXedq9FXqnMa2iC"  --datadir "data/shard1-0" --listen "0.0.0.0:9438" --externaladdress "0.0.0.0:9438" --norpcauth --rpclisten "0.0.0.0:9338" --rpcwslisten "127.0.0.1:19338"
fi
if [ "$1" == "shard1-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXZUMuDGdtWDm5cbiZbJsmbbYGWi5Y3SZNDRSk4L9TPjCEppLaef2cZFHBxcyspikPAUEiKMQa9JEQm4XNL4Q3PEiZGDWZF11nsu2m"  --datadir "data/shard1-1" --listen "0.0.0.0:9439" --externaladdress "0.0.0.0:9439" --norpcauth --rpclisten "0.0.0.0:9339" --rpcwslisten "127.0.0.1:19339"
fi
if [ "$1" == "shard1-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXjFnTYhk6iwP6Jn71TZnEYUgjCBSiXnexDWruRh7qZGaR9GWgPHwYeXCcn4e88NB3D8kEiAA4Lf6PeLj5JAzXN5W7eaxwNCbFJLFR"  --datadir "data/shard1-2" --listen "0.0.0.0:9440" --externaladdress "0.0.0.0:9440" --norpcauth --rpclisten "0.0.0.0:9340"
fi
if [ "$1" == "shard1-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnY3EN7iSeUymyhH4qeUXYSXwHXzRwDcM4G8rNPAvCsPZfuXaPfLhKLkp9PQNoeY8SbhoKh3kn7aTsaUCLoch137DEWC3MDfwe7T6hk"  --datadir "data/shard1-3" --listen "0.0.0.0:9441" --externaladdress "0.0.0.0:9441" --norpcauth --rpclisten "0.0.0.0:9341"
fi
# Shard 2
if [ "$1" == "shard2-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXGXaT4LzPbU8UZk1QFAwBi72tcA6zJWG1zKP9MEXjYxmJD1bAi2nLQiBMNYxLkxZvg8KAdYU7CSoTq3PMKdAA9FJ35HRcP4nSKuSE"  --datadir "data/shard2-0" --listen "0.0.0.0:9442" --externaladdress "0.0.0.0:9442" --norpcauth --rpclisten "0.0.0.0:9342"
fi
if [ "$1" == "shard2-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXYyWe75qLrQPg8qCyug7VtL2fB5Z8pW5RJNExaWPXTMMzJqz2gbXGk2fVw22cYM6bEEsJhezQdruBktCYXbMn2PZXxgKwfVHwBR1n"  --datadir "data/shard2-1" --listen "0.0.0.0:9443" --externaladdress "0.0.0.0:9443" --norpcauth --rpclisten "0.0.0.0:9343"
fi
if [ "$1" == "shard2-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXeQ4jd9LzvCxrfHAL3xhF48GsYS6wwd95d898fq3nzUtKi3dyM9DPQPKdn72bauvZyFSK37FPuXX8EmHYVrJt7m1fEZtsaT4Mhdzq"  --datadir "data/shard2-2" --listen "0.0.0.0:9444" --externaladdress "0.0.0.0:9444" --norpcauth --rpclisten "0.0.0.0:9344"
fi
if [ "$1" == "shard2-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXtFM8tvKE4dLJZ5KTPRhWA4xDzvp1uw4pDeHBcr2UpWagapPgagPX6QadonGUt9xXP3kj1TZLxJNby9BFcjednPWe4bd7Rp2RnZkX"  --datadir "data/shard2-3" --listen "0.0.0.0:9445" --externaladdress "0.0.0.0:9445" --norpcauth --rpclisten "0.0.0.0:9345"
fi
# Shard 3
if [ "$1" == "shard3-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXFAbV5zMbGd3JfoaXY4umTR4Lx5htd1ifxY2BHv7oAmv3QyiEiikgP81C66KNT39RaK1FZB5YYiVUEyXzwmZrQnCXjXZf2wv5FioD"  --datadir "data/shard3-0" --listen "0.0.0.0:9446" --externaladdress "0.0.0.0:9446" --norpcauth --rpclisten "0.0.0.0:9346"
fi
if [ "$1" == "shard3-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXRKj24D5csit6UWuYbKxkGPXVnb9PfEZ9xc2b2pvVundrixwvJa5g98vjGqsHmQTF3PMReLVai3Wkpvb5jZqqjNMFpxRkGiUw6i8J"  --datadir "data/shard3-1" --listen "0.0.0.0:9447" --externaladdress "0.0.0.0:9447" --norpcauth --rpclisten "0.0.0.0:9347"
fi
if [ "$1" == "shard3-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXfbnYeBVw3WuZj5FTvky3eHcPiMF7T8Nt5wdFTv8WeSUJxkABouFppC63wQ1wqkfwhmgbHV2bUxgbq4oYoLwUVmBEGEojZKvvjx6M"  --datadir "data/shard3-2" --listen "0.0.0.0:9448" --externaladdress "0.0.0.0:9448" --norpcauth --rpclisten "0.0.0.0:9348"
fi
if [ "$1" == "shard3-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnY1RGNQnZrKyfLTjvJHDQomfnQFYry31BoxPUU6pxYBjAe1iqBMkYgr96pVzwiXSYNXGCwfjKhj3UEW7uPcbw4p7FTmrDAgVY4zzdZ"  --datadir "data/shard3-3" --listen "0.0.0.0:9449" --externaladdress "0.0.0.0:9449" --norpcauth --rpclisten "0.0.0.0:9349"
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnX3Cz3ud5HG7EnM8U3apQqbtpmbAjbe5Uox3Lj7aJg85AAko91JVwXjC7wNHENWtMmFqPvQEJrYS8WhYYekDJmH1c5GBkL4YCHKV8o"  --datadir "data/beacon-0" --listen "0.0.0.0:9450" --externaladdress "0.0.0.0:9450" --norpcauth --rpclisten "0.0.0.0:9350"
fi
if [ "$1" == "beacon-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXaRvy95YLYEt78ovWCY2Azi7pCrU4v7BCHm6AjfpUNYUDMbksf6WATFjY4tHUr4g6D5bmiKgMgmjB9ih1eNHifwqdRzC6Eqv23FHD"  --datadir "data/beacon-1" --listen "0.0.0.0:9451" --externaladdress "0.0.0.0:9451" --norpcauth --rpclisten "0.0.0.0:9351"
fi
if [ "$1" == "beacon-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXbxgX9xpiJ8f4z7NhAbPY77XY9BxCEHegzhWeR2Vm19YURuxsTYqZDFkK9Nk16ERmtbXW4oGU2ww6P1WiDv2rvBwq9HgsUabhH7EB"  --datadir "data/beacon-2" --listen "0.0.0.0:9452" --externaladdress "0.0.0.0:9452" --norpcauth --rpclisten "0.0.0.0:9352"
fi
if [ "$1" == "beacon-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnY3iZhsADoE3EJMddgHEJCSCxrhqpixuqr7jwzLqsebMkK6sEuSBGWDav35tWfomGW5urs4rEoR9VrTNrmwmwFZkRvSQrSCYudiXLg"  --datadir "data/beacon-3" --listen "0.0.0.0:9453" --externaladdress "0.0.0.0:9453" --norpcauth --rpclisten "0.0.0.0:9353"
fi
# Beacon
if [ "$1" == "beacon-new-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8sSijugr8azxAiHMWS9rA22grKDv5o7AEXQ9datpT1V7N5FLHiJMjvfVnXcitL3fpj35Xt5DNnBq8iFq618X31nCgn2RjrYx5tZZWCtj"  --datadir "data/beacon-0" --listen "0.0.0.0:9460" --externaladdress "0.0.0.0:9460" --norpcauth --rpclisten "0.0.0.0:9360"
fi
if [ "$1" == "beacon-new-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8sSjSEck5J5RKWGWurVfigruYDxEzjjVPqHaTRJ57YFNo7gXBH8onUQxtdpoyFnBZrLhfGWQ4k4MNadwa6F7qYwcuFLW9R1VxTfN7q4d"  --datadir "data/beacon-1" --listen "0.0.0.0:9461" --externaladdress "0.0.0.0:9461" --norpcauth --rpclisten "0.0.0.0:9361"
fi
if [ "$1" == "beacon-new-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8sSjkAqVJi4KkbCS75GYrsag7QZYP7FTPRfZ63D1AJgzfmdHnE9sbpdJV4Kx5tN9MgbqbRYDgzER2xpgsxrHvWxNgTHHrghYwLJLfe2R"  --datadir "data/beacon-2" --listen "0.0.0.0:9462" --externaladdress "0.0.0.0:9462" --norpcauth --rpclisten "0.0.0.0:9362"
fi
if [ "$1" == "beacon-new-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8sSj637mhpaJUboUEjkXsEUQm8q82T6kND3mWtNwig71qX2aFeZegWYsLVtyxBWdiZMBoNkdJ1MZYAcWetUP8DjYFnUac4vW7kzHfYsc"  --datadir "data/beacon-3" --listen "0.0.0.0:9463" --externaladdress "0.0.0.0:9463" --norpcauth --rpclisten "0.0.0.0:9363"
fi
# FullNode testnet
if [ "$1" == "fullnode" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=5 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" GETH_NAME="http://127.0.0.1:8545" GETH_PORT="" GETH_PROTOCOL="" --relayshards "all" --datadir "data/fullnode" --listen "0.0.0.0:9433" --externaladdress "0.0.0.0:9433" --norpcauth --rpclisten "0.0.0.0:8334" --rpcwslisten 0.0.0.0:18334 --txpoolmaxtx 100000 --allowstateprune
fi
if [ "$1" == "fullnode-devnet" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=100 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "139.162.55.124:9330" GETH_NAME="https://kovan.infura.io/v3/1138a1e99b154b10bae5c382ad894361" GETH_PORT="" GETH_PROTOCOL="" --relayshards "all" --datadir "/data/devnet" --listen "0.0.0.0:9433" --externaladdress "0.0.0.0:9433" --norpcauth --rpclisten "0.0.0.0:8334" --rpcwslisten 0.0.0.0:18334 --txpoolmaxtx 100000
fi
if [ "$1" == "fullnode2-devnet" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=100 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "139.162.55.124:9330" GETH_NAME="https://kovan.infura.io/v3/1138a1e99b154b10bae5c382ad894361" GETH_PORT="" GETH_PROTOCOL="" --relayshards "all" --datadir "/data/devnet2" --listen "0.0.0.0:39433" --externaladdress "0.0.0.0:39433" --norpcauth --rpclisten "0.0.0.0:38334" --rpcwslisten 0.0.0.0:48334 --txpoolmaxtx 100000
fi
if [ "$1" == "fullnode-testnet-b" ]; then
GO111MODULE--usecoindata --coindatapre="__coins__" =on GETH_NAME=kovan.infura.io/v3/1138a1e99b154b10bae5c382ad894361 GETH_PROTOCOL=https GETH_PORT="" INCOGNITO_NETWORK_KEY=local ./incognito --testnet true --nodemode "relay" --relayshards "[0]" --externaladdress "127.0.0.1:9433" --enablewallet --wallet "wallet" --walletpassphrase "12345678" --walletautoinit --norpcauth --datadir "../testnet/fullnode" --discoverpeersaddress "testnet-bootnode.incognito.org:9330" --norpcauth --rpclisten "0.0.0.0:8334" --rpcwslisten "127.0.0.1:18338" > ../testnet/log.txt 2> ../testnet/error_log.txt &
fi
if [ "$1" == "fullnode-mainnet" ]; then
INCOGNITO_NETWORK_KEY=mainnet ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --testnet true --nodemode "relay" --relayshards "[0,1,2,3,4,5,6,7]" --externaladdress "127.0.0.1:9433" --enablewallet --wallet "wallet" --walletpassphrase "12345678" --walletautoinit --norpcauth --datadir "../inc-data/mainnet/fullnode" --discoverpeersaddress "mainnet-bootnode.incognito.org:9330" > ../inc-data/mainnet/log.txt 2> ../inc-data/mainnet/error_log.txt &
fi
######
if [ "$1" == "shard-candidate0-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f"  --datadir "data/shard-stake" --listen "127.0.0.1:9455" --externaladdress "127.0.0.1:9455" --norpcauth --rpclisten "127.0.0.1:9355"
fi
if [ "$1" == "shard-candidate0-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rrEW3NPNgU8xzbeqE7cr4WTT8JvyaQqSZyczA5hBJVvpQMTBVqNfcCdzhvquWCHH11jHihZtgyJqbdWPhWYbmmsw5aV29WSXBEsgbVX"  --datadir "data/shard-stake-2" --listen "127.0.0.1:9456" --externaladdress "127.0.0.1:9456" --norpcauth --rpclisten "127.0.0.1:9356"
fi
if [ "$1" == "shard-candidate0-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnY15HgjpbJn1es84ysseB6q9UQ5SwB6Eb82yejEQ3yzhd1dm5ShEiezdfMoEzBgvkuKcFdP5TY3SuWNHXKa1Krprsfxnk5wy7wZ6Dg"  --datadir "data/shard-stake-6" --listen "0.0.0.0:9460" --externaladdress "0.0.0.0:9460" --norpcauth --rpclisten "0.0.0.0:9360"
fi
if [ "$1" == "shard-candidate1-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8roHikeAFyuBpdCU76kXurEqrC9VYWyRyfFb6PwX6nip9KGYbwpXL78H92mUoWK2GWkA2WysgXbHqwSxnC6XCkmtxBVb3zJeCXgfcYyL"  --datadir "data/shard-stake-3" --listen "0.0.0.0:9457" --externaladdress "0.0.0.0:9457" --norpcauth --rpclisten "0.0.0.0:9357"
fi
if [ "$1" == "shard-candidate1-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rr4sE2L8WzsVNEN9WsiGcMTDCmEH9TC1ZK8517cxURRFNoWoStYQTgqXpiAMU4gzmkmnWahHdGvQqFaY1JTVsn3nHfD5Ppgz8hQDiVC"  --datadir "data/shard-stake-4" --listen "0.0.0.0:9458" --externaladdress "0.0.0.0:9458" --norpcauth --rpclisten "0.0.0.0:9358"
fi
if [ "$1" == "shard-candidate1-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rtt9Kd5LUcfXNmd7aMnQehCnKabArVB3BUk2RHVjeh88x5MJnJY4okB8JdFm4JNm4A2WjSe58qWNVkJPEFjpLHNYfKHpWfRdqyfDD9f"  --datadir "data/shard-stake-5" --listen "0.0.0.0:9459" --externaladdress "0.0.0.0:9459" --norpcauth --rpclisten "0.0.0.0:9359"
fi
if [ "$1" == "shard-candidate1-4" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnY15HgjpbJn1es84ysseB6q9UQ5SwB6Eb82yejEQ3yzhd1dm5ShEiezdfMoEzBgvkuKcFdP5TY3SuWNHXKa1Krprsfxnk5wy7wZ6Dg"  --datadir "data/shard-stake-7" --listen "0.0.0.0:9461" --externaladdress "0.0.0.0:9461" --norpcauth --rpclisten "0.0.0.0:9361"
fi
if [ "$1" == "shard-candidate1-5" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rna7M8BYBfNjNHmw3Tie6Yir9mQgp5rSRgUngTqn6A6iSRvAPex4sXsmGxVzXcpUUDfnRfRys3QrPnTHauiipdUNtj7Ef6t3mHUwiC3"  --datadir "data/shard-stake-8" --listen "0.0.0.0:9462" --externaladdress "0.0.0.0:9462" --norpcauth --rpclisten "0.0.0.0:9362"
fi

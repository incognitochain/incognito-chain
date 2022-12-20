#!/usr/bin/env bash
#GrafanaURL=http://128.199.96.206:8086/write?db=mydb
###### MULTI_MEMBERS
# Shard 0
if [ "$1" == "shard0-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnkwspf7rutf9tYXhb77N3LBTxfmU6oyE7eMQ4wyy8iH7gEZzeePsXX8fx9ZjY8ozeWySuYnWZzwACoDog4GZzViUe5kH46hGxBJzrf"  --datadir "data/shard0-0" --listen "0.0.0.0:9434" --externaladdress "0.0.0.0:9434" --norpcauth --rpclisten "0.0.0.0:9333" --rpcwslisten "0.0.0.0:19333" --relayshards ""
fi
if [ "$1" == "shard0-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnr1HRcYozjwt1Y4t1o7kAvKM876kQeN7s9jCUodhFbsrPZ1UU6vDcSSFAQMAMLz6i576gpVmyy3K4py9ao3aF7hqozCqSw6amKnUjw"  --datadir "data/shard0-1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9335" --rpcwslisten "0.0.0.0:19335" --relayshards ""
fi
if [ "$1" == "shard0-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rpiamZoiyGXTkJ8mMsLxJwgnyUUTTrfpdnpm3KckC3VjaZRzpU24jQTfmfztZUKYqdJUKUt7BrgMKMqKLjqczwffmCx9QM5D1PJShPX"  --datadir "data/shard0-2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336" --relayshards ""
fi
if [ "$1" == "shard0-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rqDSA3EThnGp2doR4zuRdgGRiGLdyYi6Deo1QTwW41nttvmo5dVEUjbE3PTW2wtWc9cthKA4ZmJjwQyRShqqXL5NQGRUhoVYjRCR5Qo"  --datadir "data/shard0-3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337" --relayshards ""
fi
if [ "$1" == "shard0-new-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rospAEaouNQgnK8vAAJGzH6ysLAGeZGqmZ5RJTT7CrF1zK8zwqVqx4DEdoD6MDTNiSK9W1vbZXtVe7vqvfEuf6LpuBbUiHvvkfF9L3X"  --datadir "data/shard0-0" --listen "0.0.0.0:9434" --externaladdress "0.0.0.0:9434" --norpcauth --rpclisten "0.0.0.0:9334" --rpcwslisten "0.0.0.0:19334"
fi
if [ "$1" == "shard0-new-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rs4QMxRdQLLyG7b5ifeBwH39TuZU92UuTXHynDKVgtE366Jd6qs99gj6gKtz46ad5NKXbaJ2UyXfxbjCouhtN6Es8ve5yyQVyEtgXod"  --datadir "data/shard0-1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9335" --rpcwslisten "0.0.0.0:19335"
fi
if [ "$1" == "shard0-new-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rtB7mXUvqSqgGJZTa84JwUJyruLXNYwrypD98t8UkAGxhBhhs6P696Z7iZ1WxdhWEFKeDbEkR5PdXNf4V8CRwmszhjUAU6AqyXQ6ME7"  --datadir "data/shard0-2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336"
fi
if [ "$1" == "shard0-new-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rwXq7wh46QCmZUuqUTaQwbrXi7mvZnPayhjaUqa2MCnGbXRxHQhxyShfRLpvQzUzqqmDGMqfmLR6R2WayyyBhXU1b6Hhz8CnZVPTBG7"  --datadir "data/shard0-3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337"
fi

if [ "$1" == "staker-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --miningkeys "12MZ4QiFoETNbdLKgRQWPMQMqsceWPKo71Jma9NzwvLTabpcDhn"  --datadir "data/shard2-0" --listen "0.0.0.0:9442" --externaladdress "0.0.0.0:9442" --norpcauth --rpclisten "0.0.0.0:9342"
fi
if [ "$1" == "staker-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --miningkeys "158ZGK5EHmoyrHEd8aA2HCaqbNQ4r45ZsnwL4Zh8mH8dueWHWs"  --datadir "data/shard2-1" --listen "0.0.0.0:9443" --externaladdress "0.0.0.0:9443" --norpcauth --rpclisten "0.0.0.0:9343"
fi
if [ "$1" == "staker-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --miningkeys "1G5Q9uGSxekPSgC1w1ZFaDJ8RxeYrekk2FtFLF33QCKNbg2V88"  --datadir "data/shard2-2" --listen "0.0.0.0:9444" --externaladdress "0.0.0.0:9444" --norpcauth --rpclisten "0.0.0.0:9344"
fi
if [ "$1" == "staker-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --miningkeys "12MZ4QiFoETNbdLKgRQWPMQMqsceWPKo71Jma9NzwvLTabpcDhn"  --datadir "data/shard2-3" --listen "0.0.0.0:9445" --externaladdress "0.0.0.0:9445" --norpcauth --rpclisten "0.0.0.0:9345"
fi
if [ "$1" == "staker-4" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --miningkeys "1cQCTV1m33LxBKpNW2SisbuJfp5VcBSEau7PE5aD16gGLAN7eq"  --datadir "data/shard3-0" --listen "0.0.0.0:9446" --externaladdress "0.0.0.0:9446" --norpcauth --rpclisten "0.0.0.0:9346"
fi

# Shard 1
if [ "$1" == "shard1-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXHTza2DB23sR5UWpKSj9BV254U5Se8FpTrvUZyXDUP1K3vuZdeukuBQKQqTGH3Nr2UGt1i84cqbYpTZx3JKYgSUXedq9FXqnMa2iC"  --datadir "data/shard1-0" --listen "0.0.0.0:9438" --externaladdress "0.0.0.0:9438" --norpcauth --rpclisten "0.0.0.0:9338" --rpcwslisten "127.0.0.1:19338" --relayshards ""
fi
if [ "$1" == "shard1-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXZUMuDGdtWDm5cbiZbJsmbbYGWi5Y3SZNDRSk4L9TPjCEppLaef2cZFHBxcyspikPAUEiKMQa9JEQm4XNL4Q3PEiZGDWZF11nsu2m"  --datadir "data/shard1-1" --listen "0.0.0.0:9439" --externaladdress "0.0.0.0:9439" --norpcauth --rpclisten "0.0.0.0:9339" --rpcwslisten "127.0.0.1:19339" --relayshards ""
fi
if [ "$1" == "shard1-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXjFnTYhk6iwP6Jn71TZnEYUgjCBSiXnexDWruRh7qZGaR9GWgPHwYeXCcn4e88NB3D8kEiAA4Lf6PeLj5JAzXN5W7eaxwNCbFJLFR"  --datadir "data/shard1-2" --listen "0.0.0.0:9440" --externaladdress "0.0.0.0:9440" --norpcauth --rpclisten "0.0.0.0:9340" --relayshards ""
fi
if [ "$1" == "shard1-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnY3EN7iSeUymyhH4qeUXYSXwHXzRwDcM4G8rNPAvCsPZfuXaPfLhKLkp9PQNoeY8SbhoKh3kn7aTsaUCLoch137DEWC3MDfwe7T6hk"  --datadir "data/shard1-3" --listen "0.0.0.0:9441" --externaladdress "0.0.0.0:9441" --norpcauth --rpclisten "0.0.0.0:9341" --relayshards ""
fi
# Shard 2
if [ "$1" == "shard2-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rnXGXaT4LzPbU8UZk1QFAwBi72tcA6zJWG1zKP9MEXjYxmJD1bAi2nLQiBMNYxLkxZvg8KAdYU7CSoTq3PMKdAA9FJ35HRcP4nSKuSE"  --datadir "data/shard2-0" --listen "0.0.0.0:9442" --externaladdress "0.0.0.0:9442" --norpcauth --rpclisten "0.0.0.0:9342"
fi
if [ "$1" == "shard2-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rnXYyWe75qLrQPg8qCyug7VtL2fB5Z8pW5RJNExaWPXTMMzJqz2gbXGk2fVw22cYM6bEEsJhezQdruBktCYXbMn2PZXxgKwfVHwBR1n"  --datadir "data/shard2-1" --listen "0.0.0.0:9443" --externaladdress "0.0.0.0:9443" --norpcauth --rpclisten "0.0.0.0:9343"
fi
if [ "$1" == "shard2-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rnXeQ4jd9LzvCxrfHAL3xhF48GsYS6wwd95d898fq3nzUtKi3dyM9DPQPKdn72bauvZyFSK37FPuXX8EmHYVrJt7m1fEZtsaT4Mhdzq"  --datadir "data/shard2-2" --listen "0.0.0.0:9444" --externaladdress "0.0.0.0:9444" --norpcauth --rpclisten "0.0.0.0:9344"
fi
if [ "$1" == "shard2-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rnXtFM8tvKE4dLJZ5KTPRhWA4xDzvp1uw4pDeHBcr2UpWagapPgagPX6QadonGUt9xXP3kj1TZLxJNby9BFcjednPWe4bd7Rp2RnZkX"  --datadir "data/shard2-3" --listen "0.0.0.0:9445" --externaladdress "0.0.0.0:9445" --norpcauth --rpclisten "0.0.0.0:9345"
fi
# Shard 3
if [ "$1" == "shard3-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rnXFAbV5zMbGd3JfoaXY4umTR4Lx5htd1ifxY2BHv7oAmv3QyiEiikgP81C66KNT39RaK1FZB5YYiVUEyXzwmZrQnCXjXZf2wv5FioD"  --datadir "data/shard3-0" --listen "0.0.0.0:9446" --externaladdress "0.0.0.0:9446" --norpcauth --rpclisten "0.0.0.0:9346"
fi
if [ "$1" == "shard3-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rnXRKj24D5csit6UWuYbKxkGPXVnb9PfEZ9xc2b2pvVundrixwvJa5g98vjGqsHmQTF3PMReLVai3Wkpvb5jZqqjNMFpxRkGiUw6i8J"  --datadir "data/shard3-1" --listen "0.0.0.0:9447" --externaladdress "0.0.0.0:9447" --norpcauth --rpclisten "0.0.0.0:9347"
fi
if [ "$1" == "shard3-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rnXfbnYeBVw3WuZj5FTvky3eHcPiMF7T8Nt5wdFTv8WeSUJxkABouFppC63wQ1wqkfwhmgbHV2bUxgbq4oYoLwUVmBEGEojZKvvjx6M"  --datadir "data/shard3-2" --listen "0.0.0.0:9448" --externaladdress "0.0.0.0:9448" --norpcauth --rpclisten "0.0.0.0:9348"
fi
if [ "$1" == "shard3-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rnY1RGNQnZrKyfLTjvJHDQomfnQFYry31BoxPUU6pxYBjAe1iqBMkYgr96pVzwiXSYNXGCwfjKhj3UEW7uPcbw4p7FTmrDAgVY4zzdZ"  --datadir "data/shard3-3" --listen "0.0.0.0:9449" --externaladdress "0.0.0.0:9449" --norpcauth --rpclisten "0.0.0.0:9349"
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rnXCerQX2RRd8KhPfsFCj2rrBYUx42FZJKgRFcdBfg36Mid3ygKyMn5LSc5LBHsxqapRaN6xMav7bGhA6VtGUzNNYuA9Y78CB5oGkti"  --datadir "data/beacon-0" --listen "0.0.0.0:9450" --externaladdress "0.0.0.0:9450" --norpcauth --rpclisten "0.0.0.0:9350"
fi
if [ "$1" == "beacon-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rnXYgxipKvTJJfHg7tQhcdmA2R1jPpCPmXg37Xi1VfgrFzWFuNy4U6828q1yfbD7VEdutD63HfVYAqL6U32joXVjqdkfUP52LnNGXda"  --datadir "data/beacon-1" --listen "0.0.0.0:9451" --externaladdress "0.0.0.0:9451" --norpcauth --rpclisten "0.0.0.0:9351"
fi
if [ "$1" == "beacon-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rnXe3Jxg5d1Rejg2fB1NwnqNsr94RCT3PX14h5NNDjrdgLeEWFkqcMNamKCHask1Gx46g5WYZDKHKx7kzLVD7h1cgvU6NxNijkyGmA9"  --datadir "data/beacon-2" --listen "0.0.0.0:9452" --externaladdress "0.0.0.0:9452" --norpcauth --rpclisten "0.0.0.0:9352"
fi
if [ "$1" == "beacon-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8rnY2gqonwhnhGD6rKeEXkbJDB7DHUtZQKC8SfLci6ABb5eCEj4o7ezWBZWaGbu7CJ1R1mrADGqmRjugg42GeA6jhaXbNDeP2HUr8udw"  --datadir "data/beacon-3" --listen "0.0.0.0:9453" --externaladdress "0.0.0.0:9453" --norpcauth --rpclisten "0.0.0.0:9353"
fi
# Beacon
if [ "$1" == "beacon-new-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8sSijugr8azxAiHMWS9rA22grKDv5o7AEXQ9datpT1V7N5FLHiJMjvfVnXcitL3fpj35Xt5DNnBq8iFq618X31nCgn2RjrYx5tZZWCtj"  --datadir "data/beacon-0" --listen "0.0.0.0:9460" --externaladdress "0.0.0.0:9460" --norpcauth --rpclisten "0.0.0.0:9360"
fi
if [ "$1" == "beacon-new-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8sSjSEck5J5RKWGWurVfigruYDxEzjjVPqHaTRJ57YFNo7gXBH8onUQxtdpoyFnBZrLhfGWQ4k4MNadwa6F7qYwcuFLW9R1VxTfN7q4d"  --datadir "data/beacon-1" --listen "0.0.0.0:9461" --externaladdress "0.0.0.0:9461" --norpcauth --rpclisten "0.0.0.0:9361"
fi
if [ "$1" == "beacon-new-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8sSjkAqVJi4KkbCS75GYrsag7QZYP7FTPRfZ63D1AJgzfmdHnE9sbpdJV4Kx5tN9MgbqbRYDgzER2xpgsxrHvWxNgTHHrghYwLJLfe2R"  --datadir "data/beacon-2" --listen "0.0.0.0:9462" --externaladdress "0.0.0.0:9462" --norpcauth --rpclisten "0.0.0.0:9362"
fi
if [ "$1" == "beacon-new-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --privatekey "112t8sSj637mhpaJUboUEjkXsEUQm8q82T6kND3mWtNwig71qX2aFeZegWYsLVtyxBWdiZMBoNkdJ1MZYAcWetUP8DjYFnUac4vW7kzHfYsc"  --datadir "data/beacon-3" --listen "0.0.0.0:9463" --externaladdress "0.0.0.0:9463" --norpcauth --rpclisten "0.0.0.0:9363"
fi
# FullNode testnet
if [ "$1" == "fullnode" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=5 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" GETH_NAME="http://127.0.0.1:8545" GETH_PORT="" GETH_PROTOCOL="" --relayshards "all" --datadir "data/fullnode" --listen "0.0.0.0:9433" --externaladdress "0.0.0.0:9433" --norpcauth --rpclisten "0.0.0.0:8334" --rpcwslisten 0.0.0.0:18334 --txpoolmaxtx 100000 --allowstateprunebyrpc
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
if [ "$1" == "shard-candidate0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnbJ16eRJqBrXMmafYCVyTPaW7vsNZPqrrA3L8q2wWxjueroosTZkfWeUzBm9ucsXXPRwjCR5rTQhjEksohxa2fmHj26AeyZUkjYnY9"  --datadir "data/shard-candidate0" --listen "127.0.0.1:9455" --externaladdress "127.0.0.1:9455" --norpcauth --rpclisten "127.0.0.1:9355"
fi
if [ "$1" == "shard-candidate1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnbLY5K98Qtu4BLJ1W7ijcRqhaPRg5RcESgLKFYir6GxKgjWUVXbwe7c65ER7fpX1Wpfg3Aac8GCgwkBBv3SMpz4SXNzgkjHmXMaC9x"  --datadir "data/shard-candidate1" --listen "127.0.0.1:9456" --externaladdress "127.0.0.1:9456" --norpcauth --rpclisten "127.0.0.1:9356"
fi
if [ "$1" == "shard-candidate2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnbiSzrMcvw95QB9a49y5aG8ttC9xaFod1B1RrDq5EwkxSN2iDfu9r4HpkCrPk6gyJqnSPYfFV27cfdvcyFWkapLApNjhSgqNCVUhSp"  --datadir "data/shard-candidate2" --listen "0.0.0.0:9457" --externaladdress "0.0.0.0:9457" --norpcauth --rpclisten "0.0.0.0:9357"
fi
if [ "$1" == "shard-candidate3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnei2en65uc6qjQgJYRL2QvrSePAXsaruFmuD6noCz3kKEawuPPHt6NcY2DqjyEu72zvHFt4xd4R8FJfC6o6y5epfwFkGnDLuGWr9z1"  --datadir "data/shard-candidate3" --listen "0.0.0.0:9458" --externaladdress "0.0.0.0:9458" --norpcauth --rpclisten "0.0.0.0:9358"
fi
if [ "$1" == "shard-candidate4" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rngLD7yLYmBdMTsExWtV28ViFnPLA5NJYffNAoKAyB3NpvWA7ESwdoxvuYYN6C7wXwebQRXtvsbVhZAUkRgp66J8UMuQusoPskp2BBU"  --datadir "data/shard-candidate4" --listen "0.0.0.0:9459" --externaladdress "0.0.0.0:9459" --norpcauth --rpclisten "0.0.0.0:9359"
fi
if [ "$1" == "shard-candidate5" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnhz7vBTPVLo8o4aii6LAiX9uxopAqf6ayiA8y2DLsVBFKfKHcdjcsdLbujVvfic4SutcFo7WZsxgG67uRDrzPnjTc9RE13dX3LP7pE"  --datadir "data/shard-candidate5" --listen "0.0.0.0:9460" --externaladdress "0.0.0.0:9460" --norpcauth --rpclisten "0.0.0.0:9360"
fi
if [ "$1" == "shard-candidate6" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rniKedaWfaDkKzFNJFKWBKfmMbNpsXJHGkr1GRYEmwHty9vCKV4F92fNhpdRyC1D5DpfDbXDzNqonANKnC6avNECrcNURfACYHrxr9i"  --datadir "data/shard-candidate6" --listen "0.0.0.0:9461" --externaladdress "0.0.0.0:9461" --norpcauth --rpclisten "0.0.0.0:9361"
fi
if [ "$1" == "shard-candidate7" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnmpoRJBXHuJ5RLX286toHvcpua1g5wyN8N5sJ9hk3fTJkCyo4ZmDatzWu7jj8RpgZfUEkMHffjQp5jqSQ9CWHXKY1j21ZwCkN53sru"  --datadir "data/shard-candidate7" --listen "0.0.0.0:9462" --externaladdress "0.0.0.0:9462" --norpcauth --rpclisten "0.0.0.0:9362"
fi
if [ "$1" == "shard-candidate8" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnqR3gMxTpGtsVVXsqpwt93W5Vz94y1TcF5osZmhVFm3ezcY162Ldbour1WjjnMLL6SuP44WmMUatnCLPCjdJVLSXHm7GcBnn37F68k"  --datadir "data/shard-candidate8" --listen "0.0.0.0:9463" --externaladdress "0.0.0.0:9463" --norpcauth --rpclisten "0.0.0.0:9363"
fi
if [ "$1" == "shard-candidate9" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnw3srMRDRDkyqzQQemRQkU8gmGCNiGM8oCFtoaTzib2ZnRHGbP1vApKqdB6KvXtDKuw98C48JiPk9PDGkaeCcXHGBhACh57NmHCs8v"  --datadir "data/shard-candidate9" --listen "0.0.0.0:9464" --externaladdress "0.0.0.0:9464" --norpcauth --rpclisten "0.0.0.0:9364"
fi
if [ "$1" == "shard-candidate10" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnxUYB8r6bm4gdxNLE22Yy2rhgt7HUQSwWzHsUkzsuD9Q1iL8r6xnLefpvNehaUU21aQgSE2iGp5axtyTvbWeJ1VfQERAQbBZY7Jd33"  --datadir "data/shard-candidate10" --listen "0.0.0.0:9465" --externaladdress "0.0.0.0:9465" --norpcauth --rpclisten "0.0.0.0:9365"
fi
if [ "$1" == "shard-candidate11" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8ro3dyq9rNiYxVXJkyTwqUFpigpwZsRpewnJa5PRCH7oRaBAPRSRzRnUjoQdWikG5Sa9FPwfomcLdiHZyQaPVxd6iXuJrj2dcgoPgYEN"  --datadir "data/shard-candidate11" --listen "0.0.0.0:9466" --externaladdress "0.0.0.0:9466" --norpcauth --rpclisten "0.0.0.0:9366"
fi
if [ "$1" == "beacon-candidate0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8ro7869VzMU3pobVp1428GWKzc6EwTR2MTaWoq6L7riCXX9fiSwH7qi6Po53L8CLB8QLe8eCLE85t3VseWj5Ktz3B4Fov2Q4A5H6o8yZ"  --datadir "data/beacon-candidate0" --listen "0.0.0.0:9467" --externaladdress "0.0.0.0:9467" --norpcauth --rpclisten "0.0.0.0:9367" --relayshards "all"
fi
if [ "$1" == "beacon-candidate1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8roC9ZD1hwDqR4NKb3K8wsWPQhknpfXEA7cvCrtBF7PWV92jdgNhRNqyJ6EHUfkAqaacJfutoQ9NELhccC6NHZNpD3XXchvMbch8x9kQ"  --datadir "data/beacon-candidate1" --listen "0.0.0.0:9468" --externaladdress "0.0.0.0:9468" --norpcauth --rpclisten "0.0.0.0:9368" --relayshards "all"
fi
if [ "$1" == "beacon-candidate2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8roCwh9gtkyrTnU6kPpVmejtLUHUbzJwXycgg2C8WTM7Mu9uK7n927HRnybCLN6mG3AoCwgzU7eLhPyYYJiW3B39ULHVmmQ6w36WePXk"  --datadir "data/beacon-candidate1" --listen "0.0.0.0:9469" --externaladdress "0.0.0.0:9469" --norpcauth --rpclisten "0.0.0.0:9369" --relayshards "all"
fi
if [ "$1" == "beacon-candidate3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8roD3c7hcfk5r8nJCzhjJhhbhvCjDUxMFyYz7kRBK5jvNirwwhzz5FwPSYYeMUUPdnECJRxzL2xAR3yALSmPKG6iCh5fPcRegWyCWJgE"  --datadir "data/beacon-candidate1" --listen "0.0.0.0:9470" --externaladdress "0.0.0.0:9470" --norpcauth --rpclisten "0.0.0.0:9370" --relayshards "all"
fi
if [ "$1" == "beacon-candidate4" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8roDLfR4Eqwz4Pf1ZZKcByF3TT4X4C2MzNpJhHHP4ZpHTGAh7VNX2TmQjhYhftKYo8ftmrRrttbqL8wJMgCdMUp8CMuFXxtVKKP9ieK2"  --datadir "data/beacon-candidate1" --listen "0.0.0.0:9471" --externaladdress "0.0.0.0:9471" --norpcauth --rpclisten "0.0.0.0:9371" --relayshards "all"
fi
if [ "$1" == "beacon-candidate5" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8roDm5xchtS1o9ajrGaEacDzmzMbifVy4KVBc1m8PRoq8fQhxGNgxLTkHe2CRZ4b9djkt9vWGPG5zxq4jsPnoTYUnZdx1qFe2CvDBW41"  --datadir "data/beacon-candidate1" --listen "0.0.0.0:9472" --externaladdress "0.0.0.0:9472" --norpcauth --rpclisten "0.0.0.0:9372" --relayshards "all"
fi
if [ "$1" == "beacon-candidate6" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --usecoindata --coindatapre="__coins__" --numindexerworkers=0 --indexeraccesstoken="0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8roFHKqEkcG9w5nzZpqEMysoUEvooanNE7SLb2HHPSvR6cnjpTodqwWpzYBjtfuskunGLcvQPU9m9QinnnHpSeUtqakfCeFL45NRRUwH"  --datadir "data/beacon-candidate1" --listen "0.0.0.0:9473" --externaladdress "0.0.0.0:9473" --norpcauth --rpclisten "0.0.0.0:9373" --relayshards "all"
fi

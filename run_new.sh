#!/usr/bin/env bash

go run *.go --miningkeys "bls:112t8rtxdpGEx5H7MHi5EXGKZhKAxgbEkRv1d8TH1cbqKzc5j4mXEKC2NUCeCSQRA3zDeUQxKUW54sAGzpqFi86JQGaPAh4Ao4Gqo7q51yGw" --nodemode "auto" --datadir "data/shard0-2" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9336"
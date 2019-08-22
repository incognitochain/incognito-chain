#!/usr/bin/env bash

go run *.go --miningkeys "bls:12v9jokoYDWFQ7731mR6QjjkvUG1Q" --nodemode "auto" --datadir "data/shard0-2" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9336"
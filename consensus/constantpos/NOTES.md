For single-node mode:
- Uncomment 'return newBlock, nil' in 'bft.go' line 88 & 101
- Change 'SHARD_NUMBER' to 1 in 'common/constants.go'
- Uncomment 'return nil' in 'blockchain/multisigs.go'
- Uncomment line 664 & 665 in 'blockchain/beacon_process.go'
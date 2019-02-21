For single-node mode:
- Uncomment 'return newBlock, nil' in 'consensus/constantpos/bft.go'
- Uncomment 'return nil' in 'blockchain/multisigs.go'
- Edit 'TestNetBeaconCommitteeSize' & 'TestNetShardCommitteeSize' or 'MainNetBeaconCommitteeSize & 'MainnNetShardCommitteeSize' to 1 in 'blockchain/params.go'
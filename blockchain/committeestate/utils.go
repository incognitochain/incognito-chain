package committeestate

//VersionByBeaconHeight get version of committee engine by beaconHeight and config of blockchain
func VersionByBeaconHeight(beaconHeight, consensusV3Height, stakingV3Height uint64) int {
	if beaconHeight >= stakingV3Height {
		return DCS_VERSION
	}
	if beaconHeight >= consensusV3Height {
		return SLASHING_VERSION
	}
	return SELF_SWAP_SHARD_VERSION
}

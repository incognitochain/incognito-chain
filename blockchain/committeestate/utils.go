package committeestate

//VersionByBeaconHeight get version of committee engine by beaconHeight and config of blockchain
func VersionByBeaconHeight(beaconHeight, consensusV3Height, beaconStateV3Height uint64) int {
	if beaconHeight >= consensusV3Height {
		if beaconHeight >= beaconStateV3Height {
			return DCS_VERSION
		} else {
			return SLASHING_VERSION
		}
	} else {
		return SELF_SWAP_SHARD_VERSION
	}
}

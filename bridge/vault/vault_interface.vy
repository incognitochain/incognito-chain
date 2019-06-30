
# External Contracts
contract Vault:
    def deposit(incognito_address: string[128]): modifying
    def parseBurnInst(inst: bytes[150]) -> (uint256, bytes32, address, uint256): constant
    def testExtract(a: bytes[150]) -> (address, uint256(wei)): constant
    def withdraw(inst: bytes[150], beaconInstPath: bytes32[4], beaconInstPathIsLeft: bool[4], beaconInstPathLen: int128, beaconInstRoot: bytes32, beaconBlkData: bytes32, beaconBlkHash: bytes32, beaconSignerPubkeys: bytes[528], beaconSignerCount: int128, beaconSignerSig: bytes32, beaconSignerPaths: bytes32[64], beaconSignerPathIsLeft: bool[64], beaconSignerPathLen: int128, bridgeInstPath: bytes32[4], bridgeInstPathIsLeft: bool[4], bridgeInstPathLen: int128, bridgeInstRoot: bytes32, bridgeBlkData: bytes32, bridgeBlkHash: bytes32, bridgeSignerPubkeys: bytes[528], bridgeSignerCount: int128, bridgeSignerSig: bytes32, bridgeSignerPaths: bytes32[64], bridgeSignerPathIsLeft: bool[64], bridgeSignerPathLen: int128): modifying
    def withdrawed(arg0: bytes32) -> bool: constant
    def incognito() -> address(Incognito_proxy): constant



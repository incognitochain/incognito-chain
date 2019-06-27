
# External Contracts
contract Incognito_proxy:
    def parseSwapInst(inst: bytes[150]) -> (uint256, bytes32): constant
    def notifyPls(v: bytes32): modifying
    def inMerkleTree(leaf: bytes32, root: bytes32, path: bytes32[4], left: bool[4], length: int128) -> bool: constant
    def verifyInst(commRoot: bytes32, instHash: bytes32, instPath: bytes32[4], instPathIsLeft: bool[4], instPathLen: int128, instRoot: bytes32, blkHash: bytes32, signerPubkeys: bytes[528], signerCount: int128, signerSig: bytes32, signerPaths: bytes32[64], signerPathIsLeft: bool[64], signerPathLen: int128) -> bool: constant
    def instructionApproved(instHash: bytes32, beaconInstPath: bytes32[4], beaconInstPathIsLeft: bool[4], beaconInstPathLen: int128, beaconInstRoot: bytes32, beaconBlkData: bytes32, beaconBlkHash: bytes32, beaconSignerPubkeys: bytes[528], beaconSignerCount: int128, beaconSignerSig: bytes32, beaconSignerPaths: bytes32[64], beaconSignerPathIsLeft: bool[64], beaconSignerPathLen: int128, bridgeInstPath: bytes32[4], bridgeInstPathIsLeft: bool[4], bridgeInstPathLen: int128, bridgeInstRoot: bytes32, bridgeBlkData: bytes32, bridgeBlkHash: bytes32, bridgeSignerPubkeys: bytes[528], bridgeSignerCount: int128, bridgeSignerSig: bytes32, bridgeSignerPaths: bytes32[64], bridgeSignerPathIsLeft: bool[64], bridgeSignerPathLen: int128) -> bool: modifying
    def swapCommittee(inst: bytes[150], beaconInstPath: bytes32[4], beaconInstPathIsLeft: bool[4], beaconInstPathLen: int128, beaconInstRoot: bytes32, beaconBlkData: bytes32, beaconBlkHash: bytes32, beaconSignerPubkeys: bytes[528], beaconSignerCount: int128, beaconSignerSig: bytes32, beaconSignerPaths: bytes32[64], beaconSignerPathIsLeft: bool[64], beaconSignerPathLen: int128, bridgeInstPath: bytes32[4], bridgeInstPathIsLeft: bool[4], bridgeInstPathLen: int128, bridgeInstRoot: bytes32, bridgeBlkData: bytes32, bridgeBlkHash: bytes32, bridgeSignerPubkeys: bytes[528], bridgeSignerCount: int128, bridgeSignerSig: bytes32, bridgeSignerPaths: bytes32[64], bridgeSignerPathIsLeft: bool[64], bridgeSignerPathLen: int128) -> bool: modifying
    def beaconCommRoot() -> bytes32: constant
    def bridgeCommRoot() -> bytes32: constant



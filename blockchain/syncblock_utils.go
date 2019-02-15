package blockchain

import (
	"time"

	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/common"
)

func getBlkNeedToGetByHash(blksHash []common.Hash, loadedData interface{}, init bool, peerID libp2p.ID) ([]common.Hash, interface{}) {
	var blksNeedToGet []common.Hash
	var blksSyncByHash *map[string]peerSyncTimestamp
	if init {
		blksSyncByHash = loadedData.(*map[string]peerSyncTimestamp)
		for _, blkHash := range blksHash {
			if timeStamp, ok := (*blksSyncByHash)[blkHash.String()]; ok {
				if time.Since(time.Unix(timeStamp.Time, 0)) > defaultMaxBlockSyncTime {
					(*blksSyncByHash)[blkHash.String()] = peerSyncTimestamp{
						Time:   time.Now().Unix(),
						PeerID: peerID,
					}
					blksNeedToGet = append(blksNeedToGet, blkHash)
				}
			} else {
				(*blksSyncByHash)[blkHash.String()] = peerSyncTimestamp{
					Time:   time.Now().Unix(),
					PeerID: peerID,
				}
				blksNeedToGet = append(blksNeedToGet, blkHash)
			}
		}
	} else {
		blksSyncByHash = &map[string]peerSyncTimestamp{}
		for _, blkHash := range blksHash {
			(*blksSyncByHash)[blkHash.String()] = peerSyncTimestamp{
				Time:   time.Now().Unix(),
				PeerID: peerID,
			}
		}
		blksNeedToGet = append(blksNeedToGet, blksHash...)
	}
	return blksNeedToGet, blksSyncByHash
}

func getBlkNeedToGetByHeight(fromHeight uint64, toHeight uint64, loadedData interface{}, init bool, peerID libp2p.ID) (map[uint64]uint64, interface{}) {
	var blkBatchsNeedToGet map[uint64]uint64
	blkBatchsNeedToGet = make(map[uint64]uint64)
	var blksSyncByHeight *map[uint64]peerSyncTimestamp
	if init {
		blksSyncByHeight = loadedData.(*map[uint64]peerSyncTimestamp)
		latestBatchBegin := uint64(0)
		for blkHeight := fromHeight; blkHeight <= toHeight; blkHeight++ {
			if timeStamp, ok := (*blksSyncByHeight)[blkHeight]; ok {
				if time.Since(time.Unix(timeStamp.Time, 0)) > defaultMaxBlockSyncTime {
					(*blksSyncByHeight)[blkHeight] = peerSyncTimestamp{
						Time:   time.Now().Unix(),
						PeerID: peerID,
					}
					if latestBatchEnd, ok := blkBatchsNeedToGet[latestBatchBegin]; !ok {
						blkBatchsNeedToGet[blkHeight] = blkHeight
						latestBatchBegin = blkHeight
					} else {
						if latestBatchEnd+1 == blkHeight {
							blkBatchsNeedToGet[latestBatchBegin] = blkHeight
						} else {
							blkBatchsNeedToGet[blkHeight] = blkHeight
							latestBatchBegin = blkHeight
						}
					}
				}
			} else {
				(*blksSyncByHeight)[blkHeight] = peerSyncTimestamp{
					Time:   time.Now().Unix(),
					PeerID: peerID,
				}
				if latestBatchEnd, ok := blkBatchsNeedToGet[latestBatchBegin]; !ok {
					blkBatchsNeedToGet[blkHeight] = blkHeight
					latestBatchBegin = blkHeight
				} else {
					if latestBatchEnd+1 == blkHeight {
						blkBatchsNeedToGet[latestBatchBegin] = blkHeight
					} else {
						blkBatchsNeedToGet[blkHeight] = blkHeight
						latestBatchBegin = blkHeight
					}
				}
			}
		}
	} else {
		blksSyncByHeight = &map[uint64]peerSyncTimestamp{}
		for blkHeight := fromHeight; blkHeight <= toHeight; blkHeight++ {
			(*blksSyncByHeight)[blkHeight] = peerSyncTimestamp{
				Time:   time.Now().Unix(),
				PeerID: peerID,
			}
		}
		blkBatchsNeedToGet[fromHeight] = toHeight
	}
	return blkBatchsNeedToGet, blksSyncByHeight
}

//GetDiffHashesOf Get unique hashes of 1st slice compare to 2nd slice
func GetDiffHashesOf(slice1 []common.Hash, slice2 []common.Hash) []common.Hash {
	var diff []common.Hash

	for _, s1 := range slice1 {
		found := false
		for _, s2 := range slice2 {
			if s1 == s2 {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, s1)
		}
	}

	return diff
}

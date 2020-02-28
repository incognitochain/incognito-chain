package syncker

//
//type S2BPool struct {
//	action            chan func()
//	BlkPoolByHash     map[string]common.BlockPoolInterface // hash -> block
//	BlkPoolByPrevHash map[string][]string                  // prevhash -> []nexthash
//}
//
//func NewS2BPool() *S2BPool {
//	pool := new(S2BPool)
//	pool.action = make(chan func())
//	pool.BlkPoolByHash = make(map[string]common.BlockPoolInterface)
//	pool.BlkPoolByPrevHash = make(map[string][]string)
//	go pool.Start()
//	return pool
//}
//
//func (pool *S2BPool) Start() {
//	ticker := time.NewTicker(time.Millisecond * 500)
//	for {
//		select {
//		case f := <-pool.action:
//			f()
//		default:
//			<-ticker.C
//			//TODO: loop through all prevhash, delete if all nextHash is deleted
//		}
//	}
//}
//
//func (pool *S2BPool) AddBlock(blk common.BlockPoolInterface) {
//	shardID := blk.GetShardID()
//	pool.action <- func() {
//		prevHash := blk.GetPrevHash()
//		hash := blk.GetHash()
//		if _, ok := pool.BlkPoolByHash[shardID][hash]; ok {
//			return
//		}
//		pool.BlkPoolByHash[shardID][hash] = blk
//		if common.IndexOfStr(hash, pool.BlkPoolByPrevHash[shardID][prevHash]) > -1 {
//			return
//		}
//		pool.BlkPoolByPrevHash[shardID][prevHash] = append(pool.BlkPoolByPrevHash[shardID][prevHash], hash)
//		fmt.Printf("Syncker: add s2b block to pool. ShardID %d Height %d", shardID, blk.GetHeight())
//	}
//}
//
//func (pool *S2BPool) RemoveBlock(hash string) {
//	pool.action <- func() {
//		if _, ok := pool.BlkPoolByHash[hash]; ok {
//			delete(pool.BlkPoolByHash, hash)
//		}
//	}
//}
//
//func (pool *BlkPool) GetNextBlock(prevhash string, shouldGetLatest bool) common.BlockPoolInterface {
//	//For multichain, we need to Get a Map
//	res := make(chan common.BlockPoolInterface)
//	pool.action <- func() {
//		hashes := pool.BlkPoolByPrevHash[prevhash][:]
//		for _, h := range hashes {
//			blk := pool.BlkPoolByHash[h]
//			if _, ok := pool.BlkPoolByPrevHash[blk.GetHash()]; shouldGetLatest || ok {
//				res <- pool.BlkPoolByHash[h]
//				return
//			}
//		}
//		res <- nil
//	}
//	return (<-res)
//}

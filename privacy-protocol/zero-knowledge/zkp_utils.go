package zkp

//GetHashOfValues get hash of n points in G append with input values
//return blake_2b(G[0]||G[1]||...||G[CM_CAPACITY-1]||<values>)
//func GenerateChallenge(values [][]byte) []byte {
//	appendStr := Elcm.G[0].CompressPoint()
//	for i := 1; i < CM_CAPACITY; i++ {
//		appendStr = append(appendStr, Elcm.G[i].CompressPoint()...)
//	}
//	for i := 0; i < len(values); i++ {
//		appendStr = append(appendStr, values[i]...)
//	}
//	hashFunc := blake2b.New256()
//	hashFunc.Write(appendStr)
//	hashValue := hashFunc.Sum(nil)
//	return hashValue
//}

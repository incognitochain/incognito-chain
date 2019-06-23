package btc

/*
Timestamp of blockHeight 1447348  is: 	1544500229
Test timestamp: 						1544500800
Timestamp of blockHeight 1447349  is: 	1544501431 <--- result
Timestamp of blockHeight 1447350  is: 1544501634
Timestamp of blockHeight 1447351  is: 1544501799

Nonce of blockHeight 1447348  is: 2463507104
Nonce of blockHeight 1447349  is: 4121500227 <--result
Nonce of blockHeight 1447350  is: 1168465373
*/

/*
Nonce of blockHeight 577265  is: 3543993892
Timestamp of blockHeight 577265  is: 	1444500696

Nonce of blockHeight 577266  is: 3374249745		<---- result
										1444500800 <---- injected timestamp here
Timestamp of blockHeight 577266  is: 1444501304 <---- match timestamp

Nonce of blockHeight 577267  is: 768127857
Timestamp of blockHeight 577267  is: 	1444501387

Nonce of blockHeight 577268  is: 3338477159
Timestamp of blockHeight 577268  is: 1444502621
*/
//func TestGetNonceByTimestamp(t *testing.T) {
//	timestamp1 := 1544500800
//	timestamp2 := 1444500800
//	res, err := GetNonceByTimestamp(int64(timestamp1))
//	if err != nil {
//		t.Errorf("Error geting nonce: %s", err)
//	}
//	if res != int64(4121500227) {
//		t.Errorf("Error geting nonce %d with err: %s", res, err)
//	}
//
//	res, err = GetNonceByTimestamp(int64(timestamp2))
//	if err != nil {
//		t.Errorf("Error geting nonce: %s", err)
//	}
//	if res != int64(3374249745) {
//		t.Errorf("Error geting nonce %d with err: %s", res, err)
//	}
//}
//
//func TestGetNonceByBlock(t *testing.T) {
//		// curl https://api.blockcypher.com/v1/btc/test3/blocks/00000000001be2d75acc520630a117874316c07fd7a724afae1a5d99038f4f4a
//		// curl https://api.blockcypher.com/v1/btc/test3/blocks/294322?start=1&limit=1
//		blockHeight := "294322"
//		flag := true
//		nonce, time, err := GetNonceOrTimeStampByBlock(blockHeight, flag)
//		if err != nil {
//			t.Errorf("Error geting nonce: %s", err)
//			return
//		}
//		if flag {
//			if nonce != 3733494575 {
//				t.Errorf("Error getting nonce in block, nonce should be 3733494575")
//			}
//		} else {
//			if time != 1412364679 {
//				t.Errorf("Error getting time in block, time should be 1396684158")
//			}
//		}
//
//}

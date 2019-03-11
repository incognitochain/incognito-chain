package candidate

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/big0t/constant-chain/utility/generateKeys/generator"
)

const random = int64(4121500227)

var ValidatorList1 = []string{
	"1Uv23eJExNd3rumtpviranhEaKHdfccksyPucLMyaDePhJPNSuKKyfhSYDKPojPJjsvCN9fKWMRksEkGBJZiV2QyWTmM2asD5RdtRHeNA",
	"1Uv4BGuq4FmYqTMEffqhVhYpFPmo6kNBeRiCtSqZCJfduoxixftH67fxzyDgSKL45STpxzciHsNSf7eP3iDQzVZhPFzHknXtpSxgsq5bu",
	"1Uv3vqr76QGCVMRLbZ9vgDZU7mTUHLZcFv7Dy1pHxiswYXnXaABEhXrzjTWDbspeN4UQjYfUbsyQBRWvvjd8X1NytXwuM4bzP2mdWsUGg",
}

var ValidatorList1OutputHash = []string{
	"d24a361d2432dd10531d44827b6f52b3ea19e1e7cf368b23370490753d7250e8",
	"ec7f39e9dc35d069876829b4c3811946d2311fc933a35ab11507f5db1e695912",
	"b356f3c125900d51a58264039bce818c34ec19f15ad908d259ee1d8919b8905f",
}

var ValidatorList1OutputShard = []string{
	"232",
	"18",
	"95",
}
var ValidatorList2 = []string{
	"1Uv3Df2nCnm8MgzCyScbVzCKiBCJG1DbdXKN8sayDX3KwaMDEH1Vw2LmxqZzNfqxoz6jooBVj4M87GgtMfAnErNLYbjLzUpUyM4W7tNCC",
	"1Uv4CPh9RxmZK4kYzAWYDgnRJCjnZ41kqvXiETqfhYsRGKH9NidXTygWUug3nisX9NpfV96Jf2GsxT49Fa2htjQjEQezqLoE5Gg3USzRo",
	"1Uv3dihoiDhShGNn9RxayCkStfhhtrZ9tCn3mJZQmr3nQ8HhtfQn3tqctVm98knugikoxuMtqXMPf5icgR5nd34JMjGA1VAAkCAa7Pj8b",
	"1Uv4ZqdY7bSrhMH3m9ZANbxpzG5aNwkkRCrmsiL5ZGxk7T65tRajy3WWDaeQAzesosEvoMgFxYd6uUnomUih1qtbbM17HuxHYMtnyV29W",
	"1Uv3Zdz1Z5VLzkjWH4ZJDEVNq6TVWFDWVhNvsUToMvsQ8sGN8ipcW2NZ5RHGGX7RN6aBTpxd5YMePbjPPaw41Ysp37yZBkUGRiQm1RQCS",
	"1Uv2zx8j8qXwPLHUPsF5184e9wzvb37FcesmNyFBG5JLapfpfcdxuXcofLdfJzAsGFyLVAWzthRVgPd2tZv95vkz4yoHXtA3Pp4koQNMA",
	"1Uv4UB5sySevxsfdJS4YGnxULbjy9xGw8UNrZwNxcRUHtoK8eo3VLidgaxiCSFNjRGpJCURsrsZSzMqTKHcCPsTDSLuNsvq7rDK3TsSrW",
	"1Uv3zvswZeABdLmQhdVNLqmAC4YCsuyGwRDJk6WPWDjm9W4EXJvkkEfXu9nT5CTbVYt18fuLcEe1rFSaPjEvcBxTT7xBu9QTv6Jvp3swH",
	"1Uv4TtbLhvhQ7tPiB977A4aCJWzQP3DvoiyVqUi4fTKzRCGZcoAUrAEcf7h5EJmpFDTmyPKnf337r213bHyJWChg9G96K3AB498iiLpSz",
	"1Uv3wH9dEgDVPReUSzKu7XU5MQtCVapkhc7gz7fwD6TR4sypknLkm5Tq49BHRUs8D6pzFpaJwuiSCwMNXKhHMtoCkiRdTfqz6bDYMzA83",
	"1Uv2HTQ838zFvvwmJTz89wzCHL4WfLcdbqZsJcGx73GAE2zGZrJMwGh6QMPuqHJr3tyoRGCrgdkESHTppTpDHgeSBBrVwJTER2NYgtR1h",
	"1Uv3QxZ2P7vMB7A42mdYoNP25U7uAnAKUpGuZgvQUnNH5xF9ajfuAxGp3b1Ysw7rYLiVrWrj94bf4aiDDWyeQSLnWfqTWAJa7cmv4sQSm",
	"1Uv2YytznHgwQGZbHkFxpExzit89RXEMGMCmiGz9huYEQvTB3rye58zCgXYj1fxMzq9378JZhiwYnTxbfM3g9JD1Vu7dprM6jFnSHEDgq",
	"1Uv3ZNQwW4w8teiH1UEUypb3MtGdiqenu39QHXecL6r7hNj4NV2VTQL4dRj4H4KVqZoC72nrNDddysMLPHYvH8SD7d8KthNtiRZiYyV2R",
	"1Uv445DRACkpxGzDVt1rT81s4dEgCkYkZK2tguNWFjj6aVzfYWkwFgrN3SQZXEGUrR1iE9jaejmSkfcUEHu9vAB2djDEUvqSA53Se7GsR",
	"1Uv3aNz1LEitNTyQjtfKhofM3WfEZffbHSacMvA5JroZi7dDAXyUJVUsUquNz5UuwkCkDRwTrGbgEJXXiuZRctKAqhZztrTVPMYEGk5in",
	"1Uv28rhQaNhh4fJNcRKLkCbrcRLd7W21WV1pmWBMQ4FTBSLcrL8Zk3K8D92VDK86EPWdHcf37RsZh1k9oWWejNxXgrPTgk6PcGwiu8qow",
}

var ValidatorList2OutputHash = []string{
	"76415d99662f010008ef9a3228a519b34ca110d2a9944f136a69920abdef132a",
	"f809bd206d7f21d9e554eba36ad093d417b0c6f36e3216546bb85ef4d16db3f0",
	"b0e0281635bd9619e4c2983b6eddfffd827ea39696c72856b221db54067a7f44",
	"b80dceb861ac136b1a8f6cff953fbb431a385234b4249c50f66b9e4652a5e570",
	"faa86e449c6b143ba2478580075b2bcb2fb02d6d9a823e8d46d7f6bd265f2e3f",
	"488415af563de3bd6eb4d79715f408fd926adb78f0a3f4d056afc2b16f429a5a",
	"8675a32309a25669cbcbeb02220f7dc464583712f06d82fb7deb2caacf8765f8",
	"f99ba300d4da78247368fe4e82e9ed05434251177a684d149704384f85cf2f85",
	"2acccd393b0b3208144409c7cbf7424c7847916d8558fb5b070d4cf1e114d668",
	"553e4576d013472bbb20b3b059c254b6dca35bd786446c0910073115c0e4c115",
	"55ae5bfea1650451a4ffd7071948aae130bbd99d936900f534687a0fb13379e6",
	"9fa6c68d1b2131e51e906e854240beb169b38a9736b2fd2c3e57d6870cb45b29",
	"72075ce0dec98fd21a55fd5a5be40b36a6b567e56f6ca1233380572898d0d17f",
	"156faa364c97f9a6a9750c0e069eb80912a4ec6ba556591bc9d4e3279149e5be",
	"b8c4ffe29db2af840e6ceb9e9d0a3b5126eb876e6f043d70659d2e3bc3ed07ec",
	"7afce91410013ff7a09ffa550e6475cfc107f36aaab32a35f0557dec1f101a3c",
	"e32fe7ef73c34dfce5baf89e9aa4f2af61ab3623c7d7749d4805d307fdd9fe6b",
}

var ValidatorList2OutputShard = []string{
	"42",
	"240",
	"68",
	"112",
	"63",
	"90",
	"248",
	"133",
	"104",
	"21",
	"230",
	"41",
	"127",
	"190",
	"236",
	"60",
	"107",
}

var ValidatorList1Swap1 = []string{
	"1Uv4BGuq4FmYqTMEffqhVhYpFPmo6kNBeRiCtSqZCJfduoxixftH67fxzyDgSKL45STpxzciHsNSf7eP3iDQzVZhPFzHknXtpSxgsq5bu",
	"1Uv3vqr76QGCVMRLbZ9vgDZU7mTUHLZcFv7Dy1pHxiswYXnXaABEhXrzjTWDbspeN4UQjYfUbsyQBRWvvjd8X1NytXwuM4bzP2mdWsUGg",
	"1Uv3Df2nCnm8MgzCyScbVzCKiBCJG1DbdXKN8sayDX3KwaMDEH1Vw2LmxqZzNfqxoz6jooBVj4M87GgtMfAnErNLYbjLzUpUyM4W7tNCC",
}

var ValidatorList2Swap1 = []string{
	"1Uv4CPh9RxmZK4kYzAWYDgnRJCjnZ41kqvXiETqfhYsRGKH9NidXTygWUug3nisX9NpfV96Jf2GsxT49Fa2htjQjEQezqLoE5Gg3USzRo",
	"1Uv3dihoiDhShGNn9RxayCkStfhhtrZ9tCn3mJZQmr3nQ8HhtfQn3tqctVm98knugikoxuMtqXMPf5icgR5nd34JMjGA1VAAkCAa7Pj8b",
	"1Uv4ZqdY7bSrhMH3m9ZANbxpzG5aNwkkRCrmsiL5ZGxk7T65tRajy3WWDaeQAzesosEvoMgFxYd6uUnomUih1qtbbM17HuxHYMtnyV29W",
	"1Uv3Zdz1Z5VLzkjWH4ZJDEVNq6TVWFDWVhNvsUToMvsQ8sGN8ipcW2NZ5RHGGX7RN6aBTpxd5YMePbjPPaw41Ysp37yZBkUGRiQm1RQCS",
	"1Uv2zx8j8qXwPLHUPsF5184e9wzvb37FcesmNyFBG5JLapfpfcdxuXcofLdfJzAsGFyLVAWzthRVgPd2tZv95vkz4yoHXtA3Pp4koQNMA",
	"1Uv4UB5sySevxsfdJS4YGnxULbjy9xGw8UNrZwNxcRUHtoK8eo3VLidgaxiCSFNjRGpJCURsrsZSzMqTKHcCPsTDSLuNsvq7rDK3TsSrW",
	"1Uv3zvswZeABdLmQhdVNLqmAC4YCsuyGwRDJk6WPWDjm9W4EXJvkkEfXu9nT5CTbVYt18fuLcEe1rFSaPjEvcBxTT7xBu9QTv6Jvp3swH",
	"1Uv4TtbLhvhQ7tPiB977A4aCJWzQP3DvoiyVqUi4fTKzRCGZcoAUrAEcf7h5EJmpFDTmyPKnf337r213bHyJWChg9G96K3AB498iiLpSz",
	"1Uv3wH9dEgDVPReUSzKu7XU5MQtCVapkhc7gz7fwD6TR4sypknLkm5Tq49BHRUs8D6pzFpaJwuiSCwMNXKhHMtoCkiRdTfqz6bDYMzA83",
	"1Uv2HTQ838zFvvwmJTz89wzCHL4WfLcdbqZsJcGx73GAE2zGZrJMwGh6QMPuqHJr3tyoRGCrgdkESHTppTpDHgeSBBrVwJTER2NYgtR1h",
	"1Uv3QxZ2P7vMB7A42mdYoNP25U7uAnAKUpGuZgvQUnNH5xF9ajfuAxGp3b1Ysw7rYLiVrWrj94bf4aiDDWyeQSLnWfqTWAJa7cmv4sQSm",
	"1Uv2YytznHgwQGZbHkFxpExzit89RXEMGMCmiGz9huYEQvTB3rye58zCgXYj1fxMzq9378JZhiwYnTxbfM3g9JD1Vu7dprM6jFnSHEDgq",
	"1Uv3ZNQwW4w8teiH1UEUypb3MtGdiqenu39QHXecL6r7hNj4NV2VTQL4dRj4H4KVqZoC72nrNDddysMLPHYvH8SD7d8KthNtiRZiYyV2R",
	"1Uv445DRACkpxGzDVt1rT81s4dEgCkYkZK2tguNWFjj6aVzfYWkwFgrN3SQZXEGUrR1iE9jaejmSkfcUEHu9vAB2djDEUvqSA53Se7GsR",
	"1Uv3aNz1LEitNTyQjtfKhofM3WfEZffbHSacMvA5JroZi7dDAXyUJVUsUquNz5UuwkCkDRwTrGbgEJXXiuZRctKAqhZztrTVPMYEGk5in",
	"1Uv28rhQaNhh4fJNcRKLkCbrcRLd7W21WV1pmWBMQ4FTBSLcrL8Zk3K8D92VDK86EPWdHcf37RsZh1k9oWWejNxXgrPTgk6PcGwiu8qow",
}

var ValidatorList1Swap2 = []string{}

var ValidatorList2Swap2 = []string{
	"1Uv4ZqdY7bSrhMH3m9ZANbxpzG5aNwkkRCrmsiL5ZGxk7T65tRajy3WWDaeQAzesosEvoMgFxYd6uUnomUih1qtbbM17HuxHYMtnyV29W",
	"1Uv3Zdz1Z5VLzkjWH4ZJDEVNq6TVWFDWVhNvsUToMvsQ8sGN8ipcW2NZ5RHGGX7RN6aBTpxd5YMePbjPPaw41Ysp37yZBkUGRiQm1RQCS",
	"1Uv2zx8j8qXwPLHUPsF5184e9wzvb37FcesmNyFBG5JLapfpfcdxuXcofLdfJzAsGFyLVAWzthRVgPd2tZv95vkz4yoHXtA3Pp4koQNMA",
	"1Uv4UB5sySevxsfdJS4YGnxULbjy9xGw8UNrZwNxcRUHtoK8eo3VLidgaxiCSFNjRGpJCURsrsZSzMqTKHcCPsTDSLuNsvq7rDK3TsSrW",
	"1Uv3zvswZeABdLmQhdVNLqmAC4YCsuyGwRDJk6WPWDjm9W4EXJvkkEfXu9nT5CTbVYt18fuLcEe1rFSaPjEvcBxTT7xBu9QTv6Jvp3swH",
	"1Uv4TtbLhvhQ7tPiB977A4aCJWzQP3DvoiyVqUi4fTKzRCGZcoAUrAEcf7h5EJmpFDTmyPKnf337r213bHyJWChg9G96K3AB498iiLpSz",
	"1Uv3wH9dEgDVPReUSzKu7XU5MQtCVapkhc7gz7fwD6TR4sypknLkm5Tq49BHRUs8D6pzFpaJwuiSCwMNXKhHMtoCkiRdTfqz6bDYMzA83",
	"1Uv2HTQ838zFvvwmJTz89wzCHL4WfLcdbqZsJcGx73GAE2zGZrJMwGh6QMPuqHJr3tyoRGCrgdkESHTppTpDHgeSBBrVwJTER2NYgtR1h",
	"1Uv3QxZ2P7vMB7A42mdYoNP25U7uAnAKUpGuZgvQUnNH5xF9ajfuAxGp3b1Ysw7rYLiVrWrj94bf4aiDDWyeQSLnWfqTWAJa7cmv4sQSm",
	"1Uv2YytznHgwQGZbHkFxpExzit89RXEMGMCmiGz9huYEQvTB3rye58zCgXYj1fxMzq9378JZhiwYnTxbfM3g9JD1Vu7dprM6jFnSHEDgq",
	"1Uv3ZNQwW4w8teiH1UEUypb3MtGdiqenu39QHXecL6r7hNj4NV2VTQL4dRj4H4KVqZoC72nrNDddysMLPHYvH8SD7d8KthNtiRZiYyV2R",
	"1Uv445DRACkpxGzDVt1rT81s4dEgCkYkZK2tguNWFjj6aVzfYWkwFgrN3SQZXEGUrR1iE9jaejmSkfcUEHu9vAB2djDEUvqSA53Se7GsR",
	"1Uv3aNz1LEitNTyQjtfKhofM3WfEZffbHSacMvA5JroZi7dDAXyUJVUsUquNz5UuwkCkDRwTrGbgEJXXiuZRctKAqhZztrTVPMYEGk5in",
	"1Uv28rhQaNhh4fJNcRKLkCbrcRLd7W21WV1pmWBMQ4FTBSLcrL8Zk3K8D92VDK86EPWdHcf37RsZh1k9oWWejNxXgrPTgk6PcGwiu8qow",
	"1Uv23eJExNd3rumtpviranhEaKHdfccksyPucLMyaDePhJPNSuKKyfhSYDKPojPJjsvCN9fKWMRksEkGBJZiV2QyWTmM2asD5RdtRHeNA",
	"1Uv4BGuq4FmYqTMEffqhVhYpFPmo6kNBeRiCtSqZCJfduoxixftH67fxzyDgSKL45STpxzciHsNSf7eP3iDQzVZhPFzHknXtpSxgsq5bu",
	"1Uv3vqr76QGCVMRLbZ9vgDZU7mTUHLZcFv7Dy1pHxiswYXnXaABEhXrzjTWDbspeN4UQjYfUbsyQBRWvvjd8X1NytXwuM4bzP2mdWsUGg",
}

//Result has been test with go runtime in https://golang.org/pkg/crypto/sha256/
func TestAssignValidator(t *testing.T) {
	for index, validator := range ValidatorList1 {
		res := calculateHash(validator, random)
		shardID := strconv.Itoa(int(res))
		if ValidatorList1OutputShard[index] != shardID {
			t.Errorf("Error calculating shardID")
		}
	}
	for index, validator := range ValidatorList2 {
		res := calculateHash(validator, random)
		shardID := strconv.Itoa(int(res))
		if ValidatorList2OutputShard[index] != shardID {
			t.Errorf("Error calculating shardID")
		}
	}
}

func TestVerifyValidator(t *testing.T) {
	for index, validator := range ValidatorList1 {
		res, err := strconv.Atoi(ValidatorList1OutputShard[index])
		if err != nil {
			t.Errorf("Error parsing shardID")
		}
		res2, err := VerifyValidator(validator, random, byte(res))
		if res2 == false {
			t.Errorf("Error verifying validator")
		}
	}
	for index, validator := range ValidatorList2 {
		res, err := strconv.Atoi(ValidatorList2OutputShard[index])
		if err != nil {
			t.Errorf("Error parsing shardID")
		}
		res2, err := VerifyValidator(validator, random, byte(res))
		if res2 == false {
			t.Errorf("Error verifying validator")
		}
	}
}

func TestAssignValidatorIntoMap(t *testing.T) {
	res, err := AssignValidator(ValidatorList1, random)
	if err != nil {
		t.Errorf("Error getting validator assign map")
	}
	for index, _ := range ValidatorList1OutputShard {
		shardID, err := strconv.Atoi(ValidatorList1OutputShard[index])
		if err != nil {
			t.Errorf("Error parsing shardID")
		}
		if res[byte(shardID)][0] != ValidatorList1[index] {
			t.Errorf("Error verifying map")
		}
	}
}

func TestSwapValidator(t *testing.T) {
	pendingValidator, currentValidator, err := SwapValidator(generator.PreSelectShardNodeTestnetSerializedPubkey, generator.PreSelectBeaconNodeTestnetSerializedPubkey, 1)
	if err != nil {
		t.Errorf("Error swapping validator %+v", err)
	}
	if reflect.DeepEqual(pendingValidator, ValidatorList2Swap1) == false {
		t.Errorf("Error swapping validator")
	}
	if reflect.DeepEqual(currentValidator, ValidatorList1Swap1) == false {
		t.Errorf("Error swapping validator")
	}

	pendingValidator, currentValidator, err = SwapValidator(generator.PreSelectShardNodeTestnetSerializedPubkey, generator.PreSelectBeaconNodeTestnetSerializedPubkey, 4)
	if err == nil {
		t.Errorf("Error %+v should be fired", err)
	}

	pendingValidator, currentValidator, err = SwapValidator(generator.PreSelectShardNodeTestnetSerializedPubkey, generator.PreSelectBeaconNodeTestnetSerializedPubkey, 0)
	if err == nil {
		t.Errorf("Error %+v should be fired", err)
	}

	pendingValidator, currentValidator, err = SwapValidator(generator.PreSelectBeaconNodeTestnetSerializedPubkey, generator.PreSelectShardNodeTestnetSerializedPubkey, 5)
	if err != nil {
		t.Errorf("Error swapping validator %+v", err)
	}
	if reflect.DeepEqual(pendingValidator, ValidatorList1Swap2) == false {
		t.Errorf("Error swapping validator")
	}
	if reflect.DeepEqual(currentValidator, ValidatorList2Swap2) == false {
		t.Errorf("Error swapping validator")
	}

}

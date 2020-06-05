package multiview

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PrivateList struct {
	KeyMap map[byte][]string `json:"KeyMap"`
}

type Scene struct {
	From     uint64              `json:"From"`
	To       uint64              `json:"To"`
	PubGroup map[string][]string `json:"PublishGroup"`
}

func LoadKeyList(filename string) *PrivateList {
	privateList := &PrivateList{}
	if filename != "" {
		jsonFile, err := os.Open(filename)
		// if we os.Open returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Successfully Opened private.json")
		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		// var result map[string]interface{}
		err = json.Unmarshal([]byte(byteValue), &privateList)
		if err != nil {
			fmt.Println(err)
		}
		// fmt.Println(privateList)
	}
	return privateList
}

func MiningKeyFromPrivateKey(priKey string) string {
	wl, _ := wallet.Base58CheckDeserialize(priKey)
	wl.KeySet.InitFromPrivateKey(&wl.KeySet.PrivateKey)
	committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(wl.KeySet.PrivateKey)), wl.KeySet.PaymentAddress.Pk)
	return committeeKey.GetMiningKeyBase58(common.BlsConsensus)
}

func ListPrivateToPayLoad(
	priList *PrivateList,
	PubGroup map[int]map[int][]int,
	From map[int]uint64,
	To map[int]uint64,
) (
	res map[string]Scene,
) {
	chainKey := ""
	res = map[string]Scene{}
	for sID, listPrivate := range priList.KeyMap {
		listPubKey := []string{}
		for _, priKey := range listPrivate {
			listPubKey = append(listPubKey, MiningKeyFromPrivateKey(priKey))
		}
		if sID == 255 {
			chainKey = "beacon"
		} else {
			chainKey = fmt.Sprintf("shard-%v", sID)
		}
		if from, ok := From[int(sID)]; ok {
			res[chainKey] = Scene{
				From:     from,
				To:       To[int(sID)],
				PubGroup: map[string][]string{},
			}
		} else {
			res[chainKey] = Scene{
				From:     0,
				To:       10000,
				PubGroup: map[string][]string{},
			}
		}
		for src, dsts := range PubGroup[int(sID)] {
			for _, dst := range dsts {
				res[chainKey].PubGroup[listPubKey[src-1]] = append(res[chainKey].PubGroup[listPubKey[src-1]], listPubKey[dst-1])
			}
		}
	}
	return res
}

func GenPayload(
	filePriKey string,
	pubGroup map[int]map[int][]int,
	From map[int]uint64,
	To map[int]uint64,
	fileOut string,
) {
	priList := LoadKeyList(filePriKey)
	scenario := ListPrivateToPayLoad(priList, pubGroup, From, To)
	res, _ := json.Marshal(scenario)
	fmt.Println(string(res))
}

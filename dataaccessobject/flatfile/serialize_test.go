package flatfile

import (
	"fmt"
	"os"
	"testing"
	"time"
)

//func SerializeShardBlock(bc *blockchain.BlockChain) {
//	ff, _ := NewFlatFile("/data/rawblock", 5000)
//	c, _, cancel := ff.ReadRecently()
//	blk := <-c
//	cancel()
//	blockHeight := 2
//	if blk != nil {
//		shardBlock := types.NewShardBlock()
//		err := json.Unmarshal(blk, shardBlock)
//		if err != nil {
//			panic(err)
//		}
//		blockHeight = int(shardBlock.GetHeight() + 1)
//	}
//
//	shardID := byte(0)
//	for {
//		blkhash, err := bc.
//			GetShardBlockHashByHeight(bc.ShardChain[shardID].
//				GetFinalView(), bc.ShardChain[shardID].GetBestView(), uint64(blockHeight))
//		if err != nil {
//			break
//		}
//		data, err := rawdbv2.GetShardBlockByHash(bc.GetShardChainDatabase(shardID), *blkhash)
//		if err != nil {
//			break
//		}
//		blockHeight++
//		ff.Append(data)
//	}
//
//}

func TestNewFlatFile(t *testing.T) {
	os.RemoveAll("./tmp")
	defer os.RemoveAll("./tmp")
	ff, _ := NewFlatFile("./tmp", 2)
	var genStr = func(s string) string {
		res := ""
		for i := 0; i < 10; i++ {
			res += s
		}
		return res
	}

	id, _ := ff.Append([]byte(genStr("1")))
	fmt.Println(id)
	ff.Append([]byte(genStr("2")))
	ff.Append([]byte(genStr("3")))
	ff.Append([]byte(genStr("4")))
	ff.Append([]byte(genStr("5")))
	ff.Append([]byte(genStr("6")))

	str := []byte(genStr("7"))
	id, _ = ff.Append(str)
	fmt.Println(id)
	fmt.Println("write", str)

	str2, err := ff.Read(id)
	fmt.Println("read", str2, err)
	ff.Truncate(5)

	c, e, _ := ff.ReadRecently(5)
	for {
		select {
		case msg := <-c:
			fmt.Println(msg)
			if msg == nil {
				return
			}
		case msg := <-e:
			fmt.Println(msg)
		}
		time.Sleep(time.Second)
	}

}

func TestFlatFileManager_Truncate(t *testing.T) {
	os.RemoveAll("./tmp")
	ff, _ := NewFlatFile("./tmp", 2)
	var genStr = func(s string) string {
		res := ""
		for i := 0; i < 10; i++ {
			res += s
		}
		return res
	}

	ff.Append([]byte(genStr("0")))
	ff.Append([]byte(genStr("1")))
	ff.Append([]byte(genStr("2")))
	ff.Append([]byte(genStr("3")))
	ff.Append([]byte(genStr("4")))
	ff.Append([]byte(genStr("5")))

	str := []byte(genStr("6"))
	ff.Append(str)

	ff.Truncate(5)

	str2, err := ff.Read(5)
	if err != nil {
		t.Fatal(err)
	}
	if string(str2) != genStr("5") {
		t.Fatalf("expect %+v, got %+v", genStr("5"), string(str2))
	}

	str2, err = ff.Read(4)
	if err != nil {
		t.Fatal(err)
	}
	if string(str2) != genStr("4") {
		t.Fatalf("expect %+v, got %+v", genStr("4"), string(str2))
	}

	_, err = ff.Read(3)
	if err == nil {
		t.Fatal(err)
	}

}

package benchmark

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"
)

type benchmarkCollector struct {
	cpuUsage    float64
	memSys      uint64
	dataSize    map[string]uint64
	interval    uint64
	lastTime    time.Time
	colelctType int //0 -> time,cpu,ram, 1: db
	fd          *os.File
}

var BenchmarkCollector = &benchmarkCollector{}

func (bm *benchmarkCollector) Init(collectType int, interval uint64) {
	bm.interval = interval
	var err error
	bm.fd, err = os.OpenFile("/data/benchmark", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	bm.colelctType = collectType
	bm.lastTime = time.Now()
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		idle0, total0 := common.GetCPUSample()
		for _ = range ticker.C {
			idle1, total1 := common.GetCPUSample()
			idleTicks := float64(idle1 - idle0)
			totalTicks := float64(total1 - total0)
			cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks
			idle0, total0 = common.GetCPUSample()
			//ema, 10 samples
			rate := float64(2) / 11
			bm.cpuUsage = cpuUsage*rate + bm.cpuUsage*(1-rate)

		}
	}()
}

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func (bm *benchmarkCollector) Collect(chainID int, blkHeight uint64) {
	if blkHeight%bm.interval == 0 {
		if bm.colelctType == 0 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			memUsage := m.Sys >> 20
			str := fmt.Sprintf("%v,%v,%v,%.0f,%v,%.6f,%v", chainID, time.Now().Unix(), blkHeight, bm.cpuUsage, memUsage, time.Since(bm.lastTime).Seconds(), time.Since(bm.lastTime).Milliseconds()/int64(bm.interval))
			bm.lastTime = time.Now()
			bm.fd.WriteString(str + "\n")
		} else {
			var db int64
			if chainID == -1 {
				db, _ = DirSize(path.Join(config.Config().DataDir, "block/beacon"))
			} else {
				db, _ = DirSize(path.Join(config.Config().DataDir, fmt.Sprintf("block/shard%v", chainID)))
			}
			str := fmt.Sprintf("%v,%v,%v,%v", chainID, time.Now().Unix(), blkHeight, db)
			bm.fd.WriteString(str + "\n")
		}

	}
}

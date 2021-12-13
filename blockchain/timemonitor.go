package blockchain

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
)

var timeMonitor *TimeMonitor
var fd *os.File

func init() {
	timeMonitor = &TimeMonitor{
		eventTime:    map[string]int64{},
		eventPercent: map[string]int64{},
	}
	fd1, err := os.OpenFile("/data/metrics", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	fd = fd1

	////disk usage
	//fs := syscall.Statfs_t{}
	//err = syscall.Statfs("/data/mainnet/block/shard0", &fs)
	//if err != nil {
	//	panic(err)
	//}
	//All := fs.Blocks * uint64(fs.Bsize)
	//Free := fs.Bfree * uint64(fs.Bsize)
	//Used := All - Free
	//timeMonitor.usageSize = Used

	if err != nil {
		panic(err)
	}
}

type TimeMonitor struct {
	currentEpoch uint64
	currentNumTx uint64
	usageSize    uint64

	startTime    int64
	lasttime     int64
	eventTime    map[string]int64
	eventPercent map[string]int64
}

func GetTimeMonitor() *TimeMonitor {
	return timeMonitor
}

func (t *TimeMonitor) Start() {
	t.startTime = time.Now().UnixNano()
	t.lasttime = t.startTime
	t.eventTime = map[string]int64{}
	t.eventPercent = map[string]int64{}
}

func (t *TimeMonitor) Tick(event string) {
	Logger.log.Info(event)
	elapseTime := time.Now().UnixNano() - t.lasttime
	t.lasttime = time.Now().UnixNano()
	t.eventTime[event] = elapseTime
}

func (t *TimeMonitor) Stop(epoch uint64, height uint64, tx uint64) {
	pStr := []string{}
	totalTime := time.Now().UnixNano() - t.startTime
	pStr = append(pStr, fmt.Sprintf("%v:%v", "totalTime", totalTime/1e6))
	pStr = append(pStr, fmt.Sprintf("%v:%v", "height", height))
	pStr = append(pStr, fmt.Sprintf("%v:%v", "epoch", epoch))
	pStr = append(pStr, fmt.Sprintf("%v:%v", "tx", tx))
	t.currentNumTx += tx
	for k, v := range t.eventTime {
		t.eventPercent[k] = (100 * v) / totalTime
		if t.eventPercent[k] > 5 {
			pStr = append(pStr, fmt.Sprintf("%v:%v:%v", "event", k, t.eventPercent[k]))
		}
	}

	newEpoch := false
	if epoch > t.currentEpoch {
		newEpoch = true
		t.currentEpoch = epoch
		pStr = append(pStr, fmt.Sprintf("%v:%v", "epochTotalTx", t.currentNumTx))
		t.currentNumTx = 0

		//disk usage
		fs := syscall.Statfs_t{}
		err := syscall.Statfs("/data/mainnet/block/shard0", &fs)
		if err != nil {
			panic(err)
		}
		All := fs.Blocks * uint64(fs.Bsize)
		Free := fs.Bfree * uint64(fs.Bsize)
		Used := All - Free
		pStr = append(pStr, fmt.Sprintf("%v:%v", "increaseSize", (Used)/(1024*1024)))
	}

	if totalTime > 10*1e6 || newEpoch {
		fd.Write([]byte(strings.Join(pStr, ",")))
		fd.Write([]byte("\n"))
		fmt.Println(strings.Join(pStr, ","))
	}

}

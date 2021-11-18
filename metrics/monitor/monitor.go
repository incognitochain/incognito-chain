package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"syscall"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metrics"

	"github.com/google/uuid"
	"github.com/incognitochain/incognito-chain/blockchain"

	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var monitorFile *os.File
var globalParam *logKV
var blockchainObj *blockchain.BlockChain

func init() {
	uid := uuid.New()
	globalParam = &logKV{param: make(map[string]interface{})}
	SetGlobalParam("UID", uid.String())
	commitID := os.Getenv("commit")
	if commitID == "" {
		commitID = "NA"
	}
	SetGlobalParam("CommitID", commitID)
	//var err error
	//monitorFile, err = os.OpenFile("/data/monitor.json", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	//if err != nil {
	//	panic("Cannot open to monitor file")
	//}

	go func() {
		ticker := time.NewTicker(40 * time.Second)
		idle0, total0 := common.GetCPUSample()
		var m runtime.MemStats
		for _ = range ticker.C {
			if blockchainObj == nil {
				time.Sleep(time.Second)
				continue
			}
			l := NewLog()
			idle1, total1 := common.GetCPUSample()
			idleTicks := float64(idle1 - idle0)
			totalTicks := float64(total1 - total0)
			cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks
			runtime.ReadMemStats(&m)

			//disk usage
			fs := syscall.Statfs_t{}
			err := syscall.Statfs("/data", &fs)
			if err == nil {
				All := fs.Blocks * uint64(fs.Bsize)
				Free := fs.Bfree * uint64(fs.Bsize)
				Used := All - Free
				l.Add("DISK_USAGE", fmt.Sprintf("%.2f", float64(Used*100)/float64(All)))
			}
			l.Add("CPU_USAGE", fmt.Sprintf("%.2f", cpuUsage), "MEM_USAGE", m.Sys>>20)
			idle0, total0 = common.GetCPUSample()
			l.Write()
		}
	}()
}

type logKV struct {
	param map[string]interface{}
	sync.RWMutex
}

func SetGlobalParam(p ...interface{}) {
	globalParam.Add(p...)
}

func SetBlockChainObj(obj *blockchain.BlockChain) {
	blockchainObj = obj
}

func NewLog(p ...interface{}) *logKV {
	nl := (&logKV{param: make(map[string]interface{})}).Add(p...)
	globalParam.RLock()
	for k, v := range globalParam.param {
		nl.param[k] = v
	}
	globalParam.RUnlock()
	return nl
}

func (s *logKV) Add(p ...interface{}) *logKV {
	if len(p) == 0 || len(p)%2 != 0 {
		return s
	}
	s.Lock()
	defer s.Unlock()
	for i, v := range p {
		if i%2 == 0 {
			s.param[v.(string)] = p[i+1]
		}
	}
	return s
}

func (s *logKV) Write() {
	s.RLock()
	defer s.RUnlock()
	//fn, f, l := getMethodName(2)
	//s.param["FILE"] = fmt.Sprintf("%s:%s", f, l)
	//r, _ := regexp.Compile("(^[^\\.]*)")
	//s.param["PACKAGE"] = fmt.Sprintf("%s", r.FindStringSubmatch(fn)[1])
	s.param["Time"] = fmt.Sprintf("%s", time.Now().Format(time.RFC3339))
	b, _ := json.Marshal(s.param)
	if v, ok := s.param["MINING_PUBKEY"]; !ok || v == "" {
		return
	}
	//io.Copy(monitorFile, bytes.NewReader(b))
	//io.Copy(monitorFile, bytes.NewReader([]byte("\n")))

	go func() {
		monitorEP := os.Getenv("MONITOR")
		if monitorEP != "" {
			req, err := http.NewRequest(http.MethodPost, monitorEP, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			if err != nil {
				metrics.IncLogger.Log.Debug("Create Request failed with err: ", err)
				return
			}
			ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
			defer cancel()
			req = req.WithContext(ctx)
			client := &http.Client{}
			client.Do(req)
		}

	}()
}

func getMethodName(depthList ...int) (string, string, string) {
	var depth int
	if depthList == nil {
		depth = 1
	} else {
		depth = depthList[0]
	}
	function, file, line, _ := runtime.Caller(depth)
	r, _ := regexp.Compile("([^/]*$)")
	r1, _ := regexp.Compile("/([^/]*$)")
	return r.FindStringSubmatch(runtime.FuncForPC(function).Name())[1], r1.FindStringSubmatch(file)[1], strconv.Itoa(line)
}

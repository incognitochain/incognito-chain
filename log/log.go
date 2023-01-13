package log

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/log/proto"
	"github.com/pkg/errors"
)

var FLogManager FeatureLogManager

type FeatureLogManager struct {
	isEnable bool
	db       incdb.Database
	locker   *sync.Mutex
}

func (f *FeatureLogManager) IsEnable() bool {
	if f.locker == nil {
		return false
	}
	f.locker.Lock()
	defer f.locker.Unlock()
	return f.isEnable

}

func (f *FeatureLogManager) Enable() error {
	if f.locker == nil {
		f.Init()
		return nil
	}
	f.locker.Lock()
	defer f.locker.Unlock()
	if f.isEnable {
		return errors.Errorf("Can not enable again")
	}
	f.isEnable = true
	return nil
}

func (f *FeatureLogManager) Disable() error {
	if f.locker == nil {
		return nil
	}
	f.locker.Lock()
	defer f.locker.Unlock()
	if !f.isEnable {
		return errors.Errorf("Can not disable again")
	}
	f.isEnable = false
	return nil
}

func (f *FeatureLogManager) Init() {
	logDB, err := incdb.Open("leveldb", filepath.Join(config.Config().DataDir, "featurelog"))
	if err != nil {
		incdb.Logger.Log.Error("could not open connection to leveldb")
		incdb.Logger.Log.Error(err)
		panic(err)
	}
	f.db = logDB
	f.locker = &sync.Mutex{}
	f.Enable()
}

func (f *FeatureLogManager) SetDatabase(db incdb.Database) {
	f.db = db
}

func (f *FeatureLogManager) GetDatabase() incdb.Database {
	return f.db
}

func GetLogKey(f *proto.FeatureLog) []byte {
	key := ""
	key += strconv.FormatUint(f.CheckPoint.BlockHeight, 10)
	blkHash := common.Hash{}.NewHashFromStr2(string(f.CheckPoint.BlockHash))
	// copy(blkHash[:], f.CheckPoint.BlockHash)
	key += "-" + blkHash.String()
	key += "-" + f.ID.String()
	key += "-" + strconv.FormatInt(f.Timestamp, 10)
	fmt.Printf("testtest %v - %v \n", blkHash.String(), key)
	return []byte(key)
}

func (f *FeatureLogManager) NewFLogWithValue(r *proto.CheckPoint, t int64, tag proto.FeatureID, value string) *proto.FeatureLog {
	fLog := proto.FeatureLog{
		Timestamp:  t,
		CheckPoint: r,
		ID:         tag,
		Data:       []byte(value),
	}
	return &fLog
}

func (f *FeatureLogManager) Store(fLog *proto.FeatureLog) error {
	if !f.IsEnable() {
		return errors.Errorf("FeatureLog is disable")
	}
	key := GetLogKey(fLog)
	value := fLog.Data
	return f.db.Put(key, value)
}

func (f *FeatureLogManager) GetFeatureLog(r *proto.CheckPoint) ([]proto.FeatureLog, error) {
	if !f.IsEnable() {
		return nil, errors.Errorf("FeatureLog is disable")
	}
	res := []proto.FeatureLog{}
	prefix := ""
	prefix += strconv.FormatUint(r.BlockHeight, 10)
	if len(r.BlockHash) != 0 {
		blkHash := common.Hash{}
		copy(blkHash[:], r.BlockHash)
		prefix += "-" + blkHash.String()
	}
	it := f.db.NewIteratorWithPrefix([]byte(prefix))
	for it.Next() {
		key := make([]byte, len(it.Key()))
		copy(key, it.Key())
		keys := strings.Split(string(key), "-")
		blkHashStr := keys[1]
		blkHash := common.Hash{}.NewHashFromStr2(blkHashStr)
		fIDStr := keys[2]
		fID, ok := proto.FeatureID_value[fIDStr]
		if !ok {
			continue
		}
		timestampStr := keys[3]
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			incdb.Logger.Log.Error(err)
			continue
		}
		fLog := proto.FeatureLog{
			Timestamp: timestamp,
			ID:        proto.FeatureID(fID),
			Data:      it.Value(),
			CheckPoint: &proto.CheckPoint{
				BlockHeight: r.BlockHeight,
				BlockHash:   blkHash.String(),
			},
		}
		res = append(res, fLog)
	}
	return res, nil
}

func (f *FeatureLogManager) GetRangeFeatureLog(r *proto.RangeCheckPoint) ([]proto.FeatureLog, error) {
	if !f.IsEnable() {
		return nil, errors.Errorf("FeatureLog is disable")
	}
	res := []proto.FeatureLog{}
	for blkHeight := r.From.BlockHeight; blkHeight <= r.To.BlockHeight; blkHeight++ {
		fLogs, err := f.GetFeatureLog(&proto.CheckPoint{BlockHeight: blkHeight, BlockHash: ""})
		if err != nil {
			incdb.Logger.Log.Error(err)
			continue
		}
		res = append(res, fLogs...)
	}
	return res, nil
}

func (f *FeatureLogManager) GetFeatureLogByFeature(rf *proto.RequestLogByFeature) ([]proto.FeatureLog, error) {
	if !f.IsEnable() {
		return nil, errors.Errorf("FeatureLog is disable")
	}
	fID := rf.ID
	res := []proto.FeatureLog{}
	fLogs, err := f.GetRangeFeatureLog(rf.CheckPoint)
	if err != nil {
		return nil, err
	}
	for _, fLog := range fLogs {
		if fLog.ID == fID {
			res = append(res, fLog)
		}
	}
	return res, nil
}

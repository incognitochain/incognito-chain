package flatfile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"os"
	"strconv"
)

type FlatFile interface {
	//append item uint64o flat file, return item index
	Append([]byte) (uint64, error)

	//read item in flatfile with specific index (return from append)
	Read(index uint64) ([]byte, error)

	//read recent data, return data channel, errpr channel, and cancel function
	ReadRecently(index uint64) (chan []byte, chan uint64, func())

	//truncate flat file system
	Truncate(lastIndex uint64) error

	//return current size of ff
	Size() uint64
}

type FlatFileManager struct {
	dataDir       string
	fileSizeLimit uint64 //number of item

	folderMap       map[uint64]bool
	currentFD       *os.File
	currentFile     uint64
	currentFileSize uint64

	parseCache *lru.Cache
	itemCache  *lru.Cache
	lock       *sync.RWMutex
}

type ReadInfo struct {
	fd     *os.File
	offset int64
	size   int64
}

func (ff *FlatFileManager) Size() uint64 {
	return ff.currentFile*ff.fileSizeLimit + ff.currentFileSize
}

func (ff *FlatFileManager) Truncate(lastIndex uint64) error {
	lastFile := lastIndex / ff.fileSizeLimit
	files, err := ioutil.ReadDir(ff.dataDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		name := filepath.Base(f.Name())
		i, err := strconv.Atoi(name)
		if err == nil {
			if uint64(i) < lastFile {

				err := os.Remove(path.Join(ff.dataDir, name))
				log.Println("truncate file ", lastIndex, uint64(i), lastFile, err)
			}
		}
	}
	//newff, err := NewFlatFile(ff.dataDir, ff.fileSizeLimit)
	//if err != nil {
	//	return nil
	//}
	//*ff = *newff
	return nil
}

func (f *FlatFileManager) PasreFile(fileID uint64) (map[uint64]ReadInfo, error) {

	f.lock.RLock()
	defer f.lock.RUnlock()

	//TODO: concert int64
	p := path.Join(f.dataDir, strconv.Itoa(int(fileID)))
	fd, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	readInfos := make(map[uint64]ReadInfo)
	offset := int64(0)
	size := 0
	for {
		b := make([]byte, 8)
		n, err := fd.ReadAt(b, offset)
		if err != nil && err.Error() != "EOF" {
			return nil, err
		}
		var result int64
		err = binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &result)
		if err != nil || n == 0 {
			break
		}

		readInf := ReadInfo{
			fd,
			offset + 8,
			result,
		}
		id := len(readInfos)
		readInfos[uint64(id)] = readInf
		offset += 8
		offset += result
		size++
	}

	f.parseCache.Add(fileID, readInfos)

	return readInfos, nil
}

func (f FlatFileManager) Read(index uint64) ([]byte, error) {

	//get from cache byte firm
	cacheByte, ok := f.itemCache.Get(index)
	if ok {
		return cacheByte.([]byte), nil
	}

	//else, parse and read
	fileID := index / f.fileSizeLimit
	itemFileIndex := index % f.fileSizeLimit
	var readInfo map[uint64]ReadInfo
	var err error

	//find parse info in cache first
	v, ok := f.parseCache.Get(fileID)
	if ok {
		if int(itemFileIndex) < len(v.(map[uint64]ReadInfo)) {
			readInfo = v.(map[uint64]ReadInfo)
		} else { //if find item index is greater than len of cache <- could be new item, parse again
			readInfo, err = f.PasreFile(fileID)
		}
	} else { //no cache, parse
		readInfo, err = f.PasreFile(fileID)
	}

	if err != nil {
		return nil, err
	}
	if int(itemFileIndex) >= len(readInfo) {
		return nil, errors.New(fmt.Sprintf("Cannot read item at index %v", index))
	}

	b := make([]byte, readInfo[itemFileIndex].size)
	readInfo[itemFileIndex].fd.ReadAt(b, int64(readInfo[itemFileIndex].offset))
	rawB, err := ioutil.ReadAll(bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	return rawB, nil
}

func (f FlatFileManager) ReadRecently(index uint64) (chan []byte, chan uint64, func()) {
	c := make(chan []byte)
	e := make(chan uint64)
	closed := false

	var cancel = func() {
		if !closed {
			close(c)
			closed = true
		}

	}
	go func() {
		fromFile := index / f.fileSizeLimit
		offset := index % f.fileSizeLimit

		for i := int64(fromFile); i >= 0; i-- {
			readInfo, err := f.PasreFile(uint64(i))
			if err != nil {
				e <- 1
				cancel()
			}
			if offset > uint64(len(readInfo)-1) {
				e <- 1
				cancel()
			}

			for j := int64(offset); j >= 0; j-- {
				rawB := make([]byte, readInfo[uint64(j)].size)
				readInfo[uint64(j)].fd.ReadAt(rawB, readInfo[uint64(j)].offset)

			LOOP:
				if !closed {
					select {
					case c <- rawB:
						continue
					default:
						time.Sleep(time.Millisecond * 10)
						goto LOOP
					}

				} else {
					return
				}

			}
		}
		cancel()
	}()

	return c, e, cancel
}

func (f *FlatFileManager) newNextFile() (*os.File, error) {
	i := 0
	if f.currentFD != nil {
		name := filepath.Base(f.currentFD.Name())
		i, _ = strconv.Atoi(name)
		i++
	}
	path := path.Join(f.dataDir, strconv.Itoa(i))
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	f.currentFD = fd
	f.currentFile = uint64(i)
	f.currentFileSize = 0
	f.folderMap[uint64(i)] = true
	return fd, nil
}

//create new file if exceed size limit
func (f *FlatFileManager) update() error {
	size, err := f.checkFileSize()
	if err != nil {
		return err
	}
	if size >= f.fileSizeLimit {
		_, err := f.newNextFile()
		if err != nil {
			return err
		}
	} else {
		f.currentFileSize = size
	}
	return nil
}

func (f *FlatFileManager) checkFileSize() (uint64, error) {
	if f.currentFD == nil {
		return 0, errors.New("Not yet open file!")
	}
	offset := uint64(0)
	size := uint64(0)
	for {
		b := make([]byte, 8)
		n, _ := f.currentFD.ReadAt(b, int64(offset))
		var result uint64
		err := binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &result)
		if err != nil || n == 0 {
			break
		}

		offset += 8
		offset += result

		size++
	}
	return size, nil
}

func (f *FlatFileManager) Append(data []byte) (uint64, error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	//append size-bytes uint64o current FD, if max -> create new file, update currentFD
	var result = make([]byte, 8)
	binary.LittleEndian.PutUint64(result, uint64(len(data)))
	_, err := f.currentFD.Write(result)
	if err != nil {
		return 0, err
	}
	_, err = f.currentFD.Write(data)

	addedItemIndex := f.currentFile*f.fileSizeLimit + f.currentFileSize
	f.currentFileSize++
	if f.currentFileSize >= f.fileSizeLimit {
		f.newNextFile()
	}
	f.itemCache.Add(addedItemIndex, data)
	return addedItemIndex, err
}

func NewFlatFile(dir string, fileBound uint64) (*FlatFileManager, error) {
	cache, _ := lru.New(4)
	itemCache, _ := lru.New(50)
	ff := &FlatFileManager{
		dataDir:       dir,
		fileSizeLimit: fileBound,
		folderMap:     make(map[uint64]bool),
		lock:          new(sync.RWMutex),
		parseCache:    cache,
		itemCache:     itemCache,
	}

	//read all file has number  in folder -> uint64o folderMap, sortedFolder
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	currentFile := -1
	for _, f := range files {
		name := filepath.Base(f.Name())
		i, err := strconv.Atoi(name)
		if err == nil {
			ff.folderMap[uint64(i)] = true
			if currentFile < i {
				currentFile = i
			}
		}
	}

	//open last file currentFD, update currentFileSize, if max -> create new file, update currentFD
	if currentFile > -1 {
		path := path.Join(ff.dataDir, strconv.Itoa(currentFile))
		fd, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
		ff.currentFD = fd
		ff.currentFile = uint64(currentFile)
		//read file size, and create new file in need
		err = ff.update()
		if err != nil {
			return nil, err
		}
	} else {
		_, err := ff.newNextFile()
		if err != nil {
			return nil, err
		}
	}

	return ff, nil
}

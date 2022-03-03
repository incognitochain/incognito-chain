package flatfile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"io/ioutil"
	"path"
	"path/filepath"
	"sync"
	"time"

	"os"
	"sort"
	"strconv"
)

type FlatFile interface {
	//append item into flat file, return item index
	Append([]byte) (int, error)

	//read item in flatfile with specific index (return from append)
	Read(index int) ([]byte, error)

	//read recent data, return data channel, errpr channel, and cancel function
	ReadRecently(index uint64) (dataChan chan []byte, err chan int, cancel func())

	//truncate flat file system
	Truncate(lastIndex int) error
}

type FlatFileManager struct {
	dataDir         string
	fileSizeLimit   int //number of item
	folderMap       map[int]bool
	sortedFolder    []int
	currentFD       *os.File
	currentFile     int
	currentFileSize int
	lock            *sync.RWMutex
	cache           *lru.Cache
}

type ReadInfo struct {
	fd     *os.File
	offset uint64
	size   uint64
}

func (ff *FlatFileManager) Truncate(lastIndex int) error {
	lastFile := lastIndex / ff.fileSizeLimit
	files, err := ioutil.ReadDir(ff.dataDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		name := filepath.Base(f.Name())
		i, err := strconv.Atoi(name)
		if err == nil {
			if i < lastFile {
				err := os.Remove(path.Join(ff.dataDir, name))
				fmt.Println(err)
			}
		}
	}
	newff, err := NewFlatFile(ff.dataDir, ff.fileSizeLimit)
	if err != nil {
		return nil
	}
	*ff = *newff
	return nil
}

func (f FlatFileManager) PasreFile(fileID int) (map[int]ReadInfo, error) {
	//TODO: cache
	f.lock.RLock()
	defer f.lock.RUnlock()
	path := path.Join(f.dataDir, strconv.Itoa(fileID))
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	readInfos := make(map[int]ReadInfo)
	offset := uint64(0)
	size := 0
	for {
		b := make([]byte, 8)
		n, _ := fd.ReadAt(b, int64(offset))
		var result uint64
		err := binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &result)
		if err != nil || n == 0 {
			break
		}

		readInf := ReadInfo{
			fd,
			offset + 8,
			result,
		}
		id := len(readInfos)
		readInfos[id] = readInf
		offset += 8
		offset += result
		size++
	}
	return readInfos, nil
}

func (f FlatFileManager) Read(index int) ([]byte, error) {
	fileID := index / f.fileSizeLimit
	itemFileIndex := index % f.fileSizeLimit

	readInfo, err := f.PasreFile(fileID)
	if err != nil {
		return nil, err
	}
	if itemFileIndex >= len(readInfo) {
		return nil, errors.New(fmt.Sprintf("Cannot read item at index %v", index))
	}

	b := make([]byte, readInfo[itemFileIndex].size)
	_, err = readInfo[itemFileIndex].fd.ReadAt(b, int64(readInfo[itemFileIndex].offset))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (f FlatFileManager) ReadRecently(index uint64) (chan []byte, chan int, func()) {
	c := make(chan []byte)
	e := make(chan int)
	closed := false

	var cancel = func() {
		if !closed {
			close(c)
			closed = true
		}

	}
	go func() {
		fromFile := index / uint64(f.fileSizeLimit)
		offset := index % uint64(f.fileSizeLimit)
		for i := int(fromFile); i >= 0; i-- {
			readInfo, err := f.PasreFile(int(i))
			if err != nil {
				e <- 1
				cancel()
			}
			if offset > uint64(len(readInfo)-1) {
				e <- 1
				cancel()
			}

			for j := int(offset); j >= 0; j-- {
				rawB := make([]byte, readInfo[int(j)].size)
				readInfo[int(j)].fd.ReadAt(rawB, int64(readInfo[int(j)].offset))

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
	f.currentFile = i
	f.currentFileSize = 0
	f.sortedFolder = append(f.sortedFolder, i)
	f.folderMap[i] = true
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

func (f *FlatFileManager) checkFileSize() (int, error) {
	if f.currentFD == nil {
		return 0, errors.New("Not yet open file!")
	}
	offset := uint64(0)
	size := 0
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

func (f *FlatFileManager) Append(data []byte) (int, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	b := bytes.NewBuffer(data)
	//append size-bytes into current FD, if max -> create new file, update currentFD
	var result = make([]byte, 8)
	binary.LittleEndian.PutUint64(result, uint64(b.Len()))
	_, err := f.currentFD.Write(result)
	if err != nil {
		return 0, err
	}
	_, err = f.currentFD.Write(b.Bytes())

	addedItemIndex := f.currentFile*f.fileSizeLimit + f.currentFileSize
	f.currentFileSize++
	if f.currentFileSize >= f.fileSizeLimit {
		f.newNextFile()
	}
	return addedItemIndex, err
}

func NewFlatFile(dir string, fileBound int) (*FlatFileManager, error) {
	ff := &FlatFileManager{
		dataDir:       dir,
		fileSizeLimit: fileBound,
		folderMap:     make(map[int]bool),
		lock:          new(sync.RWMutex),
	}

	//read all file has number  in folder -> into folderMap, sortedFolder
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
			ff.folderMap[i] = true
			ff.sortedFolder = append(ff.sortedFolder, i)
			if currentFile < i {
				currentFile = i
			}
		}
	}
	sort.Slice(ff.sortedFolder, func(i, j int) bool {
		if ff.sortedFolder[i] < ff.sortedFolder[j] {
			return true
		} else {
			return false
		}
	})
	//open last file currentFD, update currentFileSize, if max -> create new file, update currentFD
	if currentFile > -1 {
		path := path.Join(ff.dataDir, strconv.Itoa(currentFile))
		fd, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
		ff.currentFD = fd
		ff.currentFile = currentFile
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

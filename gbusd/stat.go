package gbusd

import (
	"fmt"
	"github.com/evenzhang/gbusd/gbusd/binlog"
	"sync"
	"time"
)

type StoreStatInfo struct {
	StatMap      map[string]interface{}
	storeFreqArr []int64
	blPos        uint32
	blFileName   string
	Counter      uint64
	LastSecCount uint64

	L *sync.RWMutex
}

func NewStoreStatInfo() *StoreStatInfo {
	res := &StoreStatInfo{
		StatMap:      make(map[string]interface{}),
		storeFreqArr: make([]int64, 60, 60),
		blPos:        0,
		blFileName:   "",
		Counter:      0,
		L:            new(sync.RWMutex),
	}

	for i := 0; i < 60; i++ {
		res.storeFreqArr[i] = -1
	}
	return res
}
func (this *StoreStatInfo) OnStoreExecSucc(c binlog.RCmd) {
	this.L.Lock()
	this.blPos = c.BinLogPos
	this.blFileName = c.BinLogFileName
	this.StatMap[fmt.Sprintf("rid::%d", c.ID)] = fmt.Sprintf("%s:%d|%d", c.BinLogFileName, c.BinLogPos, time.Now().Unix())
	this.Counter++
	this.L.Unlock()
}

func (this *StoreStatInfo) StatStoreFreq() (string, bool) {
	this.L.Lock()
	curTime := time.Now()
	curKey := curTime.Second()
	curKeyExist := this.statMapExist(curKey)
	this.storeFreqArr[curKey] = int64(this.Counter)

	resultKey := curTime.Add(-time.Second).Format("15:04:05")
	this.LastSecCount = 0

	if !curKeyExist && this.statMapExist(curKey-1) {
		//前天如果没有
		if !this.statMapExist(curKey - 2) {
			this.LastSecCount = this.statMapVal(curKey - 1)
		} else {
			this.LastSecCount = this.statMapVal(curKey-1) - this.statMapVal(curKey-2)
		}
		this.storeFreqArr[this.key(curKey-2)] = -1

		if this.LastSecCount > 100000 {
			fmt.Println(this.LastSecCount)
		}
		this.L.Unlock()
		return resultKey, true
	}
	this.L.Unlock()
	return resultKey, false
}

func (this *StoreStatInfo) key(sec int) int {
	return (sec + 60) % 60
}

func (this *StoreStatInfo) statMapExist(sec int) bool {
	if this.storeFreqArr[this.key(sec)] >= 0 {
		return true
	}
	return false
}

func (this *StoreStatInfo) statMapVal(sec int) uint64 {
	return uint64(this.storeFreqArr[this.key(sec)])
}
func (this *StoreStatInfo) Clone() StoreStatInfo {
	this.L.RLock()
	res := StoreStatInfo{
		StatMap:      make(map[string]interface{}),
		storeFreqArr: nil,
		blPos:        this.blPos,
		blFileName:   this.blFileName,
		Counter:      this.Counter,
		LastSecCount: this.LastSecCount,
		L:            new(sync.RWMutex),
	}

	for k, v := range this.StatMap {
		res.StatMap[k] = v
	}

	this.L.RUnlock()
	return res
}

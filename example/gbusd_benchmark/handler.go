package main

import (
	"fmt"
	"github.com/evenzhang/gbusd/gbusd/parse"
	"github.com/evenzhang/gbusd/gbusd/store"
	"github.com/evenzhang/gbusd/gserv/config"
	"github.com/evenzhang/gbusd/gserv/debug"
	"github.com/evenzhang/gbusd/gserv/log"
	"github.com/go-redis/redis"
	"time"
)

const (
	ROW_INSERT_DATA     int = 0
	ROW_UPDATA_NEW_DATA int = 1
)

type BenchMarkEventHandler struct {
	parse.BaseEventHandler
	binlogFileName string
	binLogPos      uint32
	enable         bool
	tableName      string
	redisCount     int
	counter        int
}

func (this *BenchMarkEventHandler) OnInit() bool {
	log.Debugln("BaseEventHandler OnInit")
	redisList := config.AppCfg("store", "list")
	fmt.Println(redisList)

	this.tableName = config.AppCfg("master", "table").(string)

	this.enable = true
	this.redisCount = len(config.AppCfgArray("store", "list"))
	this.counter = 0
	return true
}

func (this *BenchMarkEventHandler) OnUpdate(sender parse.ISender, ev *parse.Event) (bool, error) {
	//return this.process(sender, ev, ROW_UPDATA_NEW_DATA)
	if len(ev.Data.Rows) > 1 {
		fmt.Println("Update", ev, len(ev.Data.Rows))
	}
	return true, nil
}

func (this *BenchMarkEventHandler) OnInsert(sender parse.ISender, ev *parse.Event) (bool, error) {
	//return this.process(sender, ev, ROW_INSERT_DATA)
	fmt.Println("Insert", ev, len(ev.Data.Rows))
	return true, nil
}

func (this *BenchMarkEventHandler) process(sender parse.ISender, ev *parse.Event, rowIndex int) (bool, error) {

	if ev.BinLogFileName == this.binlogFileName && ev.BinLogPos >= this.binLogPos {
		this.enable = false
	}

	if ev.TableName != this.tableName {
		fmt.Println("process", ev.TableName, this.tableName)
		return true, nil
	}
	log.Infof("DBSTAT|U|%s|%s|%s|%d", time.Now().Format("2006-01-02 15:04:05"), ev.TableName, ev.BinLogFileName, ev.BinLogPos)

	id := ev.Data.Rows[rowIndex][0].(int32)

	if sender.(*store.EventStore) != nil {
		cmd := redis.NewBoolCmd("hset", "bench-mark", "redis", 0)
		sender.Push(ev, int(id)%this.redisCount, cmd)
	}
	this.counter++

	return true, nil
}

func (this *BenchMarkEventHandler) Enable() bool {
	log.Debugln("BenchMarkEventHandler Enable")
	return this.enable
}

func NewEventHandler(binlogFileName string, binLogPos uint32) parse.IEventHandler {
	debug.Println("NewEventHandler")
	return &BenchMarkEventHandler{binlogFileName: binlogFileName, binLogPos: binLogPos}
}

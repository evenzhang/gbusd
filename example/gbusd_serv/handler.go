package main

import (
	"fmt"
	"github.com/evenzhang/gbusd/gbusd/parse"
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

type EventHandler struct {
	parse.BaseEventHandler
	sha        string
	redisCount int64
}

func (this *EventHandler) OnInit() bool {
	log.Debugln("BaseEventHandler OnInit")
	redisList := config.AppCfg("store", "list")
	fmt.Println(redisList)
	this.redisCount = int64(len(redisList.([]interface{})))

	return true
}
func (this *EventHandler) OnUpdate(sender parse.ISender, ev *parse.Event) (bool, error) {
	return this.process(sender, ev, ROW_UPDATA_NEW_DATA)
}

func (this *EventHandler) OnInsert(sender parse.ISender, ev *parse.Event) (bool, error) {
	return this.process(sender, ev, ROW_INSERT_DATA)
}

func (this *EventHandler) process(sender parse.ISender, ev *parse.Event, rowIndex int) (bool, error) {
	if ev.TableName != "test" {
		return true, nil
	}

	id := ev.Data.Rows[rowIndex][0].(int32)
	dst_redis := int(id) % int(this.redisCount)

	cmd := redis.NewBoolCmd("hset", "gbusd_serv_test", id, time.Now().Format("2006-01-02 15:04:05"))
	sender.Push(ev, dst_redis, cmd)

	log.Debugln("STAT |", time.Now().Format("2006-01-02 15:04:05"), "|", id)
	return true, nil
}

func NewEventHandler() parse.IEventHandler {
	debug.Println("NewEventHandler")
	return &EventHandler{}
}

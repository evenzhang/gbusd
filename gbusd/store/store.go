package store

import (
	"fmt"
	"github.com/evenzhang/gbusd/gbusd/parse"
	"github.com/evenzhang/gbusd/gserv/comm"
	"github.com/evenzhang/gbusd/gserv/config"
	"github.com/evenzhang/gbusd/gserv/daemon"
	"github.com/evenzhang/gbusd/gserv/debug"
	"github.com/evenzhang/gbusd/gserv/log"
	"github.com/go-redis/redis"
	"sync"
	"time"
)

const SENDER_REDIS_MAX int = 50

type EventStore struct {
	ClientNum int
	RedisList []*redis.Client

	metaServer *MetaServer

	EventSender *parse.ISender
	CmdChanList []chan parse.RCmd

	statMap map[string]interface{}
	lockMap *sync.RWMutex
}

func NewEventStore(chanLen int) (store *EventStore, err error) {
	debug.Println("NewEventStore")

	metaServer, err := NewMetaServer()
	comm.AssertErr(err, "NewEventStore-NewMetaServer error ")

	cfg := config.AppCfgArray("store", "list")
	store = &EventStore{ClientNum: len(cfg),
		RedisList: make([]*redis.Client, SENDER_REDIS_MAX), lockMap: new(sync.RWMutex), metaServer: metaServer, statMap: make(map[string]interface{})}

	for i, v := range cfg {
		client, err := comm.NewRedisClient(v.(string))
		if err != nil {
			return nil, err
		}
		store.RedisList[i] = client
	}

	store.initChanList(len(store.RedisList), chanLen)
	return store, nil
}

func (this *EventStore) Exec(c parse.RCmd) error {
	if c.ID < 0 || c.ID >= SENDER_REDIS_MAX {
		log.Fatalf("EventStore Exec error: redis index range error (0 ~ %d),id=%d", SENDER_REDIS_MAX, c.ID)
	}

	if this.RedisList[c.ID] == nil {
		log.Fatalf("EventStore Exec error: Redislist[%d] == nil", c.ID)
	}

	err := this.RedisList[c.ID].Process(c.Cmd)
	log.Debugln("EventStore Exec: id=", c.ID, c, err)
	if err == nil {
		this.lockMap.Lock()
		this.statMap[fmt.Sprintf("rid::%d", c.ID)] = c
		this.lockMap.Unlock()
	}
	return err
}

func (this *EventStore) Push(ev *parse.Event, id int, cmder redis.Cmder) {
	cmd := parse.RCmd{
		BinLogFileName: ev.BinLogFileName,
		BinLogPos:      ev.BinLogPos,
		ID:             id,
		Cmd:            cmder,
	}
	this.CmdChanList[id] <- cmd
}

func (this *EventStore) initChanList(chanNum int, chanLen int) {
	chanList := make([]chan parse.RCmd, chanNum, chanNum)
	for i := 0; i < chanNum; i++ {
		chanList[i] = make(chan parse.RCmd, chanLen)
	}
	this.CmdChanList = chanList
}

func (this *EventStore) metaSavePos(blFileName string, blPos uint32, count uint64) bool {
	//默认重试30秒
	retryTimes := 60
	for retryTimes > 0 {
		retryTimes--
		err := this.metaServer.Save(blFileName, blPos, count)
		if err != nil {
			time.Sleep(time.Millisecond * 500)
			log.Debugln("metaSavePos-Retry ", retryTimes, blFileName, blPos)
			continue
		}
		return true
	}
	return false
}

func (this *EventStore) storeExec(c parse.RCmd) bool {
	//默认重试30秒
	retryTimes := 60
	for retryTimes > 0 {
		retryTimes--
		err := this.Exec(c)
		if err != nil {
			time.Sleep(time.Millisecond * 500)
			log.Debugln("storeExec-Retry ", retryTimes, c)
			continue
		}
		return true
	}
	return false
}

func (this *EventStore) StatLoop() {
	for {
		this.lockMap.RLock()
		var minPos *parse.RCmd = nil
		tmpMap := make(map[string]interface{})
		for k, v := range this.statMap {
			binlogInfo := v.(parse.RCmd)
			tmpMap[k] = fmt.Sprintf("%s:%d|%d", binlogInfo.BinLogFileName, binlogInfo.BinLogPos, time.Now().Unix())
			if minPos == nil {
				minPos = &binlogInfo
				continue
			}

			if minPos.BinLogFileName >= binlogInfo.BinLogFileName && minPos.BinLogPos > binlogInfo.BinLogPos {
				minPos = &binlogInfo
			}
		}
		this.lockMap.RUnlock()

		if minPos != nil {
			this.metaSavePos(minPos.BinLogFileName, minPos.BinLogPos, 0)
		}

		this.metaServer.SaveStatInfo(tmpMap)
		time.Sleep(time.Second)
	}
}
func (this *EventStore) ExecLoop() {
	this.ExecLoopFunc(nil, nil)
}
func (this *EventStore) ExecLoopFunc(onSucc func(c parse.RCmd), onError func()) {
	for i := 0; i < len(this.CmdChanList); i++ {
		go func(index int) {
			log.Infoln("Init storeLoop()")
			for {
				c := <-this.CmdChanList[index]
				log.Debugln("storeLoop", c)
				if this.storeExec(c) {
					if onSucc != nil {
						onSucc(c)
					}
				} else {
					if onError != nil {
						onError()
					} else {
						//系统下游消费失败，启动应急处理策略
						log.Errorf("storeExec error")
						daemon.Restart(true)
					}
				}
			}
		}(i)
	}
}

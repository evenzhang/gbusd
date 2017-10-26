package gbusd

import (
	"flag"
	"fmt"
	"github.com/evenzhang/gbusd/gbusd/api"
	"github.com/evenzhang/gbusd/gbusd/binlog"
	"github.com/evenzhang/gbusd/gbusd/res"
	"github.com/evenzhang/gbusd/gbusd/store"
	"github.com/evenzhang/gbusd/gserv/application"
	"github.com/evenzhang/gbusd/gserv/comm"
	"github.com/evenzhang/gbusd/gserv/daemon"
	"github.com/evenzhang/gbusd/gserv/debug"
	"github.com/evenzhang/gbusd/gserv/log"
	"os"
	"time"
)

const SENDER_CHAN_MAX int = 100000

type BusdDaemon struct {
	application.BaseDaemonServer

	master       *binlog.MasterDB
	parser       *binlog.BinlogParser
	meta         *binlog.MetaServer
	electionLog  *binlog.ElectionLogServer
	eventHandler binlog.IEventHandler
	cmdChan      chan binlog.RCmd
	store        *store.EventStore
	storeStat    *StoreStatInfo
	cliMeta      bool
}

func NewDaemonServer(handler binlog.IEventHandler) *BusdDaemon {
	return &BusdDaemon{eventHandler: handler}
}

func (this *BusdDaemon) ParserFlag() {
	flag.BoolVar(&this.cliMeta, "cli", false, "Enable Cli mode.")
}

func (this *BusdDaemon) OnInit() {
	debug.Println("qBusd", "OnInit")
	res.SetRuntimeVar("Pid", os.Getpid())

	var err error
	this.master, err = binlog.NewMaster()
	comm.AssertErr(err)

	this.cmdChan = make(chan binlog.RCmd, SENDER_CHAN_MAX)
	this.parser = binlog.NewBinlogParser(this.cmdChan)

	this.meta, err = binlog.NewMetaServer()
	comm.AssertErr(err)

	this.electionLog, err = binlog.NewElectionLogServer()
	comm.AssertErr(err)

	this.store, _ = store.NewEventStore(this.cmdChan)

	this.storeStat = NewStoreStatInfo()

	if this.cliMeta {
		this.enterCliMode()
		os.Exit(0)
	}
}

func (this *BusdDaemon) RegisterAllApi() {
	api.RegisterAllApi()
}

func (this *BusdDaemon) addElectionLog() {
	binlogInfo := ""
	data, err := this.meta.Get()
	if err == nil {
		binlogInfo = fmt.Sprintf("%s:%d", data.BinLogFileName, data.BinLogPos)
	}
	this.electionLog.Add(binlogInfo)
}

func (this *BusdDaemon) OnRun() {
	log.Info("Busd.OnRun start")
	debug.Println("Busd.OnRun start")
	res.SetRuntimeVar("ElectionStatus", "master")
	res.SetRuntimeVarNow("ElectionTime")
	res.SetRuntimeVar("MinCount", 0)

	this.addElectionLog()
	go this.storeLoop()
	go this.storeStatLoop()

	this.startReplication()
}

func (this *BusdDaemon) startReplication() {
	log.Infoln("Init startReplication()")
	exist, err := this.meta.IsExist()
	if err != nil {
		log.Fatalln("Server OnRun Error:", err)
	}
	log.Infoln("Busd.startReplication,meta exist=", exist)

	if exist {
		data, err := this.meta.Get()
		if err != nil {
			log.Fatalln("Server OnRun Error:", err)
			return
		}

		this.parser.StartReplication(this.eventHandler, data.BinLogFileName, data.BinLogPos)
	} else {
		log.Fatalf("startReplication error: meta binlog position info not found")
	}
}
func (this *BusdDaemon) storeStatLoop() {
	log.Infoln("Init storeStatLoop()")
	var oldCount uint64 = 0
	for {
		key, needSave := this.storeStat.StatStoreFreq()
		statData := this.storeStat.Clone()

		if oldCount < statData.Counter {
			if !this.metaSavePos(statData.blFileName, statData.blPos, statData.Counter) {
				//存储位置点信息，重试后最终失败，启动应急策略
				log.Errorf("metaSavePos error")
				daemon.Restart(true)
			}
			oldCount = statData.Counter
		}

		this.meta.SaveStatInfo(statData.StatMap)

		if needSave {
			this.meta.AddTpsInfo(key, statData.LastSecCount, statData.Counter)
			res.SetRuntimeVar("SecCount", statData.LastSecCount)
			debug.Println(statData)
		}

		time.Sleep(time.Second)
	}
}
func (this *BusdDaemon) storeLoop() {
	log.Infoln("Init storeLoop()")
	for {
		c := <-this.cmdChan
		log.Debugln("storeLoop", c)
		if this.storeExec(c) {
			log.Debugf("REDISSTAT | %s | %s | %d ", time.Now().Format("2006-01-02 15:04:05"), c.BinLogFileName, c.BinLogPos)
			this.storeStat.OnStoreExecSucc(c)
		} else {
			//系统下游消费失败，启动应急处理策略
			log.Errorf("storeExec error")
			daemon.Restart(true)
		}
	}
}

func (this *BusdDaemon) storeExec(c binlog.RCmd) bool {
	//默认重试30秒
	retryTimes := 60
	for retryTimes > 0 {
		retryTimes--
		err := this.store.Exec(c)
		if err != nil {
			time.Sleep(time.Millisecond * 500)
			log.Debugln("storeExec-Retry ", retryTimes, c)
			continue
		}
		return true
	}
	return false
}

func (this *BusdDaemon) metaSavePos(blFileName string, blPos uint32, count uint64) bool {
	//默认重试30秒
	retryTimes := 60
	for retryTimes > 0 {
		retryTimes--
		err := this.meta.Save(blFileName, blPos, count)
		if err != nil {
			time.Sleep(time.Millisecond * 500)
			log.Debugln("metaSavePos-Retry ", retryTimes, blFileName, blPos)
			continue
		}
		return true
	}
	return false
}

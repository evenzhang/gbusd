package gbusd

import (
	"fmt"
	"github.com/evenzhang/gbusd/gbusd/parse"
	"github.com/evenzhang/gbusd/gbusd/store"
	"github.com/evenzhang/gbusd/gserv/application"
	"github.com/evenzhang/gbusd/gserv/comm"
	"github.com/evenzhang/gbusd/gserv/debug"
	"github.com/evenzhang/gbusd/gserv/log"
	"os"
)

const SENDER_CHAN_MAX int = 100000

type Busd struct {
	application.BaseDaemonServer

	parser       *parse.BinlogParser
	eventHandler parse.IEventHandler

	master      *store.MasterDB
	meta        *store.MetaServer
	electionLog *store.ElectionLogServer
	store       *store.EventStore
	cliMeta     bool
}

func NewDaemonServer(handler parse.IEventHandler) *Busd {
	return &Busd{eventHandler: handler}
}

func (this *Busd) OnInit() {
	debug.Println("qBusd", "OnInit")
	SetRuntimeVar("Pid", os.Getpid())

	var err error
	this.master, err = store.NewMaster()
	comm.AssertErr(err)

	this.store, _ = store.NewEventStore(SENDER_CHAN_MAX)
	this.parser = parse.NewBinlogParser(this.store)

	this.meta, err = store.NewMetaServer()
	comm.AssertErr(err)

	this.electionLog, err = store.NewElectionLogServer()
	comm.AssertErr(err)

	if this.cliMeta {
		this.enterCliMode()
		os.Exit(0)
	}
}

func (this *Busd) addElectionLog() {
	binlogInfo := ""
	data, err := this.meta.Get()
	if err == nil {
		binlogInfo = fmt.Sprintf("%s:%d", data.BinLogFileName, data.BinLogPos)
	}
	this.electionLog.Add(binlogInfo)
}

func (this *Busd) OnRun() {
	log.Info("Busd.OnRun start")
	debug.Println("Busd.OnRun start")
	SetRuntimeVar("ElectionStatus", "master")
	SetRuntimeVarNow("ElectionTime")
	SetRuntimeVar("MinCount", 0)

	this.addElectionLog()

	go this.store.ExecLoop()
	go this.store.StatLoop()

	this.startReplication()
}

func (this *Busd) startReplication() {
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
		log.Fatalf("startReplication error: meta parse position info not found")
	}
}

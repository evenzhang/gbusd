package parse

import (
	"github.com/evenzhang/gbusd/gserv/config"
	"github.com/evenzhang/gbusd/gserv/log"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"golang.org/x/net/context"
	"sync"
)

var (
	lock = new(sync.RWMutex)
)

type BinlogParser struct {
	Host           string
	Port           uint16
	UserName       string
	Password       string
	ServerID       uint32
	DbName         string
	TableName      string
	binLogFileName string

	Sender ISender
	//Opts                  Options
	syncer                *replication.BinlogSyncer
	ctx                   context.Context
	cancelReplicationJobs context.CancelFunc
}

func NewBinlogParser(sender ISender) *BinlogParser {
	binlogParser := &BinlogParser{
		Host:      config.AppCfg("master", "host").(string),
		Port:      uint16(config.AppCfg("master", "port").(int)),
		Password:  config.AppCfg("master", "password").(string),
		UserName:  config.AppCfg("master", "user").(string),
		ServerID:  uint32(config.AppCfg("master", "serverid").(int)),
		DbName:    config.AppCfg("master", "dbname").(string),
		TableName: config.AppCfg("master", "table").(string)}

	binlogParser.Sender = sender
	return binlogParser
}
func (this *BinlogParser) initSyncer(binLogFileName string, binLogPos uint32) *replication.BinlogStreamer {
	// Create a parse syncer with a unique busd id, the busd id must be different from other MySQL's.
	// flavor is mysql or mariadb
	cfg := replication.BinlogSyncerConfig{
		ServerID: this.ServerID,
		Flavor:   "mysql",
		Host:     this.Host,
		Port:     this.Port,
		User:     this.UserName,
		Password: this.Password,
	}

	this.syncer = replication.NewBinlogSyncer(cfg)
	log.Debugln("BinlogParser New NewBinlogSyncer succ")
	this.binLogFileName = binLogFileName
	// Start sync with sepcified parse file and position
	streamer, _ := this.syncer.StartSync(mysql.Position{this.binLogFileName, binLogPos})
	log.Debugln("BinlogParser New StartSync succ")
	log.Infoln("BinlogParser Init succ")
	// or you can start a gtid replication like
	// streamer, _ := syncer.StartSyncGTID(gtidSet)
	// the mysql GTID set likes this "de278ad0-2106-11e4-9f8e-6edd0ca20947:1-2"
	// the mariadb GTID set likes this "0-1-100"

	this.ctx, this.cancelReplicationJobs = context.WithCancel(context.Background())

	return streamer
}

func (this *BinlogParser) StartReplication(handler IEventHandler, binLogFileName string, binLogPos uint32) error {
	streamer := this.initSyncer(binLogFileName, binLogPos)
	handler.OnInit()
	this.ProcessEvent(streamer, handler)
	handler.OnClose()
	return nil
}

func (this *BinlogParser) ProcessEvent(streamer *replication.BinlogStreamer, eventHandler IEventHandler) {
	for eventHandler.Enable() {
		ev, err := streamer.GetEvent(this.ctx)

		if err != nil {
			log.Errorln("BinlogParser.ProcessEvent", err)
			break
		}
		switch ev.Header.EventType {
		//parse 文件切换
		case replication.ROTATE_EVENT:
			rotateEvent := ev.Event.(*replication.RotateEvent)
			this.binLogFileName = string(rotateEvent.NextLogName)
			log.Infoln("rotateEvent:", string(rotateEvent.NextLogName))
			//Row式存储格式解析[QUERY_EVENT(Begin)  TABLE_MAP_EVENT  WRITE|UPDATE|DEL  XID_EVENT(Commit)]
		case replication.WRITE_ROWS_EVENTv2:
			fallthrough
		case replication.WRITE_ROWS_EVENTv1:
			event := this.createRowsEvent("INSERT", ev)
			eventHandler.OnInsert(this.Sender, event)

		case replication.UPDATE_ROWS_EVENTv2:
			fallthrough
		case replication.UPDATE_ROWS_EVENTv1:
			event := this.createRowsEvent("UPDATE", ev)
			eventHandler.OnUpdate(this.Sender, event)

		case replication.DELETE_ROWS_EVENTv2:
			fallthrough
		case replication.DELETE_ROWS_EVENTv1:
			event := this.createRowsEvent("DELETE", ev)
			eventHandler.OnDelete(this.Sender, event)
		}
	}
}
func (this *BinlogParser) createRowsEvent(action string, ev *replication.BinlogEvent) *Event {
	rev := ev.Event.(*replication.RowsEvent)
	dbName := string(rev.Table.Schema)
	tableName := string(rev.Table.Table)
	return &Event{Data: rev, Header: ev.Header, BinLogPos: ev.Header.LogPos, BinLogFileName: this.binLogFileName, DBName: dbName, TableName: tableName, Type: action}
}

func (this *BinlogParser) Close() {
	this.cancelReplicationJobs()
	this.syncer.Close()
}

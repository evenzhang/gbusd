package binlog

import (
	"github.com/go-redis/redis"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"github.com/evenzhang/gbusd/gserv/config"
	"github.com/evenzhang/gbusd/gserv/debug"
	"github.com/evenzhang/gbusd/gserv/log"
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
	TableID        uint64
	BinLogPos      uint32
	BinLogFileName string

	Sender *Sender
	//Opts                  Options
	syncer                *replication.BinlogSyncer
	ctx                   context.Context
	cancelReplicationJobs context.CancelFunc
}

type IEventHandler interface {
	OnInit() bool
	OnInsert(Sender *Sender, event *Event) (bool, error)
	OnUpdate(Sender *Sender, event *Event) (bool, error)
	OnDelete(Sender *Sender, event *Event) (bool, error)
	OnClose()
	Enable() bool
}

type Sender struct {
	CmdChan chan RCmd
}

func (this *Sender) Push(ev *Event, id int, cmder redis.Cmder) {
	cmd := RCmd{
		BinLogFileName: ev.BinLogFileName,
		BinLogPos:      ev.BinLogPos,
		ID:             id,
		Cmd:            cmder,
	}
	this.CmdChan <- cmd
}

type Event struct {
	DBName         string
	TableName      string
	Type           string
	BinLogFileName string
	BinLogPos      uint32
	Data           *replication.RowsEvent
	Header         *replication.EventHeader
}

type RCmd struct {
	BinLogFileName string
	BinLogPos      uint32
	ID             int
	Cmd            redis.Cmder
}

func (this *Event) DebugPrint() {
	debug.Printf("BEvent-DebugPrint: DBName %s|TableName %s|Type %s|TableID %d|Flags %d|ColumnCount %d\n",
		this.DBName, this.TableName, this.Type, this.Data.TableID, this.Data.Flags, this.Data.ColumnCount)
	debug.Println("EV-EVENT:", *this.Data)
}

func NewBinlogParser(cmdChan chan RCmd) *BinlogParser {
	binlogParser := &BinlogParser{
		Host:      config.AppCfg("master", "host").(string),
		Port:      uint16(config.AppCfg("master", "port").(int)),
		Password:  config.AppCfg("master", "password").(string),
		UserName:  config.AppCfg("master", "user").(string),
		ServerID:  uint32(config.AppCfg("master", "serverid").(int)),
		DbName:    config.AppCfg("master", "dbname").(string),
		TableName: config.AppCfg("master", "table").(string)}

	binlogParser.Sender = &Sender{CmdChan: cmdChan}
	return binlogParser
}
func (this *BinlogParser) InitSyncer() *replication.BinlogStreamer {
	// Create a binlog syncer with a unique busd id, the busd id must be different from other MySQL's.
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
	// Start sync with sepcified binlog file and position
	streamer, _ := this.syncer.StartSync(mysql.Position{this.BinLogFileName, this.BinLogPos})
	log.Debugln("BinlogParser New StartSync succ")
	log.Infoln("BinlogParser Init succ")
	// or you can start a gtid replication like
	// streamer, _ := syncer.StartSyncGTID(gtidSet)
	// the mysql GTID set likes this "de278ad0-2106-11e4-9f8e-6edd0ca20947:1-2"
	// the mariadb GTID set likes this "0-1-100"

	this.ctx, this.cancelReplicationJobs = context.WithCancel(context.Background())

	return streamer
}

func (this *BinlogParser) ProcessEvent(streamer *replication.BinlogStreamer, eventHandler IEventHandler) {
	var preEventType = replication.UNKNOWN_EVENT
	var preEventObj *replication.BinlogEvent

	log.Debug("ProcessEvent Loop")

	for eventHandler.Enable() {
		ev, err := streamer.GetEvent(this.ctx)

		if err != nil {
			log.Errorln("BinlogParser.ProcessEvent", err)
			break
		}
		log.Debugln("ProcessEvent.GetEvent", ev.Header.EventType)

		log.Debug(ev)

		switch ev.Header.EventType {
		//binlog 文件切换
		case replication.ROTATE_EVENT:
			rotateEvent := ev.Event.(*replication.RotateEvent)
			this.BinLogFileName = string(rotateEvent.NextLogName)
			this.BinLogPos = ev.Header.LogPos
			log.Infoln("rotateEvent:", string(rotateEvent.NextLogName))
			//Row式存储格式解析[QUERY_EVENT(Begin)  TABLE_MAP_EVENT  WRITE|UPDATE|DEL  XID_EVENT(Commit)]
		case replication.QUERY_EVENT:
			this.BinLogPos = ev.Header.LogPos
		case replication.TABLE_MAP_EVENT:
			//获取并记录TableID值
			tableEvent := ev.Event.(*replication.TableMapEvent)
			//TableID可能会变化
			if string(tableEvent.Schema) == this.DbName && string(tableEvent.Table) == this.TableName {
				this.TableID = tableEvent.TableID
			}
		case replication.XID_EVENT:
			this.BinLogPos = ev.Header.LogPos
			log.Debugf("ProcessEvent replication.XID_EVENT")
			this.processXIDEvent(eventHandler, preEventType, preEventObj)
		}

		preEventType = ev.Header.EventType
		preEventObj = ev
	}

	this.Close()
}

func (this *BinlogParser) StartReplication(handler IEventHandler, binLogFileName string, binLogPos uint32) error {
	streamer := this.InitSyncer()
	handler.OnInit()
	this.ProcessEvent(streamer, handler)
	handler.OnClose()
	return nil
}

func (this *BinlogParser) processXIDEvent(eventHandler IEventHandler, preEventType replication.EventType,
	preEventObj *replication.BinlogEvent) {

	rev := preEventObj.Event.(*replication.RowsEvent)
	dbName := string(rev.Table.Schema)
	tableName := string(rev.Table.Table)

	event := &Event{Data: rev, Header: preEventObj.Header, BinLogPos: this.BinLogPos, BinLogFileName: this.BinLogFileName, DBName: dbName, TableName: tableName}

	switch preEventType {
	case replication.WRITE_ROWS_EVENTv2:
		fallthrough
	case replication.WRITE_ROWS_EVENTv1:
		event.Type = "INSERT"
		log.Debugf("processXIDEvent INSERT")
		eventHandler.OnInsert(this.Sender, event)

	case replication.UPDATE_ROWS_EVENTv2:
		fallthrough
	case replication.UPDATE_ROWS_EVENTv1:
		event.Type = "UPDATE"
		log.Debugf("processXIDEvent UPDATE")
		eventHandler.OnUpdate(this.Sender, event)

	case replication.DELETE_ROWS_EVENTv2:
		fallthrough
	case replication.DELETE_ROWS_EVENTv1:
		event.Type = "DELETE"
		log.Debugf("processXIDEvent DELETE")
		eventHandler.OnDelete(this.Sender, event)
	}
}

func (this *BinlogParser) Close() {
	this.cancelReplicationJobs()
	this.syncer.Close()
}

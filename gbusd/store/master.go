package store

import (
	"fmt"
	"github.com/evenzhang/gbusd/gserv/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"sync"
)

type LastPosInfo struct {
	FileName string `xorm:"File"`
	Pos      uint32 `xorm:"Position"`
	LogDoDB  string `xorm:"Binlog_Do_DB"`
	IgnoreDB string `xorm:"Binlog_Ignore_DB"`
}

type FirstPosInfo struct {
	FileName  string `xorm:"Log_name"`
	StartPos  uint32 `xorm:"Pos"`
	EventType string `xorm:"Event_type"`
	ServerID  string `xorm:"Server_id"`
	EndPos    uint32 `xorm:"End_log_pos"`
	Info      string `xorm:"Info"`
}

type MasterDB struct {
	l      sync.RWMutex
	engine *xorm.Engine

	dsn      string
	Table    string
	ServerID uint32
}

func (this *MasterDB) GetPos() (*LastPosInfo, error) {
	lastPosInfo := &LastPosInfo{}
	this.l.RLock()
	this.engine.SQL("show master status").Get(lastPosInfo)
	this.l.RUnlock()
	return lastPosInfo, nil
}

func (this *MasterDB) GetFirstPos() (*FirstPosInfo, error) {
	firstPosInfo := &FirstPosInfo{}
	this.l.Lock()
	this.engine.SQL("show binlog events limit 1").Get(firstPosInfo)
	this.l.Unlock()

	return firstPosInfo, nil
}

func GetMasterDsn() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8",
		config.AppCfg("master", "user").(string),
		config.AppCfg("master", "password").(string),
		config.AppCfg("master", "host").(string),
		config.AppCfg("master", "port").(int),
		config.AppCfg("master", "dbname").(string))

	fmt.Println(dsn)
	return dsn
}
func NewMaster() (*MasterDB, error) {
	dsn := GetMasterDsn()
	return NewMasterByDsn(dsn)
}

func NewMasterByDsn(dsn string) (*MasterDB, error) {
	master := MasterDB{}

	master.l.Lock()
	defer master.l.Unlock()
	var err error
	if master.engine == nil {
		master.engine, err = xorm.NewEngine("mysql", dsn)
	}

	return &master, err
}

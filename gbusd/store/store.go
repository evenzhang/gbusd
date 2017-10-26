package store

import (
	"github.com/go-redis/redis"
	"github.com/evenzhang/gbusd/gbusd/binlog"
	"github.com/evenzhang/gbusd/gserv/comm"
	"github.com/evenzhang/gbusd/gserv/config"
	"github.com/evenzhang/gbusd/gserv/debug"
	"github.com/evenzhang/gbusd/gserv/log"
)

const SENDER_REDIS_MAX int = 50

type EventStore struct {
	ClientNum int
	RedisList []*redis.Client
	CmdChan   chan binlog.RCmd
}

func GetStoreConfig() []interface{} {
	cfg := config.AppCfg("store", "list").([]interface{})
	return cfg
}

func NewEventStore(cmdChan chan binlog.RCmd) (store *EventStore, err error) {

	debug.Println("NewEventStore")

	cfg := GetStoreConfig()
	store = &EventStore{ClientNum: len(cfg),
		RedisList: make([]*redis.Client, SENDER_REDIS_MAX), CmdChan: cmdChan}

	for i, v := range cfg {
		client, err := comm.NewRedisClient(v.(string))
		if err != nil {
			return nil, err
		}
		store.RedisList[i] = client
	}
	return store, nil
}

func (this *EventStore) Exec(c binlog.RCmd) error {
	if c.ID < 0 || c.ID >= SENDER_REDIS_MAX {
		log.Fatalf("EventStore Exec error: redis index range error (0 ~ %d),id=%d", SENDER_REDIS_MAX, c.ID)
	}

	if this.RedisList[c.ID] == nil {
		log.Fatalf("EventStore Exec error: Redislist[%d] == nil", c.ID)
	}

	err := this.RedisList[c.ID].Process(c.Cmd)
	log.Debugln("EventStore Exec: id=", c.ID, c, err)
	if err == nil {

	}
	return err
}

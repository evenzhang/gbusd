package store

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/evenzhang/gbusd/gserv/config"
	"github.com/evenzhang/gbusd/gserv/log"

	"github.com/evenzhang/gbusd/gserv/comm"
	"strings"
	"time"
)

const ELECTION_LOG_MAX int64 = 8

type ElectionLog struct {
	ServerName       string
	ElectionTime     string
	ServerAddr       string
	ServerListenInfo string
	BinlogInfo       string
}

type ElectionLogServer struct {
	client     *redis.Client
	serverName string
	serverHost string
	key        string
}

func NewElectionLogServerByAddr(addr string, serverName string, serverHost string) (*ElectionLogServer, error) {
	client, err := comm.NewRedisClient(addr)
	key := fmt.Sprintf("bus-serv::%s::elect-log::list", serverName)
	server := ElectionLogServer{client: client, serverName: serverName, serverHost: serverHost, key: key}
	return &server, err
}

func NewElectionLogServer() (*ElectionLogServer, error) {
	return NewElectionLogServerByAddr(
		config.AppCfg("meta", "addr").(string),
		config.Config().Common.ServerName,
		config.Config().Common.Listen,
	)
}

func (this *ElectionLogServer) Add(msg string) error {
	data := fmt.Sprintf("%s|%s|%s|%s|%s", this.serverName, time.Now().Format("2006-01-02 15:04:05"), comm.GetLocalAddr(), this.serverHost, msg)
	log.Debugln("MetaServer-Save Pos", this.key, data)
	this.client.LTrim(this.key, 0, ELECTION_LOG_MAX)
	return this.client.LPush(this.key, data).Err()
}

func (this *ElectionLogServer) Get() ([]ElectionLog, error) {
	logs, err := this.client.LRange(this.key, 0, ELECTION_LOG_MAX).Result()
	if err != nil {
		return nil, err
	}

	res := make([]ElectionLog, 0, len(logs))
	for _, log := range logs {
		items := strings.Split(log, "|")
		if len(items) >= 5 {
			res = append(res, ElectionLog{items[0], items[1], items[2], items[3], items[4]})
		}
	}

	return res, nil
}
func (this *ElectionLogServer) Close() {
	if this.client != nil {
		this.client.Close()
	}
}

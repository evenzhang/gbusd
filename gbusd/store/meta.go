package store

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/juju/errors"
	"github.com/evenzhang/gbusd/gserv/comm"
	"github.com/evenzhang/gbusd/gserv/config"
	"github.com/evenzhang/gbusd/gserv/log"
	"strconv"
	"strings"
	"time"
)

const META_FIELD_NAME string = "binlog_filepos"
const META_TPS_STAT_MAX int64 = 30

type MetaData struct {
	BinLogFileName string
	BinLogPos      uint32
	Count          uint64
	UpdateTime     time.Time
	BinLogTime     time.Time
}

type MetaServer struct {
	client *redis.Client
	key    string
	tpsKey string
}

func NewMetaServerByAddr(addr string, serverName string) (*MetaServer, error) {
	client, err := comm.NewRedisClient(addr)
	key := fmt.Sprintf("bus-serv::%s::meta::hash", serverName)
	tpsKey := fmt.Sprintf("bus-serv::%s::meta-tps::list", serverName)
	server := MetaServer{client: client, key: key, tpsKey: tpsKey}
	return &server, err
}

func NewMetaServer() (*MetaServer, error) {
	return NewMetaServerByAddr(
		config.AppCfg("meta", "addr").(string),
		config.Config().Common.ServerName)
}

func (this *MetaServer) Save(binLogFileName string, binLogPos uint32, count uint64) error {
	data := fmt.Sprintf("%s %d %d %d", binLogFileName, binLogPos, count, time.Now().Unix())
	log.Debugln("MetaServer-Save Pos", this.key, META_FIELD_NAME, data)
	return this.client.HSet(this.key, META_FIELD_NAME, data).Err()
}
func (this *MetaServer) Get() (*MetaData, error) {
	logInfoStr, err := this.client.HGet(this.key, META_FIELD_NAME).Result()

	if err != nil {
		return nil, err
	}

	var updateTimeUnix int64 = 0
	data := MetaData{}
	fmt.Sscanf(logInfoStr, "%s %d %d %d", &data.BinLogFileName, &data.BinLogPos, &data.Count, &updateTimeUnix)
	data.UpdateTime = time.Unix(updateTimeUnix, 0)
	return &data, nil
}
func (this *MetaServer) SaveStatInfo(fields map[string]interface{}) (string, error) {
	return this.client.HMSet(this.key, fields).Result()
}

func (this *MetaServer) GetStatInfo(id int) (string, time.Time, error) {
	stat, err := this.client.HGet(this.key, fmt.Sprintf("rid::%d", id)).Result()
	if err != nil {
		return "", time.Now(), err
	}

	items := strings.Split(stat, "|")
	if len(items) < 2 {
		return "", time.Now(), errors.New("stat info format error : " + stat)
	}

	updateTime, err := strconv.Atoi(items[1])
	return items[0], time.Unix(int64(updateTime), 0), nil
}

func (this *MetaServer) GetStatList(fields []string) ([]interface{}, error) {
	return this.client.HMGet(this.key, fields...).Result()
}

func (this *MetaServer) AddTpsInfo(key string, secCount uint64, totalCount uint64) error {
	this.client.LTrim(this.tpsKey, 0, META_TPS_STAT_MAX)
	data := fmt.Sprintf("%s|%d|%d", key, secCount, totalCount)
	return this.client.LPush(this.tpsKey, data).Err()
}

func (this *MetaServer) GetTpsList() ([]string, error) {
	list, err := this.client.LRange(this.tpsKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return list, err
}

func (this *MetaServer) IsExist() (bool, error) {
	return this.client.HExists(this.key, META_FIELD_NAME).Result()
}

func (this *MetaServer) Close() {
	if this.client != nil {
		this.client.Close()
	}
}

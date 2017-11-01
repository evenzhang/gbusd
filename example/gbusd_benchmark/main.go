package main

import (
	"flag"
	"fmt"
	"github.com/evenzhang/gbusd/gbusd"
	"github.com/evenzhang/gbusd/gbusd/parse"
	"github.com/evenzhang/gbusd/gbusd/store"
	"github.com/evenzhang/gbusd/gserv/application"
	"github.com/evenzhang/gbusd/gserv/config"
	"github.com/evenzhang/gbusd/gserv/log"
	"github.com/go-redis/redis"
	"os"
	"time"
)

func benchMarkBinlogSync(eventStore *store.EventStore) {
	binlogFileName := config.AppCfg("benchmark", "binlogFileName").(string)
	binLogPos := uint32(config.AppCfg("benchmark", "binLogPos").(int))

	master, err := store.NewMaster()
	if err != nil {
		fmt.Println("NewMaster Error:", err)
	}

	lastPosInfo, err := master.GetPos()
	if err != nil {
		fmt.Println("Get Master LastPosInfo Error:", err)
	}

	fmt.Println(lastPosInfo.FileName, lastPosInfo.Pos, binlogFileName, binLogPos)

	parser := parse.NewBinlogParser(eventStore)
	fmt.Println("begin replication", binlogFileName, binLogPos)
	parser.StartReplication(NewEventHandler(lastPosInfo.FileName, lastPosInfo.Pos), binlogFileName, binLogPos)
	fmt.Println("benchMarkBinlogSync done!")
}

func benchMarkRedis() {
	evStore, err := store.NewEventStore(gbusd.SENDER_CHAN_MAX)
	if err != nil {
		log.Fatalln("benchMarkRedis store.NewEventStore error:", err)
	}

	redisMaxCount := config.AppCfg("benchmark", "redisMaxCount").(int)

	redisCount := len(config.AppCfgArray("store", "list"))

	for i := 0; i < redisMaxCount; i++ {
		cmd := redis.NewBoolCmd("hset", "bench-mark", "redis", i)
		evStore.Exec(parse.RCmd{BinLogFileName: "", BinLogPos: 0, ID: i % redisCount, Cmd: cmd})
		log.Infof("REDISSTAT|%s|%d|%d", time.Now().Format("2006-01-02 15:04:05"), i%redisCount, i)
	}
	fmt.Println("benchMarkRedis done!")
}

func benchMarkServ() {
	servChanLen := config.AppCfg("benchmark", "servChanLen").(int)
	evStore, err := store.NewEventStore(servChanLen)
	if err != nil {
		log.Fatalln("benchMarkRedis store.NewEventStore error:", err)
	}

	go evStore.ExecLoopFunc(func(c parse.RCmd) {
		log.Infof("REDISSTAT | %s | %s | %d ", time.Now().Format("2006-01-02 15:04:05"), c.BinLogFileName, c.BinLogPos)
	}, nil)
	benchMarkBinlogSync(evStore)
	time.Sleep(time.Second * 20)
}

func setLog(name string) {
	os.Remove(config.Config().Common.LogFile + "." + name)
	log.SetLevelAndFile(config.Config().Common.LogFile+"."+name, config.Config().Common.LogLevel)
}

func main() {
	App := application.NewApplication(application.ProjConstInfo{
		Name:            gbusd.RES_PROJ_NAME,
		Desc:            gbusd.RES_PROJ_DESC,
		Version:         gbusd.RES_PROJ_VERSION,
		DefaultConfPath: gbusd.GetDefaultConf("benchmark")})

	target := ""
	flag.StringVar(&target, "t", "none", "Select target")

	opts := App.EnableAndParserFlagCmd(gbusd.RES_PROJ_USR_USAGE)
	if err := App.LoadAndMergeConfigFile(opts); err != nil {
		fmt.Println(err)
	}

	switch target {
	case "db":
		setLog("replication")
		benchMarkBinlogSync(nil)
	case "redis":
		setLog("redis")
		benchMarkRedis()
	case "serv":
		setLog("serv")
		benchMarkServ()
	default:
		fmt.Println("please select target:\tgbusd_benchmark -t db/redis/serv")
	}
}

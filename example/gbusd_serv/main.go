package main

import (
	"fmt"
	"github.com/evenzhang/gbusd/gbusd"
	"github.com/evenzhang/gbusd/gserv/application"
)

func main() {
	App := application.NewApplication(application.ProjConstInfo{
		Name:            gbusd.RES_PROJ_NAME,
		Desc:            gbusd.RES_PROJ_DESC,
		Version:         gbusd.RES_PROJ_VERSION,
		DefaultConfPath: gbusd.GetDefaultConf("serv")})

	daemon := gbusd.NewDaemonServer(NewEventHandler())
	daemon.ParserFlag()

	opts := App.EnableAndParserFlagCmd(gbusd.RES_PROJ_USR_USAGE)
	if err := App.LoadAndMergeConfigFile(opts); err != nil {
		fmt.Println(err)
	}

	App.Init(daemon)
	App.Run()
}

package application

import (
	"flag"
	"fmt"
	"github.com/evenzhang/gbusd/gserv/comm"
	"github.com/evenzhang/gbusd/gserv/config"
	"github.com/evenzhang/gbusd/gserv/curator"
	"github.com/evenzhang/gbusd/gserv/daemon"
	"github.com/evenzhang/gbusd/gserv/debug"
	"github.com/evenzhang/gbusd/gserv/log"
	. "github.com/logrusorgru/aurora"
	"github.com/samuel/go-zookeeper/zk"
	"net/http"
	"os"
	"strings"
)

var usageStr = `
Usage: %s [options]
Server Options:
    -l      <host:port>   Bind to host address (default: 0.0.0.0:18887)
    -c      <file>        Configuration file
    -s      <signal>      Send signal to server process (stop, quit, reopen, reload)
    -i                    Show Config Info
    -d                    Daemon Mode
    --pidf  <file>        PID file
    --debug               Enable debugging output
%s
Common Options:
    -h, --help            Show this message
    -v, --version         Show version
%s
\n`

type ProjConstInfo struct {
	Name            string
	Desc            string
	Version         string
	DefaultConfPath string
}

type Application struct {
	projInfo     ProjConstInfo
	daemonServer IDaemonServer
	frame        *curator.CuratorFramework
	pidfile      string
}

type IDaemonServer interface {
	OnInit()
	OnRun()
	OnClose()
	RegisterAllApi()
	RegisterSignals()
}

func NewApplication(projInfo ProjConstInfo) *Application {
	return &Application{projInfo: projInfo}
}

func (this *Application) EnableAndParserFlagCmd(userUsage string) *config.Options {
	opts := &config.Options{}

	flag.BoolVar(&opts.Common.Debug, "debug", false, "Enable Debug logging.")
	flag.BoolVar(&opts.Common.Daemon, "d", false, "Enable Daemon logging.")
	flag.BoolVar(&opts.Common.IsShowVersionInfo, "v", false, "Print version information.")
	flag.BoolVar(&opts.Common.IsShowConfigInfo, "i", false, "Show Config Info.")
	flag.StringVar(&opts.Common.ConfigFile, "c", "", "Configuration file.")
	flag.StringVar(&opts.Common.Signal, "s", "", "Send signal to server process (stop, quit, start, restart)")
	flag.StringVar(&opts.Common.PidFile, "pid", "", "File to store process pid.")
	flag.StringVar(&opts.Common.Listen, "l", "", "Bind to host address (default: 0.0.0.0:18887)")

	flag.Usage = func() {
		fmt.Println(this.projInfo.Desc)
		fmt.Printf(Usage(), this.projInfo.Name, userUsage, "")
		os.Exit(0)
	}
	flag.Parse()
	return opts
}

func Usage() string {
	return usageStr
}

func (this *Application) LoadAndMergeConfigFile(opts *config.Options) error {
	debug.Println("LoadAndMergeConfigFile start")
	configFile := comm.GetCurrentDirectory() + this.projInfo.DefaultConfPath

	if opts.Common.ConfigFile != "" {
		configFile = opts.Common.ConfigFile
	}
	opts.Common.ConfigFile = configFile

	fileOpts, err := config.ProcessConfigFile(configFile)
	if err != nil {
		fmt.Println("Fatal Error", err)
		log.Fatalln(err)
		return err
	}
	opts = config.MergeOptions(fileOpts, opts)
	opts = config.MergeOptions(config.NewDefaultOptions(), opts)
	config.Set(opts)
	debug.Println("LoadAndMergeConfigFile end")
	return nil
}

func (this *Application) processCliCommand() {
	if config.Config().Common.IsShowVersionInfo {
		fmt.Println(Sprintf(Bold(Cyan("%s Daemon Server, Version:  %s")), this.projInfo.Name, this.projInfo.Version))
		os.Exit(0)
	}
	if config.Config().Common.IsShowConfigInfo {
		config.Print()
		os.Exit(0)
	}
}

func (this *Application) Init(daemonServer IDaemonServer) {
	debug.SetDebugMode(config.Config().Common.Debug)
	debug.Println("init")
	this.processCliCommand()
	this.configureLogger()

	this.daemonServer = daemonServer
	this.daemonServer.OnInit()

	this.initSignal()
	this.startDaemon()

	this.registerAllApi()

	if err := daemon.WriteAndLockPidFile(this.getPidFilePath(), 0644); err != nil {
		comm.FatalPrintAndDie(fmt.Sprintf("Could not write pidfile(%s): %v\n", config.Config().Common.PidFile, err))
	}
	this.initCurator()
	go this.serveHttp()
}
func (this *Application) initCurator() {
	if config.Config().Election.Enable {
		var err error
		zkHost := config.Config().Election.Servers
		this.frame, err = curator.NewCuratorFramework(zkHost, func(event zk.Event) {
			debug.Println(event)
			switch event.State {
			case zk.StateDisconnected:
				debug.Println("NewCuratorFramework  StateDisconnected restart")
				log.Errorln("NewCuratorFramework  StateDisconnected restart")
				daemon.Restart(true)
			}
		})
		if err != nil {
			log.Errorln(err)
		}
	}
}
func (this *Application) initSignal() {
	debug.Println("initSignal")
	this.daemonServer.RegisterSignals()
	if config.Config().Common.Signal != "" {
		pid, err := daemon.ReadPid(this.getPidFilePath())
		comm.AssertErr(err)
		log.Infoln("ReadPid pid=", pid)

		err = daemon.ServSignals(config.Config().Common.Signal, pid)
		comm.AssertErr(err)
	} else {
		go func() {
			err := daemon.ServSignals("", os.Getpid())
			comm.AssertErr(err)
		}()
	}
}
func (this *Application) Run() {
	debug.Println("run")
	if config.Config().Election.Enable {
		leaderSelector := curator.NewLeaderSelector(this.frame, config.Config().Election.Root, func() {
			debug.Println("election succ")
			this.daemonServer.OnRun()
		})
		leaderSelector.Start()
	} else {
		this.daemonServer.OnRun()
	}
}

func (this *Application) serveHttp() {
	if err := http.ListenAndServe(config.Config().Common.Listen, nil); err != nil {
		fmt.Println("serveHttp", err)
		log.Fatalln("serveHttp", err)
	}
}

func (this *Application) configureLogger() {
	debug.Println("ConfigureLogger")
	if strings.HasPrefix(config.Config().Common.LogFile, "/") {
		log.SetLevelAndFile(config.Config().Common.LogFile, config.Config().Common.LogLevel)
	} else {
		log.SetLevelAndFile(comm.GetCurrentDirectory()+config.Config().Common.LogFile, config.Config().Common.LogLevel)
	}
}
func (this *Application) registerAllApi() {
	debug.Println("RegisterAllApi")
	this.daemonServer.RegisterAllApi()
}

func (this *Application) startDaemon() {
	debug.Println("StartDaemon", config.Config().Common.Daemon)
	if config.Config().Common.Daemon {
		if daemon.Daemon(false) == daemon.DAEMON_PARENT {
			os.Exit(0)
		}
	}
}

func (this *Application) getPidFilePath() string {
	if !strings.HasPrefix(config.Config().Common.LogFile, "/") {
		return comm.GetCurrentDirectory() + config.Config().Common.LogFile
	}
	return config.Config().Common.LogFile
}

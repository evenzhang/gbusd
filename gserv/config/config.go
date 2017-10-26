package config

import (
	"fmt"
	"github.com/go-yaml/yaml"
	. "github.com/logrusorgru/aurora"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"time"
)

type Common struct {
	ServerName        string        `yaml:"servername"`
	ConfigFile        string        `yaml:"c"`
	PingInterval      time.Duration `yaml:"ping_interval"`
	NoLog             bool          `yaml:"nolog"`
	NoSigs            bool          `yaml:"nosigs"`
	PidFile           string        `yaml:"pidfile"`
	Debug             bool          `yaml:"debug"`
	Daemon            bool          `yaml:"daemon"`
	Listen            string        `yaml:"listen"`
	IpWhiteList       string        `yaml:"ipwhitelist"`
	LogFile           string        `yaml:"applog"`
	LogLevel          string        `yaml:"loglevel"`
	ErrorLog          string        `yaml:"errlog"`
	WriteDeadline     time.Duration `yaml:"-"`
	IsShowVersionInfo bool
	IsShowConfigInfo  bool
	Signal            string
}

type ElectionConfig struct {
	StoreType string   `yaml:"type"`
	Servers   []string `yaml:"servers"`
	Root      string   `yaml:"root"`
	Enable    bool     `yaml:"enable"`
}

type Options struct {
	Common     Common                            `yaml:"common"`
	Election   ElectionConfig                    `yaml:"election"`
	SettingMap map[string]map[string]interface{} `yaml:"-"`
}

func ProcessConfigFile(configFile string) (*Options, error) {
	fi, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	data, err := ioutil.ReadAll(fi)
	if err != nil {
		return nil, err
	}

	datajson := []byte(data)
	opts := Options{}
	err = yaml.Unmarshal(datajson, &opts)
	if err != nil {
		return nil, err
	}

	m := make(map[string]map[string]interface{})

	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return nil, err
	}
	delete(m, "common")
	delete(m, "election")

	opts.SettingMap = m

	return &opts, nil
}

var (
	_config      *Options = nil
	_config_lock          = new(sync.RWMutex)
)

func Set(opts *Options) {
	_config_lock.Lock()
	_config = opts
	_config_lock.Unlock()
}

func Config() *Options {
	_config_lock.RLock()
	defer _config_lock.RUnlock()

	return _config
}

func AppCfg(section string, key string) interface{} {
	_config_lock.RLock()
	defer _config_lock.RUnlock()

	return _config.SettingMap[section][key]
}

func NewDefaultOptions() *Options {
	common := Common{
		ServerName:        "",
		ConfigFile:        "",
		PingInterval:      2 * time.Second,
		NoLog:             false,
		NoSigs:            false,
		PidFile:           "tmp/bus_pid",
		Debug:             false,
		Daemon:            false,
		Listen:            "0.0.0.0:18887",
		LogFile:           "app.log",
		LogLevel:          "info",
		ErrorLog:          "logs/err-{date}.log",
		WriteDeadline:     2 * time.Second,
		IsShowVersionInfo: false,
		IsShowConfigInfo:  false,
		IpWhiteList:       "",
		Signal:            "",
	}
	election := ElectionConfig{
		StoreType: "zookeeper",
		Servers:   []string{"10.235.34.50:2181"},
		Root:      "election-dev",
		Enable:    true,
	}
	return &Options{Common: common, Election: election}
}

func MergeOptions(old, new *Options) *Options {
	if old == nil {
		return new
	}
	if new == nil {
		return old
	}
	opts := *old

	if new.Common.ServerName != "" {
		opts.Common.ServerName = new.Common.ServerName
	}

	if new.Common.ConfigFile != "" {
		opts.Common.ConfigFile = new.Common.ConfigFile
	}
	if new.Common.PidFile != "" {
		opts.Common.PidFile = new.Common.PidFile
	}
	if new.Common.Listen != "" {
		opts.Common.Listen = new.Common.Listen
	}
	if new.Common.LogFile != "" {
		opts.Common.LogFile = new.Common.LogFile
	}
	if new.Common.LogLevel != "" {
		opts.Common.LogLevel = new.Common.LogLevel
	}
	if new.Common.ErrorLog != "" {
		opts.Common.ErrorLog = new.Common.ErrorLog
	}
	if new.Common.PingInterval != 0 {
		opts.Common.PingInterval = new.Common.PingInterval
	}
	if new.Common.NoLog != false {
		opts.Common.NoLog = new.Common.NoLog
	}
	if new.Common.NoSigs != false {
		opts.Common.NoSigs = new.Common.NoSigs
	}
	if new.Common.Debug != false {
		opts.Common.Debug = new.Common.Debug
	}
	if new.Common.Daemon != false {
		opts.Common.Daemon = new.Common.Daemon
	}
	if new.Common.IsShowVersionInfo != false {
		opts.Common.IsShowVersionInfo = new.Common.IsShowVersionInfo
	}
	if new.Common.IsShowConfigInfo != false {
		opts.Common.IsShowConfigInfo = new.Common.IsShowConfigInfo
	}
	if new.Common.Signal != "" {
		opts.Common.Signal = new.Common.Signal
	}
	if new.Common.IpWhiteList != "" {
		opts.Common.IpWhiteList = new.Common.IpWhiteList
	}

	if new.Election.Root != "" {
		opts.Election = new.Election
	}

	if new.SettingMap != nil {
		opts.SettingMap = new.SettingMap
	}

	return &opts
}

func Print() {
	opts := Config()
	fmt.Println(Bold(Cyan("Server Config Info:")))
	fmt.Println(Bold(Red("[Config.Common]        ")), Gray("-----------------------"))

	fmt.Println(Bold(Green("    ConfigFile        ")), Bold(Red("  [string]  ")), opts.Common.ConfigFile)
	fmt.Println(Bold(Green("    PingInterval      ")), Bold(Red("  [string]  ")), opts.Common.PingInterval)
	fmt.Println(Bold(Green("    NoLog             ")), Bold(Red("  [ bool ]  ")), opts.Common.NoLog)
	fmt.Println(Bold(Green("    NoSigs            ")), Bold(Red("  [ bool ]  ")), opts.Common.NoSigs)
	fmt.Println(Bold(Green("    PidFile           ")), Bold(Red("  [string]  ")), opts.Common.PidFile)
	fmt.Println(Bold(Green("    Debug             ")), Bold(Red("  [ bool ]  ")), opts.Common.Debug)
	fmt.Println(Bold(Green("    Daemon            ")), Bold(Red("  [ bool ]  ")), opts.Common.Daemon)
	fmt.Println(Bold(Green("    Listen            ")), Bold(Red("  [string]  ")), opts.Common.Listen)
	fmt.Println(Bold(Green("    IpWhiteList       ")), Bold(Red("  [string]  ")), opts.Common.IpWhiteList)
	fmt.Println(Bold(Green("    LogFile           ")), Bold(Red("  [string]  ")), opts.Common.LogFile)
	fmt.Println(Bold(Green("    LogLevel          ")), Bold(Red("  [string]  ")), opts.Common.LogLevel)
	fmt.Println(Bold(Green("    ErrorLog          ")), Bold(Red("  [string]  ")), opts.Common.ErrorLog)
	fmt.Println(Bold(Green("    WriteDeadline     ")), Bold(Red("  [timedr]  ")), opts.Common.WriteDeadline)
	fmt.Println(Bold(Green("    IsShowVersionInfo ")), Bold(Red("  [ bool ]  ")), opts.Common.IsShowVersionInfo)
	fmt.Println(Bold(Green("    IsShowConfigInfo  ")), Bold(Red("  [ bool ]  ")), opts.Common.IsShowConfigInfo)
	fmt.Println(Bold(Green("    Signal            ")), Bold(Red("  [string]  ")), opts.Common.Signal)

	fmt.Println(Bold(Red("\n[Config.Election]       ")), Gray("-----------------------"))
	fmt.Println(Bold(Green("    StoreType         ")), Bold(Red("  [string]  ")), opts.Election.StoreType)
	fmt.Println(Bold(Green("    Servers           ")), Bold(Red("  [[]str ]  ")), opts.Election.Servers)
	fmt.Println(Bold(Green("    Root              ")), Bold(Red("  [string]  ")), opts.Election.Root)
	fmt.Println(Bold(Green("    Master            ")), Bold(Red("  [ bool ]  ")), opts.Election.Enable)

	for section, cfgmap := range opts.SettingMap {
		fmt.Println(Sprintf(Bold(Red("[Config.%-16s")), Sprintf("%s]", section)), Gray("-----------------------"))
		for k, v := range cfgmap {
			vtype := reflect.TypeOf(v)
			if vtype.Kind() == reflect.Slice {
				vval := reflect.ValueOf(v)
				vallen := vval.Len()

				fmt.Println(Sprintf(Bold(Green("    %-18s")), k), Sprintf(Bold(Red("  [%-6s]  ")), vtype.Kind()), Sprintf(Bold(Green("#len = %d")), vallen))

				//for _, item := range v {
				for i := 0; i < vallen; i++ {
					fmt.Println(Sprintf("%37s", "-"), vval.Index(i))
				}
			} else {
				fmt.Println(Sprintf(Bold(Green("    %-18s")), k), Sprintf(Bold(Red("  [%-6s]  ")), vtype.Kind()), v)
			}
		}
	}
	fmt.Println(Bold(Cyan("\n------------------------------------------------\nDone!")))
}

package api

import (
	"encoding/json"
	"github.com/juju/errors"
	"github.com/evenzhang/gbusd/gbusd/binlog"
	"github.com/evenzhang/gbusd/gbusd/res"
	"github.com/evenzhang/gbusd/gbusd/store"
	"github.com/evenzhang/gbusd/gserv/config"
	"github.com/evenzhang/gbusd/gserv/daemon"
	"github.com/evenzhang/gbusd/gserv/log"
	"net/http"
	"strings"
	"time"
)

type ApiData struct {
	Code int
	Msg  string
	Data interface{}
}

func renderSucc(w http.ResponseWriter, val interface{}) {
	data := ApiData{10000, "succ", val}
	json, err := json.Marshal(data)
	if err != nil {
		log.Debugln(err)
	}
	w.Write(json)
}
func renderError(w http.ResponseWriter, msg string, val interface{}) {
	data := ApiData{10002, msg, val}
	json, err := json.Marshal(data)
	if err != nil {
		log.Debugln(err)
	}
	w.Write(json)
}

func apiGetVersion(w http.ResponseWriter, r *http.Request) {
	if err := checkLoginIP(r); err != nil {
		renderError(w, err.Error(), "")
		return
	}
	renderSucc(w, res.RES_PROJ_VERSION)
}

func apiGetConfig(w http.ResponseWriter, r *http.Request) {
	if err := checkLoginIP(r); err != nil {
		renderError(w, err.Error(), "")
		return
	}

	data := make(map[string]interface{})
	data["common"] = config.Config().Common
	data["election"] = config.Config().Election
	data["meta"] = config.AppCfg("meta", "addr").(string)

	data["master"] = binlog.GetMasterDsn()

	data["store"] = store.GetStoreConfig()
	data["version"] = res.RES_PROJ_VERSION
	data["runtime"] = res.GetAllRuntime()

	renderSucc(w, data)
}

func apiRestart(w http.ResponseWriter, r *http.Request) {
	if err := checkLoginIP(r); err != nil {
		renderError(w, err.Error(), "")
		return
	}

	renderSucc(w, "Server Restarting ... ...")
	go func() {
		time.Sleep(time.Microsecond * 300)
		daemon.Restart(true)
	}()
}

func checkLoginIP(r *http.Request) error {
	ipWhiteList := strings.TrimSpace(config.Config().Common.IpWhiteList)
	//白名单为空，默认不校验IP
	if ipWhiteList == "*" || ipWhiteList == "" {
		return nil
	}

	ip := strings.Split(r.RemoteAddr, ":")[0]
	if strings.Contains(";"+ipWhiteList+";", ";"+ip+";") {
		return nil
	}
	return errors.Errorf("Access denied for ip[%s]", ip)
}

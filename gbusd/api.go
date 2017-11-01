package gbusd

import (
	"encoding/json"
	"github.com/evenzhang/gbusd/gbusd/store"
	"github.com/evenzhang/gbusd/gserv/config"
	"github.com/evenzhang/gbusd/gserv/daemon"
	"github.com/evenzhang/gbusd/gserv/log"
	"github.com/juju/errors"
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
	renderSucc(w, RES_PROJ_VERSION)
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

	data["master"] = store.GetMasterDsn()

	data["store"] = config.AppCfgArray("store", "list")
	data["version"] = RES_PROJ_VERSION
	data["runtime"] = GetAllRuntime()

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

func RegisterAllApi() {
	http.HandleFunc("/api/version", apiGetVersion)
	http.HandleFunc("/api/conf", apiGetConfig)
	http.HandleFunc("/api/restart", apiRestart)
}

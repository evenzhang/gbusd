package res

import (
	"fmt"
	"sync"
	"time"
)

const (
	RES_PROJ_NAME              = "busd"
	RES_PROJ_DESC              = "busd is daemon server"
	RES_PROJ_VERSION           = "beta 1.00"
	RES_PROJ_DEFAULT_CONF_PATH = "conf/app.yml"
	RES_PROJ_USR_USAGE         = "    --cli                 Enable cli mode"
)

var (
	_runtimeInfo  = make(map[string]interface{})
	_runtime_lock = new(sync.RWMutex)
)

func init() {
	SetRuntimeVarNow("BootTime")
	SetRuntimeVar("ElectionStatus", "slave")
}

func SetRuntimeVarNow(key string) {
	_runtime_lock.Lock()
	_runtimeInfo[key] = time.Now().Format("2006-01-02 15:04:05")
	_runtime_lock.Unlock()
}

func SetRuntimeVar(key string, val interface{}) {
	_runtime_lock.Lock()
	_runtimeInfo[key] = val
	_runtime_lock.Unlock()
}
func GetRuntimeVar(key string) interface{} {
	_runtime_lock.RLock()
	defer _runtime_lock.RUnlock()

	return _runtimeInfo[key]
}

func GetDefaultConf(fileName string) string {
	return fmt.Sprintf("conf/app_%s.yml", fileName)
}

func GetAllRuntime() map[string]interface{} {
	_runtime_lock.RLock()
	defer _runtime_lock.RUnlock()

	return _runtimeInfo
}

package debug

import (
	"fmt"
	"sync"
)

var (
	_debugEnable bool          = false
	_lock        *sync.RWMutex = new(sync.RWMutex)
)

func SetDebugMode(debug bool) {
	_lock.Lock()
	_debugEnable = debug
	_lock.Unlock()
}

func Printf(format string, args ...interface{}) {
	_lock.RLock()
	debugEnable := _debugEnable
	_lock.RUnlock()

	if debugEnable {
		fmt.Printf(format, args...)
	}

}

func Println(args ...interface{}) {
	_lock.RLock()
	debugEnable := _debugEnable
	_lock.RUnlock()

	if debugEnable {
		fmt.Println(args...)
	}
}

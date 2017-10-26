package application

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/evenzhang/gbusd/gserv/daemon"
	"github.com/evenzhang/gbusd/gserv/debug"
	"github.com/evenzhang/gbusd/gserv/log"
	"net/http"
	"os"
	"syscall"
)

type BaseDaemonServer struct {
}

func (this *BaseDaemonServer) OnClose() {
	debug.Println("Busd", "OnClose")
}

func (this *BaseDaemonServer) RegisterAllApi() {
	debug.Println("Busd", "RegisterAllApi")
	http.HandleFunc("/", httpHandler)
}

func (this *BaseDaemonServer) RegisterSignals() {
	RegisterDefaultSignals()
}

func (this *BaseDaemonServer) OnCuratorEvent(event zk.Event) {
	DefaultCuratorEvent(event)
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "go-daemon-pid: %d", os.Getpid())
}

func DefaultCuratorEvent(event zk.Event) {
	debug.Println("Busd", "OnCuratorEvent")
	switch event.State {
	case zk.StateDisconnected:
		fmt.Println("NewCuratorFramework  StateDisconnected restart")
		log.Errorln("NewCuratorFramework  StateDisconnected restart")
		daemon.Restart(true)
	}
}
func RegisterDefaultSignals() {
	debug.Println("Busd", "RegisterSignals")
	daemon.AddCommand("restart", syscall.SIGUSR2, func() {
		fmt.Printf("AddCommand-RESTART")
		daemon.Restart(true)
	})
}

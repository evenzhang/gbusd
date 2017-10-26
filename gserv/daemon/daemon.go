package daemon

import (
	"fmt"
	"github.com/juju/errors"
	"github.com/kardianos/osext"
	d "github.com/sevlyar/go-daemon"
	"os"
	"os/signal"
	"syscall"
)

/*
eg:

if Daemon() == DAEMON_PARENT {
	os.Exit(0)
}


if Restart() == DAEMON_PARENT {
	os.Exit(0)
}
*/

var (
	pidfile *d.LockFile
)

const (
	DAEMON_CHILD  = 1
	DAEMON_PARENT = 0
	DAEMON_ERROR  = -1
)

const (
	_MARK_NAME  = "_GO_DAEMON"
	_MARK_VALUE = "1"
)

func Daemon(autoExit bool) int {
	res := DAEMON_CHILD
	if os.Getenv(_MARK_NAME) == "" {
		os.Setenv(_MARK_NAME, _MARK_VALUE)
		res = Restart(autoExit)
	}

	return res
}
func Restart(autoExit bool) int {
	attr := &os.ProcAttr{
		Dir:   "./",
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	abspath, err := osext.Executable()
	if err != nil {
		return DAEMON_ERROR
	}

	child, err := os.StartProcess(abspath, os.Args, attr)

	if err != nil {
		return DAEMON_ERROR
	}

	if child != nil {
		if autoExit {
			os.Exit(1)
		}
		return DAEMON_PARENT
	}
	fmt.Println("DAEMON_CHILD")
	return DAEMON_CHILD
}

// SignalHandlerFunc is the interface for signal handler functions.
type SignalHandlerFunc func()

var handlers = make(map[os.Signal]SignalHandlerFunc)
var flags = make(map[string]syscall.Signal)

func AddCommand(cmdstr string, sig syscall.Signal, handler func()) {
	flags[cmdstr] = sig
	handlers[sig] = handler
}

func ServSignals(flag string, pid int) error {
	if flag != "" {
		if signum, exist := flags[flag]; exist {
			syscall.Kill(pid, signum)
			os.Exit(1)
		}
		return errors.New("Signal handler not found")
	}
	signals := make([]os.Signal, 0, 10)
	for k, _ := range handlers {
		signals = append(signals, k)
	}

	ch := make(chan os.Signal, 10)
	signal.Notify(ch, signals...)

	for sig := range ch {
		handlers[sig]()
	}

	signal.Stop(ch)
	return nil
}

func ReadPid(name string) (pid int, err error) {
	return d.ReadPidFile(name)
}

func WriteAndLockPidFile(name string, perm os.FileMode) (err error) {
	if pidfile != nil {
		return nil
	}

	if pidfile, err = d.OpenLockFile(name, perm); err != nil {
		return err
	}
	if err = pidfile.Lock(); err != nil {
		return err
	}
	if err = pidfile.WritePid(); err != nil {
		return err
	}
	return err
}

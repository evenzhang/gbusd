/*

	listener := curator.NewLeaderSelectorListener(func(CuratorFramework client) error {
		// this callback will get called when you are the leader
		// do whatever leader work you need to and only exit
		// this method when you want to relinquish leadership
	}))

	selector := curator.NewLeaderSelector(client, path, listener)
	selector.AutoRequeue()  // not required, but this is behavior that you will probably expect
	selector.Start()


eg:
	log.SetLevelAndFile("daemonServer.applog", "debug")

	frame, err := NewCuratorFramework("10.235.34.50:2181", func(event zk.Event) {
		fmt.Println(event)
		switch event.State {
		case zk.StateDisconnected:
			fmt.Println("NewCuratorFramework  StateDisconnected restart")
			log.Errorln("NewCuratorFramework  StateDisconnected restart")
			daemon.Restart(true)
		}
	})

	if err != nil {
		log.Errorln(err)
	}

	leaderSelector := NewLeaderSelector(frame, "election", func() {
		fmt.Println("election succ")
	})
	leaderSelector.Start()

*/

package curator

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/evenzhang/gbusd/gserv/debug"
	"github.com/evenzhang/gbusd/gserv/log"
	"strconv"
	"strings"
	"time"
)

type CuratorFramework struct {
	Conn             *zk.Conn
	EventChan        <-chan zk.Event
	EventFunc        func(zk.Event)
	IsStateConnected bool
}

type LeaderSelector struct {
	Frame     *CuratorFramework
	LeaderFun func()
	rootPath  string
	guid      string
}

func NewCuratorFramework(zkHost []string, eventFun func(zk.Event)) (*CuratorFramework, error) {
	z, connChan, err := zk.Connect(zkHost, time.Second)
	if err != nil {
		log.Errorln("NewCuratorFramework", err)
	}
	frame := &CuratorFramework{Conn: z, EventChan: connChan, EventFunc: eventFun, IsStateConnected: false}

	go func() {
		checkConnHealthCounter := 0
		for {
			connEvent := <-connChan
			eventFun(connEvent)
			switch connEvent.State {
			case zk.StateConnected:
				checkConnHealthCounter = 0
				frame.IsStateConnected = true
			case zk.StateConnecting:
				checkConnHealthCounter++
				//检测超时
				if checkConnHealthCounter >= 3 {
					frame.IsStateConnected = false
					ev := zk.Event{Type: zk.EventNotWatching, State: zk.StateDisconnected, Path: "", Err: nil}
					eventFun(ev)
				}
				continue
			case zk.StateExpired:
				frame.IsStateConnected = false
			case zk.StateAuthFailed:
				frame.IsStateConnected = false
			default:
				checkConnHealthCounter = 0
			}

			time.Sleep(time.Microsecond * 100)
		}
	}()

	return frame, err
}

func NewLeaderSelector(frame *CuratorFramework, path string, leaderFun func()) *LeaderSelector {
	debug.Println("NewLeaderSelector")
	return &LeaderSelector{Frame: frame, rootPath: path, LeaderFun: leaderFun, guid: ""}
}

func (this *LeaderSelector) Start() {
	for !this.Frame.IsStateConnected {
		time.Sleep(time.Second)
	}

	this.Frame.CreateZkNode(this.rootPath)
	this.election(this.Frame)
	log.Infoln("Leader election Succ")
	this.LeaderFun()

}

func isLeader(children []string, seq int64, guid string) bool {
	// Get all the children and check who has the smallest guid
	// If this is the only child left then it becomes the leader
	if len(children) == 1 {
		log.Debugf("Only child left.")
		return true
	}

	var smallest int64 = 0
	for _, child := range children {
		childSeq, _ := strconv.ParseInt(child[strings.LastIndex(child, "-")+1:], 10, 16)

		if smallest == 0 || childSeq < smallest {
			smallest = childSeq
			log.Debugf("SMALLEST %s: %d", guid, smallest)
		}
	}

	// If the current guid is the smallest, then current process becomes
	// the new leader and we can break out of this election loop
	if seq <= smallest {
		log.Debugf("Seq %d elected leader!", seq)
		return true
	}
	return false
}

func (this *LeaderSelector) election(frame *CuratorFramework) (err error) {
	path := fmt.Sprintf("/%s/", this.rootPath)
	this.guid, err = frame.Conn.CreateProtectedEphemeralSequential(path, []byte{}, zk.WorldACL(zk.PermAll))
	if err != nil {
		log.Debugf("Error creating EphemeralSequential node: %s", err)
		panic(err)
	}
	log.Debugf("GUID: %s", this.guid)

	seq, _ := strconv.ParseInt(this.guid[strings.LastIndex(this.guid, "-")+1:], 10, 16)

	path = fmt.Sprintf("/%s", this.rootPath)

	for {
		// STEP 1: Watch all children
		children, _, channel, err := frame.Conn.ChildrenW(path)
		if err != nil {
			log.Debugf("Error watching node: %s", err)
		}
		if isLeader(children, seq, this.guid) {
			break
		}

		//block here and listen for events
		log.Debugf("Waiting on channel...")
		event := <-channel
		log.Debugf("EVENT: %+v", event)
		if event.Type == zk.EventNodeChildrenChanged {
			children, _, err := frame.Conn.Children(path)
			if err != nil {
				log.Debugf("Error getting children: %s", err)
				panic(err)
			}

			if isLeader(children, seq, this.guid) {
				break
			}
		}
	}

	log.Debugf("Election ended %s", this.guid)
	return err
}

func (this *CuratorFramework) CreateZkNode(path string) {
	z := this.Conn
	nodes := strings.Split(path, "/")
	toCreate := ""
	for i := 0; i < len(nodes); i++ {
		toCreate = toCreate + "/" + nodes[i]

		if exists, _, _ := z.Exists(toCreate); !exists {
			if _, err := z.Create(toCreate, []byte{}, 0, zk.WorldACL(zk.PermAll)); err != nil {
				panic(err)
			}
		}
		log.Debugf("toCreate: %s", toCreate)
	}

}

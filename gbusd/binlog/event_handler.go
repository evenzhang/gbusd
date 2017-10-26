package binlog

import (
	"github.com/evenzhang/gbusd/gserv/log"
)

type BaseEventHandler struct {
}

func (this *BaseEventHandler) OnInit() bool {
	log.Debugln("BaseEventHandler OnInit")
	return true
}

func (this *BaseEventHandler) OnInsert(sender *Sender, ev *Event) (bool, error) {
	log.Debugln("BaseEventHandler OnInsert", sender, ev)
	return true, nil
}

func (this *BaseEventHandler) OnUpdate(sender *Sender, ev *Event) (bool, error) {
	log.Debugln("BaseEventHandler OnUpdate", sender, ev)
	return true, nil
}

func (this *BaseEventHandler) OnDelete(sender *Sender, ev *Event) (bool, error) {
	log.Debugln("BaseEventHandler OnDelete", sender, ev)
	return true, nil
}

func (this *BaseEventHandler) OnClose() {
	log.Debugln("BaseEventHandler OnClose")
}

func (this *BaseEventHandler) Enable() bool {
	log.Debugln("BaseEventHandler Enable")
	return true
}

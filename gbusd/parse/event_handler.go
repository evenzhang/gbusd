package parse

import (
	"github.com/evenzhang/gbusd/gserv/log"
)

type IEventHandler interface {
	OnInit() bool
	OnInsert(Sender ISender, event *Event) (bool, error)
	OnUpdate(Sender ISender, event *Event) (bool, error)
	OnDelete(Sender ISender, event *Event) (bool, error)
	OnClose()
	Enable() bool
}

type BaseEventHandler struct {
}

func (this *BaseEventHandler) OnInit() bool {
	log.Debugln("BaseEventHandler OnInit")
	return true
}

func (this *BaseEventHandler) OnInsert(sender ISender, ev *Event) (bool, error) {
	log.Debugln("BaseEventHandler OnInsert", sender, ev)
	return true, nil
}

func (this *BaseEventHandler) OnUpdate(sender ISender, ev *Event) (bool, error) {
	log.Debugln("BaseEventHandler OnUpdate", sender, ev)
	return true, nil
}

func (this *BaseEventHandler) OnDelete(sender ISender, ev *Event) (bool, error) {
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

package transfer

import (
	"errors"
	"log"
	"strings"
)

var (
	Evt_Create = "EVT_CREATE"
	Evt_Remove = "EVT_REMOVE"
	Evt_RemoveDir = "EVT_REMOVE_DIR"
)

type TransferEvent struct {
	EventType string
	FileName  string
	IsDir     bool
}

func NewTransferEvent(eventType string, fileName string) *TransferEvent {
	return &TransferEvent{
		EventType: eventType,
		FileName:  fileName,
	}
}

func ParseEvent(eventString string) (*TransferEvent, error) {
	log.Println("eventString: ", eventString)
	evt := strings.Split(eventString, ":")
	if len(evt) != 3 {
		return nil, errors.New("invalid event string")
	}
	// todo 这里没必要重复判断dir了
	if (evt[0] == Evt_Create || evt[0] == Evt_Remove || evt[0] == Evt_RemoveDir) && len(evt[2]) > 0 {
		event := NewTransferEvent(evt[0], evt[2])
		if evt[1] == "DIR" {
			event.IsDir = true
		} else {
			event.IsDir = false
		}
		return event,nil
	}
	return nil, errors.New("invalid event string: " + eventString, )
}

func (e *TransferEvent) String() string {
	var filetype string
	if e.IsDir {
		filetype = "DIR"
	} else {
		filetype = "FILE"
	}
	return e.EventType + ":" + filetype + ":" + e.FileName
}

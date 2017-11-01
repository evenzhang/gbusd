package parse

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/siddontang/go-mysql/replication"
)

type ISender interface {
	Push(ev *Event, id int, cmder redis.Cmder)
}

type RCmd struct {
	BinLogFileName string
	BinLogPos      uint32
	ID             int
	Cmd            redis.Cmder
}

type Event struct {
	DBName         string
	TableName      string
	Type           string
	BinLogFileName string
	BinLogPos      uint32
	Data           *replication.RowsEvent
	Header         *replication.EventHeader
}

func PrintEventInfo(ev *replication.BinlogEvent) {
	eventTypeNameList := []string{"UNKNOWN_EVENT", "START_EVENT_V3", "QUERY_EVENT", "STOP_EVENT", "ROTATE_EVENT",
		"INTVAR_EVENT", "LOAD_EVENT", "SLAVE_EVENT", "CREATE_FILE_EVENT", "APPEND_BLOCK_EVENT", "EXEC_LOAD_EVENT",
		"DELETE_FILE_EVENT", "NEW_LOAD_EVENT", "RAND_EVENT", "USER_VAR_EVENT", "FORMAT_DESCRIPTION_EVENT", "XID_EVENT",
		"BEGIN_LOAD_QUERY_EVENT", "EXECUTE_LOAD_QUERY_EVENT", "TABLE_MAP_EVENT", "WRITE_ROWS_EVENTv0", "UPDATE_ROWS_EVENTv0",
		"DELETE_ROWS_EVENTv0", "WRITE_ROWS_EVENTv1", "UPDATE_ROWS_EVENTv1", "DELETE_ROWS_EVENTv1", "INCIDENT_EVENT", "HEARTBEAT_EVENT",
		"IGNORABLE_EVENT", "ROWS_QUERY_EVENT", "WRITE_ROWS_EVENTv2", "UPDATE_ROWS_EVENTv2", "DELETE_ROWS_EVENTv2",
		"GTID_EVENT", "ANONYMOUS_GTID_EVENT", "PREVIOUS_GTIDS_EVENT"}
	fmt.Println("DefaultProcessEvent", ev.Header.EventType, eventTypeNameList[ev.Header.EventType], ev.Header.LogPos, ev.Header.LogPos-ev.Header.EventSize)
}

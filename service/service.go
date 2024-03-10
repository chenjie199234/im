package service

import (
	"github.com/chenjie199234/im/dao"
	"github.com/chenjie199234/im/service/chat"
	"github.com/chenjie199234/im/service/status"
)

// SvcStatus one specify sub service
var SvcStatus *status.Service

// SvcChat one specify sub service
var SvcChat *chat.Service

// StartService start the whole service
func StartService() error {
	if e := dao.NewApi(); e != nil {
		return e
	}
	//start sub service
	SvcStatus = status.Start()
	SvcChat = chat.Start()
	return nil
}

// StopService stop the whole service
func StopService() {
	//stop sub service
	SvcStatus.Stop()
}


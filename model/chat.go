package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MsgInfo struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	ChatKey     string             `bson:"chat_key"`
	Sender      string             `bson:"sender"`
	Msg         string             `bson:"msg"`
	Extra       string             `bson:"extra"`
	MsgIndex    uint32             `bson:"msg_index"`
	RecallIndex uint32             `bson:"recall_index"` //this field has a sparse index
}
type Ack struct {
	ChatKey      string `bson:"chat_key"`
	Acker        string `bson:"acker"`
	ReadMsgIndex uint32 `bson:"read_msg_index"`
}

type IMIndex struct {
	MsgIndex    uint32
	RecallIndex uint32
	AckIndex    uint32
}
type IMSession struct {
	//RemoteAddr's IP will be different with the RealIP when the raw connection is a websocket connection and there is proxy before this server
	RemoteAddr string // ip:port
	RealIP     string // ip
	Gate       string
	Netlag     uint64 //nano seconds
}

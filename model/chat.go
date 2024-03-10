package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MsgInfo struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	ChatKey     string             `bson:"chat_key"`
	Sender      primitive.ObjectID `bson:"sender"`
	Msg         string             `bson:"msg"`
	Extra       string             `bson:"extra"`
	MsgIndex    uint64             `bson:"msg_index"`
	RecallIndex uint64             `bson:"recall_index"` //this field has a sparse index
}
type Ack struct {
	ChatKey      string             `bson:"chat_key"`
	Acker        primitive.ObjectID `bson:"acker"`
	ReadMsgIndex uint64             `bson:"read_msg_index"`
}

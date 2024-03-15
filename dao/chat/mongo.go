package chat

import (
	"context"
	"encoding/binary"

	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (d *Dao) MongoGetMaxMsgIndex(ctx context.Context, chatkey string) (uint32, error) {
	filter := bson.M{"chat_key": chatkey}
	opts := options.FindOne().SetSort(bson.M{"msg_index": -1}).SetProjection(bson.M{"msg_index": 1})
	msg := &model.MsgInfo{}
	if e := d.mongo.Database("im").Collection("msg").FindOne(ctx, filter, opts).Decode(msg); e != nil && e != mongo.ErrNoDocuments {
		return 0, e
	} else if e != nil {
		//mongo.ErrNoDocuments
		return 0, nil
	} else {
		return msg.MsgIndex, nil
	}
}
func (d *Dao) MongoGetMaxRecallIndex(ctx context.Context, chatkey string) (uint32, error) {
	filter := bson.M{"chat_key": chatkey, "recall_index": bson.M{"$exists": true}}
	opts := options.FindOne().SetSort(bson.M{"recall_index": -1}).SetProjection(bson.M{"recall_index": 1})
	msg := &model.MsgInfo{}
	if e := d.mongo.Database("im").Collection("msg").FindOne(ctx, filter, opts).Decode(msg); e != nil && e != mongo.ErrNoDocuments {
		return 0, e
	} else if e != nil {
		//mongo.ErrNoDocuments
		return 0, nil
	} else {
		return msg.RecallIndex, nil
	}
}

// mintimestamp: unit second
func (d *Dao) MongoGetMsgs(ctx context.Context, chatkey, direction string, startMsgIndex, num, mintimestamp uint32) ([]*model.MsgInfo, error) {
	filter := bson.M{"chat_key": chatkey}
	opts := options.Find().SetLimit(int64(num))
	if direction == "after" {
		filter["msg_index"] = bson.M{"$gte": startMsgIndex}
		opts = opts.SetSort(bson.M{"msg_index": 1})
	} else {
		filter["msg_index"] = bson.M{"$lte": startMsgIndex}
		opts = opts.SetSort(bson.M{"msg_index": -1})
	}
	if mintimestamp != 0 {
		var minobjid primitive.ObjectID
		binary.BigEndian.PutUint32(minobjid[0:4], mintimestamp)
		filter["_id"] = bson.M{"$gte": minobjid}
	}
	cursor, e := d.mongo.Database("im").Collection("msg").Find(ctx, filter, opts)
	if e != nil {
		return nil, e
	}
	r := make([]*model.MsgInfo, 0, num)
	if e := cursor.All(ctx, &r); e != nil {
		return nil, e
	}
	return r, nil
}

// mintimestamp: unit second
// return key:recall index,value:msg index
func (d *Dao) MongoGetRecalls(ctx context.Context, chatkey, direction string, startRecallIndex, num, mintimestamp uint32) (map[uint32]uint32, error) {
	filter := bson.M{"chat_key": chatkey}
	opts := options.Find().SetLimit(int64(num)).SetProjection(bson.M{"msg_index": 1, "recall_index": 1})
	if direction == "after" {
		filter["recall_index"] = bson.M{"$gte": startRecallIndex}
		opts = opts.SetSort(bson.M{"recall_index": 1})
	} else {
		filter["recall_index"] = bson.M{"$lte": startRecallIndex}
		opts = opts.SetSort(bson.M{"recall_index": -1})
	}
	if mintimestamp != 0 {
		var minobjid primitive.ObjectID
		binary.BigEndian.PutUint32(minobjid[0:4], mintimestamp)
		filter["_id"] = bson.M{"$gte": minobjid}
	}
	cursor, e := d.mongo.Database("im").Collection("msg").Find(ctx, filter, opts)
	if e != nil {
		return nil, e
	}
	tmp := make([]*model.MsgInfo, 0, num)
	if e := cursor.All(ctx, &tmp); e != nil {
		return nil, e
	}
	r := make(map[uint32]uint32, num)
	for _, v := range tmp {
		r[v.RecallIndex] = v.MsgIndex
	}
	return r, nil
}
func (d *Dao) MongoSend(ctx context.Context, msg *model.MsgInfo) error {
	for {
		tmp := &model.MsgInfo{}
		filter := bson.M{"chat_key": msg.ChatKey}
		opts := options.FindOne().SetSort(bson.M{"msg_index": -1})
		if e := d.mongo.Database("im").Collection("msg").FindOne(ctx, filter, opts).Decode(tmp); e != nil && e != mongo.ErrNoDocuments {
			return e
		} else if e != nil {
			//mongo.ErrNoDocuments
			msg.MsgIndex = 1
		} else {
			msg.MsgIndex = tmp.MsgIndex + 1
		}
		doc := bson.M{"chat_key": msg.ChatKey, "sender": msg.Sender, "msg": msg.Msg, "extra": msg.Extra, "msg_index": msg.MsgIndex}
		if r, e := d.mongo.Database("im").Collection("msg").InsertOne(ctx, doc); e != nil && !mongo.IsDuplicateKeyError(e) {
			return e
		} else if e != nil {
			//DuplicateKeyError
			continue
		} else {
			msg.ID = r.InsertedID.(primitive.ObjectID)
		}
		return d.MongoAck(ctx, msg.Sender, msg.ChatKey, msg.MsgIndex)
	}
}
func (d *Dao) MongoRecall(ctx context.Context, sender, chatkey string, msgindex uint32) (recallindex uint32, e error) {
	for {
		tmp := &model.MsgInfo{}
		filter := bson.M{"chat_key": chatkey}
		opts := options.FindOne().SetSort(bson.M{"recall_index": -1})
		if e = d.mongo.Database("im").Collection("msg").FindOne(ctx, filter, opts).Decode(tmp); e != nil && e != mongo.ErrNoDocuments {
			return
		} else if e != nil {
			//mongo.ErrNoDocuments
			recallindex = 1
		} else {
			recallindex = tmp.RecallIndex + 1
		}
		filter["msg_index"] = msgindex
		filter["sender"] = sender
		updater := bson.M{"$set": bson.M{"recall_index": recallindex}}
		var r *mongo.UpdateResult
		if r, e = d.mongo.Database("im").Collection("msg").UpdateOne(ctx, filter, updater); e != nil && mongo.IsDuplicateKeyError(e) {
			//DuplicateKeyError
			continue
		}
		if e != nil || r.MatchedCount != 0 {
			return
		}
		delete(filter, "sender")
		var count int64
		if count, e = d.mongo.Database("im").Collection("msg").CountDocuments(ctx, filter); e != nil {
			return
		}
		if count == 0 {
			e = ecode.ErrMsgNotExist
		} else {
			e = ecode.ErrMsgOwnerWrong
		}
		return
	}
}
func (d *Dao) MongoGetAck(ctx context.Context, acker, chatkey string) (uint32, error) {
	filter := bson.M{"chat_key": chatkey, "acker": acker}
	ack := &model.Ack{}
	if e := d.mongo.Database("im").Collection("ack").FindOne(ctx, filter).Decode(ack); e != nil && e != mongo.ErrNoDocuments {
		return 0, e
	} else if e != nil {
		//mongo.ErrNoDocuments
		return 0, nil
	} else {
		return ack.ReadMsgIndex, nil
	}
}
func (d *Dao) MongoAck(ctx context.Context, acker string, chatkey string, msgindex uint32) error {
	filter := bson.M{"chat_key": chatkey, "acker": acker}
	updater := bson.M{"$max": bson.M{"read_msg_index": msgindex}}
	_, e := d.mongo.Database("im").Collection("ack").UpdateOne(ctx, filter, updater, options.Update().SetUpsert(true))
	return e
}

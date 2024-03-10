package relation

import (
	"context"

	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// friend's relation should be add first,because the request will not be consumed if failed and user will retry accept the request
func (d *Dao) MongoAcceptMakeFriend(ctx context.Context, userid, friendid primitive.ObjectID) (username, friendname string, e error) {
	filter := bson.M{
		"$or": bson.A{
			bson.M{"main": userid, "main_type": "user", "sub": primitive.NilObjectID, "sub_type": "user"},
			bson.M{"main": friendid, "main_type": "user", "sub": primitive.NilObjectID, "sub_type": "user"},
		},
	}
	var cursor *mongo.Cursor
	if cursor, e = d.mongo.Database("im").Collection("relation").Find(ctx, filter); e != nil {
		return
	}
	all := make([]*model.Relation, 0, cursor.RemainingBatchLength())
	if e = cursor.All(ctx, &all); e != nil {
		return
	}
	if len(all) != 2 {
		e = ecode.ErrDBDataBroken
		return
	}
	if all[0].Main.Hex() == userid.Hex() {
		username = all[0].Name
		friendname = all[1].Name
	} else {
		username = all[1].Name
		friendname = all[0].Name
	}
	if username == "" || friendname == "" {
		//user's or group's name can't be empty in this system
		e = ecode.ErrDBDataBroken
		return
	}

	opts := options.Update().SetUpsert(true)

	//use update with upsert to ignore DuplicateKeyError
	filter = bson.M{"main": friendid, "main_type": "user", "sub": userid, "sub_type": "user"}
	_, e = d.mongo.Database("im").Collection("relation").UpdateOne(ctx, filter, bson.M{"$set": bson.M{"name": username}}, opts)
	if e != nil {
		return
	}
	filter["main"] = userid
	filter["sub"] = friendid
	_, e = d.mongo.Database("im").Collection("relation").UpdateOne(ctx, filter, bson.M{"$set": bson.M{"name": friendname}}, opts)
	return
}

// group's relation should be add first,because the request will not be consumed if failed and user will retry accept the request
func (d *Dao) MongoAcceptGroupInvite(ctx context.Context, userid, groupid primitive.ObjectID) (username, groupname string, e error) {
	filter := bson.M{
		"$or": bson.A{
			bson.M{"main": userid, "main_type": "user", "sub": primitive.NilObjectID, "sub_type": "user"},
			bson.M{"main": groupid, "main_type": "group", "sub": primitive.NilObjectID, "sub_type": "user"},
		},
	}
	var cursor *mongo.Cursor
	if cursor, e = d.mongo.Database("im").Collection("relation").Find(ctx, filter); e != nil {
		return
	}
	all := make([]*model.Relation, 0, cursor.RemainingBatchLength())
	if e = cursor.All(ctx, &all); e != nil {
		return
	}
	if len(all) != 2 {
		e = ecode.ErrDBDataBroken
		return
	}
	if all[0].Main.Hex() == userid.Hex() {
		username = all[0].Name
		groupname = all[1].Name
	} else {
		username = all[1].Name
		groupname = all[0].Name
	}
	if username == "" || groupname == "" {
		//user's or group's name can't be empty in this system
		e = ecode.ErrDBDataBroken
		return
	}

	opts := options.Update().SetUpsert(true)

	//use update with upsert to ignore DuplicateKeyError
	filter = bson.M{"main": groupid, "main_type": "group", "sub": userid, "sub_type": "user"}
	_, e = d.mongo.Database("im").Collection("relation").UpdateOne(ctx, filter, bson.M{"$set": bson.M{"name": username}}, opts)
	if e != nil {
		return
	}
	filter["main"] = userid
	filter["main_type"] = "user"
	filter["sub"] = groupid
	filter["sub_type"] = "group"
	_, e = d.mongo.Database("im").Collection("relation").UpdateOne(ctx, filter, bson.M{"$set": bson.M{"name": groupname}}, opts)
	return
}

// user's relation should be add first,because the request will not be consumed if failed and group's admin will retry accept the request
func (d *Dao) MongoAcceptGroupApply(ctx context.Context, groupid, userid primitive.ObjectID) (username, groupname string, e error) {
	filter := bson.M{
		"$or": bson.A{
			bson.M{"main": userid, "main_type": "user", "sub": primitive.NilObjectID, "sub_type": "user"},
			bson.M{"main": groupid, "main_type": "group", "sub": primitive.NilObjectID, "sub_type": "user"},
		},
	}
	var cursor *mongo.Cursor
	if cursor, e = d.mongo.Database("im").Collection("relation").Find(ctx, filter); e != nil {
		return
	}
	all := make([]*model.Relation, 0, cursor.RemainingBatchLength())
	if e = cursor.All(ctx, &all); e != nil {
		return
	}
	if len(all) != 2 {
		e = ecode.ErrDBDataBroken
		return
	}
	if all[0].Main.Hex() == userid.Hex() {
		username = all[0].Name
		groupname = all[1].Name
	} else {
		username = all[1].Name
		groupname = all[0].Name
	}
	if username == "" || groupname == "" {
		//user's or group's name can't be empty in this system
		e = ecode.ErrDBDataBroken
		return
	}

	opts := options.Update().SetUpsert(true)

	//use update with upsert to ignore DuplicateKeyError
	filter = bson.M{"main": userid, "main_type": "user", "sub": groupid, "sub_type": "group"}
	_, e = d.mongo.Database("im").Collection("relation").UpdateOne(ctx, filter, bson.M{"$set": bson.M{"name": groupname}}, opts)
	if e != nil {
		return
	}
	filter["main"] = groupid
	filter["main_type"] = "group"
	filter["sub"] = userid
	filter["sub_type"] = "user"
	_, e = d.mongo.Database("im").Collection("relation").UpdateOne(ctx, filter, bson.M{"$set": bson.M{"main": username}}, opts)
	return
}

// friend's relation should be deleted first,because the user will retry
func (d *Dao) MongoDelFriend(ctx context.Context, userid, friendid primitive.ObjectID) error {
	filter := bson.M{"main": friendid, "main_type": "user", "sub": userid, "sub_type": "user"}
	_, e := d.mongo.Database("im").Collection("relation").DeleteOne(ctx, filter)
	if e != nil {
		return e
	}
	filter["main"] = userid
	filter["main_type"] = "user"
	filter["sub"] = friendid
	filter["sub_type"] = "user"
	_, e = d.mongo.Database("im").Collection("relation").DeleteOne(ctx, filter)
	return e
}

// group's relation should be deleted first,because the user will retry
func (d *Dao) MongoLeaveGroup(ctx context.Context, userid, groupid primitive.ObjectID) error {
	filter := bson.M{"main": groupid, "main_type": "group", "sub": userid, "sub_type": "user"}
	_, e := d.mongo.Database("im").Collection("relation").DeleteOne(ctx, filter)
	if e != nil {
		return e
	}
	filter["main"] = userid
	filter["main_type"] = "user"
	filter["sub"] = groupid
	filter["sub_type"] = "group"
	_, e = d.mongo.Database("im").Collection("relation").DeleteOne(ctx, filter)
	return e
}

// user's relation should be deleted first,because the group will retry
func (d *Dao) MongoKickGroup(ctx context.Context, groupid, userid primitive.ObjectID) error {
	filter := bson.M{"main": userid, "main_type": "user", "sub": groupid, "sub_type": "group"}
	_, e := d.mongo.Database("im").Collection("relation").DeleteOne(ctx, filter)
	if e != nil {
		return e
	}
	filter["main"] = groupid
	filter["main_type"] = "group"
	filter["sub"] = userid
	filter["sub_type"] = "user"
	_, e = d.mongo.Database("im").Collection("relation").DeleteOne(ctx, filter)
	return e
}

func (d *Dao) MongoCreateGroup(ctx context.Context, groupname string, owner primitive.ObjectID, starters ...primitive.ObjectID) (primitive.ObjectID, error) {
	if groupname == "" {
		return primitive.NilObjectID, ecode.ErrReq
	}
	or := bson.A{}
	if !owner.IsZero() {
		or = append(or, bson.M{"main": owner, "main_type": "user", "sub": primitive.NilObjectID, "sub_type": "user"})
	}
	for _, starter := range starters {
		if starter.IsZero() {
			continue
		}
		or = append(or, bson.M{"main": starter, "main_type": "user", "sub": primitive.NilObjectID, "sub_type": "user"})
	}
	if len(or) == 0 {
		return primitive.NilObjectID, ecode.ErrReq
	}
	cursor, e := d.mongo.Database("im").Collection("relation").Find(ctx, bson.M{"$or": or})
	if e != nil {
		return primitive.NilObjectID, e
	}
	all := make([]*model.Relation, 0, len(or))
	if e = cursor.All(ctx, &all); e != nil {
		return primitive.NilObjectID, e
	}

	docs := bson.A{&model.Relation{
		Main:     primitive.NilObjectID,
		MainType: "group",
		Sub:      primitive.NilObjectID,
		SubType:  "user",
		Name:     groupname,
		Duty:     1,
	}}
	if !owner.IsZero() {
		var username string
		for _, v := range all {
			if v.Main.Hex() == owner.Hex() {
				username = v.Name
				break
			}
		}
		if username == "" {
			return primitive.NilObjectID, ecode.ErrUserNotExist
		}
		docs = append(docs, &model.Relation{
			Main:     owner,
			MainType: "user",
			Sub:      primitive.NilObjectID,
			SubType:  "group",
			Name:     groupname,
		})
		docs = append(docs, &model.Relation{
			Main:     primitive.NilObjectID,
			MainType: "group",
			Sub:      owner,
			SubType:  "user",
			Name:     username,
			Duty:     2,
		})
	}
	for _, starter := range starters {
		if starter.IsZero() {
			continue
		}
		var username string
		for _, v := range all {
			if v.Main.Hex() == starter.Hex() {
				username = v.Name
				break
			}
		}
		if username == "" {
			return primitive.NilObjectID, ecode.ErrUserNotExist
		}
		docs = append(docs, &model.Relation{
			Main:     starter,
			MainType: "user",
			Sub:      primitive.NilObjectID,
			SubType:  "group",
			Name:     groupname,
		})
		docs = append(docs, &model.Relation{
			Main:     primitive.NilObjectID,
			MainType: "group",
			Sub:      starter,
			SubType:  "user",
			Name:     username,
		})
	}
	for {
		var groupid primitive.ObjectID
		for {
			groupid = primitive.NewObjectID()
			filter := bson.M{"main": groupid, "main_type": "group", "sub": primitive.NilObjectID}
			if count, e := d.mongo.Database("im").Collection("relation").CountDocuments(ctx, filter); e != nil {
				return groupid, e
			} else if count != 0 {
				continue
			}
			break
		}
		for _, v := range docs {
			doc := v.(*model.Relation)
			if doc.MainType == "group" && doc.Main.IsZero() {
				doc.Main = groupid
			}
			if doc.SubType == "group" && doc.Sub.IsZero() {
				doc.Sub = groupid
			}
		}
		if _, e := d.mongo.Database("im").Collection("relation").InsertMany(ctx, docs); e != nil && !mongo.IsDuplicateKeyError(e) {
			return groupid, e
		} else if e != nil {
			//ignore DuplicateKeyError
			continue
		}
		return groupid, nil
	}
}

func (d *Dao) MongoGetUserRelations(ctx context.Context, userid primitive.ObjectID) ([]*model.RelationTarget, error) {
	if userid.IsZero() {
		return nil, ecode.ErrReq
	}
	filter := bson.M{
		"main":      userid,
		"main_type": "user",
	}
	opts := options.Find().SetProjection(bson.M{
		"sub":      1,
		"sub_type": 1,
		"name":     1,
	})
	cursor, e := d.mongo.Database("im").Collection("relation").Find(ctx, filter, opts)
	if e != nil {
		return nil, e
	}
	all := make([]*model.RelationTarget, 0, cursor.RemainingBatchLength())
	if e = cursor.All(ctx, &all); e != nil {
		return nil, e
	}
	if len(all) == 0 {
		return nil, ecode.ErrUserNotExist
	}
	return all, nil
}

func (d *Dao) MongoGetUserName(ctx context.Context, userid primitive.ObjectID) (string, error) {
	if userid.IsZero() {
		return "", ecode.ErrReq
	}
	filter := bson.M{"main": userid, "main_type": "user", "sub": primitive.NilObjectID}
	opts := options.FindOne().SetProjection(bson.M{"name": 1})
	user := &model.Relation{}
	e := d.mongo.Database("im").Collection("relation").FindOne(ctx, filter, opts).Decode(user)
	if e != nil && e == mongo.ErrNoDocuments {
		e = ecode.ErrUserNotExist
	}
	return user.Name, e
}

// user update self's name
func (d *Dao) MongoSetUserName(ctx context.Context, userid primitive.ObjectID, name string) error {
	if userid.IsZero() || name == "" {
		return ecode.ErrReq
	}
	filter := bson.M{"main": userid, "main_type": "user", "sub": primitive.NilObjectID, "sub_type": "user"}
	updater := bson.M{"$set": bson.M{"name": name}}
	opts := options.Update().SetUpsert(true)
	_, e := d.mongo.Database("im").Collection("relation").UpdateOne(ctx, filter, updater, opts)
	return e
}

func (d *Dao) MongoGetGroupMembers(ctx context.Context, groupid primitive.ObjectID) ([]*model.RelationTarget, error) {
	if groupid.IsZero() {
		return nil, ecode.ErrReq
	}
	filter := bson.M{
		"main":      groupid,
		"main_type": "group",
	}
	opts := options.Find().SetProjection(bson.M{
		"sub":      1,
		"sub_type": 1,
		"name":     1,
		"duty":     1,
	})
	cursor, e := d.mongo.Database("im").Collection("relation").Find(ctx, filter, opts)
	if e != nil {
		return nil, e
	}
	all := make([]*model.RelationTarget, 0, cursor.RemainingBatchLength())
	if e = cursor.All(ctx, &all); e != nil {
		return nil, e
	}
	if len(all) == 0 {
		return nil, ecode.ErrGroupNotExist
	}
	return all, nil
}

func (d *Dao) MongoGetGroupName(ctx context.Context, groupid primitive.ObjectID) (string, error) {
	if groupid.IsZero() {
		return "", ecode.ErrReq
	}
	filter := bson.M{"main": groupid, "main_type": "group", "sub": primitive.NilObjectID}
	opts := options.FindOne().SetProjection(bson.M{"name": 1})
	group := &model.Relation{}
	e := d.mongo.Database("im").Collection("relation").FindOne(ctx, filter, opts).Decode(group)
	if e != nil && e == mongo.ErrNoDocuments {
		e = ecode.ErrGroupNotExist
	}
	return group.Name, e
}

// group owner update the group's name
func (d *Dao) MongoSetGroupName(ctx context.Context, groupid primitive.ObjectID, name string) error {
	if groupid.IsZero() || name == "" {
		return ecode.ErrReq
	}
	filter := bson.M{"main": groupid, "main_type": "group", "sub": primitive.NilObjectID, "sub_type": "user"}
	r, e := d.mongo.Database("im").Collection("relation").UpdateOne(ctx, filter, bson.M{"$set": bson.M{"name": name}})
	if e != nil {
		return e
	}
	if r.MatchedCount == 0 {
		return ecode.ErrGroupNotExist
	}
	return nil
}

// user update the target's name in user's view
func (d *Dao) MongoUpdateUserRelationName(ctx context.Context, userid, target primitive.ObjectID, targetType, newname string) (*model.RelationTarget, error) {
	filter := bson.M{"main": userid, "main_type": "user", "sub": target, "sub_type": targetType}
	updater := bson.M{"$set": bson.M{"name": newname}}
	if e := d.mongo.Database("im").Collection("relation").FindOneAndUpdate(ctx, filter, updater).Err(); e != nil {
		if e == mongo.ErrNoDocuments {
			e = ecode.ErrNotFriends
		}
		return nil, e
	}
	return &model.RelationTarget{
		Target:     target,
		TargetType: targetType,
		Name:       newname,
	}, nil
}

// user update self's name in group
func (d *Dao) MongoUpdateNameInGroup(ctx context.Context, userid, groupid primitive.ObjectID, newname string) (*model.RelationTarget, error) {
	filter := bson.M{"main": groupid, "main_type": "group", "sub": userid, "sub_type": "user"}
	updater := bson.M{"$set": bson.M{"name": newname}}
	r := &model.RelationTarget{}
	if e := d.mongo.Database("im").Collection("relation").FindOneAndUpdate(ctx, filter, updater).Decode(r); e != nil {
		if e == mongo.ErrNoDocuments {
			e = ecode.ErrNotInGroup
		}
		return nil, e
	}
	r.Name = newname
	return r, nil
}

// group admin update another user's duty in group
func (d *Dao) MongoUpdateDutyInGroup(ctx context.Context, userid, groupid primitive.ObjectID, newduty uint8) (*model.RelationTarget, error) {
	filter := bson.M{"main": groupid, "main_type": "group", "sub": userid, "sub_type": "user"}
	updater := bson.M{"$set": bson.M{"duty": newduty}}
	r := &model.RelationTarget{}
	if e := d.mongo.Database("im").Collection("relation").FindOneAndUpdate(ctx, filter, updater).Decode(r); e != nil {
		if e == mongo.ErrNoDocuments {
			e = ecode.ErrGroupMemberNotExist
		}
		return nil, e
	}
	r.Duty = newduty
	return r, nil
}

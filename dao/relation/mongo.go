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
func (d *Dao) MongoAcceptMakeFriend(ctx context.Context, userid string, friendid string) (username, friendname string, e error) {
	filter := bson.M{
		"$or": bson.A{
			bson.M{"main": userid, "main_type": "user", "sub": "", "sub_type": "user"},
			bson.M{"main": friendid, "main_type": "user", "sub": "", "sub_type": "user"},
		},
	}
	var cursor *mongo.Cursor
	if cursor, e = d.mongo.Database("im").Collection("relation").Find(ctx, filter); e != nil {
		return
	}
	all := make([]*model.Relation, 0, 2)
	if e = cursor.All(ctx, &all); e != nil {
		return
	}
	if len(all) != 2 {
		e = ecode.ErrDBDataBroken
		return
	}
	if all[0].Main == userid {
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
func (d *Dao) MongoAcceptGroupInvite(ctx context.Context, userid, groupid string) (username, groupname string, e error) {
	filter := bson.M{
		"$or": bson.A{
			bson.M{"main": userid, "main_type": "user", "sub": "", "sub_type": "user"},
			bson.M{"main": groupid, "main_type": "group", "sub": "", "sub_type": "user"},
		},
	}
	var cursor *mongo.Cursor
	if cursor, e = d.mongo.Database("im").Collection("relation").Find(ctx, filter); e != nil {
		return
	}
	all := make([]*model.Relation, 0, 2)
	if e = cursor.All(ctx, &all); e != nil {
		return
	}
	if len(all) != 2 {
		e = ecode.ErrDBDataBroken
		return
	}
	if all[0].Main == userid {
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
func (d *Dao) MongoAcceptGroupApply(ctx context.Context, groupid, userid string) (username, groupname string, e error) {
	filter := bson.M{
		"$or": bson.A{
			bson.M{"main": userid, "main_type": "user", "sub": "", "sub_type": "user"},
			bson.M{"main": groupid, "main_type": "group", "sub": "", "sub_type": "user"},
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
	if all[0].Main == userid {
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
	_, e = d.mongo.Database("im").Collection("relation").UpdateOne(ctx, filter, bson.M{"$set": bson.M{"name": username}}, opts)
	return
}

// friend's relation should be deleted first,because the user will retry
func (d *Dao) MongoDelFriend(ctx context.Context, userid, friendid string) error {
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
func (d *Dao) MongoLeaveGroup(ctx context.Context, userid, groupid string) error {
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
func (d *Dao) MongoKickGroup(ctx context.Context, groupid, userid string) error {
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

func (d *Dao) MongoCreateGroup(ctx context.Context, groupname, owner string, starters ...string) (string, error) {
	if groupname == "" {
		return "", ecode.ErrReq
	}
	or := bson.A{}
	if owner != "" {
		or = append(or, bson.M{"main": owner, "main_type": "user", "sub": "", "sub_type": "user"})
	}
	for _, starter := range starters {
		if starter == "" {
			continue
		}
		or = append(or, bson.M{"main": starter, "main_type": "user", "sub": "", "sub_type": "user"})
	}
	if len(or) == 0 {
		return "", ecode.ErrReq
	}
	cursor, e := d.mongo.Database("im").Collection("relation").Find(ctx, bson.M{"$or": or})
	if e != nil {
		return "", e
	}
	all := make([]*model.Relation, 0, len(or))
	if e = cursor.All(ctx, &all); e != nil {
		return "", e
	}

	docs := bson.A{&model.Relation{
		Main:     "",
		MainType: "group",
		Sub:      "",
		SubType:  "user",
		Name:     groupname,
		Duty:     1,
	}}
	if owner != "" {
		var username string
		for _, v := range all {
			if v.Main == owner {
				username = v.Name
				break
			}
		}
		if username == "" {
			return "", ecode.ErrUserNotExist
		}
		docs = append(docs, &model.Relation{
			Main:     owner,
			MainType: "user",
			Sub:      "",
			SubType:  "group",
			Name:     groupname,
		})
		docs = append(docs, &model.Relation{
			Main:     "",
			MainType: "group",
			Sub:      owner,
			SubType:  "user",
			Name:     username,
			Duty:     2,
		})
	}
	for _, starter := range starters {
		if starter == "" {
			continue
		}
		var username string
		for _, v := range all {
			if v.Main == starter {
				username = v.Name
				break
			}
		}
		if username == "" {
			return "", ecode.ErrUserNotExist
		}
		docs = append(docs, &model.Relation{
			Main:     starter,
			MainType: "user",
			Sub:      "",
			SubType:  "group",
			Name:     groupname,
		})
		docs = append(docs, &model.Relation{
			Main:     "",
			MainType: "group",
			Sub:      starter,
			SubType:  "user",
			Name:     username,
		})
	}
	for {
		var groupid string
		for {
			groupid = primitive.NewObjectID().Hex()
			filter := bson.M{"main": groupid, "main_type": "group", "sub": "", "sub_type": "user"}
			if count, e := d.mongo.Database("im").Collection("relation").CountDocuments(ctx, filter); e != nil {
				return groupid, e
			} else if count != 0 {
				continue
			}
			break
		}
		for _, v := range docs {
			doc := v.(*model.Relation)
			if doc.MainType == "group" && doc.Main == "" {
				doc.Main = groupid
			}
			if doc.SubType == "group" && doc.Sub == "" {
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

func (d *Dao) MongoGetUserRelations(ctx context.Context, userid string) ([]*model.RelationTarget, error) {
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

// user update self's name
func (d *Dao) MongoSetUserName(ctx context.Context, userid, name string) error {
	if userid == "" || name == "" {
		return ecode.ErrReq
	}
	filter := bson.M{"main": userid, "main_type": "user", "sub": "", "sub_type": "user"}
	updater := bson.M{"$set": bson.M{"name": name}}
	opts := options.Update().SetUpsert(true)
	_, e := d.mongo.Database("im").Collection("relation").UpdateOne(ctx, filter, updater, opts)
	return e
}

func (d *Dao) MongoGetGroupMembers(ctx context.Context, groupid string) ([]*model.RelationTarget, error) {
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

// group owner update the group's name
func (d *Dao) MongoSetGroupName(ctx context.Context, groupid, name string) error {
	if groupid == "" || name == "" {
		return ecode.ErrReq
	}
	filter := bson.M{"main": groupid, "main_type": "group", "sub": "", "sub_type": "user"}
	updater := bson.M{"$set": bson.M{"name": name}}
	r, e := d.mongo.Database("im").Collection("relation").UpdateOne(ctx, filter, updater)
	if e != nil {
		return e
	}
	if r.MatchedCount == 0 {
		return ecode.ErrGroupNotExist
	}
	return nil
}

// user update the target's name in user's view
func (d *Dao) MongoUpdateUserRelationName(ctx context.Context, userid, target, targetType, newname string) (*model.RelationTarget, error) {
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
func (d *Dao) MongoUpdateNameInGroup(ctx context.Context, userid, groupid, newname string) (*model.RelationTarget, error) {
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
func (d *Dao) MongoUpdateDutyInGroup(ctx context.Context, userid, groupid string, newduty uint8) (*model.RelationTarget, error) {
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

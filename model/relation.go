package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MainType:user + SubType:user => user's friend
// MainType:user + SubType:group => user joined group
// MainType:group + SubType:user => group's member
type Relation struct {
	Main     primitive.ObjectID `bson:"main"`
	MainType string             `bson:"main_type"` //user or group
	Sub      primitive.ObjectID `bson:"sub"`
	SubType  string             `bson:"sub_type"` //user or group
	Name     string             `bson:"name"`
	Duty     uint8              `bson:"duty"` //this field is used when MainType is group and Sub is not empty,0-normal,1-system owner,2-owner,3-admin
}
type RelationTarget struct {
	Target     primitive.ObjectID `bson:"sub"`      //empty means this is the self
	TargetType string             `bson:"sub_type"` //user or group
	Name       string             `bson:"name"`
	Duty       uint8              `bson:"duty"`
}

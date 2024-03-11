package model

// MainType:user + SubType:user => user's friend
// MainType:user + SubType:group => user joined group
// MainType:group + SubType:user => group's member
type Relation struct {
	Main     string `bson:"main"`
	MainType string `bson:"main_type"` //user or group
	Sub      string `bson:"sub"`       //empty means this is the self
	SubType  string `bson:"sub_type"`  //user or group
	Name     string `bson:"name"`
	Duty     uint8  `bson:"duty"` //this field is used when MainType is group,0-normal,1-system owner,2-owner,3-admin
}
type RelationTarget struct {
	Target     string `bson:"sub"`      //empty means this is the self
	TargetType string `bson:"sub_type"` //user or group
	Name       string `bson:"name"`
	Duty       uint8  `bson:"duty"`
}

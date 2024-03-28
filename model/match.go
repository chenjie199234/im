package model

//key:activity name,value: element1:min people num,element2:max people num
var ActivitiesCN map[string][2]uint8 = map[string][2]uint8{
	"摄影": {2, 6},
	"爬山": {2, 10},
	"露营": {2, 10},
	"喝酒": {2, 6},
	"聚餐": {2, 10},
	"骑行": {2, 10},
	"摩旅": {2, 10},
	"自驾": {2, 30},
}

var CitiesCN map[string][]string = map[string][]string{
	"上海": {"浦东", "徐汇", "长宁", "普陀", "杨浦", "闵行", "松江", "宝山", "嘉定", "青浦", "奉贤", "金山"},
}

package model

// Room 房间模型
type Room struct {
	BaseModel `bson:",inline"`
	Name      string `bson:"name" json:"name"`
	Status    int    `bson:"status" json:"status"`
}

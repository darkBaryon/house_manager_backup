package model

import "go.mongodb.org/mongo-driver/v2/bson"

const (
	CollectionUserAuth = "hs_usr_auth"
)

// UserAuth 对应 hs_usr_auth，微信身份认证绑定记录。
type UserAuth struct {
	CommonFields `bson:",inline"`

	UserID       bson.ObjectID `bson:"user_id" json:"userId"`
	AuthProvider AuthProvider  `bson:"auth_provider" json:"authProvider"`
	OpenID       string        `bson:"openid" json:"openid"`
	UnionID      string        `bson:"unionid" json:"unionid"`
	LastLoginAt  int64         `bson:"last_login_at" json:"lastLoginAt"`
	LastLoginIP  string        `bson:"last_login_ip" json:"lastLoginIp"`
}

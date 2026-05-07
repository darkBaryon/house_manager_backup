package model

import "go.mongodb.org/mongo-driver/v2/bson"

const (
	CollectionUser           = "hs_usr_user"
	CollectionUserAuth       = "hs_usr_auth"
	CollectionUserProfileExt = "hs_usr_profile_ext"

	AuthProviderWechat = "wechat"
	SourceChannelMini  = "miniapp"
)

// User 对应 hs_usr_user，C 端用户主档。
type User struct {
	CommonFields `bson:",inline"`

	Nickname      string `bson:"nickname" json:"nickname"`
	Avatar        string `bson:"avatar" json:"avatar"`
	Phone         string `bson:"phone" json:"phone"`
	City          string `bson:"city" json:"city"`
	SourceChannel string `bson:"source_channel" json:"sourceChannel"`
	LastActiveAt  int64  `bson:"last_active_at" json:"lastActiveAt"`
}

// UserAuth 对应 hs_usr_auth，微信身份认证绑定记录。
type UserAuth struct {
	CommonFields `bson:",inline"`

	UserID       bson.ObjectID `bson:"user_id" json:"userId"`
	AuthProvider string        `bson:"auth_provider" json:"authProvider"`
	OpenID       string        `bson:"openid" json:"openid"`
	UnionID      string        `bson:"unionid" json:"unionid"`
	LastLoginAt  int64         `bson:"last_login_at" json:"lastLoginAt"`
	LastLoginIP  string        `bson:"last_login_ip" json:"lastLoginIp"`
}

// UserProfileExt 对应 hs_usr_profile_ext，用户找房偏好扩展信息。
type UserProfileExt struct {
	CommonFields `bson:",inline"`

	UserID            bson.ObjectID `bson:"user_id" json:"userId"`
	BudgetMin         int           `bson:"budget_min" json:"budgetMin"`
	BudgetMax         int           `bson:"budget_max" json:"budgetMax"`
	PreferredAreas    []string      `bson:"preferred_areas" json:"preferredAreas"`
	PreferredRentMode string        `bson:"preferred_rent_mode" json:"preferredRentMode"`
	MoveInPlan        string        `bson:"move_in_plan" json:"moveInPlan"`
	Remark            string        `bson:"remark" json:"remark"`
}

package user

import "go.mongodb.org/mongo-driver/v2/bson"

type ProfileInput struct {
	UserID bson.ObjectID
}

type UpdateProfileInput struct {
	UserID            bson.ObjectID
	Nickname          string
	Avatar            string
	City              string
	BudgetMin         int
	BudgetMax         int
	PreferredAreas    []string
	PreferredRentMode string
	MoveInPlan        string
	Remark            string
}

type DashboardInput struct {
	UserID bson.ObjectID
}

type Profile struct {
	UserID            string
	Nickname          string
	Avatar            string
	Phone             string
	City              string
	BudgetMin         int
	BudgetMax         int
	PreferredAreas    []string
	PreferredRentMode string
	MoveInPlan        string
	Remark            string
}

type Dashboard struct {
	FavoriteCount           int64
	HistoryCount            int64
	PlanCount               int64
	UnreadNotificationCount int64
}

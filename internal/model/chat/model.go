package chat

import (
	commonmodel "house-manager/internal/model/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	CollectionSession = "hs_chat_session"
	CollectionMessage = "hs_chat_message"
)

const (
	SessionStatusActive = 1
	SessionStatusEnded  = 2
)

const (
	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"
)

type Session struct {
	commonmodel.CommonFields `bson:",inline"`

	UserID        bson.ObjectID `bson:"user_id" json:"userId"`
	SessionStatus int           `bson:"session_status" json:"sessionStatus"`
	StartedAt     int64         `bson:"started_at" json:"startedAt"`
	LastActiveAt  int64         `bson:"last_active_at" json:"lastActiveAt"`
	EndedAt       int64         `bson:"ended_at" json:"endedAt"`
	LastSeq       int           `bson:"last_seq" json:"lastSeq"`
}

type Message struct {
	commonmodel.CommonFields `bson:",inline"`

	SessionID bson.ObjectID `bson:"session_id" json:"sessionId"`
	UserID    bson.ObjectID `bson:"user_id" json:"userId"`
	Seq       int           `bson:"seq" json:"seq"`
	Role      string        `bson:"role" json:"role"`
	Content   string        `bson:"content" json:"content"`
}

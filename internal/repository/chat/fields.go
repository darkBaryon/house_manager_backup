package chat

import "house-manager/internal/repository/common"

const (
	fieldID            common.Field = "_id"
	fieldSessionID     common.Field = "session_id"
	fieldUserID        common.Field = "user_id"
	fieldSessionStatus common.Field = "session_status"
	fieldStatus        common.Field = "status"
	fieldLastActiveAt  common.Field = "last_active_at"
	fieldUpdatedAt     common.Field = "updated_at"
	fieldLastSeq       common.Field = "last_seq"
	fieldVersion       common.Field = "version"
	fieldEndedAt       common.Field = "ended_at"
	fieldSeq           common.Field = "seq"
)

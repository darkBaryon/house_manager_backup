package chat

import (
	"context"
	"fmt"
	chatmodel "house-manager/internal/model/chat"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SessionRepository struct {
	*common.Repository[chatmodel.Session]
}

func NewSessionRepository(client *dbmongo.Client) *SessionRepository {
	return &SessionRepository{
		Repository: common.NewRepository[chatmodel.Session](client.Collection(chatmodel.CollectionSession)),
	}
}

func (r *SessionRepository) Create(ctx context.Context, entity *chatmodel.Session) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create chat session: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *SessionRepository) FindActiveByUserID(ctx context.Context, userID bson.ObjectID, minLastActiveAt int64) (*chatmodel.Session, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("find active chat session: userID is required")
	}
	filter := bson.M{
		string(fieldUserID):        userID,
		string(fieldSessionStatus): chatmodel.SessionStatusActive,
		string(fieldStatus):        commonmodel.StatusActive,
	}
	if minLastActiveAt > 0 {
		filter[string(fieldLastActiveAt)] = bson.M{"$gte": minLastActiveAt}
	}
	sessions, err := r.FindMany(ctx, filter, options.Find().SetSort(bson.D{{Key: string(fieldLastActiveAt), Value: -1}}).SetLimit(1))
	if err != nil {
		return nil, fmt.Errorf("find active chat session: %w", err)
	}
	if len(sessions) == 0 {
		return nil, nil
	}
	return &sessions[0], nil
}

func (r *SessionRepository) TouchActive(ctx context.Context, sessionID bson.ObjectID, lastActiveAt int64) error {
	if sessionID.IsZero() {
		return fmt.Errorf("touch active chat session: sessionID is required")
	}
	if lastActiveAt <= 0 {
		return fmt.Errorf("touch active chat session: lastActiveAt is required")
	}
	now := time.Now().Unix()
	update := common.NewUpdateDoc().
		Set(fieldLastActiveAt, lastActiveAt).
		Set(fieldUpdatedAt, now).
		Inc(fieldVersion, 1)
	matched, err := r.UpdateOneBy(ctx, common.And(
		common.Eq(fieldID, sessionID),
		common.Eq(fieldSessionStatus, chatmodel.SessionStatusActive),
		common.Active(),
	), update)
	if err != nil {
		return fmt.Errorf("touch active chat session: %w", err)
	}
	if !matched {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *SessionRepository) End(ctx context.Context, sessionID bson.ObjectID, endedAt int64) error {
	if sessionID.IsZero() {
		return fmt.Errorf("end chat session: sessionID is required")
	}
	if endedAt <= 0 {
		return fmt.Errorf("end chat session: endedAt is required")
	}
	now := time.Now().Unix()
	update := common.NewUpdateDoc().
		Set(fieldSessionStatus, chatmodel.SessionStatusEnded).
		Set(fieldEndedAt, endedAt).
		Set(fieldUpdatedAt, now).
		Inc(fieldVersion, 1)
	matched, err := r.UpdateOneBy(ctx, common.And(
		common.Eq(fieldID, sessionID),
		common.Active(),
	), update)
	if err != nil {
		return fmt.Errorf("end chat session: %w", err)
	}
	if !matched {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *SessionRepository) AllocateMessageSeq(ctx context.Context, sessionID bson.ObjectID) (int, error) {
	if sessionID.IsZero() {
		return 0, fmt.Errorf("allocate chat message seq: sessionID is required")
	}
	now := time.Now().Unix()
	update := common.NewUpdateDoc().
		Set(fieldUpdatedAt, now).
		Inc(fieldLastSeq, 1).
		Inc(fieldVersion, 1)
	session, err := r.FindOneAndUpdateBy(ctx, common.And(
		common.Eq(fieldID, sessionID),
		common.Eq(fieldSessionStatus, chatmodel.SessionStatusActive),
		common.Active(),
	), update, common.ReturnAfter())
	if err != nil {
		return 0, fmt.Errorf("allocate chat message seq: %w", err)
	}
	if session == nil {
		return 0, fmt.Errorf("allocate chat message seq: %w", mongo.ErrNoDocuments)
	}
	if session.LastSeq <= 0 {
		return 0, fmt.Errorf("allocate chat message seq: invalid allocated seq %d", session.LastSeq)
	}
	return session.LastSeq, nil
}

func (r *SessionRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex("user_id_1_session_status_1_last_active_at_-1",
			common.IndexKey(fieldUserID, common.SortAsc),
			common.IndexKey(fieldSessionStatus, common.SortAsc),
			common.IndexKey(fieldLastActiveAt, common.SortDesc),
		),
		common.NewIndex("user_id_1_last_active_at_-1",
			common.IndexKey(fieldUserID, common.SortAsc),
			common.IndexKey(fieldLastActiveAt, common.SortDesc),
		),
	); err != nil {
		return fmt.Errorf("ensure chat session indexes: %w", err)
	}
	return nil
}

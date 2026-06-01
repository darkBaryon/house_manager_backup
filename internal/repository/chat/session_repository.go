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
		"user_id":        userID,
		"session_status": chatmodel.SessionStatusActive,
		"status":         commonmodel.StatusActive,
	}
	if minLastActiveAt > 0 {
		filter["last_active_at"] = bson.M{"$gte": minLastActiveAt}
	}
	var session chatmodel.Session
	if err := r.Collection.FindOne(ctx, filter, options.FindOne().SetSort(bson.D{{Key: "last_active_at", Value: -1}})).Decode(&session); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("find active chat session: %w", err)
	}
	return &session, nil
}

func (r *SessionRepository) TouchActive(ctx context.Context, sessionID bson.ObjectID, lastActiveAt int64) error {
	if sessionID.IsZero() {
		return fmt.Errorf("touch active chat session: sessionID is required")
	}
	if lastActiveAt <= 0 {
		return fmt.Errorf("touch active chat session: lastActiveAt is required")
	}
	now := time.Now().Unix()
	update := bson.M{
		"$set": bson.M{
			"last_active_at": lastActiveAt,
			"updated_at":     now,
		},
		"$inc": bson.M{"version": 1},
	}
	res, err := r.Collection.UpdateOne(ctx, bson.M{
		"_id":            sessionID,
		"session_status": chatmodel.SessionStatusActive,
		"status":         commonmodel.StatusActive,
	}, update)
	if err != nil {
		return fmt.Errorf("touch active chat session: %w", err)
	}
	if res.MatchedCount == 0 {
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
	update := bson.M{
		"$set": bson.M{
			"session_status": chatmodel.SessionStatusEnded,
			"ended_at":       endedAt,
			"updated_at":     now,
		},
		"$inc": bson.M{"version": 1},
	}
	res, err := r.Collection.UpdateOne(ctx, bson.M{
		"_id":    sessionID,
		"status": commonmodel.StatusActive,
	}, update)
	if err != nil {
		return fmt.Errorf("end chat session: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *SessionRepository) AllocateMessageSeq(ctx context.Context, sessionID bson.ObjectID) (int, error) {
	if sessionID.IsZero() {
		return 0, fmt.Errorf("allocate chat message seq: sessionID is required")
	}
	now := time.Now().Unix()
	update := bson.M{
		"$set": bson.M{
			"updated_at": now,
		},
		"$inc": bson.M{
			"last_seq": 1,
			"version":  1,
		},
	}
	var session chatmodel.Session
	if err := r.Collection.FindOneAndUpdate(ctx,
		bson.M{
			"_id":            sessionID,
			"session_status": chatmodel.SessionStatusActive,
			"status":         commonmodel.StatusActive,
		},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&session); err != nil {
		return 0, fmt.Errorf("allocate chat message seq: %w", err)
	}
	if session.LastSeq <= 0 {
		return 0, fmt.Errorf("allocate chat message seq: invalid allocated seq %d", session.LastSeq)
	}
	return session.LastSeq, nil
}

func (r *SessionRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "session_status", Value: 1}, {Key: "last_active_at", Value: -1}},
			Options: options.Index().SetName("user_id_1_session_status_1_last_active_at_-1"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "last_active_at", Value: -1}},
			Options: options.Index().SetName("user_id_1_last_active_at_-1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure chat session indexes: %w", err)
	}
	return nil
}

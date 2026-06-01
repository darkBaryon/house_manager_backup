package chat

import (
	"context"
	"fmt"
	chatmodel "house-manager/internal/model/chat"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MessageRepository struct {
	*common.Repository[chatmodel.Message]
}

func NewMessageRepository(client *dbmongo.Client) *MessageRepository {
	return &MessageRepository{
		Repository: common.NewRepository[chatmodel.Message](client.Collection(chatmodel.CollectionMessage)),
	}
}

func (r *MessageRepository) Create(ctx context.Context, entity *chatmodel.Message) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create chat message: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *MessageRepository) ListRecentBySessionID(ctx context.Context, sessionID bson.ObjectID, limit int) ([]chatmodel.Message, error) {
	if sessionID.IsZero() {
		return nil, fmt.Errorf("list recent chat messages: sessionID is required")
	}
	if limit <= 0 {
		return []chatmodel.Message{}, nil
	}
	items, err := r.FindMany(ctx,
		bson.M{"session_id": sessionID, "status": commonmodel.StatusActive},
		options.Find().SetSort(bson.D{{Key: "seq", Value: -1}}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, fmt.Errorf("list recent chat messages: %w", err)
	}
	reverseMessages(items)
	return items, nil
}

func (r *MessageRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "session_id", Value: 1},
				{Key: "seq", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("session_id_1_seq_1_status_1_unique").SetUnique(true),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure chat message indexes: %w", err)
	}
	return nil
}

func reverseMessages(items []chatmodel.Message) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}

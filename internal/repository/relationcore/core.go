package relationcore

import (
	"context"
	"fmt"
	"time"

	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	fieldStatus    common.Field = "status"
	fieldCreatedAt common.Field = "created_at"
	fieldUpdatedAt common.Field = "updated_at"
	fieldVersion   common.Field = "version"
)

type repository[T any] interface {
	Insert(context.Context, *T) error
	FindManyBy(context.Context, common.Filter, ...common.QueryOption) ([]T, error)
	UpdateManyBy(context.Context, common.Filter, common.UpdateDoc) (int64, error)
	UpsertOneBy(context.Context, common.Filter, common.UpdateDoc) (bool, error)
	DeleteAllBy(context.Context, common.Filter) error
}

type Config[T any] struct {
	LeftField       common.Field
	RightField      common.Field
	AssignedByField common.Field
	AssignedAtField common.Field
	Validate        func(*T) error
}

type Core[T any] struct {
	repo            repository[T]
	leftField       common.Field
	rightField      common.Field
	assignedByField common.Field
	assignedAtField common.Field
	validate        func(*T) error
}

func NewCore[T any](repo *common.Repository[T], cfg Config[T]) *Core[T] {
	return newCore[T](repo, cfg)
}

func newCore[T any](repo repository[T], cfg Config[T]) *Core[T] {
	return &Core[T]{
		repo:            repo,
		leftField:       cfg.LeftField,
		rightField:      cfg.RightField,
		assignedByField: cfg.AssignedByField,
		assignedAtField: cfg.AssignedAtField,
		validate:        cfg.Validate,
	}
}

func (c *Core[T]) CreateMany(ctx context.Context, leftID bson.ObjectID, rightIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	objectIDs := common.CompactObjectIDs(rightIDs)
	if len(objectIDs) == 0 {
		return nil
	}
	now := time.Now().Unix()
	for _, rightID := range objectIDs {
		item, err := c.newEntity(leftID, rightID, assignedBy, now)
		if err != nil {
			return err
		}
		if c.validate != nil {
			if err := c.validate(item); err != nil {
				return err
			}
		}
		if err := c.repo.Insert(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func (c *Core[T]) ReplaceByLeft(ctx context.Context, leftID bson.ObjectID, rightIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	objectIDs := common.CompactObjectIDs(rightIDs)
	now := time.Now().Unix()
	disableFilters := []common.Filter{
		common.Eq(c.leftField, leftID),
		common.Active(),
	}
	if len(objectIDs) > 0 {
		disableFilters = append(disableFilters, common.Nin(c.rightField, objectIDs))
	}
	if _, err := c.repo.UpdateManyBy(ctx, common.And(disableFilters...), common.NewUpdateDoc().
		Set(fieldStatus, commonmodel.StatusDeleted).
		Set(fieldUpdatedAt, now).
		Inc(fieldVersion, 1),
	); err != nil {
		return err
	}

	for _, rightID := range objectIDs {
		if _, err := c.repo.UpsertOneBy(ctx, c.pairFilter(leftID, rightID), common.NewUpdateDoc().
			Set(c.leftField, leftID).
			Set(c.rightField, rightID).
			Set(c.assignedByField, assignedBy).
			Set(c.assignedAtField, now).
			Set(fieldStatus, commonmodel.StatusActive).
			Set(fieldUpdatedAt, now).
			Inc(fieldVersion, 1).
			SetOnInsert(fieldCreatedAt, now),
		); err != nil {
			return err
		}
	}
	return nil
}

func (c *Core[T]) ListActiveByLeft(ctx context.Context, leftID bson.ObjectID) ([]T, error) {
	return c.repo.FindManyBy(ctx, common.And(
		common.Eq(c.leftField, leftID),
		common.Active(),
	))
}

func (c *Core[T]) ListActiveByLefts(ctx context.Context, leftIDs []bson.ObjectID) ([]T, error) {
	objectIDs := common.CompactObjectIDs(leftIDs)
	if len(objectIDs) == 0 {
		return []T{}, nil
	}
	return c.repo.FindManyBy(ctx, common.And(
		common.In(c.leftField, objectIDs),
		common.Active(),
	))
}

func (c *Core[T]) ListActiveByRight(ctx context.Context, rightID bson.ObjectID) ([]T, error) {
	return c.repo.FindManyBy(ctx, common.And(
		common.Eq(c.rightField, rightID),
		common.Active(),
	))
}

func (c *Core[T]) RollbackByLeft(ctx context.Context, leftID bson.ObjectID) error {
	return c.repo.DeleteAllBy(ctx, common.Eq(c.leftField, leftID))
}

func (c *Core[T]) pairFilter(leftID, rightID bson.ObjectID) common.Filter {
	return common.And(
		common.Eq(c.leftField, leftID),
		common.Eq(c.rightField, rightID),
	)
}

func (c *Core[T]) newEntity(leftID, rightID, assignedBy bson.ObjectID, assignedAt int64) (*T, error) {
	doc := bson.M{
		string(c.leftField):       leftID,
		string(c.rightField):      rightID,
		string(c.assignedByField): assignedBy,
		string(c.assignedAtField): assignedAt,
	}
	raw, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshal relation entity: %w", err)
	}
	var entity T
	if err := bson.Unmarshal(raw, &entity); err != nil {
		return nil, fmt.Errorf("unmarshal relation entity: %w", err)
	}
	return &entity, nil
}

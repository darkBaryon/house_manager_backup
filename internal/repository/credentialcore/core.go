package credentialcore

import (
	"context"
	"time"

	authmodel "house-manager/internal/model/auth"
	"house-manager/internal/repository/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	fieldAuthType    common.Field = "auth_type"
	fieldLastLoginAt common.Field = "last_login_at"
	fieldLastLoginIP common.Field = "last_login_ip"
	fieldUpdatedAt   common.Field = "updated_at"
	fieldVersion     common.Field = "version"
)

type repository[T any] interface {
	Insert(context.Context, *T) error
	FindOneBy(context.Context, common.Filter) (*T, error)
	FindManyBy(context.Context, common.Filter, ...common.QueryOption) ([]T, error)
	UpdateByID(context.Context, bson.ObjectID, common.UpdateDoc) error
	DeleteAllBy(context.Context, common.Filter) error
}

type Config[T any] struct {
	OwnerField common.Field
	Normalize  func(*T)
	Validate   func(*T) error
}

type Core[T any] struct {
	repo       repository[T]
	ownerField common.Field
	normalize  func(*T)
	validate   func(*T) error
}

func NewCore[T any](repo *common.Repository[T], cfg Config[T]) *Core[T] {
	return newCore[T](repo, cfg)
}

func newCore[T any](repo repository[T], cfg Config[T]) *Core[T] {
	return &Core[T]{
		repo:       repo,
		ownerField: cfg.OwnerField,
		normalize:  cfg.Normalize,
		validate:   cfg.Validate,
	}
}

func (c *Core[T]) Create(ctx context.Context, entity *T) error {
	if c.normalize != nil {
		c.normalize(entity)
	}
	if c.validate != nil {
		if err := c.validate(entity); err != nil {
			return err
		}
	}
	return c.repo.Insert(ctx, entity)
}

func (c *Core[T]) FindActivePasswordByOwner(ctx context.Context, ownerID bson.ObjectID) (*T, error) {
	return c.repo.FindOneBy(ctx, c.activePasswordFilter(ownerID))
}

func (c *Core[T]) FindActivePasswordByOwners(ctx context.Context, ownerIDs []bson.ObjectID) ([]T, error) {
	objectIDs := common.CompactObjectIDs(ownerIDs)
	if len(objectIDs) == 0 {
		return []T{}, nil
	}
	return c.repo.FindManyBy(ctx, common.And(
		common.In(c.ownerField, objectIDs),
		common.Eq(fieldAuthType, authmodel.PasswordAuthTypePassword),
		common.Active(),
	))
}

func (c *Core[T]) TouchLastLogin(ctx context.Context, authID bson.ObjectID, loginIP string) error {
	now := time.Now().Unix()
	return c.repo.UpdateByID(ctx, authID, common.NewUpdateDoc().
		Set(fieldLastLoginAt, now).
		Set(fieldLastLoginIP, loginIP).
		Set(fieldUpdatedAt, now).
		Inc(fieldVersion, 1))
}

func (c *Core[T]) RollbackCreateByOwner(ctx context.Context, ownerID bson.ObjectID) error {
	return c.repo.DeleteAllBy(ctx, common.Eq(c.ownerField, ownerID))
}

func (c *Core[T]) activePasswordFilter(ownerID bson.ObjectID) common.Filter {
	return common.And(
		common.Eq(c.ownerField, ownerID),
		common.Eq(fieldAuthType, authmodel.PasswordAuthTypePassword),
		common.Active(),
	)
}

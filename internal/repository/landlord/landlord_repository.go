package landlord

import (
	"context"
	"fmt"
	"strings"
	"time"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type LandlordRepository struct {
	*common.Repository[authmodel.Landlord]
}

const (
	landlordFieldID        common.Field = "_id"
	landlordFieldPhone     common.Field = "phone"
	landlordFieldStatus    common.Field = "status"
	landlordFieldCreatedAt common.Field = "created_at"
	landlordFieldUpdatedAt common.Field = "updated_at"
	landlordFieldVersion   common.Field = "version"
)

type ListFilter struct {
	Phone  string
	Status int
	Skip   int64
	Limit  int64
}

func NewLandlordRepository(client *dbmongo.Client) *LandlordRepository {
	return &LandlordRepository{
		Repository: common.NewRepository[authmodel.Landlord](client.Collection(authmodel.CollectionLandlord)),
	}
}

func (r *LandlordRepository) Create(ctx context.Context, landlord *authmodel.Landlord) error {
	if err := landlord.ValidateForCreate(); err != nil {
		return fmt.Errorf("create landlord: %w", err)
	}
	normalizeLandlord(landlord)
	return r.Insert(ctx, landlord)
}

func (r *LandlordRepository) FindActiveByPhone(ctx context.Context, phone string) (*authmodel.Landlord, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, fmt.Errorf("find landlord by phone: phone is required")
	}
	landlord, err := r.FindUniqueBy(ctx, common.And(
		common.Eq(landlordFieldPhone, phone),
		common.Eq(landlordFieldStatus, commonmodel.StatusActive),
	))
	if err != nil {
		if strings.Contains(err.Error(), "multiple documents found") {
			return nil, fmt.Errorf("find landlord by phone: multiple active landlord records found")
		}
		return nil, fmt.Errorf("find landlord by phone: %w", err)
	}
	return landlord, nil
}

func (r *LandlordRepository) FindByPhone(ctx context.Context, phone string) (*authmodel.Landlord, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, fmt.Errorf("find landlord by phone: phone is required")
	}
	landlord, err := r.FindUniqueBy(ctx, common.Eq(landlordFieldPhone, phone))
	if err != nil {
		if strings.Contains(err.Error(), "multiple documents found") {
			return nil, fmt.Errorf("find landlord by phone: multiple landlord records found")
		}
		return nil, fmt.Errorf("find landlord by phone: %w", err)
	}
	return landlord, nil
}

func (r *LandlordRepository) FindActiveByID(ctx context.Context, id bson.ObjectID) (*authmodel.Landlord, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find landlord by id: id is required")
	}
	landlord, err := r.FindOne(ctx, bson.M{
		"_id":    id,
		"status": commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("find landlord by id: %w", err)
	}
	return landlord, nil
}

func (r *LandlordRepository) FindByID(ctx context.Context, id bson.ObjectID) (*authmodel.Landlord, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find landlord by id: id is required")
	}
	landlord, err := r.FindOne(ctx, bson.M{
		"_id": id,
		"status": bson.M{"$in": []int{
			commonmodel.StatusActive,
			commonmodel.StatusDeleted,
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("find landlord by id: %w", err)
	}
	return landlord, nil
}

func (r *LandlordRepository) List(ctx context.Context, input ListFilter) ([]authmodel.Landlord, int64, error) {
	filters := []common.Filter{common.Eq(landlordFieldStatus, input.Status)}
	if input.Phone = strings.TrimSpace(input.Phone); input.Phone != "" {
		filters = append(filters, common.Eq(landlordFieldPhone, input.Phone))
	}
	filter := common.And(filters...)
	total, err := r.CountBy(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count landlords: %w", err)
	}
	opts := []common.QueryOption{
		common.SortBy(landlordFieldCreatedAt, common.SortDesc),
		common.SortBy(landlordFieldID, common.SortDesc),
	}
	if input.Skip > 0 {
		opts = append(opts, common.Skip(input.Skip))
	}
	if input.Limit > 0 {
		opts = append(opts, common.Limit(input.Limit))
	}
	items, err := r.FindManyBy(ctx, filter, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("list landlords: %w", err)
	}
	return items, total, nil
}

func (r *LandlordRepository) UpdateFields(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update landlord fields: id is required")
	}
	if len(fields) == 0 {
		return fmt.Errorf("update landlord fields: fields is required")
	}
	fields = common.CloneBSONMap(fields)
	fields["updated_at"] = time.Now().Unix()
	update := common.NewUpdateDoc().Inc(landlordFieldVersion, 1)
	for key, value := range fields {
		update = update.Set(common.Field(key), value)
	}
	matched, err := r.UpdateOneBy(ctx, common.And(
		common.Eq(landlordFieldID, id),
		common.In(landlordFieldStatus, []int{
			commonmodel.StatusActive,
			commonmodel.StatusDeleted,
		}),
	), update)
	if err != nil {
		return fmt.Errorf("update landlord fields: %w", err)
	}
	if !matched {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *LandlordRepository) RollbackCreate(ctx context.Context, id bson.ObjectID) error {
	if id.IsZero() {
		return fmt.Errorf("rollback landlord create: id is required")
	}
	if err := r.DeleteOneBy(ctx, common.Eq(landlordFieldID, id)); err != nil {
		return fmt.Errorf("rollback landlord create: %w", err)
	}
	return nil
}

func normalizeLandlord(landlord *authmodel.Landlord) {
	if landlord == nil {
		return
	}
	landlord.Phone = strings.TrimSpace(landlord.Phone)
}

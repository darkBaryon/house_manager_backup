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
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type LandlordRepository struct {
	*common.Repository[authmodel.Landlord]
}

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
	items, err := r.FindMany(ctx, bson.M{
		"phone":  phone,
		"status": commonmodel.StatusActive,
	}, options.Find().SetLimit(2))
	if err != nil {
		return nil, fmt.Errorf("find landlord by phone: %w", err)
	}
	if len(items) > 1 {
		return nil, fmt.Errorf("find landlord by phone: multiple active landlord records found")
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (r *LandlordRepository) FindByPhone(ctx context.Context, phone string) (*authmodel.Landlord, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, fmt.Errorf("find landlord by phone: phone is required")
	}
	items, err := r.FindMany(ctx, bson.M{"phone": phone}, options.Find().SetLimit(2))
	if err != nil {
		return nil, fmt.Errorf("find landlord by phone: %w", err)
	}
	if len(items) > 1 {
		return nil, fmt.Errorf("find landlord by phone: multiple landlord records found")
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
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
	filter := bson.M{"status": input.Status}
	if input.Phone = strings.TrimSpace(input.Phone); input.Phone != "" {
		filter["phone"] = input.Phone
	}
	total, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count landlords: %w", err)
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}})
	if input.Skip > 0 {
		opts.SetSkip(input.Skip)
	}
	if input.Limit > 0 {
		opts.SetLimit(input.Limit)
	}
	items, err := r.FindMany(ctx, filter, opts)
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
	fields = cloneBsonM(fields)
	fields["updated_at"] = time.Now().Unix()
	res, err := r.Collection.UpdateOne(ctx, bson.M{
		"_id": id,
		"status": bson.M{"$in": []int{
			commonmodel.StatusActive,
			commonmodel.StatusDeleted,
		}},
	}, bson.M{
		"$set": fields,
		"$inc": bson.M{"version": 1},
	})
	if err != nil {
		return fmt.Errorf("update landlord fields: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *LandlordRepository) RollbackCreate(ctx context.Context, id bson.ObjectID) error {
	if id.IsZero() {
		return fmt.Errorf("rollback landlord create: id is required")
	}
	if _, err := r.Collection.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
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

func cloneBsonM(src bson.M) bson.M {
	if src == nil {
		return nil
	}
	dst := make(bson.M, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

package adm

import (
	"context"
	"fmt"
	"regexp"
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

type StaffRepository struct {
	*common.Repository[authmodel.AdmStaff]
}

type StaffListFilter struct {
	Keyword  string
	Phone    string
	Status   int
	StaffIDs []bson.ObjectID
	Skip     int64
	Limit    int64
}

func NewStaffRepository(client *dbmongo.Client) *StaffRepository {
	return &StaffRepository{
		Repository: common.NewRepository[authmodel.AdmStaff](client.Collection(authmodel.CollectionAdmStaff)),
	}
}

func (r *StaffRepository) Create(ctx context.Context, staff *authmodel.AdmStaff) error {
	normalizeStaff(staff)
	if err := staff.ValidateForCreate(); err != nil {
		return fmt.Errorf("create staff: %w", err)
	}
	return r.Insert(ctx, staff)
}

func (r *StaffRepository) FindByPhone(ctx context.Context, phone string) (*authmodel.AdmStaff, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, fmt.Errorf("find staff by phone: phone is required")
	}
	items, err := r.FindMany(ctx, bson.M{
		"phone": phone,
	}, options.Find().SetLimit(2))
	if err != nil {
		return nil, fmt.Errorf("find staff by phone: %w", err)
	}
	if len(items) > 1 {
		return nil, fmt.Errorf("find staff by phone: multiple staff records found")
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (r *StaffRepository) FindActiveByPhone(ctx context.Context, phone string) (*authmodel.AdmStaff, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, fmt.Errorf("find staff by phone: phone is required")
	}
	items, err := r.FindMany(ctx, bson.M{
		"phone":  phone,
		"status": commonmodel.StatusActive,
	}, options.Find().SetLimit(2))
	if err != nil {
		return nil, fmt.Errorf("find staff by phone: %w", err)
	}
	if len(items) > 1 {
		return nil, fmt.Errorf("find staff by phone: multiple active staff records found")
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (r *StaffRepository) FindActiveByID(ctx context.Context, id bson.ObjectID) (*authmodel.AdmStaff, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find staff by id: id is required")
	}
	staff, err := r.FindOne(ctx, bson.M{
		"_id":    id,
		"status": commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("find staff by id: %w", err)
	}
	return staff, nil
}

func (r *StaffRepository) FindByID(ctx context.Context, id bson.ObjectID) (*authmodel.AdmStaff, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find staff by id: id is required")
	}
	staff, err := r.FindOne(ctx, bson.M{
		"_id": id,
		"status": bson.M{"$in": []int{
			commonmodel.StatusActive,
			commonmodel.StatusDeleted,
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("find staff by id: %w", err)
	}
	return staff, nil
}

func (r *StaffRepository) List(ctx context.Context, input StaffListFilter) ([]authmodel.AdmStaff, int64, error) {
	filter := bson.M{"status": input.Status}
	if input.Phone = strings.TrimSpace(input.Phone); input.Phone != "" {
		filter["phone"] = input.Phone
	}
	if input.Keyword = strings.TrimSpace(input.Keyword); input.Keyword != "" {
		filter["name"] = bson.Regex{Pattern: regexp.QuoteMeta(input.Keyword), Options: "i"}
	}
	if input.StaffIDs != nil {
		staffIDs := compactObjectIDs(input.StaffIDs)
		if len(staffIDs) == 0 {
			return []authmodel.AdmStaff{}, 0, nil
		}
		filter["_id"] = bson.M{"$in": staffIDs}
	}

	total, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count staff list: %w", err)
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
		return nil, 0, fmt.Errorf("list staff: %w", err)
	}
	return items, total, nil
}

func (r *StaffRepository) UpdateFields(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update staff fields: id is required")
	}
	if len(fields) == 0 {
		return fmt.Errorf("update staff fields: fields is required")
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
		return fmt.Errorf("update staff fields: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *StaffRepository) RollbackCreate(ctx context.Context, id bson.ObjectID) error {
	if id.IsZero() {
		return fmt.Errorf("rollback staff create: id is required")
	}
	if _, err := r.Collection.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return fmt.Errorf("rollback staff create: %w", err)
	}
	return nil
}

func normalizeStaff(staff *authmodel.AdmStaff) {
	if staff == nil {
		return
	}
	staff.Name = strings.TrimSpace(staff.Name)
	staff.Phone = strings.TrimSpace(staff.Phone)
	staff.Email = strings.TrimSpace(staff.Email)
	staff.Department = strings.TrimSpace(staff.Department)
	staff.JobTitle = strings.TrimSpace(staff.JobTitle)
	staff.ContactQRCode = strings.TrimSpace(staff.ContactQRCode)
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

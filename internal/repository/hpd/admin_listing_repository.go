package hpd

import (
	"context"
	"fmt"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AdminListingListFilter struct {
	OwnerLandlordID bson.ObjectID
	AssetMode       hpdmodel.HpdAssetMode
	City            string
	District        string
	RoomStatus      *hmdmodel.RoomStatus
	ListingStatus   *hpdmodel.HpdListingStatus
	AuditStatus     *hpdmodel.HpdAuditStatus
	Skip            int64
	Limit           int64
}

type AdminListingRepository struct {
	*common.Repository[hpdmodel.HpdAdminListing]
}

func NewAdminListingRepository(client *dbmongo.Client) *AdminListingRepository {
	return &AdminListingRepository{
		Repository: common.NewRepository[hpdmodel.HpdAdminListing](client.Collection(hpdmodel.CollectionHpdAdminListing)),
	}
}

func (r *AdminListingRepository) Create(ctx context.Context, entity *hpdmodel.HpdAdminListing) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hpd admin listing: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *AdminListingRepository) FindByID(ctx context.Context, id bson.ObjectID) (*hpdmodel.HpdAdminListing, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hpd admin listing by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
}

func (r *AdminListingRepository) FindByListingID(ctx context.Context, listingID bson.ObjectID) (*hpdmodel.HpdAdminListing, error) {
	if listingID.IsZero() {
		return nil, fmt.Errorf("find hpd admin listing by listingID: listingID is required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"listing_id": listingID}))
}

func (r *AdminListingRepository) FindBySource(ctx context.Context, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID) (*hpdmodel.HpdAdminListing, error) {
	if !sourceType.Valid() || sourceID.IsZero() {
		return nil, fmt.Errorf("find hpd admin listing by source: sourceType and sourceID are required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"source_type": sourceType, "source_id": sourceID}))
}

func (r *AdminListingRepository) ListAdmin(ctx context.Context, input AdminListingListFilter) ([]hpdmodel.HpdAdminListing, int64, error) {
	filter, err := adminListingListFilter(input)
	if err != nil {
		return nil, 0, fmt.Errorf("list hpd admin listings: %w", err)
	}
	total, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count hpd admin listings: %w", err)
	}
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}, {Key: "_id", Value: -1}})
	if input.Skip > 0 {
		opts.SetSkip(input.Skip)
	}
	if input.Limit > 0 {
		opts.SetLimit(input.Limit)
	}
	items, err := r.FindMany(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list hpd admin listings: %w", err)
	}
	return items, total, nil
}

func (r *AdminListingRepository) UpsertByListingID(ctx context.Context, entity *hpdmodel.HpdAdminListing) (*hpdmodel.HpdAdminListing, error) {
	if err := entity.ValidateForCreate(); err != nil {
		return nil, fmt.Errorf("upsert hpd admin listing by listingID: %w", err)
	}
	filter := activeFilter(bson.M{"listing_id": entity.ListingID})
	if _, err := r.UpsertFields(ctx, filter, adminListingFields(entity)); err != nil {
		return nil, fmt.Errorf("upsert hpd admin listing by listingID: %w", err)
	}
	return r.FindByListingID(ctx, entity.ListingID)
}

func (r *AdminListingRepository) UpdateProjectionFields(ctx context.Context, listingID bson.ObjectID, fields bson.M) error {
	if listingID.IsZero() {
		return fmt.Errorf("update hpd admin listing projection fields: listingID is required")
	}
	safeFields, err := pickAllowedFields(fields, adminProjectionFields)
	if err != nil {
		return fmt.Errorf("update hpd admin listing projection fields: %w", err)
	}
	if err := hpdmodel.ValidateHpdUpdateFields(safeFields); err != nil {
		return fmt.Errorf("update hpd admin listing projection fields: %w", err)
	}

	setFields := cloneBsonM(safeFields)
	setFields["updated_at"] = time.Now().Unix()
	update := bson.M{
		"$set": setFields,
		"$inc": bson.M{"version": 1},
	}
	res, err := r.Collection.UpdateOne(ctx, activeFilter(bson.M{"listing_id": listingID}), update)
	if err != nil {
		return fmt.Errorf("update hpd admin listing projection fields: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func adminListingListFilter(input AdminListingListFilter) (bson.M, error) {
	if input.AssetMode != "" && !input.AssetMode.Valid() {
		return nil, fmt.Errorf("assetMode is invalid")
	}
	if input.RoomStatus != nil && !input.RoomStatus.Valid() {
		return nil, fmt.Errorf("roomStatus is invalid")
	}
	if input.ListingStatus != nil && !input.ListingStatus.Valid() {
		return nil, fmt.Errorf("listingStatus is invalid")
	}
	if input.AuditStatus != nil && !input.AuditStatus.Valid() {
		return nil, fmt.Errorf("auditStatus is invalid")
	}

	fields := bson.M{}
	if !input.OwnerLandlordID.IsZero() {
		fields["owner_landlord_id"] = input.OwnerLandlordID
	}
	if input.AssetMode != "" {
		fields["asset_mode"] = input.AssetMode
	}
	if city := strings.TrimSpace(input.City); city != "" {
		fields["city"] = city
	}
	if district := strings.TrimSpace(input.District); district != "" {
		fields["district"] = district
	}
	if input.RoomStatus != nil {
		fields["room_status"] = *input.RoomStatus
	}
	if input.ListingStatus != nil {
		fields["listing_status"] = *input.ListingStatus
	}
	if input.AuditStatus != nil {
		fields["audit_status"] = *input.AuditStatus
	}
	return activeFilter(fields), nil
}

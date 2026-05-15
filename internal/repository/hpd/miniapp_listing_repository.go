package hpd

import (
	"context"
	"fmt"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MiniappListingSearchFilter struct {
	City         string
	District     string
	BizArea      string
	RentMode     hmdmodel.RentMode
	AssetMode    hpdmodel.HpdAssetMode
	Keyword      string
	FeatureFlags []string
	PriceMin     int
	PriceMax     int
	Skip         int64
	Limit        int64
}

type MiniappListingRepository struct {
	*common.Repository[hpdmodel.HpdMiniappListing]
}

func NewMiniappListingRepository(client *dbmongo.Client) *MiniappListingRepository {
	return &MiniappListingRepository{
		Repository: common.NewRepository[hpdmodel.HpdMiniappListing](client.Collection(hpdmodel.CollectionHpdMiniappListing)),
	}
}

func (r *MiniappListingRepository) Create(ctx context.Context, entity *hpdmodel.HpdMiniappListing) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hpd miniapp listing: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *MiniappListingRepository) FindByID(ctx context.Context, id bson.ObjectID) (*hpdmodel.HpdMiniappListing, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hpd miniapp listing by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
}

func (r *MiniappListingRepository) FindByListingID(ctx context.Context, listingID bson.ObjectID) (*hpdmodel.HpdMiniappListing, error) {
	if listingID.IsZero() {
		return nil, fmt.Errorf("find hpd miniapp listing by listingID: listingID is required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"listing_id": listingID}))
}

func (r *MiniappListingRepository) FindBySource(ctx context.Context, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID) (*hpdmodel.HpdMiniappListing, error) {
	if !sourceType.Valid() || sourceID.IsZero() {
		return nil, fmt.Errorf("find hpd miniapp listing by source: sourceType and sourceID are required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"source_type": sourceType, "source_id": sourceID}))
}

func (r *MiniappListingRepository) FindOnlineDetail(ctx context.Context, listingID bson.ObjectID) (*hpdmodel.HpdMiniappListing, error) {
	if listingID.IsZero() {
		return nil, fmt.Errorf("find hpd miniapp online detail: listingID is required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{
		"listing_id": listingID,
		"is_online":  hpdmodel.HpdOnlineStatusYes,
	}))
}

func (r *MiniappListingRepository) FindOnlineByListingIDs(ctx context.Context, listingIDs []bson.ObjectID) ([]hpdmodel.HpdMiniappListing, error) {
	if len(listingIDs) == 0 {
		return []hpdmodel.HpdMiniappListing{}, nil
	}
	ids := make([]bson.ObjectID, 0, len(listingIDs))
	for _, id := range listingIDs {
		if !id.IsZero() {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return []hpdmodel.HpdMiniappListing{}, nil
	}
	return r.FindMany(ctx, activeFilter(bson.M{
		"listing_id": bson.M{"$in": ids},
		"is_online":  hpdmodel.HpdOnlineStatusYes,
	}))
}

func (r *MiniappListingRepository) UpsertByListingID(ctx context.Context, entity *hpdmodel.HpdMiniappListing) (*hpdmodel.HpdMiniappListing, error) {
	if err := entity.ValidateForCreate(); err != nil {
		return nil, fmt.Errorf("upsert hpd miniapp listing by listingID: %w", err)
	}
	filter := activeFilter(bson.M{"listing_id": entity.ListingID})
	if _, err := r.UpsertFields(ctx, filter, miniappListingFields(entity)); err != nil {
		return nil, fmt.Errorf("upsert hpd miniapp listing by listingID: %w", err)
	}
	return r.FindByListingID(ctx, entity.ListingID)
}

func (r *MiniappListingRepository) UpdateProjectionFields(ctx context.Context, listingID bson.ObjectID, fields bson.M) error {
	if listingID.IsZero() {
		return fmt.Errorf("update hpd miniapp listing projection fields: listingID is required")
	}
	safeFields, err := pickAllowedFields(fields, miniappProjectionFields)
	if err != nil {
		return fmt.Errorf("update hpd miniapp listing projection fields: %w", err)
	}
	if err := hpdmodel.ValidateHpdUpdateFields(safeFields); err != nil {
		return fmt.Errorf("update hpd miniapp listing projection fields: %w", err)
	}

	setFields := cloneBsonM(safeFields)
	setFields["updated_at"] = time.Now().Unix()
	update := bson.M{
		"$set": setFields,
		"$inc": bson.M{"version": 1},
	}
	res, err := r.Collection.UpdateOne(ctx, activeFilter(bson.M{"listing_id": listingID}), update)
	if err != nil {
		return fmt.Errorf("update hpd miniapp listing projection fields: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *MiniappListingRepository) SearchMiniapp(ctx context.Context, search MiniappListingSearchFilter) ([]hpdmodel.HpdMiniappListing, error) {
	filter, err := miniappSearchFilter(search)
	if err != nil {
		return nil, fmt.Errorf("search hpd miniapp listings: %w", err)
	}

	findOptions := options.Find().SetSort(bson.D{
		{Key: "weight_score", Value: -1},
		{Key: "updated_at", Value: -1},
	})
	if search.Skip > 0 {
		findOptions.SetSkip(search.Skip)
	}
	if search.Limit > 0 {
		findOptions.SetLimit(search.Limit)
	}
	return r.FindMany(ctx, filter, findOptions)
}

func (r *MiniappListingRepository) CountMiniapp(ctx context.Context, search MiniappListingSearchFilter) (int64, error) {
	filter, err := miniappSearchFilter(search)
	if err != nil {
		return 0, fmt.Errorf("count hpd miniapp listings: %w", err)
	}

	total, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count hpd miniapp listings: %w", err)
	}
	return total, nil
}

func miniappSearchFilter(search MiniappListingSearchFilter) (bson.M, error) {
	if search.PriceMin < 0 || search.PriceMax < 0 {
		return nil, fmt.Errorf("price range must be non-negative")
	}
	if search.PriceMax > 0 && search.PriceMin > search.PriceMax {
		return nil, fmt.Errorf("priceMin must be less than or equal to priceMax")
	}
	if search.RentMode != "" && !search.RentMode.Valid() {
		return nil, fmt.Errorf("rentMode is invalid")
	}
	if search.AssetMode != "" && !search.AssetMode.Valid() {
		return nil, fmt.Errorf("assetMode is invalid")
	}

	fields := bson.M{"is_online": hpdmodel.HpdOnlineStatusYes}
	if search.City != "" {
		fields["city"] = search.City
	}
	if search.District != "" {
		fields["district"] = search.District
	}
	if search.BizArea != "" {
		fields["biz_area"] = search.BizArea
	}
	if search.RentMode != "" {
		fields["rent_mode"] = search.RentMode
	}
	if search.AssetMode != "" {
		fields["asset_mode"] = search.AssetMode
	}
	if len(search.FeatureFlags) > 0 {
		flags := make([]string, 0, len(search.FeatureFlags))
		for _, flag := range search.FeatureFlags {
			flag = strings.TrimSpace(flag)
			if flag != "" {
				flags = append(flags, flag)
			}
		}
		if len(flags) > 0 {
			fields["feature_flags"] = bson.M{"$all": flags}
		}
	}
	if keyword := strings.TrimSpace(search.Keyword); keyword != "" {
		pattern := regexp.QuoteMeta(keyword)
		fields["$or"] = bson.A{
			bson.M{"title": bson.Regex{Pattern: pattern, Options: "i"}},
			bson.M{"community_name": bson.Regex{Pattern: pattern, Options: "i"}},
			bson.M{"building_or_community_name": bson.Regex{Pattern: pattern, Options: "i"}},
			bson.M{"address_text": bson.Regex{Pattern: pattern, Options: "i"}},
			bson.M{"subway_station": bson.Regex{Pattern: pattern, Options: "i"}},
		}
	}
	if search.PriceMin > 0 || search.PriceMax > 0 {
		priceFilter := bson.M{}
		if search.PriceMin > 0 {
			priceFilter["$gte"] = search.PriceMin
		}
		if search.PriceMax > 0 {
			priceFilter["$lte"] = search.PriceMax
		}
		fields["price"] = priceFilter
	}
	return activeFilter(fields), nil
}

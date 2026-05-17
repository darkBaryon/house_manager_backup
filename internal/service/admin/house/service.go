package house

import (
	"context"
	"fmt"
	"strings"

	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	hpdrepo "house-manager/internal/repository/hpd"
	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	adminListings adminListingRepository
}

type adminListingRepository interface {
	ListAdmin(ctx context.Context, input hpdrepo.AdminListingListFilter) ([]hpdmodel.HpdAdminListing, int64, error)
	FindByListingID(ctx context.Context, listingID bson.ObjectID) (*hpdmodel.HpdAdminListing, error)
}

func NewService(adminListings *hpdrepo.AdminListingRepository) *Service {
	return newService(adminListings)
}

func newService(adminListings adminListingRepository) *Service {
	return &Service{adminListings: adminListings}
}

func (s *Service) List(ctx context.Context, input ListInput) (*ListResult, error) {
	normalized, filter, err := normalizeListInput(input)
	if err != nil {
		return nil, err
	}
	if s == nil || s.adminListings == nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取房源列表失败，请稍后重试", fmt.Errorf("后台房源服务未正确初始化"))
	}

	items, total, err := s.adminListings.ListAdmin(ctx, hpdrepo.AdminListingListFilter{
		OwnerLandlordID: filter.ProviderID,
		AssetMode:       filter.AssetMode,
		City:            normalized.City,
		District:        normalized.District,
		RoomStatus:      filter.RoomStatus,
		ListingStatus:   filter.ListingStatus,
		AuditStatus:     filter.AuditStatus,
		Skip:            int64((normalized.Page - 1) * normalized.PageSize),
		Limit:           int64(normalized.PageSize),
	})
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取房源列表失败，请稍后重试", err)
	}
	return &ListResult{
		List:     toListItems(items),
		Page:     normalized.Page,
		PageSize: normalized.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) Detail(ctx context.Context, input DetailInput) (*DetailResult, error) {
	listingID, err := parseListingID(input.ListingID)
	if err != nil {
		return nil, err
	}
	if s == nil || s.adminListings == nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取房源详情失败，请稍后重试", fmt.Errorf("后台房源服务未正确初始化"))
	}

	item, err := s.adminListings.FindByListingID(ctx, listingID)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取房源详情失败，请稍后重试", err)
	}
	if item == nil {
		return nil, errcode.NotFound.WithError(fmt.Errorf("房源不存在或已删除"))
	}
	return &DetailResult{House: toHouseSummary(*item)}, nil
}

type normalizedListFilter struct {
	ProviderID    bson.ObjectID
	AssetMode     hpdmodel.HpdAssetMode
	RoomStatus    *hmdmodel.RoomStatus
	ListingStatus *hpdmodel.HpdListingStatus
	AuditStatus   *hpdmodel.HpdAuditStatus
}

func normalizeListInput(input ListInput) (ListInput, normalizedListFilter, error) {
	var filter normalizedListFilter
	input.ProviderID = strings.TrimSpace(input.ProviderID)
	input.AssetMode = strings.TrimSpace(input.AssetMode)
	input.City = strings.TrimSpace(input.City)
	input.District = strings.TrimSpace(input.District)
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	if input.ProviderID != "" {
		id, err := bson.ObjectIDFromHex(input.ProviderID)
		if err != nil || id.IsZero() {
			return input, filter, errcode.InvalidParam.WithError(fmt.Errorf("房源筛选参数不正确"))
		}
		filter.ProviderID = id
	}
	if input.AssetMode != "" {
		filter.AssetMode = hpdmodel.HpdAssetMode(input.AssetMode)
		if !filter.AssetMode.Valid() {
			return input, filter, errcode.InvalidParam.WithError(fmt.Errorf("房源筛选参数不正确"))
		}
	}
	if input.RoomStatus != nil {
		status := hmdmodel.RoomStatus(*input.RoomStatus)
		if !status.Valid() {
			return input, filter, errcode.InvalidParam.WithError(fmt.Errorf("房源筛选参数不正确"))
		}
		filter.RoomStatus = &status
	}
	if input.ListingStatus != nil {
		status := hpdmodel.HpdListingStatus(*input.ListingStatus)
		if !status.Valid() {
			return input, filter, errcode.InvalidParam.WithError(fmt.Errorf("房源筛选参数不正确"))
		}
		filter.ListingStatus = &status
	}
	if input.AuditStatus != nil {
		status := hpdmodel.HpdAuditStatus(*input.AuditStatus)
		if !status.Valid() {
			return input, filter, errcode.InvalidParam.WithError(fmt.Errorf("房源筛选参数不正确"))
		}
		filter.AuditStatus = &status
	}
	return input, filter, nil
}

func parseListingID(value string) (bson.ObjectID, error) {
	id, err := bson.ObjectIDFromHex(strings.TrimSpace(value))
	if err != nil || id.IsZero() {
		return bson.NilObjectID, errcode.InvalidParam.WithError(fmt.Errorf("房源参数不正确"))
	}
	return id, nil
}

func toListItems(items []hpdmodel.HpdAdminListing) []ListItem {
	if len(items) == 0 {
		return []ListItem{}
	}
	result := make([]ListItem, 0, len(items))
	for _, item := range items {
		result = append(result, ListItem{HouseSummary: toHouseSummary(item)})
	}
	return result
}

func toHouseSummary(item hpdmodel.HpdAdminListing) HouseSummary {
	return HouseSummary{
		ListingID:         objectIDHex(item.ListingID),
		SourceType:        string(item.SourceType),
		SourceID:          objectIDHex(item.SourceID),
		AssetMode:         string(item.AssetMode),
		ProviderID:        objectIDHex(item.OwnerLandlordID),
		ProviderPhone:     item.OwnerPhoneSnapshot,
		LandlordName:      item.LandlordNameSnapshot,
		RootType:          string(item.RootType),
		RootID:            objectIDHex(item.RootID),
		ProjectID:         objectIDHex(item.ProjectID),
		ProjectName:       item.ProjectName,
		BuildingID:        objectIDHex(item.BuildingID),
		BuildingName:      item.BuildingName,
		RoomTypeID:        objectIDHex(item.RoomTypeID),
		RoomTypeName:      item.RoomTypeName,
		DecentralizedID:   objectIDHex(item.DecentralizedID),
		CommunityName:     item.CommunityName,
		RentMode:          string(item.RentMode),
		City:              item.City,
		District:          item.District,
		BizArea:           item.BizArea,
		AddressText:       item.AddressText,
		RoomNo:            item.RoomNo,
		Title:             item.Title,
		Price:             item.Price,
		PriceText:         item.PriceText,
		LayoutText:        item.LayoutText,
		AreaSize:          item.AreaSize,
		RoomStatus:        int(item.RoomStatus),
		ListingStatus:     int(item.ListingStatus),
		AuditStatus:       int(item.AuditStatus),
		IsOnline:          int(item.IsOnline),
		LatestAuditTaskID: objectIDHex(item.LatestAuditTaskID),
		LatestSubmittedAt: item.LatestSubmittedAt,
		LatestReviewedAt:  item.LatestReviewedAt,
		ReviewerStaffID:   objectIDHex(item.ReviewerStaffID),
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

func objectIDHex(id bson.ObjectID) string {
	if id.IsZero() {
		return ""
	}
	return id.Hex()
}

func publicSystemError(code int, message string, cause error) *errcode.Error {
	return errcode.New(code, message).WithError(cause)
}

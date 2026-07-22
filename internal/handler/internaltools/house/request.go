package house

import (
	"fmt"
	"strings"

	housesvc "house-manager/internal/service/miniapp/house"
	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const internalToolDefaultPageSize = 5
const internalToolMaxPageSize = 10

type searchRequest struct {
	SessionID string        `json:"session_id" binding:"required"`
	Payload   searchPayload `json:"payload" binding:"required"`
}

type searchPayload struct {
	City          string   `json:"city"`
	District      string   `json:"district"`
	BizArea       string   `json:"biz_area"`
	RentMode      string   `json:"rent_mode"`
	AssetMode     string   `json:"asset_mode"`
	MinPrice      int      `json:"min_price"`
	MaxPrice      int      `json:"max_price"`
	RoomCount     *int     `json:"room_count"`
	HallCount     *int     `json:"hall_count"`
	BathroomCount *int     `json:"bathroom_count"`
	KitchenCount  *int     `json:"kitchen_count"`
	Keyword       string   `json:"keyword"`
	FeatureFlags  []string `json:"feature_flags"`
	Page          int      `json:"page"`
	PageSize      int      `json:"page_size"`
}

func (r searchRequest) validate() error {
	if strings.TrimSpace(r.SessionID) == "" {
		return errcode.InvalidParam.WithError(fmt.Errorf("session_id is required"))
	}
	return nil
}

func (r searchRequest) toServiceInput() housesvc.SearchInput {
	pageSize := r.Payload.PageSize
	if pageSize <= 0 {
		pageSize = internalToolDefaultPageSize
	}
	if pageSize > internalToolMaxPageSize {
		pageSize = internalToolMaxPageSize
	}
	return housesvc.SearchInput{
		City:          strings.TrimSpace(r.Payload.City),
		District:      strings.TrimSpace(r.Payload.District),
		BizArea:       strings.TrimSpace(r.Payload.BizArea),
		RentMode:      strings.TrimSpace(r.Payload.RentMode),
		AssetMode:     strings.TrimSpace(r.Payload.AssetMode),
		MinPrice:      r.Payload.MinPrice,
		MaxPrice:      r.Payload.MaxPrice,
		RoomCount:     r.Payload.RoomCount,
		HallCount:     r.Payload.HallCount,
		BathroomCount: r.Payload.BathroomCount,
		KitchenCount:  r.Payload.KitchenCount,
		Keyword:       strings.TrimSpace(r.Payload.Keyword),
		FeatureFlags:  r.Payload.FeatureFlags,
		Page:          r.Payload.Page,
		PageSize:      pageSize,
	}
}

type publicDetailRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	ListingID string `json:"listing_id" binding:"required"`
}

func (r publicDetailRequest) validate() (bson.ObjectID, error) {
	if strings.TrimSpace(r.SessionID) == "" {
		return bson.NilObjectID, errcode.InvalidParam.WithError(fmt.Errorf("session_id is required"))
	}
	listingID, err := bson.ObjectIDFromHex(strings.TrimSpace(r.ListingID))
	if err != nil {
		return bson.NilObjectID, errcode.InvalidParam.WithError(fmt.Errorf("invalid listing_id %q", r.ListingID))
	}
	return listingID, nil
}

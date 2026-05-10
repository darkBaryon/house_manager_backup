package house

import housesvc "house-manager/internal/service/miniapp/house"

type searchRequest struct {
	City         string   `json:"city"`
	District     string   `json:"district"`
	BizArea      string   `json:"biz_area"`
	RentMode     string   `json:"rent_mode"`
	AssetMode    string   `json:"asset_mode"`
	MinPrice     int      `json:"min_price"`
	MaxPrice     int      `json:"max_price"`
	Keyword      string   `json:"keyword"`
	FeatureFlags []string `json:"feature_flags"`
	Page         int      `json:"page"`
	PageSize     int      `json:"page_size"`
}

func (r searchRequest) toServiceInput() housesvc.SearchInput {
	return housesvc.SearchInput{
		City:         r.City,
		District:     r.District,
		BizArea:      r.BizArea,
		RentMode:     r.RentMode,
		AssetMode:    r.AssetMode,
		MinPrice:     r.MinPrice,
		MaxPrice:     r.MaxPrice,
		Keyword:      r.Keyword,
		FeatureFlags: r.FeatureFlags,
		Page:         r.Page,
		PageSize:     r.PageSize,
	}
}

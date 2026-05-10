package favorite

type listingRequest struct {
	ListingID string `json:"listing_id" binding:"required"`
}

type listRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

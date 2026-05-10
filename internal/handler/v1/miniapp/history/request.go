package history

type addRequest struct {
	ListingID string `json:"listing_id" binding:"required"`
	Source    string `json:"source" binding:"required"`
}

type listRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

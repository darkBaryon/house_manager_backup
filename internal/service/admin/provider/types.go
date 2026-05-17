package provider

type ProviderSummary struct {
	ProviderID        string
	Phone             string
	Status            int
	CreatedByStaffID  string
	UpdatedByStaffID  string
	CreatedAt         int64
	UpdatedAt         int64
	PasswordUpdatedAt int64
	LastLoginAt       int64
	LastLoginIP       string
}

type CreateInput struct {
	OperatorStaffID string
	Phone           string
	Password        string
}

type CreateResult struct {
	Provider ProviderSummary
}

type ListInput struct {
	Phone    string
	Status   *int
	Page     int
	PageSize int
}

type ListItem struct {
	ProviderSummary
}

type ListResult struct {
	List     []ListItem
	Page     int
	PageSize int
	Total    int64
}

type DetailInput struct {
	ProviderID string
}

type DetailResult struct {
	Provider ProviderSummary
}

type UpdateInput struct {
	OperatorStaffID string
	ProviderID      string
	Phone           string
}

type UpdateResult struct {
	Provider ProviderSummary
}

type DisableInput struct {
	OperatorStaffID string
	ProviderID      string
}

type DisableResult struct {
	Success bool
}

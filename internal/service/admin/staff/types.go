package staff

type CreateInput struct {
	OperatorStaffID string
	Name            string
	Phone           string
	Password        string
	Email           string
	Department      string
	JobTitle        string
	ContactQRCode   string
	RoleIDs         []string
}

type RoleSummary struct {
	RoleID   string
	RoleCode string
	RoleName string
}

type StaffSummary struct {
	StaffID       string
	Name          string
	Phone         string
	Email         string
	Department    string
	JobTitle      string
	ContactQRCode string
	Roles         []RoleSummary
	RoleNames     []string
	Status        int
	CreatedAt     int64
	UpdatedAt     int64
}

type CreateResult struct {
	Staff StaffSummary
}

type ListInput struct {
	Keyword  string
	Phone    string
	RoleID   string
	Status   *int
	Page     int
	PageSize int
}

type StaffListItem struct {
	StaffSummary
	LastLoginAt int64
	LastLoginIP string
}

type ListResult struct {
	List     []StaffListItem
	Page     int
	PageSize int
	Total    int64
}

type DetailInput struct {
	StaffID string
}

type StaffDetail struct {
	StaffSummary
	CreatedByStaffID  string
	PasswordUpdatedAt int64
	LastLoginAt       int64
	LastLoginIP       string
}

type DetailResult struct {
	Staff StaffDetail
}

type UpdateInput struct {
	OperatorStaffID string
	StaffID         string
	Name            *string
	Email           *string
	Department      *string
	JobTitle        *string
	ContactQRCode   *string
	Status          *int
	RoleIDs         *[]string
}

type UpdateResult struct {
	Staff StaffDetail
}

type DisableInput struct {
	OperatorStaffID string
	StaffID         string
}

type DisableResult struct {
	Success bool
}

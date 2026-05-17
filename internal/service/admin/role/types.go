package role

type ListInput struct {
	Keyword  string
	Page     int
	PageSize int
}

type PermissionSummary struct {
	PermissionID   string
	PermissionCode string
	PermissionName string
	Module         string
	Action         string
}

type RoleSummary struct {
	RoleID          string
	RoleName        string
	RoleCode        string
	Description     string
	PermissionCodes []string
	PermissionNames []string
	IsSystem        int
	Status          int
	CreatedAt       int64
	UpdatedAt       int64
}

type RoleDetail struct {
	RoleSummary
	Permissions []PermissionSummary
}

type ListItem struct {
	RoleSummary
}

type ListResult struct {
	List     []ListItem
	Page     int
	PageSize int
	Total    int64
}

type DetailInput struct {
	RoleID string
}

type DetailResult struct {
	Role RoleDetail
}

type CreateInput struct {
	OperatorStaffID string
	RoleName        string
	RoleCode        string
	Description     string
	PermissionCodes []string
}

type CreateResult struct {
	Role RoleDetail
}

type UpdateInput struct {
	OperatorStaffID string
	RoleID          string
	RoleName        *string
	Description     *string
	PermissionCodes *[]string
}

type UpdateResult struct {
	Role RoleDetail
}

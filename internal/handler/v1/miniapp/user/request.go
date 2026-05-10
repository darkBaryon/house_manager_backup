package user

type updateProfileRequest struct {
	Nickname          *string   `json:"nickname"`
	Avatar            *string   `json:"avatar"`
	City              *string   `json:"city"`
	BudgetMin         *int      `json:"budget_min"`
	BudgetMax         *int      `json:"budget_max"`
	PreferredAreas    *[]string `json:"preferred_areas"`
	PreferredRentMode *string   `json:"preferred_rent_mode"`
	MoveInPlan        *string   `json:"move_in_plan"`
	Remark            *string   `json:"remark"`
}

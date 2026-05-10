package user

import usersvc "house-manager/internal/service/miniapp/user"

type profileResponse struct {
	UserID            string   `json:"user_id"`
	Nickname          string   `json:"nickname"`
	Avatar            string   `json:"avatar"`
	Phone             string   `json:"phone"`
	City              string   `json:"city"`
	BudgetMin         int      `json:"budget_min"`
	BudgetMax         int      `json:"budget_max"`
	PreferredAreas    []string `json:"preferred_areas"`
	PreferredRentMode string   `json:"preferred_rent_mode"`
	MoveInPlan        string   `json:"move_in_plan"`
	Remark            string   `json:"remark"`
}

type dashboardResponse struct {
	FavoriteCount           int64 `json:"favorite_count"`
	HistoryCount            int64 `json:"history_count"`
	PlanCount               int64 `json:"plan_count"`
	UnreadNotificationCount int64 `json:"unread_notification_count"`
}

func toProfileResponse(profile *usersvc.Profile) profileResponse {
	if profile == nil {
		return profileResponse{PreferredAreas: []string{}}
	}
	return profileResponse{
		UserID:            profile.UserID,
		Nickname:          profile.Nickname,
		Avatar:            profile.Avatar,
		Phone:             profile.Phone,
		City:              profile.City,
		BudgetMin:         profile.BudgetMin,
		BudgetMax:         profile.BudgetMax,
		PreferredAreas:    profile.PreferredAreas,
		PreferredRentMode: profile.PreferredRentMode,
		MoveInPlan:        profile.MoveInPlan,
		Remark:            profile.Remark,
	}
}

func toDashboardResponse(dashboard *usersvc.Dashboard) dashboardResponse {
	if dashboard == nil {
		return dashboardResponse{}
	}
	return dashboardResponse{
		FavoriteCount:           dashboard.FavoriteCount,
		HistoryCount:            dashboard.HistoryCount,
		PlanCount:               dashboard.PlanCount,
		UnreadNotificationCount: dashboard.UnreadNotificationCount,
	}
}

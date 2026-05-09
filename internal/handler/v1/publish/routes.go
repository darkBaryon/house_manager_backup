package publish

import "github.com/gin-gonic/gin"

func (h *PublishHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/publish/create_centralized_project", h.CreateCentralizedProject)
	rg.POST("/publish/centralized_project_detail", h.CentralizedProjectDetail)
	rg.POST("/publish/update_centralized_project", h.UpdateCentralizedProject)
	rg.POST("/publish/list_centralized_projects", h.ListCentralizedProjects)

	rg.POST("/publish/create_building", h.CreateBuilding)
	rg.POST("/publish/building_detail", h.BuildingDetail)
	rg.POST("/publish/update_building", h.UpdateBuilding)
	rg.POST("/publish/list_buildings_by_project", h.ListBuildingsByProject)

	rg.POST("/publish/create_room_type", h.CreateRoomType)
	rg.POST("/publish/room_type_detail", h.RoomTypeDetail)
	rg.POST("/publish/update_room_type", h.UpdateRoomType)
	rg.POST("/publish/list_room_types_by_project", h.ListRoomTypesByProject)
	rg.POST("/publish/list_room_types_by_building", h.ListRoomTypesByBuilding)

	rg.POST("/publish/create_centralized_room", h.CreateCentralizedRoom)
	rg.POST("/publish/centralized_room_detail", h.CentralizedRoomDetail)
	rg.POST("/publish/update_centralized_room", h.UpdateCentralizedRoom)
	rg.POST("/publish/update_centralized_room_status", h.UpdateCentralizedRoomStatus)
	rg.POST("/publish/list_centralized_rooms_by_project", h.ListCentralizedRoomsByProject)
	rg.POST("/publish/list_centralized_rooms_by_building", h.ListCentralizedRoomsByBuilding)

	rg.POST("/publish/create_decentralized_community", h.CreateDecentralizedCommunity)
	rg.POST("/publish/decentralized_community_detail", h.DecentralizedCommunityDetail)
	rg.POST("/publish/update_decentralized_community", h.UpdateDecentralizedCommunity)
	rg.POST("/publish/list_decentralized_communities", h.ListDecentralizedCommunities)
	rg.POST("/publish/list_decentralized_communities_by_city", h.ListDecentralizedCommunities)
	rg.POST("/publish/list_decentralized_communities_by_district", h.ListDecentralizedCommunities)

	rg.POST("/publish/create_decentralized_room", h.CreateDecentralizedRoom)
	rg.POST("/publish/decentralized_room_detail", h.DecentralizedRoomDetail)
	rg.POST("/publish/update_decentralized_room", h.UpdateDecentralizedRoom)
	rg.POST("/publish/update_decentralized_room_status", h.UpdateDecentralizedRoomStatus)
	rg.POST("/publish/list_decentralized_rooms_by_community", h.ListDecentralizedRoomsByCommunity)
}

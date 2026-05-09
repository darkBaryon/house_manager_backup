package publish

import (
	"context"
	"fmt"
	"strings"

	"house-manager/internal/model"
	repohmd "house-manager/internal/repository/hmd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type HmdService struct {
	centralizedRepo         *repohmd.CentralizedRepository
	buildingRepo            *repohmd.BuildingRepository
	decentralizedRepo       *repohmd.DecentralizedRepository
	roomTypeCentralizedRepo *repohmd.RoomTypeCentralizedRepository
	roomCentralizedRepo     *repohmd.RoomCentralizedRepository
	roomDecentralizedRepo   *repohmd.RoomDecentralizedRepository
}

func NewHmdService(
	centralizedRepo *repohmd.CentralizedRepository,
	buildingRepo *repohmd.BuildingRepository,
	decentralizedRepo *repohmd.DecentralizedRepository,
	roomTypeCentralizedRepo *repohmd.RoomTypeCentralizedRepository,
	roomCentralizedRepo *repohmd.RoomCentralizedRepository,
	roomDecentralizedRepo *repohmd.RoomDecentralizedRepository,
) *HmdService {
	return &HmdService{
		centralizedRepo:         centralizedRepo,
		buildingRepo:            buildingRepo,
		decentralizedRepo:       decentralizedRepo,
		roomTypeCentralizedRepo: roomTypeCentralizedRepo,
		roomCentralizedRepo:     roomCentralizedRepo,
		roomDecentralizedRepo:   roomDecentralizedRepo,
	}
}

func (s *HmdService) CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*model.HmdCentralized, error) {
	entity := &model.HmdCentralized{
		ProjectName: strings.TrimSpace(input.ProjectName),
		ProjectCode: strings.TrimSpace(input.ProjectCode),
		City:        strings.TrimSpace(input.City),
		District:    strings.TrimSpace(input.District),
		AddressText: strings.TrimSpace(input.AddressText),
		Geo:         toGeoPoint(input.Geo),
		BrandName:   strings.TrimSpace(input.BrandName),
	}

	existing, err := s.centralizedRepo.FindByProjectCode(ctx, entity.ProjectCode)
	if err != nil {
		return nil, fmt.Errorf("create centralized project: find existing project code: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("create centralized project: projectCode already exists")
	}

	if err := s.centralizedRepo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("create centralized project: %w", err)
	}
	return entity, nil
}

func (s *HmdService) GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error) {
	return s.requireCentralizedProject(ctx, id, "get centralized project")
}

func (s *HmdService) ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]model.HmdCentralized, error) {
	city := strings.TrimSpace(input.City)
	district := strings.TrimSpace(input.District)
	if city == "" {
		return nil, fmt.Errorf("list centralized projects: city is required")
	}
	if district != "" {
		return s.centralizedRepo.ListByCityAndDistrict(ctx, city, district)
	}
	return s.centralizedRepo.ListByCity(ctx, city)
}

func (s *HmdService) UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*model.HmdCentralized, error) {
	project, err := s.requireCentralizedProject(ctx, input.ID, "update centralized project")
	if err != nil {
		return nil, err
	}

	fields := bsonFields(
		"project_name", strings.TrimSpace(input.ProjectName),
		"city", strings.TrimSpace(input.City),
		"district", strings.TrimSpace(input.District),
		"address_text", strings.TrimSpace(input.AddressText),
		"geo", toGeoPoint(input.Geo),
		"brand_name", strings.TrimSpace(input.BrandName),
	)
	if err := s.centralizedRepo.UpdateBaseInfo(ctx, project.ID, fields); err != nil {
		return nil, fmt.Errorf("update centralized project: %w", err)
	}
	return s.requireCentralizedProject(ctx, project.ID, "update centralized project")
}

func (s *HmdService) CreateBuilding(ctx context.Context, input CreateBuildingInput) (*model.HmdBuilding, error) {
	project, err := s.requireCentralizedProject(ctx, input.ProjectID, "create building")
	if err != nil {
		return nil, err
	}

	entity := &model.HmdBuilding{
		ProjectID:         project.ID,
		BuildingName:      strings.TrimSpace(input.BuildingName),
		BuildingCode:      strings.TrimSpace(input.BuildingCode),
		FloorTotal:        input.FloorTotal,
		ManagerName:       strings.TrimSpace(input.ManagerName),
		ManagerPhone:      strings.TrimSpace(input.ManagerPhone),
		Photos:            cloneStringSlice(input.Photos),
		ListingFacilities: cloneStringSlice(input.ListingFacilities),
	}

	existing, err := s.buildingRepo.FindByBuildingCode(ctx, entity.BuildingCode)
	if err != nil {
		return nil, fmt.Errorf("create building: find existing building code: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("create building: buildingCode already exists")
	}

	if err := s.buildingRepo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("create building: %w", err)
	}
	return entity, nil
}

func (s *HmdService) GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error) {
	return s.requireBuilding(ctx, id, "get building")
}

func (s *HmdService) ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error) {
	if _, err := s.requireCentralizedProject(ctx, projectID, "list buildings by project"); err != nil {
		return nil, err
	}
	return s.buildingRepo.ListByProjectID(ctx, projectID)
}

func (s *HmdService) UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*model.HmdBuilding, error) {
	building, err := s.requireBuilding(ctx, input.ID, "update building")
	if err != nil {
		return nil, err
	}

	fields := bsonFields(
		"building_name", strings.TrimSpace(input.BuildingName),
		"floor_total", input.FloorTotal,
		"manager_name", strings.TrimSpace(input.ManagerName),
		"manager_phone", strings.TrimSpace(input.ManagerPhone),
		"photos", cloneStringSlice(input.Photos),
		"listing_facilities", cloneStringSlice(input.ListingFacilities),
	)
	if err := s.buildingRepo.UpdateBaseInfo(ctx, building.ID, fields); err != nil {
		return nil, fmt.Errorf("update building: %w", err)
	}
	return s.requireBuilding(ctx, building.ID, "update building")
}

func (s *HmdService) CreateRoomType(ctx context.Context, input CreateRoomTypeInput) (*model.HmdRoomTypeCentralized, error) {
	projectID, buildingID, err := s.validateRoomTypeOwnership(ctx, input.ProjectID, input.BuildingID, "create room type")
	if err != nil {
		return nil, err
	}

	roomTypeName := strings.TrimSpace(input.RoomTypeName)
	if err := s.ensureRoomTypeNameAvailable(ctx, projectID, buildingID, roomTypeName, bson.NilObjectID, "create room type"); err != nil {
		return nil, err
	}

	entity := &model.HmdRoomTypeCentralized{
		ProjectID:       projectID,
		BuildingID:      buildingID,
		RoomTypeName:    roomTypeName,
		RoomCount:       input.RoomCount,
		HallCount:       input.HallCount,
		BathroomCount:   input.BathroomCount,
		KitchenCount:    input.KitchenCount,
		AreaSize:        input.AreaSize,
		Orientation:     strings.TrimSpace(input.Orientation),
		DecorationLevel: strings.TrimSpace(input.DecorationLevel),
		PaymentCycle:    strings.TrimSpace(input.PaymentCycle),
		Rent:            input.Rent,
		Deposit:         input.Deposit,
		ServiceFee:      input.ServiceFee,
		AgencyFeeMode:   strings.TrimSpace(input.AgencyFeeMode),
		AgencyFeeValue:  input.AgencyFeeValue,
		Images:          toTaggedImages(input.Images),
		RoomFacilities:  cloneStringSlice(input.RoomFacilities),
	}

	if err := s.roomTypeCentralizedRepo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("create room type: %w", err)
	}
	return entity, nil
}

func (s *HmdService) GetRoomType(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error) {
	return s.requireRoomType(ctx, id, "get room type")
}

func (s *HmdService) ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	if _, err := s.requireCentralizedProject(ctx, projectID, "list room types by project"); err != nil {
		return nil, err
	}
	return s.roomTypeCentralizedRepo.ListByProjectID(ctx, projectID)
}

func (s *HmdService) ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	if _, err := s.requireBuilding(ctx, buildingID, "list room types by building"); err != nil {
		return nil, err
	}
	return s.roomTypeCentralizedRepo.ListByBuildingID(ctx, buildingID)
}

func (s *HmdService) UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*model.HmdRoomTypeCentralized, error) {
	roomType, err := s.requireRoomType(ctx, input.ID, "update room type")
	if err != nil {
		return nil, err
	}

	roomTypeName := strings.TrimSpace(input.RoomTypeName)
	if err := s.ensureRoomTypeNameAvailable(ctx, roomType.ProjectID, roomType.BuildingID, roomTypeName, roomType.ID, "update room type"); err != nil {
		return nil, err
	}

	fields := bsonFields(
		"room_type_name", roomTypeName,
		"room_count", input.RoomCount,
		"hall_count", input.HallCount,
		"bathroom_count", input.BathroomCount,
		"kitchen_count", input.KitchenCount,
		"area_size", input.AreaSize,
		"orientation", strings.TrimSpace(input.Orientation),
		"decoration_level", strings.TrimSpace(input.DecorationLevel),
		"payment_cycle", strings.TrimSpace(input.PaymentCycle),
		"rent", input.Rent,
		"deposit", input.Deposit,
		"service_fee", input.ServiceFee,
		"agency_fee_mode", strings.TrimSpace(input.AgencyFeeMode),
		"agency_fee_value", input.AgencyFeeValue,
		"images", toTaggedImages(input.Images),
		"room_facilities", cloneStringSlice(input.RoomFacilities),
	)
	if err := s.roomTypeCentralizedRepo.UpdateBaseInfo(ctx, roomType.ID, fields); err != nil {
		return nil, fmt.Errorf("update room type: %w", err)
	}
	return s.requireRoomType(ctx, roomType.ID, "update room type")
}

func (s *HmdService) CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*model.HmdRoomCentralized, error) {
	project, building, roomType, err := s.validateCentralizedRoomDependencies(ctx, input.ProjectID, input.BuildingID, input.RoomTypeID, "create centralized room")
	if err != nil {
		return nil, err
	}

	roomNo := strings.TrimSpace(input.RoomNo)
	existing, err := s.roomCentralizedRepo.FindByBuildingAndRoomNo(ctx, building.ID, roomNo)
	if err != nil {
		return nil, fmt.Errorf("create centralized room: find existing room: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("create centralized room: roomNo already exists under building")
	}

	entity := &model.HmdRoomCentralized{
		ProjectID:         project.ID,
		BuildingID:        building.ID,
		RoomTypeID:        roomTypeID(roomType),
		RoomNo:            roomNo,
		FloorNo:           input.FloorNo,
		RentMode:          strings.TrimSpace(input.RentMode),
		LayoutText:        strings.TrimSpace(input.LayoutText),
		AreaSize:          input.AreaSize,
		Orientation:       strings.TrimSpace(input.Orientation),
		DecorationLevel:   strings.TrimSpace(input.DecorationLevel),
		PaymentCycle:      strings.TrimSpace(input.PaymentCycle),
		Rent:              input.Rent,
		Deposit:           input.Deposit,
		ServiceFee:        input.ServiceFee,
		AgencyFeeMode:     strings.TrimSpace(input.AgencyFeeMode),
		AgencyFeeValue:    input.AgencyFeeValue,
		ViewingTimeRule:   strings.TrimSpace(input.ViewingTimeRule),
		StartRentRule:     strings.TrimSpace(input.StartRentRule),
		Images:            toTaggedImages(input.Images),
		RoomFacilities:    cloneStringSlice(input.RoomFacilities),
		ListingFacilities: cloneStringSlice(input.ListingFacilities),
	}

	if err := s.roomCentralizedRepo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("create centralized room: %w", err)
	}
	return entity, nil
}

func (s *HmdService) GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error) {
	return s.requireCentralizedRoom(ctx, id, "get centralized room")
}

func (s *HmdService) ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	if _, err := s.requireCentralizedProject(ctx, projectID, "list centralized rooms by project"); err != nil {
		return nil, err
	}
	return s.roomCentralizedRepo.ListByProjectID(ctx, projectID)
}

func (s *HmdService) ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	if _, err := s.requireBuilding(ctx, buildingID, "list centralized rooms by building"); err != nil {
		return nil, err
	}
	return s.roomCentralizedRepo.ListByBuildingID(ctx, buildingID)
}

func (s *HmdService) UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*model.HmdRoomCentralized, error) {
	room, err := s.requireCentralizedRoom(ctx, input.ID, "update centralized room")
	if err != nil {
		return nil, err
	}

	roomNo := strings.TrimSpace(input.RoomNo)
	existing, err := s.roomCentralizedRepo.FindByBuildingAndRoomNo(ctx, room.BuildingID, roomNo)
	if err != nil {
		return nil, fmt.Errorf("update centralized room: find existing room: %w", err)
	}
	if existing != nil && existing.ID != room.ID {
		return nil, fmt.Errorf("update centralized room: roomNo already exists under building")
	}

	fields := bsonFields(
		"room_no", roomNo,
		"floor_no", input.FloorNo,
		"rent_mode", strings.TrimSpace(input.RentMode),
		"layout_text", strings.TrimSpace(input.LayoutText),
		"area_size", input.AreaSize,
		"orientation", strings.TrimSpace(input.Orientation),
		"decoration_level", strings.TrimSpace(input.DecorationLevel),
		"payment_cycle", strings.TrimSpace(input.PaymentCycle),
		"rent", input.Rent,
		"deposit", input.Deposit,
		"service_fee", input.ServiceFee,
		"agency_fee_mode", strings.TrimSpace(input.AgencyFeeMode),
		"agency_fee_value", input.AgencyFeeValue,
		"viewing_time_rule", strings.TrimSpace(input.ViewingTimeRule),
		"start_rent_rule", strings.TrimSpace(input.StartRentRule),
		"images", toTaggedImages(input.Images),
		"room_facilities", cloneStringSlice(input.RoomFacilities),
		"listing_facilities", cloneStringSlice(input.ListingFacilities),
	)
	if err := s.roomCentralizedRepo.UpdateBaseInfo(ctx, room.ID, fields); err != nil {
		return nil, fmt.Errorf("update centralized room: %w", err)
	}
	return s.requireCentralizedRoom(ctx, room.ID, "update centralized room")
}

func (s *HmdService) UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*model.HmdRoomCentralized, error) {
	room, err := s.requireCentralizedRoom(ctx, input.ID, "update centralized room status")
	if err != nil {
		return nil, err
	}
	if err := s.roomCentralizedRepo.UpdateStatus(ctx, room.ID, input.RoomStatus); err != nil {
		return nil, fmt.Errorf("update centralized room status: %w", err)
	}
	return s.requireCentralizedRoom(ctx, room.ID, "update centralized room status")
}

func (s *HmdService) CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*model.HmdDecentralized, error) {
	entity := &model.HmdDecentralized{
		CommunityName: strings.TrimSpace(input.CommunityName),
		City:          strings.TrimSpace(input.City),
		District:      strings.TrimSpace(input.District),
		BizArea:       strings.TrimSpace(input.BizArea),
		AddressText:   strings.TrimSpace(input.AddressText),
		Geo:           toGeoPoint(input.Geo),
		SubwayStation: strings.TrimSpace(input.SubwayStation),
	}

	existing, err := s.decentralizedRepo.FindByCommunity(ctx, entity.City, entity.District, entity.CommunityName)
	if err != nil {
		return nil, fmt.Errorf("create decentralized community: find existing community: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("create decentralized community: community already exists in city and district")
	}

	if err := s.decentralizedRepo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("create decentralized community: %w", err)
	}
	return entity, nil
}

func (s *HmdService) GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*model.HmdDecentralized, error) {
	return s.requireDecentralized(ctx, id, "get decentralized community")
}

func (s *HmdService) ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]model.HmdDecentralized, error) {
	city := strings.TrimSpace(input.City)
	district := strings.TrimSpace(input.District)
	if city == "" {
		return nil, fmt.Errorf("list decentralized communities: city is required")
	}
	if district != "" {
		return s.decentralizedRepo.ListByDistrict(ctx, city, district)
	}
	return s.decentralizedRepo.ListByCity(ctx, city)
}

func (s *HmdService) UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*model.HmdDecentralized, error) {
	community, err := s.requireDecentralized(ctx, input.ID, "update decentralized community")
	if err != nil {
		return nil, err
	}

	city := strings.TrimSpace(input.City)
	district := strings.TrimSpace(input.District)
	communityName := strings.TrimSpace(input.CommunityName)
	existing, err := s.decentralizedRepo.FindByCommunity(ctx, city, district, communityName)
	if err != nil {
		return nil, fmt.Errorf("update decentralized community: find existing community: %w", err)
	}
	if existing != nil && existing.ID != community.ID {
		return nil, fmt.Errorf("update decentralized community: community already exists in city and district")
	}

	fields := bsonFields(
		"community_name", communityName,
		"city", city,
		"district", district,
		"biz_area", strings.TrimSpace(input.BizArea),
		"address_text", strings.TrimSpace(input.AddressText),
		"geo", toGeoPoint(input.Geo),
		"subway_station", strings.TrimSpace(input.SubwayStation),
	)
	if err := s.decentralizedRepo.UpdateBaseInfo(ctx, community.ID, fields); err != nil {
		return nil, fmt.Errorf("update decentralized community: %w", err)
	}
	return s.requireDecentralized(ctx, community.ID, "update decentralized community")
}

func (s *HmdService) CreateDecentralizedRoom(ctx context.Context, input CreateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error) {
	community, err := s.requireDecentralized(ctx, input.DecentralizedID, "create decentralized room")
	if err != nil {
		return nil, err
	}

	var roomType *model.HmdRoomTypeCentralized
	if !input.RoomTypeID.IsZero() {
		roomType, err = s.requireRoomType(ctx, input.RoomTypeID, "create decentralized room")
		if err != nil {
			return nil, err
		}
	}

	roomNo := strings.TrimSpace(input.RoomNo)
	existing, err := s.roomDecentralizedRepo.FindByDecentralizedAndRoomNo(ctx, community.ID, roomNo)
	if err != nil {
		return nil, fmt.Errorf("create decentralized room: find existing room: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("create decentralized room: roomNo already exists under community")
	}

	entity := &model.HmdRoomDecentralized{
		DecentralizedID:   community.ID,
		RoomTypeID:        roomTypeID(roomType),
		RoomNo:            roomNo,
		FloorNo:           input.FloorNo,
		RentMode:          strings.TrimSpace(input.RentMode),
		LayoutText:        strings.TrimSpace(input.LayoutText),
		AreaSize:          input.AreaSize,
		Orientation:       strings.TrimSpace(input.Orientation),
		DecorationLevel:   strings.TrimSpace(input.DecorationLevel),
		PaymentCycle:      strings.TrimSpace(input.PaymentCycle),
		Rent:              input.Rent,
		Deposit:           input.Deposit,
		ServiceFee:        input.ServiceFee,
		AgencyFeeMode:     strings.TrimSpace(input.AgencyFeeMode),
		AgencyFeeValue:    input.AgencyFeeValue,
		ViewingTimeRule:   strings.TrimSpace(input.ViewingTimeRule),
		StartRentRule:     strings.TrimSpace(input.StartRentRule),
		Images:            toTaggedImages(input.Images),
		RoomFacilities:    cloneStringSlice(input.RoomFacilities),
		ListingFacilities: cloneStringSlice(input.ListingFacilities),
	}

	if err := s.roomDecentralizedRepo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("create decentralized room: %w", err)
	}
	return entity, nil
}

func (s *HmdService) GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error) {
	return s.requireDecentralizedRoom(ctx, id, "get decentralized room")
}

func (s *HmdService) ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error) {
	if _, err := s.requireDecentralized(ctx, decentralizedID, "list decentralized rooms by community"); err != nil {
		return nil, err
	}
	return s.roomDecentralizedRepo.ListByDecentralizedID(ctx, decentralizedID)
}

func (s *HmdService) UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error) {
	room, err := s.requireDecentralizedRoom(ctx, input.ID, "update decentralized room")
	if err != nil {
		return nil, err
	}

	roomNo := strings.TrimSpace(input.RoomNo)
	existing, err := s.roomDecentralizedRepo.FindByDecentralizedAndRoomNo(ctx, room.DecentralizedID, roomNo)
	if err != nil {
		return nil, fmt.Errorf("update decentralized room: find existing room: %w", err)
	}
	if existing != nil && existing.ID != room.ID {
		return nil, fmt.Errorf("update decentralized room: roomNo already exists under community")
	}

	fields := bsonFields(
		"room_no", roomNo,
		"floor_no", input.FloorNo,
		"rent_mode", strings.TrimSpace(input.RentMode),
		"layout_text", strings.TrimSpace(input.LayoutText),
		"area_size", input.AreaSize,
		"orientation", strings.TrimSpace(input.Orientation),
		"decoration_level", strings.TrimSpace(input.DecorationLevel),
		"payment_cycle", strings.TrimSpace(input.PaymentCycle),
		"rent", input.Rent,
		"deposit", input.Deposit,
		"service_fee", input.ServiceFee,
		"agency_fee_mode", strings.TrimSpace(input.AgencyFeeMode),
		"agency_fee_value", input.AgencyFeeValue,
		"viewing_time_rule", strings.TrimSpace(input.ViewingTimeRule),
		"start_rent_rule", strings.TrimSpace(input.StartRentRule),
		"images", toTaggedImages(input.Images),
		"room_facilities", cloneStringSlice(input.RoomFacilities),
		"listing_facilities", cloneStringSlice(input.ListingFacilities),
	)
	if err := s.roomDecentralizedRepo.UpdateBaseInfo(ctx, room.ID, fields); err != nil {
		return nil, fmt.Errorf("update decentralized room: %w", err)
	}
	return s.requireDecentralizedRoom(ctx, room.ID, "update decentralized room")
}

func (s *HmdService) UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*model.HmdRoomDecentralized, error) {
	room, err := s.requireDecentralizedRoom(ctx, input.ID, "update decentralized room status")
	if err != nil {
		return nil, err
	}
	if err := s.roomDecentralizedRepo.UpdateStatus(ctx, room.ID, input.RoomStatus); err != nil {
		return nil, fmt.Errorf("update decentralized room status: %w", err)
	}
	return s.requireDecentralizedRoom(ctx, room.ID, "update decentralized room status")
}

func (s *HmdService) requireCentralizedProject(ctx context.Context, id bson.ObjectID, action string) (*model.HmdCentralized, error) {
	project, err := s.centralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", action, err)
	}
	if project == nil {
		return nil, fmt.Errorf("%s: centralized project not found", action)
	}
	return project, nil
}

func (s *HmdService) requireBuilding(ctx context.Context, id bson.ObjectID, action string) (*model.HmdBuilding, error) {
	building, err := s.buildingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", action, err)
	}
	if building == nil {
		return nil, fmt.Errorf("%s: building not found", action)
	}
	return building, nil
}

func (s *HmdService) requireRoomType(ctx context.Context, id bson.ObjectID, action string) (*model.HmdRoomTypeCentralized, error) {
	roomType, err := s.roomTypeCentralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", action, err)
	}
	if roomType == nil {
		return nil, fmt.Errorf("%s: room type not found", action)
	}
	return roomType, nil
}

func (s *HmdService) requireCentralizedRoom(ctx context.Context, id bson.ObjectID, action string) (*model.HmdRoomCentralized, error) {
	room, err := s.roomCentralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", action, err)
	}
	if room == nil {
		return nil, fmt.Errorf("%s: centralized room not found", action)
	}
	return room, nil
}

func (s *HmdService) requireDecentralized(ctx context.Context, id bson.ObjectID, action string) (*model.HmdDecentralized, error) {
	community, err := s.decentralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", action, err)
	}
	if community == nil {
		return nil, fmt.Errorf("%s: decentralized community not found", action)
	}
	return community, nil
}

func (s *HmdService) requireDecentralizedRoom(ctx context.Context, id bson.ObjectID, action string) (*model.HmdRoomDecentralized, error) {
	room, err := s.roomDecentralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", action, err)
	}
	if room == nil {
		return nil, fmt.Errorf("%s: decentralized room not found", action)
	}
	return room, nil
}

func (s *HmdService) validateRoomTypeOwnership(ctx context.Context, projectID, buildingID bson.ObjectID, action string) (bson.ObjectID, bson.ObjectID, error) {
	if projectID.IsZero() && buildingID.IsZero() {
		return bson.NilObjectID, bson.NilObjectID, fmt.Errorf("%s: projectID or buildingID is required", action)
	}

	if !projectID.IsZero() {
		if _, err := s.requireCentralizedProject(ctx, projectID, action); err != nil {
			return bson.NilObjectID, bson.NilObjectID, err
		}
	}

	if !buildingID.IsZero() {
		building, err := s.requireBuilding(ctx, buildingID, action)
		if err != nil {
			return bson.NilObjectID, bson.NilObjectID, err
		}
		if !projectID.IsZero() && building.ProjectID != projectID {
			return bson.NilObjectID, bson.NilObjectID, fmt.Errorf("%s: building does not belong to project", action)
		}
		if projectID.IsZero() {
			projectID = building.ProjectID
		}
	}

	return projectID, buildingID, nil
}

func (s *HmdService) validateCentralizedRoomDependencies(ctx context.Context, projectID, buildingID, roomTypeID bson.ObjectID, action string) (*model.HmdCentralized, *model.HmdBuilding, *model.HmdRoomTypeCentralized, error) {
	project, err := s.requireCentralizedProject(ctx, projectID, action)
	if err != nil {
		return nil, nil, nil, err
	}

	building, err := s.requireBuilding(ctx, buildingID, action)
	if err != nil {
		return nil, nil, nil, err
	}
	if building.ProjectID != project.ID {
		return nil, nil, nil, fmt.Errorf("%s: building does not belong to project", action)
	}

	if roomTypeID.IsZero() {
		return project, building, nil, nil
	}

	roomType, err := s.requireRoomType(ctx, roomTypeID, action)
	if err != nil {
		return nil, nil, nil, err
	}
	if !roomType.ProjectID.IsZero() && roomType.ProjectID != project.ID {
		return nil, nil, nil, fmt.Errorf("%s: room type does not belong to project", action)
	}
	if !roomType.BuildingID.IsZero() && roomType.BuildingID != building.ID {
		return nil, nil, nil, fmt.Errorf("%s: room type does not belong to building", action)
	}

	return project, building, roomType, nil
}

func (s *HmdService) ensureRoomTypeNameAvailable(ctx context.Context, projectID, buildingID bson.ObjectID, roomTypeName string, currentID bson.ObjectID, action string) error {
	if roomTypeName == "" {
		return fmt.Errorf("%s: roomTypeName is required", action)
	}

	if !projectID.IsZero() {
		existing, err := s.roomTypeCentralizedRepo.FindByProjectAndName(ctx, projectID, roomTypeName)
		if err != nil {
			return fmt.Errorf("%s: find existing room type by project: %w", action, err)
		}
		if existing != nil && existing.ID != currentID {
			return fmt.Errorf("%s: roomTypeName already exists under project", action)
		}
	}

	if !buildingID.IsZero() {
		roomTypes, err := s.roomTypeCentralizedRepo.ListByBuildingID(ctx, buildingID)
		if err != nil {
			return fmt.Errorf("%s: list room types by building: %w", action, err)
		}
		for _, roomType := range roomTypes {
			if roomType.RoomTypeName == roomTypeName && roomType.ID != currentID {
				return fmt.Errorf("%s: roomTypeName already exists under building", action)
			}
		}
	}

	return nil
}

func roomTypeID(roomType *model.HmdRoomTypeCentralized) bson.ObjectID {
	if roomType == nil {
		return bson.NilObjectID
	}
	return roomType.ID
}

func bsonFields(kv ...any) bson.M {
	fields := make(bson.M, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok || key == "" {
			continue
		}
		fields[key] = kv[i+1]
	}
	return fields
}

func toGeoPoint(input *GeoPointInput) *model.GeoPoint {
	if input == nil {
		return nil
	}
	return &model.GeoPoint{
		Lng: input.Lng,
		Lat: input.Lat,
	}
}

func toTaggedImages(inputs []TaggedImageInput) []model.TaggedImage {
	if len(inputs) == 0 {
		return nil
	}
	images := make([]model.TaggedImage, 0, len(inputs))
	for _, item := range inputs {
		images = append(images, model.TaggedImage{
			URL: strings.TrimSpace(item.URL),
			Tag: strings.TrimSpace(item.Tag),
		})
	}
	return images
}

func cloneStringSlice(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, strings.TrimSpace(item))
	}
	return out
}

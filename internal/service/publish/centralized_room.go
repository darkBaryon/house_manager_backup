package publish

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	"house-manager/pkg/errcode"
)

type centralizedRoomService struct {
	hmd       centralizedRoomScopeDomain
	publisher mutationPublisher
	access    publishAccessService
}

func newCentralizedRoomService(hmd centralizedRoomScopeDomain, publisher mutationPublisher, access publishAccessService) *centralizedRoomService {
	return &centralizedRoomService{hmd: hmd, publisher: publisher, access: access}
}

func (s *centralizedRoomService) CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*hmdmodel.HmdRoomCentralized, error) {
	logPublishInfo(ctx, "publish.room.create.start", "project_id", input.ProjectID.Hex(), "building_id", input.BuildingID.Hex(), "room_no", input.RoomNo)
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.room.create.success", "publish.room.create.failed", err)
		return nil, err
	}
	if err := s.requireBuildingAccess(ctx, scope, input.BuildingID, "create centralized room"); err != nil {
		logPublishWarn(ctx, "publish.room.create.denied", "building_id", input.BuildingID.Hex(), "error", err)
		return nil, err
	}
	input, err = s.mergeRoomTypeTemplate(ctx, input)
	if err != nil {
		logPublishResult(ctx, "publish.room.create.success", "publish.room.create.failed", err, "building_id", input.BuildingID.Hex(), "room_type_id", input.RoomTypeID.Hex(), "room_no", input.RoomNo)
		return nil, err
	}
	result, err := s.hmd.CreateCentralizedRoom(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	if err != nil {
		logPublishResult(ctx, "publish.room.create.success", "publish.room.create.failed", err, "building_id", input.BuildingID.Hex(), "room_no", input.RoomNo)
		return nil, err
	}
	logPublishInfo(ctx, "publish.room.create.success", "room_id", entity.ID.Hex(), "project_id", entity.ProjectID.Hex(), "room_no", entity.RoomNo)
	return entity, nil
}

func (s *centralizedRoomService) mergeRoomTypeTemplate(ctx context.Context, input CreateCentralizedRoomInput) (CreateCentralizedRoomInput, error) {
	if input.RoomTypeID.IsZero() {
		return input, nil
	}

	roomType, err := s.hmd.GetRoomType(ctx, input.RoomTypeID)
	if err != nil {
		return input, err
	}
	if roomType == nil {
		return input, errcode.NotFound.WithError(fmt.Errorf("房型不存在"))
	}
	if !roomType.ProjectID.IsZero() && roomType.ProjectID != input.ProjectID {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("房型不属于当前项目"))
	}
	if !roomType.BuildingID.IsZero() && roomType.BuildingID != input.BuildingID {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("房型不属于当前楼栋"))
	}

	if strings.TrimSpace(input.LayoutText) == "" {
		input.LayoutText = layoutTextFromRoomType(roomType)
	}
	if input.AreaSize == 0 {
		input.AreaSize = roomType.AreaSize
	}
	if strings.TrimSpace(input.Orientation) == "" {
		input.Orientation = string(roomType.Orientation)
	}
	if strings.TrimSpace(input.DecorationLevel) == "" {
		input.DecorationLevel = string(roomType.DecorationLevel)
	}
	if strings.TrimSpace(input.PaymentCycle) == "" {
		input.PaymentCycle = string(roomType.PaymentCycle)
	}
	if input.Rent == 0 {
		input.Rent = roomType.Rent
	}
	if input.Deposit == 0 {
		input.Deposit = roomType.Deposit
	}
	if input.ServiceFee == 0 {
		input.ServiceFee = roomType.ServiceFee
	}
	if strings.TrimSpace(input.AgencyFeeMode) == "" {
		input.AgencyFeeMode = string(roomType.AgencyFeeMode)
	}
	if input.AgencyFeeValue == 0 {
		input.AgencyFeeValue = roomType.AgencyFeeValue
	}
	if len(input.Images) == 0 {
		input.Images = taggedImageInputsFromModel(roomType.Images)
	}
	if len(input.RoomFacilities) == 0 {
		input.RoomFacilities = roomFacilitiesFromModel(roomType.RoomFacilities)
	}
	return input, nil
}

func layoutTextFromRoomType(roomType *hmdmodel.HmdRoomTypeCentralized) string {
	if roomType == nil {
		return ""
	}
	var parts []string
	if roomType.RoomCount > 0 {
		parts = append(parts, fmt.Sprintf("%d室", roomType.RoomCount))
	}
	if roomType.HallCount > 0 {
		parts = append(parts, fmt.Sprintf("%d厅", roomType.HallCount))
	}
	if roomType.BathroomCount > 0 {
		parts = append(parts, fmt.Sprintf("%d卫", roomType.BathroomCount))
	}
	if roomType.KitchenCount > 0 {
		parts = append(parts, fmt.Sprintf("%d厨", roomType.KitchenCount))
	}
	return strings.Join(parts, "")
}

func taggedImageInputsFromModel(images []hmdmodel.TaggedImage) []TaggedImageInput {
	if len(images) == 0 {
		return nil
	}
	result := make([]TaggedImageInput, 0, len(images))
	for _, image := range images {
		result = append(result, TaggedImageInput{
			URL: image.URL,
			Tag: string(image.Tag),
		})
	}
	return result
}

func roomFacilitiesFromModel(facilities []hmdmodel.RoomFacility) []string {
	if len(facilities) == 0 {
		return nil
	}
	result := make([]string, 0, len(facilities))
	for _, facility := range facilities {
		result = append(result, string(facility))
	}
	return result
}

func (s *centralizedRoomService) GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error) {
	logPublishInfo(ctx, "publish.room.detail.start", "room_id", id.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.room.detail.success", "publish.room.detail.failed", err, "room_id", id.Hex())
		return nil, err
	}
	room, err := s.hmd.GetCentralizedRoom(ctx, id)
	if err != nil {
		logPublishResult(ctx, "publish.room.detail.success", "publish.room.detail.failed", err, "room_id", id.Hex())
		return nil, err
	}
	if room != nil {
		if err := requireCentralizedProjectAccess(ctx, scope, room.ProjectID, "get centralized room"); err != nil {
			logPublishWarn(ctx, "publish.room.detail.denied", "room_id", id.Hex(), "project_id", room.ProjectID.Hex(), "error", err)
			return nil, err
		}
	}
	logPublishInfo(ctx, "publish.room.detail.success", "room_id", id.Hex(), "found", room != nil)
	return room, nil
}

func (s *centralizedRoomService) ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	logPublishInfo(ctx, "publish.room.list_by_project.start", "project_id", projectID.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.room.list_by_project.success", "publish.room.list_by_project.failed", err, "project_id", projectID.Hex())
		return nil, err
	}
	allowed, err := canAccessCentralizedProject(ctx, scope, projectID)
	if err != nil {
		logPublishResult(ctx, "publish.room.list_by_project.success", "publish.room.list_by_project.failed", err, "project_id", projectID.Hex())
		return nil, err
	}
	if !allowed {
		logPublishWarn(ctx, "publish.room.list_by_project.denied", "project_id", projectID.Hex())
		return []hmdmodel.HmdRoomCentralized{}, nil
	}
	rooms, err := s.hmd.ListCentralizedRoomsByProject(ctx, projectID)
	logPublishResult(ctx, "publish.room.list_by_project.success", "publish.room.list_by_project.failed", err, "project_id", projectID.Hex(), "result_count", len(rooms))
	return rooms, err
}

func (s *centralizedRoomService) ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	logPublishInfo(ctx, "publish.room.list_by_building.start", "building_id", buildingID.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.room.list_by_building.success", "publish.room.list_by_building.failed", err, "building_id", buildingID.Hex())
		return nil, err
	}
	allowed, err := s.canAccessBuilding(ctx, scope, buildingID)
	if err != nil {
		logPublishResult(ctx, "publish.room.list_by_building.success", "publish.room.list_by_building.failed", err, "building_id", buildingID.Hex())
		return nil, err
	}
	if !allowed {
		logPublishWarn(ctx, "publish.room.list_by_building.denied", "building_id", buildingID.Hex())
		return []hmdmodel.HmdRoomCentralized{}, nil
	}
	rooms, err := s.hmd.ListCentralizedRoomsByBuilding(ctx, buildingID)
	logPublishResult(ctx, "publish.room.list_by_building.success", "publish.room.list_by_building.failed", err, "building_id", buildingID.Hex(), "result_count", len(rooms))
	return rooms, err
}

func (s *centralizedRoomService) UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*hmdmodel.HmdRoomCentralized, error) {
	logPublishInfo(ctx, "publish.room.update.start", "room_id", input.ID.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.room.update.success", "publish.room.update.failed", err, "room_id", input.ID.Hex())
		return nil, err
	}
	room, err := s.hmd.GetCentralizedRoom(ctx, input.ID)
	if err != nil {
		logPublishResult(ctx, "publish.room.update.success", "publish.room.update.failed", err, "room_id", input.ID.Hex())
		return nil, err
	}
	if room == nil {
		err := scopeNotFound("update centralized room")
		logPublishWarn(ctx, "publish.room.update.denied", "room_id", input.ID.Hex(), "error", err)
		return nil, err
	}
	if err := requireCentralizedProjectAccess(ctx, scope, room.ProjectID, "update centralized room"); err != nil {
		logPublishWarn(ctx, "publish.room.update.denied", "room_id", input.ID.Hex(), "project_id", room.ProjectID.Hex(), "error", err)
		return nil, err
	}
	result, err := s.hmd.UpdateCentralizedRoom(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	logPublishResult(ctx, "publish.room.update.success", "publish.room.update.failed", err, "room_id", input.ID.Hex())
	return entity, err
}

func (s *centralizedRoomService) UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*hmdmodel.HmdRoomCentralized, error) {
	logPublishInfo(ctx, "publish.room.update_status.start", "room_id", input.ID.Hex(), "room_status", input.RoomStatus)
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.room.update_status.success", "publish.room.update_status.failed", err, "room_id", input.ID.Hex())
		return nil, err
	}
	room, err := s.hmd.GetCentralizedRoom(ctx, input.ID)
	if err != nil {
		logPublishResult(ctx, "publish.room.update_status.success", "publish.room.update_status.failed", err, "room_id", input.ID.Hex())
		return nil, err
	}
	if room == nil {
		err := scopeNotFound("update centralized room status")
		logPublishWarn(ctx, "publish.room.update_status.denied", "room_id", input.ID.Hex(), "error", err)
		return nil, err
	}
	if err := requireCentralizedProjectAccess(ctx, scope, room.ProjectID, "update centralized room status"); err != nil {
		logPublishWarn(ctx, "publish.room.update_status.denied", "room_id", input.ID.Hex(), "project_id", room.ProjectID.Hex(), "error", err)
		return nil, err
	}
	result, err := s.hmd.UpdateCentralizedRoomStatus(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	logPublishResult(ctx, "publish.room.update_status.success", "publish.room.update_status.failed", err, "room_id", input.ID.Hex(), "room_status", input.RoomStatus)
	return entity, err
}

func (s *centralizedRoomService) canAccessBuilding(ctx context.Context, scope PublishScope, buildingID bson.ObjectID) (bool, error) {
	building, err := s.hmd.GetBuilding(ctx, buildingID)
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	if building == nil {
		return false, nil
	}
	return canAccessCentralizedProject(ctx, scope, building.ProjectID)
}

func (s *centralizedRoomService) requireBuildingAccess(ctx context.Context, scope PublishScope, buildingID bson.ObjectID, action string) error {
	allowed, err := s.canAccessBuilding(ctx, scope, buildingID)
	if err != nil {
		return err
	}
	if !allowed {
		return scopeNotFound(action)
	}
	return nil
}

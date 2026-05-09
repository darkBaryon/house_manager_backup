package hpd

import (
	"context"
	"testing"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestMiniappProjectorRefreshBuildingFansOutToRooms(t *testing.T) {
	buildingID := bson.NewObjectID()
	projectID := bson.NewObjectID()
	roomAID := bson.NewObjectID()
	roomBID := bson.NewObjectID()

	hpdListings := &fanoutHpdListingRepo{}
	miniappListings := &fanoutMiniappListingRepo{}
	rooms := &fanoutCentralizedRoomRepo{
		rooms: map[bson.ObjectID]*model.HmdRoomCentralized{
			roomAID: {
				CommonFields: model.CommonFields{ID: roomAID},
				ProjectID:    projectID,
				BuildingID:   buildingID,
				RoomNo:       "801",
				RentMode:     model.RentModeWhole,
				Rent:         5200,
				RoomStatus:   model.RoomStatusAvailable,
			},
			roomBID: {
				CommonFields: model.CommonFields{ID: roomBID},
				ProjectID:    projectID,
				BuildingID:   buildingID,
				RoomNo:       "802",
				RentMode:     model.RentModeWhole,
				Rent:         5300,
				RoomStatus:   model.RoomStatusAvailable,
			},
		},
		listByBuildingID: map[bson.ObjectID][]model.HmdRoomCentralized{
			buildingID: {
				{CommonFields: model.CommonFields{ID: roomAID}},
				{CommonFields: model.CommonFields{ID: roomBID}},
			},
		},
	}
	projector := NewMiniappProjector(
		hpdListings,
		miniappListings,
		&fanoutCentralizedRepo{project: &model.HmdCentralized{
			CommonFields: model.CommonFields{ID: projectID},
			ProjectName:  "泊寓南山",
			City:         "深圳",
		}},
		&fanoutBuildingRepo{building: &model.HmdBuilding{
			CommonFields: model.CommonFields{ID: buildingID},
			ProjectID:    projectID,
			BuildingName: "A座",
		}},
		nil,
		&fanoutRoomTypeRepo{},
		rooms,
		nil,
	)

	if err := projector.RefreshBuilding(context.Background(), buildingID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hpdListings.upsertedSources) != 2 {
		t.Fatalf("expected two listing upserts, got %#v", hpdListings.upsertedSources)
	}
	if len(miniappListings.upserted) != 2 {
		t.Fatalf("expected two miniapp upserts, got %#v", miniappListings.upserted)
	}
}

type fanoutHpdListingRepo struct {
	upsertedSources []bson.ObjectID
}

func (f *fanoutHpdListingRepo) FindByID(ctx context.Context, id bson.ObjectID) (*model.HpdListing, error) {
	return nil, nil
}

func (f *fanoutHpdListingRepo) FindBySource(ctx context.Context, sourceType model.HpdSourceType, sourceID bson.ObjectID) (*model.HpdListing, error) {
	return nil, nil
}

func (f *fanoutHpdListingRepo) UpsertBySource(ctx context.Context, entity *model.HpdListing) (*model.HpdListing, error) {
	f.upsertedSources = append(f.upsertedSources, entity.SourceID)
	return &model.HpdListing{
		CommonFields:  model.CommonFields{ID: bson.NewObjectID()},
		SourceType:    entity.SourceType,
		SourceID:      entity.SourceID,
		AssetMode:     entity.AssetMode,
		ListingStatus: model.HpdListingStatusPublished,
	}, nil
}

func (f *fanoutHpdListingRepo) UpdateLifecycleFields(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	return nil
}

func (f *fanoutHpdListingRepo) UpdateStatus(ctx context.Context, id bson.ObjectID, listingStatus model.HpdListingStatus) error {
	return nil
}

type fanoutMiniappListingRepo struct {
	upserted []*model.HpdMiniappListing
}

func (f *fanoutMiniappListingRepo) UpsertByListingID(ctx context.Context, entity *model.HpdMiniappListing) (*model.HpdMiniappListing, error) {
	f.upserted = append(f.upserted, entity)
	return entity, nil
}

type fanoutCentralizedRepo struct {
	project *model.HmdCentralized
}

func (f *fanoutCentralizedRepo) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error) {
	return f.project, nil
}

type fanoutBuildingRepo struct {
	building *model.HmdBuilding
}

func (f *fanoutBuildingRepo) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error) {
	return f.building, nil
}

type fanoutRoomTypeRepo struct{}

func (f *fanoutRoomTypeRepo) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error) {
	return nil, nil
}

type fanoutCentralizedRoomRepo struct {
	rooms            map[bson.ObjectID]*model.HmdRoomCentralized
	listByProjectID  map[bson.ObjectID][]model.HmdRoomCentralized
	listByBuildingID map[bson.ObjectID][]model.HmdRoomCentralized
	listByRoomTypeID map[bson.ObjectID][]model.HmdRoomCentralized
}

func (f *fanoutCentralizedRoomRepo) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error) {
	return f.rooms[id], nil
}

func (f *fanoutCentralizedRoomRepo) ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	return f.listByProjectID[projectID], nil
}

func (f *fanoutCentralizedRoomRepo) ListByBuildingID(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	return f.listByBuildingID[buildingID], nil
}

func (f *fanoutCentralizedRoomRepo) ListByRoomTypeID(ctx context.Context, roomTypeID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	return f.listByRoomTypeID[roomTypeID], nil
}

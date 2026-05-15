package listingprojection

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	commonmodel "house-manager/internal/model/common"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	"testing"
)

func TestMiniappProjectorRefreshBuildingFansOutToRooms(t *testing.T) {
	buildingID := bson.NewObjectID()
	projectID := bson.NewObjectID()
	roomAID := bson.NewObjectID()
	roomBID := bson.NewObjectID()

	hpdListings := &fanoutHpdListingRepo{}
	miniappListings := &fanoutMiniappListingRepo{}
	rooms := &fanoutCentralizedRoomRepo{
		rooms: map[bson.ObjectID]*hmdmodel.HmdRoomCentralized{
			roomAID: {
				CommonFields: commonmodel.CommonFields{ID: roomAID},
				ProjectID:    projectID,
				BuildingID:   buildingID,
				RoomNo:       "801",
				RentMode:     hmdmodel.RentModeWhole,
				Rent:         5200,
				RoomStatus:   hmdmodel.RoomStatusAvailable,
			},
			roomBID: {
				CommonFields: commonmodel.CommonFields{ID: roomBID},
				ProjectID:    projectID,
				BuildingID:   buildingID,
				RoomNo:       "802",
				RentMode:     hmdmodel.RentModeWhole,
				Rent:         5300,
				RoomStatus:   hmdmodel.RoomStatusAvailable,
			},
		},
		listByBuildingID: map[bson.ObjectID][]hmdmodel.HmdRoomCentralized{
			buildingID: {
				{CommonFields: commonmodel.CommonFields{ID: roomAID}},
				{CommonFields: commonmodel.CommonFields{ID: roomBID}},
			},
		},
	}
	projector := NewMiniappProjector(
		hpdListings,
		miniappListings,
		&fanoutCentralizedRepo{project: &hmdmodel.HmdCentralized{
			CommonFields: commonmodel.CommonFields{ID: projectID},
			ProjectName:  "泊寓南山",
			City:         "深圳",
		}},
		&fanoutBuildingRepo{building: &hmdmodel.HmdBuilding{
			CommonFields: commonmodel.CommonFields{ID: buildingID},
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

func (f *fanoutHpdListingRepo) FindByID(ctx context.Context, id bson.ObjectID) (*hpdmodel.HpdListing, error) {
	return nil, nil
}

func (f *fanoutHpdListingRepo) FindBySource(ctx context.Context, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID) (*hpdmodel.HpdListing, error) {
	return nil, nil
}

func (f *fanoutHpdListingRepo) UpsertBySource(ctx context.Context, entity *hpdmodel.HpdListing) (*hpdmodel.HpdListing, error) {
	f.upsertedSources = append(f.upsertedSources, entity.SourceID)
	return &hpdmodel.HpdListing{
		CommonFields:  commonmodel.CommonFields{ID: bson.NewObjectID()},
		SourceType:    entity.SourceType,
		SourceID:      entity.SourceID,
		AssetMode:     entity.AssetMode,
		ListingStatus: hpdmodel.HpdListingStatusPublished,
	}, nil
}

func (f *fanoutHpdListingRepo) UpdateLifecycleFields(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	return nil
}

func (f *fanoutHpdListingRepo) UpdateStatus(ctx context.Context, id bson.ObjectID, listingStatus hpdmodel.HpdListingStatus) error {
	return nil
}

type fanoutMiniappListingRepo struct {
	upserted []*hpdmodel.HpdMiniappListing
}

func (f *fanoutMiniappListingRepo) UpsertByListingID(ctx context.Context, entity *hpdmodel.HpdMiniappListing) (*hpdmodel.HpdMiniappListing, error) {
	f.upserted = append(f.upserted, entity)
	return entity, nil
}

type fanoutCentralizedRepo struct {
	project *hmdmodel.HmdCentralized
}

func (f *fanoutCentralizedRepo) FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdCentralized, error) {
	return f.project, nil
}

type fanoutBuildingRepo struct {
	building *hmdmodel.HmdBuilding
}

func (f *fanoutBuildingRepo) FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdBuilding, error) {
	return f.building, nil
}

type fanoutRoomTypeRepo struct{}

func (f *fanoutRoomTypeRepo) FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomTypeCentralized, error) {
	return nil, nil
}

type fanoutCentralizedRoomRepo struct {
	rooms            map[bson.ObjectID]*hmdmodel.HmdRoomCentralized
	listByProjectID  map[bson.ObjectID][]hmdmodel.HmdRoomCentralized
	listByBuildingID map[bson.ObjectID][]hmdmodel.HmdRoomCentralized
	listByRoomTypeID map[bson.ObjectID][]hmdmodel.HmdRoomCentralized
}

func (f *fanoutCentralizedRoomRepo) FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error) {
	return f.rooms[id], nil
}

func (f *fanoutCentralizedRoomRepo) ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	return f.listByProjectID[projectID], nil
}

func (f *fanoutCentralizedRoomRepo) ListByBuildingID(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	return f.listByBuildingID[buildingID], nil
}

func (f *fanoutCentralizedRoomRepo) ListByRoomTypeID(ctx context.Context, roomTypeID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	return f.listByRoomTypeID[roomTypeID], nil
}

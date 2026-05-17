package provider

import (
	"context"
	"errors"
	"testing"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	landlordrepo "house-manager/internal/repository/landlord"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateCreatesProviderAndPasswordAuth(t *testing.T) {
	operatorID := bson.NewObjectID()
	landlordRepo := &fakeLandlordRepository{}
	authRepo := &fakeLandlordAuthRepository{}
	svc := newService(landlordRepo, authRepo)

	result, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: operatorID.Hex(),
		Phone:           " 13800000000 ",
		Password:        " secret123 ",
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if result.Provider.Phone != "13800000000" || result.Provider.CreatedByStaffID != operatorID.Hex() {
		t.Fatalf("unexpected provider: %#v", result.Provider)
	}
	if landlordRepo.created == nil || landlordRepo.created.CreatedByStaffID != operatorID {
		t.Fatalf("expected landlord created by operator, got %#v", landlordRepo.created)
	}
	if authRepo.created == nil || authRepo.created.LandlordID != landlordRepo.created.ID {
		t.Fatalf("expected auth for landlord, got %#v", authRepo.created)
	}
	if bcrypt.CompareHashAndPassword([]byte(authRepo.created.PasswordHash), []byte("secret123")) != nil {
		t.Fatalf("expected bcrypt password hash")
	}
}

func TestCreateRejectsDuplicatePhone(t *testing.T) {
	existing := &authmodel.Landlord{Phone: "13800000000"}
	existing.ID = bson.NewObjectID()
	svc := newService(&fakeLandlordRepository{existing: existing}, &fakeLandlordAuthRepository{})

	_, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: bson.NewObjectID().Hex(),
		Phone:           "13800000000",
		Password:        "secret123",
	})
	assertProviderErr(t, err, errcode.AlreadyExists.Code, "发房方手机号已存在，请更换后重试")
}

func TestCreateRollsBackWhenAuthCreateFails(t *testing.T) {
	landlordRepo := &fakeLandlordRepository{}
	authRepo := &fakeLandlordAuthRepository{err: errors.New("mongo down")}
	svc := newService(landlordRepo, authRepo)

	_, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: bson.NewObjectID().Hex(),
		Phone:           "13800000000",
		Password:        "secret123",
	})
	assertProviderErr(t, err, errcode.DatabaseError.Code, "创建发房方失败，请稍后重试")
	if landlordRepo.rollbackID != landlordRepo.created.ID || authRepo.rollbackLandlordID != landlordRepo.created.ID {
		t.Fatalf("expected rollback, landlord=%s auth=%s created=%s", landlordRepo.rollbackID.Hex(), authRepo.rollbackLandlordID.Hex(), landlordRepo.created.ID.Hex())
	}
}

func TestListReturnsProviderWithLoginSummary(t *testing.T) {
	providerID := bson.NewObjectID()
	operatorID := bson.NewObjectID()
	status := commonmodel.StatusDeleted
	landlord := authmodel.Landlord{Phone: "13800000000", CreatedByStaffID: operatorID}
	landlord.ID = providerID
	landlord.Status = status
	landlord.CreatedAt = 111
	landlord.UpdatedAt = 222
	authRecord := authmodel.LandlordAuth{
		LandlordID:        providerID,
		PasswordUpdatedAt: 333,
		LastLoginAt:       444,
		LastLoginIP:       "127.0.0.1",
	}
	svc := newService(
		&fakeLandlordRepository{listItems: []authmodel.Landlord{landlord}, listTotal: 1},
		&fakeLandlordAuthRepository{auths: []authmodel.LandlordAuth{authRecord}},
	)

	result, err := svc.List(context.Background(), ListInput{Phone: " 13800000000 ", Status: &status, Page: -1, PageSize: 999})
	if err != nil {
		t.Fatalf("list provider: %v", err)
	}
	if result.Page != 1 || result.PageSize != 100 || result.Total != 1 {
		t.Fatalf("unexpected page result: %#v", result)
	}
	if len(result.List) != 1 {
		t.Fatalf("expected one provider, got %#v", result.List)
	}
	if result.List[0].ProviderID != providerID.Hex() || result.List[0].LastLoginAt != 444 {
		t.Fatalf("unexpected provider item: %#v", result.List[0])
	}
}

func TestUpdateChangesPhoneAndInvalidatesPublishSession(t *testing.T) {
	operatorID := bson.NewObjectID()
	providerID := bson.NewObjectID()
	landlord := &authmodel.Landlord{Phone: "13800000000"}
	landlord.ID = providerID
	landlord.Status = commonmodel.StatusActive
	sessionStore := &fakeProviderSessionInvalidator{}
	svc := newService(&fakeLandlordRepository{detail: landlord}, &fakeLandlordAuthRepository{}, sessionStore)

	result, err := svc.Update(context.Background(), UpdateInput{
		OperatorStaffID: operatorID.Hex(),
		ProviderID:      providerID.Hex(),
		Phone:           "13800000001",
	})
	if err != nil {
		t.Fatalf("update provider: %v", err)
	}
	if result.Provider.Phone != "13800000001" {
		t.Fatalf("unexpected result: %#v", result.Provider)
	}
	if sessionStore.invalidated.PrincipalType != session.PrincipalTypeLandlord ||
		sessionStore.invalidated.PrincipalID != providerID.Hex() ||
		sessionStore.invalidated.Terminal != session.TerminalPublish {
		t.Fatalf("expected publish landlord session invalidated, got %#v", sessionStore.invalidated)
	}
}

func TestDisableMarksProviderDeletedAndInvalidatesPublishSession(t *testing.T) {
	operatorID := bson.NewObjectID()
	providerID := bson.NewObjectID()
	landlord := &authmodel.Landlord{Phone: "13800000000"}
	landlord.ID = providerID
	landlord.Status = commonmodel.StatusActive
	landlordRepo := &fakeLandlordRepository{detail: landlord}
	sessionStore := &fakeProviderSessionInvalidator{}
	svc := newService(landlordRepo, &fakeLandlordAuthRepository{}, sessionStore)

	result, err := svc.Disable(context.Background(), DisableInput{
		OperatorStaffID: operatorID.Hex(),
		ProviderID:      providerID.Hex(),
	})
	if err != nil {
		t.Fatalf("disable provider: %v", err)
	}
	if !result.Success || landlordRepo.detail.Status != commonmodel.StatusDeleted {
		t.Fatalf("expected provider disabled, result=%#v detail=%#v", result, landlordRepo.detail)
	}
	if sessionStore.invalidated.PrincipalID != providerID.Hex() {
		t.Fatalf("expected session invalidated, got %#v", sessionStore.invalidated)
	}
}

func TestListRejectsInvalidStatus(t *testing.T) {
	status := 2
	svc := newService(&fakeLandlordRepository{}, &fakeLandlordAuthRepository{})
	_, err := svc.List(context.Background(), ListInput{Status: &status})
	assertProviderErr(t, err, errcode.InvalidParam.Code, "发房方状态参数不正确")
}

func assertProviderErr(t *testing.T, err error, code int, message string) {
	t.Helper()
	got := errcode.FromError(err)
	if got == nil {
		t.Fatalf("expected errcode, got %v", err)
	}
	if got.Code != code || got.PublicMessage() != message {
		t.Fatalf("expected code=%d message=%q, got code=%d message=%q err=%v", code, message, got.Code, got.PublicMessage(), err)
	}
}

type fakeLandlordRepository struct {
	created    *authmodel.Landlord
	existing   *authmodel.Landlord
	detail     *authmodel.Landlord
	listInput  landlordrepo.ListFilter
	listItems  []authmodel.Landlord
	listTotal  int64
	rollbackID bson.ObjectID
	err        error
}

func (f *fakeLandlordRepository) Create(ctx context.Context, landlord *authmodel.Landlord) error {
	if f.err != nil {
		return f.err
	}
	if landlord.ID.IsZero() {
		landlord.ID = bson.NewObjectID()
	}
	landlord.Status = commonmodel.StatusActive
	landlord.CreatedAt = 111
	landlord.UpdatedAt = 111
	copied := *landlord
	f.created = &copied
	f.detail = &copied
	return nil
}

func (f *fakeLandlordRepository) FindByPhone(ctx context.Context, phone string) (*authmodel.Landlord, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.existing, nil
}

func (f *fakeLandlordRepository) FindByID(ctx context.Context, id bson.ObjectID) (*authmodel.Landlord, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.detail, nil
}

func (f *fakeLandlordRepository) List(ctx context.Context, input landlordrepo.ListFilter) ([]authmodel.Landlord, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	f.listInput = input
	return f.listItems, f.listTotal, nil
}

func (f *fakeLandlordRepository) UpdateFields(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if f.err != nil {
		return f.err
	}
	if f.detail == nil || f.detail.ID != id {
		return nil
	}
	if value, ok := fields["phone"].(string); ok {
		f.detail.Phone = value
	}
	if value, ok := fields["status"].(int); ok {
		f.detail.Status = value
	}
	if value, ok := fields["updated_by_staff_id"].(bson.ObjectID); ok {
		f.detail.UpdatedByStaffID = value
	}
	return nil
}

func (f *fakeLandlordRepository) RollbackCreate(ctx context.Context, id bson.ObjectID) error {
	f.rollbackID = id
	return nil
}

type fakeLandlordAuthRepository struct {
	created            *authmodel.LandlordAuth
	auths              []authmodel.LandlordAuth
	rollbackLandlordID bson.ObjectID
	err                error
}

func (f *fakeLandlordAuthRepository) Create(ctx context.Context, auth *authmodel.LandlordAuth) error {
	if f.err != nil {
		return f.err
	}
	if auth.ID.IsZero() {
		auth.ID = bson.NewObjectID()
	}
	copied := *auth
	f.created = &copied
	f.auths = append(f.auths, copied)
	return nil
}

func (f *fakeLandlordAuthRepository) FindActivePasswordByLandlordIDs(ctx context.Context, landlordIDs []bson.ObjectID) ([]authmodel.LandlordAuth, error) {
	if f.err != nil {
		return nil, f.err
	}
	allowed := make(map[bson.ObjectID]struct{}, len(landlordIDs))
	for _, id := range landlordIDs {
		allowed[id] = struct{}{}
	}
	result := make([]authmodel.LandlordAuth, 0, len(f.auths))
	for _, authRecord := range f.auths {
		if _, ok := allowed[authRecord.LandlordID]; ok {
			result = append(result, authRecord)
		}
	}
	return result, nil
}

func (f *fakeLandlordAuthRepository) RollbackCreateByLandlordID(ctx context.Context, landlordID bson.ObjectID) error {
	f.rollbackLandlordID = landlordID
	return nil
}

type fakeProviderSessionInvalidator struct {
	invalidated session.Principal
	err         error
}

func (f *fakeProviderSessionInvalidator) InvalidatePrincipal(ctx context.Context, principal session.Principal) error {
	if f.err != nil {
		return f.err
	}
	f.invalidated = principal
	return nil
}

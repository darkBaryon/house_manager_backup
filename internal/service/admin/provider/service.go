package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	landlordrepo "house-manager/internal/repository/landlord"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = bcrypt.DefaultCost

var mainlandPhonePattern = regexp.MustCompile(`^1[3-9][0-9]{9}$`)

type Service struct {
	landlordRepo     landlordRepository
	landlordAuthRepo landlordAuthRepository
	sessionStore     sessionInvalidator
}

type landlordRepository interface {
	Create(ctx context.Context, landlord *authmodel.Landlord) error
	FindByPhone(ctx context.Context, phone string) (*authmodel.Landlord, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*authmodel.Landlord, error)
	List(ctx context.Context, input landlordrepo.ListFilter) ([]authmodel.Landlord, int64, error)
	UpdateFields(ctx context.Context, id bson.ObjectID, fields bson.M) error
	RollbackCreate(ctx context.Context, id bson.ObjectID) error
}

type landlordAuthRepository interface {
	Create(ctx context.Context, auth *authmodel.LandlordAuth) error
	FindActivePasswordByLandlordIDs(ctx context.Context, landlordIDs []bson.ObjectID) ([]authmodel.LandlordAuth, error)
	RollbackCreateByLandlordID(ctx context.Context, landlordID bson.ObjectID) error
}

type sessionInvalidator interface {
	InvalidatePrincipal(ctx context.Context, principal session.Principal) error
}

func NewService(
	landlordRepo *landlordrepo.LandlordRepository,
	landlordAuthRepo *landlordrepo.LandlordAuthRepository,
	sessionStore *session.Store,
) *Service {
	return newService(landlordRepo, landlordAuthRepo, sessionStore)
}

func newService(
	landlordRepo landlordRepository,
	landlordAuthRepo landlordAuthRepository,
	sessionStores ...sessionInvalidator,
) *Service {
	var sessionStore sessionInvalidator
	if len(sessionStores) > 0 {
		sessionStore = sessionStores[0]
	}
	return &Service{
		landlordRepo:     landlordRepo,
		landlordAuthRepo: landlordAuthRepo,
		sessionStore:     sessionStore,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*CreateResult, error) {
	normalized, err := normalizeCreateInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.requireRepos("创建发房方失败，请稍后重试"); err != nil {
		return nil, err
	}
	operatorID, err := parseOperatorID(normalized.OperatorStaffID)
	if err != nil {
		return nil, err
	}

	existing, err := s.landlordRepo.FindByPhone(ctx, normalized.Phone)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "创建发房方失败，请稍后重试", err)
	}
	if existing != nil {
		return nil, errcode.AlreadyExists.WithError(fmt.Errorf("发房方手机号已存在，请更换后重试"))
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(normalized.Password), bcryptCost)
	if err != nil {
		return nil, publicSystemError(errcode.InternalError.Code, "创建发房方失败，请稍后重试", err)
	}
	now := time.Now().Unix()
	landlord := &authmodel.Landlord{
		Phone:            normalized.Phone,
		CreatedByStaffID: operatorID,
	}
	if err := s.landlordRepo.Create(ctx, landlord); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errcode.AlreadyExists.WithError(fmt.Errorf("发房方手机号已存在，请更换后重试"))
		}
		return nil, publicSystemError(errcode.DatabaseError.Code, "创建发房方失败，请稍后重试", err)
	}

	authRecord := &authmodel.LandlordAuth{
		LandlordID:        landlord.ID,
		AuthType:          authmodel.PasswordAuthTypePassword,
		PasswordHash:      string(passwordHash),
		PasswordUpdatedAt: now,
	}
	if err := s.landlordAuthRepo.Create(ctx, authRecord); err != nil {
		return nil, s.rollbackCreate(ctx, landlord.ID, err)
	}

	return &CreateResult{Provider: toProviderSummary(landlord, authRecord)}, nil
}

func (s *Service) List(ctx context.Context, input ListInput) (*ListResult, error) {
	normalized, err := normalizeListInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.requireRepos("获取发房方列表失败，请稍后重试"); err != nil {
		return nil, err
	}

	landlords, total, err := s.landlordRepo.List(ctx, landlordrepo.ListFilter{
		Phone:  normalized.Phone,
		Status: *normalized.Status,
		Skip:   int64((normalized.Page - 1) * normalized.PageSize),
		Limit:  int64(normalized.PageSize),
	})
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取发房方列表失败，请稍后重试", err)
	}
	if len(landlords) == 0 {
		return &ListResult{List: []ListItem{}, Page: normalized.Page, PageSize: normalized.PageSize, Total: total}, nil
	}

	authRecords, err := s.landlordAuthRepo.FindActivePasswordByLandlordIDs(ctx, collectLandlordIDs(landlords))
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取发房方列表失败，请稍后重试", err)
	}
	return &ListResult{
		List:     buildListItems(landlords, authRecords),
		Page:     normalized.Page,
		PageSize: normalized.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) Detail(ctx context.Context, input DetailInput) (*DetailResult, error) {
	providerID, err := parseProviderID(input.ProviderID)
	if err != nil {
		return nil, err
	}
	if err := s.requireRepos("获取发房方详情失败，请稍后重试"); err != nil {
		return nil, err
	}
	detail, err := s.detailByID(ctx, providerID, "获取发房方详情失败，请稍后重试")
	if err != nil {
		return nil, err
	}
	return &DetailResult{Provider: *detail}, nil
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (*UpdateResult, error) {
	normalized, err := normalizeUpdateInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.requireRepos("更新发房方失败，请稍后重试"); err != nil {
		return nil, err
	}
	providerID, err := parseProviderID(normalized.ProviderID)
	if err != nil {
		return nil, err
	}
	operatorID, err := parseOperatorID(normalized.OperatorStaffID)
	if err != nil {
		return nil, err
	}

	landlord, err := s.landlordRepo.FindByID(ctx, providerID)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "更新发房方失败，请稍后重试", err)
	}
	if landlord == nil {
		return nil, errcode.NotFound.WithError(fmt.Errorf("发房方不存在或已删除"))
	}
	existing, err := s.landlordRepo.FindByPhone(ctx, normalized.Phone)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "更新发房方失败，请稍后重试", err)
	}
	if existing != nil && existing.ID != providerID {
		return nil, errcode.AlreadyExists.WithError(fmt.Errorf("发房方手机号已存在，请更换后重试"))
	}
	if err := s.landlordRepo.UpdateFields(ctx, providerID, bson.M{
		"phone":               normalized.Phone,
		"updated_by_staff_id": operatorID,
	}); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.NotFound.WithError(fmt.Errorf("发房方不存在或已删除"))
		}
		return nil, publicSystemError(errcode.DatabaseError.Code, "更新发房方失败，请稍后重试", err)
	}
	if err := s.invalidateProviderSession(ctx, providerID, "更新发房方失败，请稍后重试"); err != nil {
		return nil, err
	}

	detail, err := s.detailByID(ctx, providerID, "更新发房方失败，请稍后重试")
	if err != nil {
		return nil, err
	}
	return &UpdateResult{Provider: *detail}, nil
}

func (s *Service) Disable(ctx context.Context, input DisableInput) (*DisableResult, error) {
	providerID, err := parseProviderID(input.ProviderID)
	if err != nil {
		return nil, err
	}
	operatorID, err := parseOperatorID(input.OperatorStaffID)
	if err != nil {
		return nil, err
	}
	if err := s.requireRepos("禁用发房方失败，请稍后重试"); err != nil {
		return nil, err
	}
	landlord, err := s.landlordRepo.FindByID(ctx, providerID)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "禁用发房方失败，请稍后重试", err)
	}
	if landlord == nil {
		return nil, errcode.NotFound.WithError(fmt.Errorf("发房方不存在或已删除"))
	}
	if err := s.landlordRepo.UpdateFields(ctx, providerID, bson.M{
		"status":              commonmodel.StatusDeleted,
		"updated_by_staff_id": operatorID,
	}); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.NotFound.WithError(fmt.Errorf("发房方不存在或已删除"))
		}
		return nil, publicSystemError(errcode.DatabaseError.Code, "禁用发房方失败，请稍后重试", err)
	}
	if err := s.invalidateProviderSession(ctx, providerID, "禁用发房方失败，请稍后重试"); err != nil {
		return nil, err
	}
	return &DisableResult{Success: true}, nil
}

func (s *Service) detailByID(ctx context.Context, providerID bson.ObjectID, systemMessage string) (*ProviderSummary, error) {
	landlord, err := s.landlordRepo.FindByID(ctx, providerID)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, systemMessage, err)
	}
	if landlord == nil {
		return nil, errcode.NotFound.WithError(fmt.Errorf("发房方不存在或已删除"))
	}
	authRecords, err := s.landlordAuthRepo.FindActivePasswordByLandlordIDs(ctx, []bson.ObjectID{providerID})
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, systemMessage, err)
	}
	var authRecord *authmodel.LandlordAuth
	if len(authRecords) > 0 {
		authRecord = &authRecords[0]
	}
	summary := toProviderSummary(landlord, authRecord)
	return &summary, nil
}

func (s *Service) rollbackCreate(ctx context.Context, providerID bson.ObjectID, cause error) error {
	var rollbackErrs []error
	if s.landlordAuthRepo != nil {
		if err := s.landlordAuthRepo.RollbackCreateByLandlordID(ctx, providerID); err != nil {
			rollbackErrs = append(rollbackErrs, err)
		}
	}
	if s.landlordRepo != nil {
		if err := s.landlordRepo.RollbackCreate(ctx, providerID); err != nil {
			rollbackErrs = append(rollbackErrs, err)
		}
	}
	if len(rollbackErrs) > 0 {
		return publicSystemError(errcode.DatabaseError.Code, "创建发房方失败，请稍后重试", fmt.Errorf("create failed: %w; rollback failed: %w", cause, errors.Join(rollbackErrs...)))
	}
	return publicSystemError(errcode.DatabaseError.Code, "创建发房方失败，请稍后重试", cause)
}

func (s *Service) invalidateProviderSession(ctx context.Context, providerID bson.ObjectID, message string) error {
	if s.sessionStore == nil {
		return nil
	}
	if err := s.sessionStore.InvalidatePrincipal(ctx, session.Principal{
		PrincipalType: session.PrincipalTypeLandlord,
		PrincipalID:   providerID.Hex(),
		Terminal:      session.TerminalPublish,
	}); err != nil {
		return publicSystemError(errcode.CacheError.Code, message, err)
	}
	return nil
}

func (s *Service) requireRepos(message string) error {
	if s.landlordRepo == nil || s.landlordAuthRepo == nil {
		return publicSystemError(errcode.DatabaseError.Code, message, fmt.Errorf("后台发房方服务未正确初始化"))
	}
	return nil
}

func normalizeCreateInput(input CreateInput) (CreateInput, error) {
	input.OperatorStaffID = strings.TrimSpace(input.OperatorStaffID)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Password = strings.TrimSpace(input.Password)
	if input.Phone == "" || input.Password == "" {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("请输入发房方手机号和初始密码"))
	}
	if !mainlandPhonePattern.MatchString(input.Phone) {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("请输入正确的发房方手机号"))
	}
	return input, nil
}

func normalizeListInput(input ListInput) (ListInput, error) {
	input.Phone = strings.TrimSpace(input.Phone)
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}
	if input.Status == nil {
		status := commonmodel.StatusActive
		input.Status = &status
	}
	if *input.Status != commonmodel.StatusActive && *input.Status != commonmodel.StatusDeleted {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("发房方状态参数不正确"))
	}
	return input, nil
}

func normalizeUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.OperatorStaffID = strings.TrimSpace(input.OperatorStaffID)
	input.ProviderID = strings.TrimSpace(input.ProviderID)
	input.Phone = strings.TrimSpace(input.Phone)
	if input.Phone == "" {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("请输入发房方手机号"))
	}
	if !mainlandPhonePattern.MatchString(input.Phone) {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("请输入正确的发房方手机号"))
	}
	return input, nil
}

func parseProviderID(value string) (bson.ObjectID, error) {
	id, err := bson.ObjectIDFromHex(strings.TrimSpace(value))
	if err != nil || id.IsZero() {
		return bson.NilObjectID, errcode.InvalidParam.WithError(fmt.Errorf("发房方参数不正确"))
	}
	return id, nil
}

func parseOperatorID(value string) (bson.ObjectID, error) {
	id, err := bson.ObjectIDFromHex(strings.TrimSpace(value))
	if err != nil || id.IsZero() {
		return bson.NilObjectID, errcode.Unauthorized.WithError(fmt.Errorf("当前登录状态无效，请重新登录"))
	}
	return id, nil
}

func toProviderSummary(landlord *authmodel.Landlord, authRecord *authmodel.LandlordAuth) ProviderSummary {
	status := landlord.Status
	if status == commonmodel.StatusUnspecified {
		status = commonmodel.StatusActive
	}
	summary := ProviderSummary{
		ProviderID: landlord.ID.Hex(),
		Phone:      landlord.Phone,
		Status:     status,
		CreatedAt:  landlord.CreatedAt,
		UpdatedAt:  landlord.UpdatedAt,
	}
	if !landlord.CreatedByStaffID.IsZero() {
		summary.CreatedByStaffID = landlord.CreatedByStaffID.Hex()
	}
	if !landlord.UpdatedByStaffID.IsZero() {
		summary.UpdatedByStaffID = landlord.UpdatedByStaffID.Hex()
	}
	if authRecord != nil {
		summary.PasswordUpdatedAt = authRecord.PasswordUpdatedAt
		summary.LastLoginAt = authRecord.LastLoginAt
		summary.LastLoginIP = authRecord.LastLoginIP
	}
	return summary
}

func collectLandlordIDs(items []authmodel.Landlord) []bson.ObjectID {
	result := make([]bson.ObjectID, 0, len(items))
	for _, item := range items {
		if !item.ID.IsZero() {
			result = append(result, item.ID)
		}
	}
	return result
}

func buildListItems(landlords []authmodel.Landlord, authRecords []authmodel.LandlordAuth) []ListItem {
	authByLandlordID := make(map[bson.ObjectID]authmodel.LandlordAuth, len(authRecords))
	for _, authRecord := range authRecords {
		authByLandlordID[authRecord.LandlordID] = authRecord
	}
	result := make([]ListItem, 0, len(landlords))
	for _, landlord := range landlords {
		var authRecord *authmodel.LandlordAuth
		if item, ok := authByLandlordID[landlord.ID]; ok {
			authRecord = &item
		}
		result = append(result, ListItem{ProviderSummary: toProviderSummary(&landlord, authRecord)})
	}
	return result
}

func publicSystemError(code int, message string, cause error) *errcode.Error {
	return errcode.New(code, message).WithError(cause)
}

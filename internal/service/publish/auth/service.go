package auth

import (
	"context"
	"fmt"
	authmodel "house-manager/internal/model/auth"
	landlordrepo "house-manager/internal/repository/landlord"
	"house-manager/pkg/applog"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"
	"log/slog"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	landlordRepo     landlordRepository
	landlordAuthRepo landlordAuthRepository
	sessionStore     *session.Store
}

type landlordRepository interface {
	FindActiveByPhone(ctx context.Context, phone string) (*authmodel.Landlord, error)
	FindActiveByID(ctx context.Context, id bson.ObjectID) (*authmodel.Landlord, error)
}

type landlordAuthRepository interface {
	FindActivePasswordByLandlordID(ctx context.Context, landlordID bson.ObjectID) (*authmodel.LandlordAuth, error)
	TouchLastLogin(ctx context.Context, authID bson.ObjectID, loginIP string) error
}

func NewService(
	landlordRepo *landlordrepo.LandlordRepository,
	landlordAuthRepo *landlordrepo.LandlordAuthRepository,
	sessionStore *session.Store,
) *Service {
	return newService(
		landlordRepo,
		landlordAuthRepo,
		sessionStore,
	)
}

func newService(
	landlordRepo landlordRepository,
	landlordAuthRepo landlordAuthRepository,
	sessionStore *session.Store,
) *Service {
	return &Service{
		landlordRepo:     landlordRepo,
		landlordAuthRepo: landlordAuthRepo,
		sessionStore:     sessionStore,
	}
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	phone := strings.TrimSpace(input.Phone)
	password := strings.TrimSpace(input.Password)
	slog.InfoContext(ctx, "publish.auth.login.start", "phone", applog.Phone(phone))
	if phone == "" {
		err := errcode.InvalidParam.WithError(fmt.Errorf("手机号不能为空"))
		applog.Result(ctx, "publish.auth.login.success", "publish.auth.login.failed", err, "phone", applog.Phone(phone))
		return nil, err
	}
	if password == "" {
		err := errcode.InvalidParam.WithError(fmt.Errorf("密码不能为空"))
		applog.Result(ctx, "publish.auth.login.success", "publish.auth.login.failed", err, "phone", applog.Phone(phone))
		return nil, err
	}
	if s.landlordRepo == nil || s.landlordAuthRepo == nil || s.sessionStore == nil {
		err := errcode.DatabaseError.WithError(fmt.Errorf("房东登录服务未正确初始化"))
		applog.Result(ctx, "publish.auth.login.success", "publish.auth.login.failed", err, "phone", applog.Phone(phone))
		return nil, err
	}

	landlord, err := s.landlordRepo.FindActiveByPhone(ctx, phone)
	if err != nil {
		dbErr := errcode.DatabaseError.WithError(err)
		applog.Result(ctx, "publish.auth.login.success", "publish.auth.login.failed", dbErr, "phone", applog.Phone(phone))
		return nil, dbErr
	}
	if landlord == nil {
		authErr := errcode.Unauthorized.WithError(fmt.Errorf("手机号或密码错误"))
		applog.Result(ctx, "publish.auth.login.success", "publish.auth.login.failed", authErr, "phone", applog.Phone(phone))
		return nil, authErr
	}
	slog.InfoContext(ctx, "publish.auth.login.landlord_found", "landlord_id", landlord.ID.Hex(), "phone", applog.Phone(landlord.Phone))
	authRecord, err := s.landlordAuthRepo.FindActivePasswordByLandlordID(ctx, landlord.ID)
	if err != nil {
		dbErr := errcode.DatabaseError.WithError(err)
		applog.Result(ctx, "publish.auth.login.success", "publish.auth.login.failed", dbErr, "landlord_id", landlord.ID.Hex())
		return nil, dbErr
	}
	if authRecord == nil || bcrypt.CompareHashAndPassword([]byte(authRecord.PasswordHash), []byte(password)) != nil {
		authErr := errcode.Unauthorized.WithError(fmt.Errorf("手机号或密码错误"))
		applog.Result(ctx, "publish.auth.login.success", "publish.auth.login.failed", authErr, "landlord_id", landlord.ID.Hex())
		return nil, authErr
	}

	authSession, err := s.buildSession(landlord)
	if err != nil {
		applog.Result(ctx, "publish.auth.login.success", "publish.auth.login.failed", err, "landlord_id", landlord.ID.Hex())
		return nil, err
	}
	token, err := s.sessionStore.CreatePrincipal(ctx, authSession.Principal)
	if err != nil {
		cacheErr := errcode.CacheError.WithError(err)
		applog.Result(ctx, "publish.auth.login.success", "publish.auth.login.failed", cacheErr, "landlord_id", landlord.ID.Hex())
		return nil, cacheErr
	}
	if err := s.landlordAuthRepo.TouchLastLogin(ctx, authRecord.ID, input.LoginIP); err != nil {
		applog.Result(ctx, "publish.auth.login.success", "publish.auth.login.failed", errcode.DatabaseError.WithError(err), "landlord_id", landlord.ID.Hex(), "step", "touch_last_login")
	}
	slog.InfoContext(ctx, "publish.auth.login.session_created", "landlord_id", landlord.ID.Hex())

	result := &LoginResult{
		Token:       token,
		AuthSession: *authSession,
	}
	slog.InfoContext(ctx, "publish.auth.login.success", "landlord_id", landlord.ID.Hex())
	return result, nil
}

func (s *Service) Session(ctx context.Context, principal session.Principal) (*AuthSession, error) {
	slog.InfoContext(ctx, "publish.auth.session.start")
	if principal.Terminal != session.TerminalPublish || principal.PrincipalType != session.PrincipalTypeLandlord {
		err := errcode.Unauthorized
		applog.Result(ctx, "publish.auth.session.success", "publish.auth.session.failed", err, "principal_id", principal.PrincipalID)
		return nil, err
	}
	landlordID, err := bson.ObjectIDFromHex(principal.PrincipalID)
	if err != nil {
		authErr := errcode.Unauthorized.WithError(err)
		applog.Result(ctx, "publish.auth.session.success", "publish.auth.session.failed", authErr, "principal_id", principal.PrincipalID)
		return nil, authErr
	}
	landlord, err := s.landlordRepo.FindActiveByID(ctx, landlordID)
	if err != nil {
		dbErr := errcode.DatabaseError.WithError(err)
		applog.Result(ctx, "publish.auth.session.success", "publish.auth.session.failed", dbErr, "landlord_id", landlordID.Hex())
		return nil, dbErr
	}
	if landlord == nil {
		authErr := errcode.Unauthorized.WithError(fmt.Errorf("登录会话已失效，请重新登录"))
		applog.Result(ctx, "publish.auth.session.success", "publish.auth.session.failed", authErr, "landlord_id", landlordID.Hex())
		return nil, authErr
	}
	authSession, err := s.buildSession(landlord)
	applog.Result(ctx, "publish.auth.session.success", "publish.auth.session.failed", err, "landlord_id", landlordID.Hex())
	return authSession, err
}

func (s *Service) Logout(ctx context.Context, token string) error {
	slog.InfoContext(ctx, "publish.auth.logout.start")
	if strings.TrimSpace(token) == "" {
		err := errcode.Unauthorized
		applog.Result(ctx, "publish.auth.logout.success", "publish.auth.logout.failed", err)
		return err
	}
	if err := s.sessionStore.Delete(ctx, token); err != nil {
		cacheErr := errcode.CacheError.WithError(err)
		applog.Result(ctx, "publish.auth.logout.success", "publish.auth.logout.failed", cacheErr)
		return cacheErr
	}
	slog.InfoContext(ctx, "publish.auth.logout.success")
	return nil
}

func (s *Service) buildSession(landlord *authmodel.Landlord) (*AuthSession, error) {
	if landlord == nil || landlord.ID.IsZero() {
		return nil, errcode.DatabaseError.WithError(fmt.Errorf("房东主体数据无效"))
	}
	principal := session.Principal{
		PrincipalType:   session.PrincipalTypeLandlord,
		PrincipalID:     landlord.ID.Hex(),
		Terminal:        session.TerminalPublish,
		Phone:           landlord.Phone,
		RoleCodes:       []string{},
		PermissionCodes: []string{},
	}
	return &AuthSession{
		Principal: principal,
		Subject: Subject{
			ID:    landlord.ID.Hex(),
			Phone: landlord.Phone,
		},
	}, nil
}

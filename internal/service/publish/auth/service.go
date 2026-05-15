package auth

import (
	"context"
	"fmt"
	authmodel "house-manager/internal/model/auth"
	publishauthrepo "house-manager/internal/repository/publish_auth"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	ownerUserRepo ownerUserRepository
	sessionStore  *session.Store
	env           string
}

type ownerUserRepository interface {
	FindActiveByPhone(ctx context.Context, phone string) (*authmodel.User, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*authmodel.User, error)
}

func NewService(
	ownerUserRepo *publishauthrepo.OwnerUserRepository,
	sessionStore *session.Store,
	env string,
) *Service {
	return newService(
		ownerUserRepo,
		sessionStore,
		env,
	)
}

func newService(
	ownerUserRepo ownerUserRepository,
	sessionStore *session.Store,
	env string,
) *Service {
	return &Service{
		ownerUserRepo: ownerUserRepo,
		sessionStore:  sessionStore,
		env:           env,
	}
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	phone := strings.TrimSpace(input.Phone)
	logAuthInfo(ctx, "publish.auth.login.start", "phone", maskPhone(phone))
	if phone == "" {
		err := errcode.InvalidParam.WithError(fmt.Errorf("phone is required"))
		logAuthResult(ctx, "publish.auth.login.success", "publish.auth.login.failed", err, "phone", maskPhone(phone))
		return nil, err
	}
	if s.env != "local" {
		err := errcode.Forbidden.WithError(fmt.Errorf("phone-only publish login is local-only"))
		logAuthResult(ctx, "publish.auth.login.success", "publish.auth.login.failed", err, "phone", maskPhone(phone))
		return nil, err
	}
	if s.ownerUserRepo == nil || s.sessionStore == nil {
		err := errcode.DatabaseError.WithError(fmt.Errorf("publish auth dependency is nil"))
		logAuthResult(ctx, "publish.auth.login.success", "publish.auth.login.failed", err, "phone", maskPhone(phone))
		return nil, err
	}

	user, err := s.ownerUserRepo.FindActiveByPhone(ctx, phone)
	if err != nil {
		dbErr := errcode.DatabaseError.WithError(err)
		logAuthResult(ctx, "publish.auth.login.success", "publish.auth.login.failed", dbErr, "phone", maskPhone(phone))
		return nil, dbErr
	}
	if user == nil {
		authErr := errcode.Unauthorized.WithError(fmt.Errorf("user not found"))
		logAuthResult(ctx, "publish.auth.login.success", "publish.auth.login.failed", authErr, "phone", maskPhone(phone))
		return nil, authErr
	}
	logAuthInfo(ctx, "publish.auth.login.user_found", "user_id", user.ID.Hex(), "phone", maskPhone(user.Phone))

	authSession, err := s.buildSession(user)
	if err != nil {
		logAuthResult(ctx, "publish.auth.login.success", "publish.auth.login.failed", err, "user_id", user.ID.Hex())
		return nil, err
	}
	token, err := s.sessionStore.CreatePrincipal(ctx, authSession.Principal)
	if err != nil {
		cacheErr := errcode.CacheError.WithError(err)
		logAuthResult(ctx, "publish.auth.login.success", "publish.auth.login.failed", cacheErr, "user_id", user.ID.Hex())
		return nil, cacheErr
	}
	logAuthInfo(ctx, "publish.auth.login.session_created", "user_id", user.ID.Hex())

	result := &LoginResult{
		Token:       token,
		AuthSession: *authSession,
	}
	logAuthInfo(ctx, "publish.auth.login.success", "user_id", user.ID.Hex())
	return result, nil
}

func (s *Service) Session(ctx context.Context, principal session.Principal) (*AuthSession, error) {
	logAuthInfo(ctx, "publish.auth.session.start")
	if principal.Terminal != session.TerminalPublish || principal.PrincipalType != session.PrincipalTypeUser {
		err := errcode.Unauthorized
		logAuthResult(ctx, "publish.auth.session.success", "publish.auth.session.failed", err, "principal_id", principal.PrincipalID)
		return nil, err
	}
	userID, err := bson.ObjectIDFromHex(principal.PrincipalID)
	if err != nil {
		authErr := errcode.Unauthorized.WithError(err)
		logAuthResult(ctx, "publish.auth.session.success", "publish.auth.session.failed", authErr, "principal_id", principal.PrincipalID)
		return nil, authErr
	}
	user, err := s.ownerUserRepo.FindByID(ctx, userID)
	if err != nil {
		dbErr := errcode.DatabaseError.WithError(err)
		logAuthResult(ctx, "publish.auth.session.success", "publish.auth.session.failed", dbErr, "user_id", userID.Hex())
		return nil, dbErr
	}
	if user == nil {
		authErr := errcode.Unauthorized.WithError(fmt.Errorf("user not found"))
		logAuthResult(ctx, "publish.auth.session.success", "publish.auth.session.failed", authErr, "user_id", userID.Hex())
		return nil, authErr
	}
	authSession, err := s.buildSession(user)
	logAuthResult(ctx, "publish.auth.session.success", "publish.auth.session.failed", err, "user_id", userID.Hex())
	return authSession, err
}

func (s *Service) Logout(ctx context.Context, token string) error {
	logAuthInfo(ctx, "publish.auth.logout.start")
	if strings.TrimSpace(token) == "" {
		err := errcode.Unauthorized
		logAuthResult(ctx, "publish.auth.logout.success", "publish.auth.logout.failed", err)
		return err
	}
	if err := s.sessionStore.Delete(ctx, token); err != nil {
		cacheErr := errcode.CacheError.WithError(err)
		logAuthResult(ctx, "publish.auth.logout.success", "publish.auth.logout.failed", cacheErr)
		return cacheErr
	}
	logAuthInfo(ctx, "publish.auth.logout.success")
	return nil
}

func (s *Service) buildSession(user *authmodel.User) (*AuthSession, error) {
	if user == nil || user.ID.IsZero() {
		return nil, errcode.DatabaseError.WithError(fmt.Errorf("user is required"))
	}
	principal := session.Principal{
		PrincipalType:   session.PrincipalTypeUser,
		PrincipalID:     user.ID.Hex(),
		Terminal:        session.TerminalPublish,
		Phone:           user.Phone,
		RoleCodes:       []string{},
		PermissionCodes: []string{},
	}
	return &AuthSession{
		Principal: principal,
		Subject: Subject{
			ID:    user.ID.Hex(),
			Name:  user.Nickname,
			Phone: user.Phone,
		},
	}, nil
}

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
	if phone == "" {
		return nil, errcode.InvalidParam.WithError(fmt.Errorf("phone is required"))
	}
	if s.env != "local" {
		return nil, errcode.Forbidden.WithError(fmt.Errorf("phone-only publish login is local-only"))
	}
	if s.ownerUserRepo == nil || s.sessionStore == nil {
		return nil, errcode.DatabaseError.WithError(fmt.Errorf("publish auth dependency is nil"))
	}

	user, err := s.ownerUserRepo.FindActiveByPhone(ctx, phone)
	if err != nil {
		return nil, errcode.DatabaseError.WithError(err)
	}
	if user == nil {
		return nil, errcode.Unauthorized.WithError(fmt.Errorf("user not found"))
	}

	authSession, err := s.buildSession(user)
	if err != nil {
		return nil, err
	}
	token, err := s.sessionStore.CreatePrincipal(ctx, authSession.Principal)
	if err != nil {
		return nil, errcode.CacheError.WithError(err)
	}

	return &LoginResult{
		Token:       token,
		AuthSession: *authSession,
	}, nil
}

func (s *Service) Session(ctx context.Context, principal session.Principal) (*AuthSession, error) {
	if principal.Terminal != session.TerminalPublish || principal.PrincipalType != session.PrincipalTypeUser {
		return nil, errcode.Unauthorized
	}
	userID, err := bson.ObjectIDFromHex(principal.PrincipalID)
	if err != nil {
		return nil, errcode.Unauthorized.WithError(err)
	}
	user, err := s.ownerUserRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errcode.DatabaseError.WithError(err)
	}
	if user == nil {
		return nil, errcode.Unauthorized.WithError(fmt.Errorf("user not found"))
	}
	return s.buildSession(user)
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return errcode.Unauthorized
	}
	if err := s.sessionStore.Delete(ctx, token); err != nil {
		return errcode.CacheError.WithError(err)
	}
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

package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"house-manager/pkg/cache"
)

const keyPrefix = "hs:sess:"

type principalContextKey struct{}

const (
	PrincipalTypeUser     = "user"
	PrincipalTypeLandlord = "landlord"
	PrincipalTypeStaff    = "staff"

	TerminalMiniapp = "miniapp"
	TerminalPublish = "publish"
	TerminalAdmin   = "admin"
)

// Principal 是 Redis session 中保存的结构化登录身份。
type Principal struct {
	PrincipalType   string   `json:"principal_type"`
	PrincipalID     string   `json:"principal_id"`
	Terminal        string   `json:"terminal"`
	Phone           string   `json:"phone"`
	RoleCodes       []string `json:"role_codes"`
	PermissionCodes []string `json:"permission_codes"`
	SessionVersion  string   `json:"session_version,omitempty"`
}

// Store Redis session 存储
type Store struct {
	client cache.Client
	ttl    time.Duration
}

// NewStore 创建 session store（复用 cache.Client 接口）
func NewStore(client cache.Client, ttl time.Duration) *Store {
	return &Store{client: client, ttl: ttl}
}

// Create 创建小程序 user session，返回 opaque token。
func (s *Store) Create(ctx context.Context, userId string) (string, error) {
	return s.CreateMiniappUser(ctx, userId, "")
}

// CreateMiniappUser 创建小程序用户 session，返回 opaque token。
func (s *Store) CreateMiniappUser(ctx context.Context, userID, phone string) (string, error) {
	return s.CreatePrincipal(ctx, Principal{
		PrincipalType: PrincipalTypeUser,
		PrincipalID:   userID,
		Terminal:      TerminalMiniapp,
		Phone:         phone,
	})
}

// CreatePrincipal 创建结构化 principal session，返回 opaque token。
func (s *Store) CreatePrincipal(ctx context.Context, principal Principal) (string, error) {
	if err := principal.Validate(); err != nil {
		return "", err
	}
	normalized := principal.normalized()
	version, err := s.currentVersion(ctx, normalized)
	if err != nil {
		return "", fmt.Errorf("get session version: %w", err)
	}
	normalized.SessionVersion = version
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("marshal principal: %w", err)
	}
	if err := s.client.Set(ctx, keyPrefix+token, string(payload), s.ttl); err != nil {
		return "", fmt.Errorf("save session: %w", err)
	}
	return token, nil
}

// Get 根据 token 获取 userId，空字符串表示 session 不存在或不是 user principal。
func (s *Store) Get(ctx context.Context, token string) (string, error) {
	return s.GetUserID(ctx, token)
}

// GetUserID 根据 token 获取 user principal ID。
func (s *Store) GetUserID(ctx context.Context, token string) (string, error) {
	principal, err := s.GetPrincipal(ctx, token)
	if err != nil || principal == nil {
		return "", err
	}
	if principal.PrincipalType != PrincipalTypeUser {
		return "", nil
	}
	return principal.PrincipalID, nil
}

// GetPrincipal 根据 token 获取结构化 principal，nil 表示 session 不存在。
func (s *Store) GetPrincipal(ctx context.Context, token string) (*Principal, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, nil
	}
	key := keyPrefix + token
	val, err := s.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, cache.ErrNil) {
			return nil, nil
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	var principal Principal
	if err := json.Unmarshal([]byte(val), &principal); err != nil {
		return nil, fmt.Errorf("decode session: %w", err)
	}
	if err := principal.Validate(); err != nil {
		return nil, fmt.Errorf("validate session: %w", err)
	}
	normalized := principal.normalized()
	currentVersion, err := s.currentVersion(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("get session version: %w", err)
	}
	if currentVersion != "" && normalized.SessionVersion != currentVersion {
		return nil, nil
	}
	if s.ttl > 0 {
		payload, err := json.Marshal(normalized)
		if err != nil {
			return nil, fmt.Errorf("marshal session: %w", err)
		}
		if err := s.client.Set(ctx, key, string(payload), s.ttl); err != nil {
			return nil, fmt.Errorf("refresh session: %w", err)
		}
	}
	return &normalized, nil
}

// Delete 删除 session
func (s *Store) Delete(ctx context.Context, token string) error {
	return s.client.Del(ctx, keyPrefix+token)
}

// InvalidatePrincipal makes previously issued sessions for the same principal invalid.
func (s *Store) InvalidatePrincipal(ctx context.Context, principal Principal) error {
	if err := principal.Validate(); err != nil {
		return err
	}
	version := fmt.Sprintf("%d", time.Now().UnixNano())
	if err := s.client.Set(ctx, versionKey(principal.normalized()), version, s.ttl); err != nil {
		return fmt.Errorf("invalidate principal sessions: %w", err)
	}
	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (p Principal) Validate() error {
	if strings.TrimSpace(p.PrincipalType) == "" {
		return fmt.Errorf("principal_type is required")
	}
	if strings.TrimSpace(p.PrincipalID) == "" {
		return fmt.Errorf("principal_id is required")
	}
	if strings.TrimSpace(p.Terminal) == "" {
		return fmt.Errorf("terminal is required")
	}
	switch p.PrincipalType {
	case PrincipalTypeUser, PrincipalTypeLandlord, PrincipalTypeStaff:
	default:
		return fmt.Errorf("principal_type is invalid")
	}
	switch p.Terminal {
	case TerminalMiniapp, TerminalPublish, TerminalAdmin:
	default:
		return fmt.Errorf("terminal is invalid")
	}
	return nil
}

func (p Principal) normalized() Principal {
	p.PrincipalType = strings.TrimSpace(p.PrincipalType)
	p.PrincipalID = strings.TrimSpace(p.PrincipalID)
	p.Terminal = strings.TrimSpace(p.Terminal)
	p.Phone = strings.TrimSpace(p.Phone)
	p.SessionVersion = strings.TrimSpace(p.SessionVersion)
	if p.RoleCodes == nil {
		p.RoleCodes = []string{}
	}
	if p.PermissionCodes == nil {
		p.PermissionCodes = []string{}
	}
	return p
}

func (s *Store) currentVersion(ctx context.Context, principal Principal) (string, error) {
	version, err := s.client.Get(ctx, versionKey(principal.normalized()))
	if err != nil {
		if errors.Is(err, cache.ErrNil) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(version), nil
}

func versionKey(principal Principal) string {
	normalized := principal.normalized()
	return fmt.Sprintf("%sver:%s:%s:%s", keyPrefix, normalized.Terminal, normalized.PrincipalType, normalized.PrincipalID)
}

// ContextWithPrincipal 把结构化登录身份放入 request context。
func ContextWithPrincipal(ctx context.Context, principal Principal) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	normalized := principal.normalized()
	return context.WithValue(ctx, principalContextKey{}, normalized)
}

// PrincipalFromContext 从 request context 读取结构化登录身份。
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	if ctx == nil {
		return Principal{}, false
	}
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	if !ok {
		return Principal{}, false
	}
	if err := principal.Validate(); err != nil {
		return Principal{}, false
	}
	return principal.normalized(), true
}

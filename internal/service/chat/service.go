package chat

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	chatmodel "house-manager/internal/model/chat"

	"house-manager/pkg/errcode"
)

const defaultRecentMessageLimit = 20

type Service struct {
	ai                 AIResponder
	sessions           sessionRepository
	messages           messageRepository
	runtimeContext     runtimeContextStore
	sessionIdleTimeout time.Duration
	recentMessageLimit int
	now                func() time.Time
}

type sessionRepository interface {
	FindActiveByUserID(ctx context.Context, userID bson.ObjectID, minLastActiveAt int64) (*chatmodel.Session, error)
	Create(ctx context.Context, entity *chatmodel.Session) error
	TouchActive(ctx context.Context, sessionID bson.ObjectID, lastActiveAt int64) error
	End(ctx context.Context, sessionID bson.ObjectID, endedAt int64) error
	AllocateMessageSeq(ctx context.Context, sessionID bson.ObjectID) (int, error)
}

type messageRepository interface {
	Create(ctx context.Context, entity *chatmodel.Message) error
	ListRecentBySessionID(ctx context.Context, sessionID bson.ObjectID, limit int) ([]chatmodel.Message, error)
}

type runtimeContextStore interface {
	Get(ctx context.Context, sessionID bson.ObjectID) (map[string]any, error)
	Set(ctx context.Context, sessionID bson.ObjectID, runtimeContext map[string]any) error
}

func NewService(ai AIResponder, sessions sessionRepository, messages messageRepository, runtimeContext runtimeContextStore, cfg ServiceConfig) *Service {
	return newService(ai, sessions, messages, runtimeContext, cfg)
}

func newService(ai AIResponder, sessions sessionRepository, messages messageRepository, runtimeContext runtimeContextStore, cfg ServiceConfig) *Service {
	recentLimit := cfg.RecentMessageLimit
	if recentLimit <= 0 {
		recentLimit = defaultRecentMessageLimit
	}
	return &Service{
		ai:                 ai,
		sessions:           sessions,
		messages:           messages,
		runtimeContext:     runtimeContext,
		sessionIdleTimeout: cfg.SessionIdleTimeout,
		recentMessageLimit: recentLimit,
		now:                time.Now,
	}
}

func (s *Service) Send(ctx context.Context, input SendInput) (*SendResult, error) {
	startedAt := time.Now()
	message := strings.TrimSpace(input.Message)
	if message == "" {
		return nil, errcode.InvalidParam.WithError(fmt.Errorf("message 不能为空"))
	}
	userID, err := bson.ObjectIDFromHex(strings.TrimSpace(input.UserID))
	if err != nil {
		return nil, errcode.InvalidParam.WithError(fmt.Errorf("user_id is invalid"))
	}
	requestID := strings.TrimSpace(input.RequestID)
	if requestID == "" {
		requestID = requestIDFromContext(ctx)
	}
	if requestID == "" {
		requestID = bson.NewObjectID().Hex()
	}
	slog.InfoContext(ctx, "chat.send.start",
		"request_id", requestID,
		"user_id", userID.Hex(),
		"new_session", input.NewSession,
		"message_chars", len(message),
	)

	session, err := s.openSession(ctx, userID, input.NewSession)
	if err != nil {
		return nil, err
	}
	if err := s.appendMessage(ctx, session.ID, userID, chatmodel.MessageRoleUser, message); err != nil {
		return nil, err
	}
	recentMessages, err := s.recentMessages(ctx, session.ID)
	if err != nil {
		return nil, err
	}
	runtimeContext, err := s.loadRuntimeContext(ctx, session.ID)
	if err != nil {
		return nil, err
	}

	if s == nil || s.ai == nil {
		return nil, errcode.InternalError.WithError(fmt.Errorf("AI 回复服务未配置"))
	}
	aiStartedAt := time.Now()
	slog.InfoContext(ctx, "chat.send.ai_start",
		"request_id", requestID,
		"session_id", session.ID.Hex(),
		"recent_message_count", len(recentMessages),
		"context_keys", len(runtimeContext),
	)
	aiOutput, err := s.ai.Respond(ctx, AIRespondInput{
		RequestID:      requestID,
		SessionID:      session.ID.Hex(),
		Message:        message,
		RecentMessages: recentMessages,
		RuntimeContext: runtimeContext,
	})
	if err != nil {
		slog.ErrorContext(ctx, "chat.send.python_failed",
			"request_id", requestID,
			"session_id", session.ID.Hex(),
			"duration", time.Since(aiStartedAt),
			"error", err,
		)
		return nil, errcode.InternalError.WithError(fmt.Errorf("AI 回复生成失败，请稍后再试"))
	}
	slog.InfoContext(ctx, "chat.send.ai_success",
		"request_id", requestID,
		"session_id", session.ID.Hex(),
		"duration", time.Since(aiStartedAt),
		"sentence_count", len(aiOutput.Sentences),
		"house_count", len(aiOutput.HouseList),
	)
	if strings.TrimSpace(aiOutput.AssistantMessage) == "" {
		return nil, errcode.InternalError.WithError(fmt.Errorf("AI 回复为空，请稍后再试"))
	}
	if err := s.appendMessage(ctx, session.ID, userID, chatmodel.MessageRoleAssistant, aiOutput.AssistantMessage); err != nil {
		return nil, err
	}
	if err := s.saveRuntimeContext(ctx, session.ID, aiOutput.UpdatedContext); err != nil {
		return nil, err
	}
	if err := s.touchSession(ctx, session.ID); err != nil {
		return nil, err
	}

	sentences := aiOutput.Sentences
	if len(sentences) == 0 && strings.TrimSpace(aiOutput.AssistantMessage) != "" {
		sentences = []string{aiOutput.AssistantMessage}
	}
	houseList := aiOutput.HouseList
	if houseList == nil {
		houseList = []HouseItem{}
	}
	slog.InfoContext(ctx, "chat.send.success",
		"request_id", requestID,
		"session_id", session.ID.Hex(),
		"duration", time.Since(startedAt),
		"sentence_count", len(sentences),
		"house_count", len(houseList),
	)
	return &SendResult{
		SessionID:        session.ID.Hex(),
		AssistantMessage: aiOutput.AssistantMessage,
		Sentences:        sentences,
		HouseList:        houseList,
	}, nil
}

func (s *Service) openSession(ctx context.Context, userID bson.ObjectID, forceNew bool) (*chatmodel.Session, error) {
	if s == nil || s.sessions == nil {
		return nil, errcode.DatabaseError.WithError(fmt.Errorf("chat session repository is nil"))
	}
	now := s.now().Unix()
	minLastActiveAt := int64(0)
	if s.sessionIdleTimeout > 0 {
		minLastActiveAt = now - int64(s.sessionIdleTimeout.Seconds())
	}

	active, err := s.sessions.FindActiveByUserID(ctx, userID, 0)
	if err != nil {
		return nil, errcode.DatabaseError.WithError(fmt.Errorf("find active chat session: %w", err))
	}
	if active != nil {
		expired := minLastActiveAt > 0 && active.LastActiveAt < minLastActiveAt
		if !forceNew && !expired {
			return active, nil
		}
		if err := s.sessions.End(ctx, active.ID, now); err != nil {
			return nil, errcode.DatabaseError.WithError(fmt.Errorf("end active chat session: %w", err))
		}
	}

	session := &chatmodel.Session{
		UserID:        userID,
		SessionStatus: chatmodel.SessionStatusActive,
		StartedAt:     now,
		LastActiveAt:  now,
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, errcode.DatabaseError.WithError(fmt.Errorf("create chat session: %w", err))
	}
	return session, nil
}

func (s *Service) appendMessage(ctx context.Context, sessionID, userID bson.ObjectID, role string, content string) error {
	if s == nil || s.messages == nil {
		return errcode.DatabaseError.WithError(fmt.Errorf("chat message repository is nil"))
	}
	seq, err := s.sessions.AllocateMessageSeq(ctx, sessionID)
	if err != nil {
		return errcode.DatabaseError.WithError(fmt.Errorf("allocate chat message seq: %w", err))
	}
	message := &chatmodel.Message{
		SessionID: sessionID,
		UserID:    userID,
		Seq:       seq,
		Role:      role,
		Content:   content,
	}
	if err := s.messages.Create(ctx, message); err != nil {
		return errcode.DatabaseError.WithError(fmt.Errorf("create chat message: %w", err))
	}
	return nil
}

func (s *Service) recentMessages(ctx context.Context, sessionID bson.ObjectID) ([]RecentMessage, error) {
	items, err := s.messages.ListRecentBySessionID(ctx, sessionID, s.recentMessageLimit)
	if err != nil {
		return nil, errcode.DatabaseError.WithError(fmt.Errorf("list recent chat messages: %w", err))
	}
	out := make([]RecentMessage, 0, len(items))
	for _, item := range items {
		out = append(out, RecentMessage{Role: item.Role, Content: item.Content})
	}
	return out, nil
}

func (s *Service) loadRuntimeContext(ctx context.Context, sessionID bson.ObjectID) (map[string]any, error) {
	if s == nil || s.runtimeContext == nil {
		return nil, errcode.CacheError.WithError(fmt.Errorf("runtime context store is nil"))
	}
	runtimeContext, err := s.runtimeContext.Get(ctx, sessionID)
	if err != nil {
		return nil, errcode.CacheError.WithError(fmt.Errorf("get runtime context: %w", err))
	}
	if runtimeContext == nil {
		return defaultRuntimeContext(), nil
	}
	return runtimeContext, nil
}

func (s *Service) saveRuntimeContext(ctx context.Context, sessionID bson.ObjectID, runtimeContext map[string]any) error {
	if s == nil || s.runtimeContext == nil {
		return errcode.CacheError.WithError(fmt.Errorf("runtime context store is nil"))
	}
	if runtimeContext == nil {
		runtimeContext = defaultRuntimeContext()
	}
	if err := s.runtimeContext.Set(ctx, sessionID, runtimeContext); err != nil {
		return errcode.CacheError.WithError(fmt.Errorf("set runtime context: %w", err))
	}
	return nil
}

func (s *Service) touchSession(ctx context.Context, sessionID bson.ObjectID) error {
	if err := s.sessions.TouchActive(ctx, sessionID, s.now().Unix()); err != nil {
		return errcode.DatabaseError.WithError(fmt.Errorf("touch active chat session: %w", err))
	}
	return nil
}

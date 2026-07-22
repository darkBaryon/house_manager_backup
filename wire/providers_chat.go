package wire

import (
	"context"

	"house-manager/internal/config"
	chathandler "house-manager/internal/handler/v1/miniapp/chat"
	"house-manager/internal/integration/pythonchat"
	repochat "house-manager/internal/repository/chat"
	chatsvc "house-manager/internal/service/chat"
	dbmongo "house-manager/pkg/database/mongo"
	dbredis "house-manager/pkg/database/redis"

	"github.com/google/wire"
)

func newChatAIResponder(cfg *config.Config) (chatsvc.AIResponder, error) {
	timeout, err := cfg.AIChat.RespondTimeoutDuration()
	if err != nil {
		return nil, err
	}
	return pythonchat.NewClient(pythonchat.Config{
		BaseURL:       cfg.AIChat.PythonBaseURL,
		InternalToken: cfg.AIChat.InternalToken,
		Timeout:       timeout,
	}), nil
}

func newChatSessionRepository(ctx context.Context, client *dbmongo.Client) (*repochat.SessionRepository, error) {
	repo := repochat.NewSessionRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newChatMessageRepository(ctx context.Context, client *dbmongo.Client) (*repochat.MessageRepository, error) {
	repo := repochat.NewMessageRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newRuntimeContextStore(cfg *config.Config, client *dbredis.Client) (*chatsvc.RuntimeContextStore, error) {
	ttl, err := cfg.AIChat.RuntimeContextTTLDuration()
	if err != nil {
		return nil, err
	}
	return chatsvc.NewRuntimeContextStore(client, ttl), nil
}

func newChatService(
	cfg *config.Config,
	ai chatsvc.AIResponder,
	sessions *repochat.SessionRepository,
	messages *repochat.MessageRepository,
	runtimeContext *chatsvc.RuntimeContextStore,
) (*chatsvc.Service, error) {
	idleTimeout, err := cfg.AIChat.SessionIdleTimeoutDuration()
	if err != nil {
		return nil, err
	}
	return chatsvc.NewService(ai, sessions, messages, runtimeContext, chatsvc.ServiceConfig{
		SessionIdleTimeout: idleTimeout,
		RecentMessageLimit: cfg.AIChat.RecentMessageLimit,
	}), nil
}

var ChatSet = wire.NewSet(
	newChatAIResponder,
	newChatSessionRepository,
	newChatMessageRepository,
	newRuntimeContextStore,
	newChatService,
	chathandler.NewHandler,
)

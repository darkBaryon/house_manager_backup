package chat

import (
	"context"
	"testing"
	"time"

	chatmodel "house-manager/internal/model/chat"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type fakeSessionRepository struct {
	active  *chatmodel.Session
	ended   []bson.ObjectID
	touched []bson.ObjectID
	created []*chatmodel.Session
}

func (r *fakeSessionRepository) FindActiveByUserID(ctx context.Context, userID bson.ObjectID, minLastActiveAt int64) (*chatmodel.Session, error) {
	return r.active, nil
}

func (r *fakeSessionRepository) Create(ctx context.Context, entity *chatmodel.Session) error {
	if entity.ID.IsZero() {
		entity.ID = bson.NewObjectID()
	}
	r.created = append(r.created, entity)
	r.active = entity
	return nil
}

func (r *fakeSessionRepository) TouchActive(ctx context.Context, sessionID bson.ObjectID, lastActiveAt int64) error {
	r.touched = append(r.touched, sessionID)
	return nil
}

func (r *fakeSessionRepository) End(ctx context.Context, sessionID bson.ObjectID, endedAt int64) error {
	r.ended = append(r.ended, sessionID)
	return nil
}

func (r *fakeSessionRepository) AllocateMessageSeq(ctx context.Context, sessionID bson.ObjectID) (int, error) {
	if r.active != nil && r.active.ID == sessionID {
		r.active.LastSeq++
		return r.active.LastSeq, nil
	}
	return 0, nil
}

type fakeMessageRepository struct {
	items []chatmodel.Message
}

func (r *fakeMessageRepository) Create(ctx context.Context, entity *chatmodel.Message) error {
	r.items = append(r.items, *entity)
	return nil
}

func (r *fakeMessageRepository) ListRecentBySessionID(ctx context.Context, sessionID bson.ObjectID, limit int) ([]chatmodel.Message, error) {
	out := make([]chatmodel.Message, 0, len(r.items))
	for _, item := range r.items {
		if item.SessionID == sessionID {
			out = append(out, item)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}

type fakeRuntimeContextStore struct {
	value map[string]any
	saved map[string]any
}

func (s *fakeRuntimeContextStore) Get(ctx context.Context, sessionID bson.ObjectID) (map[string]any, error) {
	if s.value == nil {
		return defaultRuntimeContext(), nil
	}
	return s.value, nil
}

func (s *fakeRuntimeContextStore) Set(ctx context.Context, sessionID bson.ObjectID, runtimeContext map[string]any) error {
	s.saved = runtimeContext
	return nil
}

type fakeAIResponder struct {
	input  AIRespondInput
	output AIRespondOutput
	err    error
}

func (r *fakeAIResponder) Respond(ctx context.Context, input AIRespondInput) (AIRespondOutput, error) {
	r.input = input
	if r.err != nil {
		return AIRespondOutput{}, r.err
	}
	return r.output, nil
}

func TestServiceSendCallsAIResponder(t *testing.T) {
	userID := bson.NewObjectID()
	sessions := &fakeSessionRepository{}
	messages := &fakeMessageRepository{}
	runtimeContext := &fakeRuntimeContextStore{value: map[string]any{
		"dialogue_state": "IDLE",
		"requirement":    map[string]any{"district": "南山"},
	}}
	ai := &fakeAIResponder{output: AIRespondOutput{
		AssistantMessage: "可以，我先帮你看看。",
		Sentences:        []string{"可以，我先帮你看看。"},
		HouseList:        []HouseItem{},
		UpdatedContext:   map[string]any{"dialogue_state": "IDLE"},
	}}
	service := newService(
		ai,
		sessions,
		messages,
		runtimeContext,
		ServiceConfig{SessionIdleTimeout: 30 * time.Minute, RecentMessageLimit: 20},
	)
	fixedNow := time.Unix(1775451571, 0)
	service.now = func() time.Time { return fixedNow }
	result, err := service.Send(t.Context(), SendInput{
		RequestID: "req-1",
		UserID:    userID.Hex(),
		Message:   " 南山单间 ",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if ai.input.RequestID != "req-1" || ai.input.SessionID == "" {
		t.Fatalf("unexpected AI identity: %+v", ai.input)
	}
	if ai.input.Message != "南山单间" {
		t.Fatalf("message = %q", ai.input.Message)
	}
	if ai.input.RuntimeContext["dialogue_state"] != "IDLE" {
		t.Fatalf("dialogue state = %#v", ai.input.RuntimeContext["dialogue_state"])
	}
	if ai.input.RuntimeContext["requirement"].(map[string]any)["district"] != "南山" {
		t.Fatalf("runtime context = %#v", ai.input.RuntimeContext)
	}
	if len(ai.input.RecentMessages) != 1 || ai.input.RecentMessages[0].Role != chatmodel.MessageRoleUser || ai.input.RecentMessages[0].Content != "南山单间" {
		t.Fatalf("recent messages = %+v", ai.input.RecentMessages)
	}
	if result.AssistantMessage != "可以，我先帮你看看。" {
		t.Fatalf("assistant message = %q", result.AssistantMessage)
	}
	if result.SessionID != sessions.active.ID.Hex() {
		t.Fatalf("session id = %q, want %q", result.SessionID, sessions.active.ID.Hex())
	}
	if len(messages.items) != 2 {
		t.Fatalf("messages count = %d", len(messages.items))
	}
	if messages.items[0].Role != chatmodel.MessageRoleUser || messages.items[1].Role != chatmodel.MessageRoleAssistant {
		t.Fatalf("message roles = %+v", messages.items)
	}
	if messages.items[0].Seq != 1 || messages.items[1].Seq != 2 {
		t.Fatalf("message seq = %+v", messages.items)
	}
	if len(sessions.touched) != 1 || sessions.touched[0] != sessions.active.ID {
		t.Fatalf("touched sessions = %+v", sessions.touched)
	}
	if runtimeContext.saved["dialogue_state"] != "IDLE" {
		t.Fatalf("saved runtime context = %#v", runtimeContext.saved)
	}
}

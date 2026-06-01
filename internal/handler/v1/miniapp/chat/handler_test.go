package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	chatsvc "house-manager/internal/service/chat"
	"house-manager/pkg/errcode"
	"house-manager/pkg/requestlog"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type fakeChatService struct {
	input  chatsvc.SendInput
	result *chatsvc.SendResult
	err    error
}

func (s *fakeChatService) Send(ctx context.Context, input chatsvc.SendInput) (*chatsvc.SendResult, error) {
	s.input = input
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}

func TestSendUsesCurrentUserAndReturnsResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := bson.NewObjectID()
	service := &fakeChatService{result: &chatsvc.SendResult{
		SessionID:        bson.NewObjectID().Hex(),
		AssistantMessage: "可以，我先帮你看看。",
		Sentences:        []string{"可以，我先帮你看看。"},
		HouseList: []chatsvc.HouseItem{{
			ListingID: userID.Hex(),
			Title:     "测试房源",
			Price:     4200,
		}},
	}}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userId", userID.Hex())
		requestlog.SetRequestID(c, "req-1")
		c.Next()
	})
	(&Handler{service: service}).RegisterRoutes(router.Group(""))

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chat/send", bytes.NewReader([]byte(`{"message":" 南山单间 ","new_session":true}`)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", resp.Code, resp.Body.String())
	}
	if service.input.UserID != userID.Hex() || service.input.RequestID != "req-1" || service.input.Message != "南山单间" || !service.input.NewSession {
		t.Fatalf("unexpected service input: %#v", service.input)
	}
	var got response.Response
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Code != 0 {
		t.Fatalf("code = %d error = %s", got.Code, got.Error)
	}
	data := got.Data.(map[string]any)
	if data["assistant_message"] != "可以，我先帮你看看。" {
		t.Fatalf("data = %#v", data)
	}
}

func TestSendRequiresCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeChatService{}
	router := gin.New()
	(&Handler{service: service}).RegisterRoutes(router.Group(""))

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chat/send", bytes.NewReader([]byte(`{"message":"hi"}`)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body = %s", resp.Code, resp.Body.String())
	}
	if service.input.Message != "" {
		t.Fatalf("service should not be called, got %#v", service.input)
	}
	var got response.Response
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Code != errcode.Unauthorized.Code {
		t.Fatalf("code = %d", got.Code)
	}
}

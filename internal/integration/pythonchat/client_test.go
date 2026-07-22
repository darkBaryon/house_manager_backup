package pythonchat

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	chatservice "house-manager/internal/service/chat"
)

func TestClientRespondSendsInternalContract(t *testing.T) {
	var gotPath string
	var gotToken string
	var gotRequestID string
	var gotReq respondRequest
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotToken = r.Header.Get("Internal-Token")
		gotRequestID = r.Header.Get("Request-ID")
		var err error
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request: %v", err)
		}
		if err := json.Unmarshal(gotBody, &gotReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(envelope{
			Code: 0,
			Data: &respondData{
				AssistantMessage: "可以，我先帮你看看。",
				Sentences:        []string{"可以，我先帮你看看。"},
				HouseList:        []houseItem{},
				UpdatedContext:   map[string]any{"dialogue_state": "IDLE"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, InternalToken: "token-1", Timeout: time.Second})
	output, err := client.Respond(t.Context(), chatservice.AIRespondInput{
		RequestID: "req-1",
		SessionID: "session-1",
		Message:   "南山单间",
		RecentMessages: []chatservice.RecentMessage{
			{Role: "user", Content: "南山单间"},
		},
		RuntimeContext: map[string]any{"dialogue_state": "IDLE"},
	})
	if err != nil {
		t.Fatalf("Respond() error = %v", err)
	}
	if gotPath != respondPath {
		t.Fatalf("path = %q", gotPath)
	}
	if gotToken != "token-1" {
		t.Fatalf("internal token = %q", gotToken)
	}
	if gotRequestID != "req-1" {
		t.Fatalf("request id header = %q", gotRequestID)
	}
	if bytes.Contains(gotBody, []byte("request_id")) || bytes.Contains(gotBody, []byte("user_id")) {
		t.Fatalf("body must not contain request_id/user_id: %s", string(gotBody))
	}
	if gotReq.SessionID != "session-1" || gotReq.Message != "南山单间" {
		t.Fatalf("request = %+v", gotReq)
	}
	if output.AssistantMessage != "可以，我先帮你看看。" {
		t.Fatalf("assistant message = %q", output.AssistantMessage)
	}
}

func TestClientRespondHidesHTTPErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal trace with prompt text", http.StatusBadGateway)
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, InternalToken: "token-1", Timeout: time.Second})
	_, err := client.Respond(t.Context(), chatservice.AIRespondInput{RequestID: "req-1"})
	if err == nil {
		t.Fatal("expected error")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected pythonchat.Error, got %T", err)
	}
	if clientErr.Kind != ErrorKindHTTPStatus || clientErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("unexpected error = %+v", clientErr)
	}
	if strings.Contains(err.Error(), "internal trace") || strings.Contains(err.Error(), "prompt text") {
		t.Fatalf("error leaks response body: %v", err)
	}
}

func TestClientRespondMapsEnvelopeCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(envelope{Code: 20001, Error: "reply generation failed"})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, InternalToken: "token-1", Timeout: time.Second})
	_, err := client.Respond(t.Context(), chatservice.AIRespondInput{RequestID: "req-1"})
	if err == nil {
		t.Fatal("expected error")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected pythonchat.Error, got %T", err)
	}
	if clientErr.Kind != ErrorKindEnvelopeCode || clientErr.Code != 20001 {
		t.Fatalf("unexpected error = %+v", clientErr)
	}
}

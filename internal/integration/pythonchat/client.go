package pythonchat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	chatservice "house-manager/internal/service/chat"
)

const respondPath = "/internal/chat/respond"

type Client struct {
	baseURL       string
	internalToken string
	httpClient    *http.Client
}

var _ chatservice.AIResponder = (*Client)(nil)

func NewClient(cfg Config) *Client {
	return &Client{
		baseURL:       strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		internalToken: strings.TrimSpace(cfg.InternalToken),
		httpClient:    &http.Client{Timeout: cfg.Timeout},
	}
}

func (c *Client) Respond(ctx context.Context, input chatservice.AIRespondInput) (chatservice.AIRespondOutput, error) {
	if err := c.validate(); err != nil {
		return chatservice.AIRespondOutput{}, err
	}
	startedAt := time.Now()
	slog.InfoContext(ctx, "pythonchat.respond.start",
		"request_id", input.RequestID,
		"session_id", input.SessionID,
		"recent_message_count", len(input.RecentMessages),
	)
	req, err := c.newHTTPRequest(ctx, input)
	if err != nil {
		return chatservice.AIRespondOutput{}, err
	}
	body, err := c.do(req)
	if err != nil {
		slog.ErrorContext(ctx, "pythonchat.respond.failed",
			"request_id", input.RequestID,
			"session_id", input.SessionID,
			"duration", time.Since(startedAt),
			"error", err,
		)
		return chatservice.AIRespondOutput{}, err
	}
	output, err := decodeRespondEnvelope(body)
	if err != nil {
		slog.ErrorContext(ctx, "pythonchat.respond.failed",
			"request_id", input.RequestID,
			"session_id", input.SessionID,
			"duration", time.Since(startedAt),
			"error", err,
		)
		return chatservice.AIRespondOutput{}, err
	}
	slog.InfoContext(ctx, "pythonchat.respond.success",
		"request_id", input.RequestID,
		"session_id", input.SessionID,
		"duration", time.Since(startedAt),
		"sentence_count", len(output.Sentences),
		"house_count", len(output.HouseList),
	)
	return output, nil
}

func (c *Client) validate() error {
	if c == nil || c.baseURL == "" {
		return newError(ErrorKindConfig, fmt.Errorf("base url is empty"))
	}
	if c.internalToken == "" {
		return newError(ErrorKindConfig, fmt.Errorf("internal token is empty"))
	}
	if c.httpClient == nil {
		return newError(ErrorKindConfig, fmt.Errorf("http client is nil"))
	}
	return nil
}

func (c *Client) newHTTPRequest(ctx context.Context, input chatservice.AIRespondInput) (*http.Request, error) {
	body, err := json.Marshal(newRespondRequest(input))
	if err != nil {
		return nil, newError(ErrorKindRequestBuild, fmt.Errorf("marshal request: %w", err))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+respondPath, bytes.NewReader(body))
	if err != nil {
		return nil, newError(ErrorKindRequestBuild, fmt.Errorf("create request: %w", err))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Request-ID", input.RequestID)
	req.Header.Set("Internal-Token", c.internalToken)
	return req, nil
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, newError(ErrorKindCall, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, newError(ErrorKindCall, fmt.Errorf("read response: %w", err))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &Error{Kind: ErrorKindHTTPStatus, StatusCode: resp.StatusCode}
	}
	return body, nil
}

func decodeRespondEnvelope(body []byte) (chatservice.AIRespondOutput, error) {
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return chatservice.AIRespondOutput{}, newError(ErrorKindEnvelopeDecode, err)
	}
	if env.Code != 0 {
		return chatservice.AIRespondOutput{}, &Error{Kind: ErrorKindEnvelopeCode, Code: env.Code}
	}
	if env.Data == nil {
		return chatservice.AIRespondOutput{}, &Error{Kind: ErrorKindEmptyData}
	}
	return toServiceOutput(env.Data), nil
}

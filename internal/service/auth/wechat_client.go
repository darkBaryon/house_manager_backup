package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WechatConfig struct {
	AppID   string
	Secret  string
	APIBase string
}

type WechatClient struct {
	appID      string
	secret     string
	apiBase    string
	httpClient *http.Client
}

func NewWechatClient(cfg WechatConfig) (*WechatClient, error) {
	apiBase := strings.TrimRight(strings.TrimSpace(cfg.APIBase), "/")
	if apiBase == "" {
		apiBase = "https://api.weixin.qq.com"
	}

	return &WechatClient{
		appID:      cfg.AppID,
		secret:     cfg.Secret,
		apiBase:    apiBase,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *WechatClient) ExchangeLoginCode(ctx context.Context, code string) (openID, unionID string, err error) {
	if strings.TrimSpace(code) == "" {
		return "", "", fmt.Errorf("wechat code is required")
	}
	if err := c.validateConfig(); err != nil {
		return "", "", err
	}

	u, err := url.Parse(c.apiBase + "/sns/jscode2session")
	if err != nil {
		return "", "", fmt.Errorf("build wechat login url: %w", err)
	}

	query := u.Query()
	query.Set("appid", c.appID)
	query.Set("secret", c.secret)
	query.Set("js_code", strings.TrimSpace(code))
	query.Set("grant_type", "authorization_code")
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", "", fmt.Errorf("build wechat login request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("call wechat login api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("wechat login api status: %s", resp.Status)
	}

	var data struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		OpenID  string `json:"openid"`
		UnionID string `json:"unionid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", fmt.Errorf("decode wechat login response: %w", err)
	}
	if data.ErrCode != 0 {
		return "", "", fmt.Errorf("wechat login api error: %s", data.ErrMsg)
	}
	if strings.TrimSpace(data.OpenID) == "" {
		return "", "", fmt.Errorf("wechat login response missing openid")
	}
	return data.OpenID, data.UnionID, nil
}

func (c *WechatClient) ExchangePhoneCode(ctx context.Context, phoneCode string) (string, error) {
	if strings.TrimSpace(phoneCode) == "" {
		return "", fmt.Errorf("wechat phone code is required")
	}
	if err := c.validateConfig(); err != nil {
		return "", err
	}

	accessToken, err := c.fetchAccessToken(ctx)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(c.apiBase + "/wxa/business/getuserphonenumber")
	if err != nil {
		return "", fmt.Errorf("build wechat phone url: %w", err)
	}

	query := u.Query()
	query.Set("access_token", accessToken)
	u.RawQuery = query.Encode()

	body, err := json.Marshal(map[string]string{"code": strings.TrimSpace(phoneCode)})
	if err != nil {
		return "", fmt.Errorf("marshal wechat phone request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build wechat phone request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call wechat phone api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("wechat phone api status: %s", resp.Status)
	}

	var data struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		PhoneInfo struct {
			PurePhoneNumber string `json:"purePhoneNumber"`
			PhoneNumber     string `json:"phoneNumber"`
		} `json:"phone_info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode wechat phone response: %w", err)
	}
	if data.ErrCode != 0 {
		return "", fmt.Errorf("wechat phone api error: %s", data.ErrMsg)
	}

	phone := strings.TrimSpace(data.PhoneInfo.PurePhoneNumber)
	if phone == "" {
		phone = strings.TrimSpace(data.PhoneInfo.PhoneNumber)
	}
	if phone == "" {
		return "", fmt.Errorf("wechat phone response missing phone number")
	}
	return normalizePhone(phone), nil
}

func (c *WechatClient) fetchAccessToken(ctx context.Context) (string, error) {
	u, err := url.Parse(c.apiBase + "/cgi-bin/token")
	if err != nil {
		return "", fmt.Errorf("build wechat access token url: %w", err)
	}

	query := u.Query()
	query.Set("grant_type", "client_credential")
	query.Set("appid", c.appID)
	query.Set("secret", c.secret)
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("build wechat access token request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call wechat access token api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("wechat access token api status: %s", resp.Status)
	}

	var data struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode wechat access token response: %w", err)
	}
	if data.ErrCode != 0 {
		return "", fmt.Errorf("wechat access token api error: %s", data.ErrMsg)
	}
	if strings.TrimSpace(data.AccessToken) == "" {
		return "", fmt.Errorf("wechat access token response missing access_token")
	}
	return data.AccessToken, nil
}

func (c *WechatClient) validateConfig() error {
	if strings.TrimSpace(c.appID) == "" {
		return fmt.Errorf("wechat appid is required")
	}
	if strings.TrimSpace(c.secret) == "" {
		return fmt.Errorf("wechat secret is required")
	}
	return nil
}

func normalizePhone(phone string) string {
	var builder strings.Builder
	for _, ch := range strings.TrimSpace(phone) {
		if ch >= '0' && ch <= '9' {
			builder.WriteRune(ch)
		}
	}
	return builder.String()
}

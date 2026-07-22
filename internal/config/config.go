package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig `mapstructure:"server"`
	MongoDB MongoConfig  `mapstructure:"mongodb"`
	Redis   RedisConfig  `mapstructure:"redis"`
	Wechat  WechatConfig `mapstructure:"wechat"`
	Auth    AuthConfig   `mapstructure:"auth"`
	AIChat  AIChatConfig `mapstructure:"ai_chat"`
	Log     LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type MongoConfig struct {
	Addrs                  []string `mapstructure:"addrs"`
	Database               string   `mapstructure:"database"`
	AuthSource             string   `mapstructure:"auth_source"`
	Username               string   `mapstructure:"username"`
	Password               string   `mapstructure:"password"`
	PoolSize               uint64   `mapstructure:"pool_size"`
	MinPoolSize            uint64   `mapstructure:"min_pool_size"`
	ConnectTimeout         string   `mapstructure:"connect_timeout"`
	SocketTimeout          string   `mapstructure:"socket_timeout"`
	ServerSelectionTimeout string   `mapstructure:"server_selection_timeout"`
	RetryReads             bool     `mapstructure:"retry_reads"`
	RetryWrites            bool     `mapstructure:"retry_writes"`
	ReplicaSet             string   `mapstructure:"replica_set"`
}

type RedisConfig struct {
	Addrs        []string `mapstructure:"addrs"`
	Password     string   `mapstructure:"password"`
	DB           int      `mapstructure:"db"`
	PoolSize     int      `mapstructure:"pool_size"`
	MinIdleConns int      `mapstructure:"min_idle_conns"`
	ConnTimeout  string   `mapstructure:"conn_timeout"`
	ReadTimeout  string   `mapstructure:"read_timeout"`
	WriteTimeout string   `mapstructure:"write_timeout"`
	MaxRetries   int      `mapstructure:"max_retries"`
	ClusterMode  bool     `mapstructure:"cluster_mode"`
}

type WechatConfig struct {
	AppID   string `mapstructure:"appid"`
	Secret  string `mapstructure:"secret"`
	APIBase string `mapstructure:"api_base"`
}

type AuthConfig struct {
	SessionTTL string `mapstructure:"session_ttl"`
}

type AIChatConfig struct {
	PythonBaseURL      string `mapstructure:"python_base_url"`
	InternalToken      string `mapstructure:"internal_token"`
	RespondTimeout     string `mapstructure:"respond_timeout"`
	ToolTimeout        string `mapstructure:"tool_timeout"`
	SessionIdleTimeout string `mapstructure:"session_idle_timeout"`
	RecentMessageLimit int    `mapstructure:"recent_message_limit"`
	RuntimeContextTTL  string `mapstructure:"runtime_context_ttl"`
}

const (
	defaultAuthSessionTTL           = 7 * 24 * time.Hour
	defaultAIChatRespondTimeout     = 120 * time.Second
	defaultAIChatToolTimeout        = 5 * time.Second
	defaultAIChatSessionIdleTimeout = 30 * time.Minute
	defaultAIChatRuntimeContextTTL  = 24 * time.Hour
)

type LogConfig struct {
	Level     string         `mapstructure:"level"`
	Format    string         `mapstructure:"format"`
	AddSource bool           `mapstructure:"add_source"`
	Service   string         `mapstructure:"service"`
	Env       string         `mapstructure:"env"`
	Fields    map[string]any `mapstructure:"fields"`
}

func Load(path string) (*Config, error) {
	loadDotEnv()

	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	bindSensitiveEnv(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return cfg, nil
}

func (c MongoConfig) ConnectTimeoutDuration() (time.Duration, error) {
	return parseDuration("mongodb.connect_timeout", c.ConnectTimeout)
}

func (c MongoConfig) SocketTimeoutDuration() (time.Duration, error) {
	return parseDuration("mongodb.socket_timeout", c.SocketTimeout)
}

func (c MongoConfig) ServerSelectionTimeoutDuration() (time.Duration, error) {
	return parseDuration("mongodb.server_selection_timeout", c.ServerSelectionTimeout)
}

func (c RedisConfig) ConnTimeoutDuration() (time.Duration, error) {
	return parseDuration("redis.conn_timeout", c.ConnTimeout)
}

func (c RedisConfig) ReadTimeoutDuration() (time.Duration, error) {
	return parseDuration("redis.read_timeout", c.ReadTimeout)
}

func (c RedisConfig) WriteTimeoutDuration() (time.Duration, error) {
	return parseDuration("redis.write_timeout", c.WriteTimeout)
}

func (c AuthConfig) SessionTTLDuration() (time.Duration, error) {
	if strings.TrimSpace(c.SessionTTL) == "" {
		return defaultAuthSessionTTL, nil
	}
	return parseDuration("auth.session_ttl", c.SessionTTL)
}

func (c AIChatConfig) RespondTimeoutDuration() (time.Duration, error) {
	if strings.TrimSpace(c.RespondTimeout) == "" {
		return defaultAIChatRespondTimeout, nil
	}
	return parseDuration("ai_chat.respond_timeout", c.RespondTimeout)
}

func (c AIChatConfig) ToolTimeoutDuration() (time.Duration, error) {
	if strings.TrimSpace(c.ToolTimeout) == "" {
		return defaultAIChatToolTimeout, nil
	}
	return parseDuration("ai_chat.tool_timeout", c.ToolTimeout)
}

func (c AIChatConfig) SessionIdleTimeoutDuration() (time.Duration, error) {
	if strings.TrimSpace(c.SessionIdleTimeout) == "" {
		return defaultAIChatSessionIdleTimeout, nil
	}
	return parseDuration("ai_chat.session_idle_timeout", c.SessionIdleTimeout)
}

func (c AIChatConfig) RuntimeContextTTLDuration() (time.Duration, error) {
	if strings.TrimSpace(c.RuntimeContextTTL) == "" {
		return defaultAIChatRuntimeContextTTL, nil
	}
	return parseDuration("ai_chat.runtime_context_ttl", c.RuntimeContextTTL)
}

func loadDotEnv() {
	_ = godotenv.Load()
}

func bindSensitiveEnv(v *viper.Viper) {
	_ = v.BindEnv("mongodb.username", "MONGODB_USERNAME")
	_ = v.BindEnv("mongodb.password", "MONGODB_PASSWORD")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("wechat.appid", "WECHAT_APPID")
	_ = v.BindEnv("wechat.secret", "WECHAT_SECRET")
	_ = v.BindEnv("wechat.api_base", "WECHAT_API_BASE")
	_ = v.BindEnv("auth.session_ttl", "AUTH_SESSION_TTL")
	_ = v.BindEnv("ai_chat.python_base_url", "AI_CHAT_PYTHON_BASE_URL")
	_ = v.BindEnv("ai_chat.internal_token", "AI_CHAT_INTERNAL_TOKEN")
	_ = v.BindEnv("ai_chat.respond_timeout", "AI_CHAT_RESPOND_TIMEOUT")
	_ = v.BindEnv("ai_chat.tool_timeout", "AI_CHAT_TOOL_TIMEOUT")
	_ = v.BindEnv("ai_chat.session_idle_timeout", "AI_CHAT_SESSION_IDLE_TIMEOUT")
	_ = v.BindEnv("ai_chat.recent_message_limit", "AI_CHAT_RECENT_MESSAGE_LIMIT")
	_ = v.BindEnv("ai_chat.runtime_context_ttl", "AI_CHAT_RUNTIME_CONTEXT_TTL")
}

func parseDuration(name, value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	return d, nil
}

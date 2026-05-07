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
	v.AutomaticEnv()
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

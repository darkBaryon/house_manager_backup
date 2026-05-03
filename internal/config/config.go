package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MongoDB MongoConfig `mapstructure:"mongodb"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Log    LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type MongoConfig struct {
	Addrs                 []string `mapstructure:"addrs"`
	Database              string   `mapstructure:"database"`
	AuthSource            string   `mapstructure:"auth_source"`
	Username              string   `mapstructure:"username"`
	Password              string   `mapstructure:"password"`
	PoolSize              uint64   `mapstructure:"pool_size"`
	MinPoolSize           uint64   `mapstructure:"min_pool_size"`
	ConnectTimeout        int      `mapstructure:"connect_timeout"`
	SocketTimeout         int      `mapstructure:"socket_timeout"`
	ServerSelectionTimeout int     `mapstructure:"server_selection_timeout"`
	MaxRetries            int      `mapstructure:"max_retries"`
	ReplicaSet            string   `mapstructure:"replica_set"`
}

type RedisConfig struct {
	Addrs        []string `mapstructure:"addrs"`
	Password     string   `mapstructure:"password"`
	DB           int      `mapstructure:"db"`
	PoolSize     int      `mapstructure:"pool_size"`
	MinIdleConns int      `mapstructure:"min_idle_conns"`
	ConnTimeout  int      `mapstructure:"conn_timeout"`
	ReadTimeout  int      `mapstructure:"read_timeout"`
	WriteTimeout int      `mapstructure:"write_timeout"`
	MaxRetries   int      `mapstructure:"max_retries"`
	ClusterMode  bool     `mapstructure:"cluster_mode"`
}

type LogConfig struct {
	Level     string         `mapstructure:"level"`
	Format    string         `mapstructure:"format"`
	AddSource bool           `mapstructure:"add_source"`
	Service   string         `mapstructure:"service"`
	Env       string         `mapstructure:"env"`
	Fields    map[string]any `mapstructure:"fields"`
}

func (c *Config) Redacted() any {
	if c == nil {
		return nil
	}

	return map[string]any{
		"server": c.Server,
		"mongodb": map[string]any{
			"addrs":                    c.MongoDB.Addrs,
			"database":                 c.MongoDB.Database,
			"auth_source":              c.MongoDB.AuthSource,
			"username":                 c.MongoDB.Username,
			"password":                 redactSecret(c.MongoDB.Password),
			"pool_size":                c.MongoDB.PoolSize,
			"min_pool_size":            c.MongoDB.MinPoolSize,
			"connect_timeout":          c.MongoDB.ConnectTimeout,
			"socket_timeout":           c.MongoDB.SocketTimeout,
			"server_selection_timeout": c.MongoDB.ServerSelectionTimeout,
			"max_retries":              c.MongoDB.MaxRetries,
			"replica_set":              c.MongoDB.ReplicaSet,
		},
		"redis": map[string]any{
			"addrs":          c.Redis.Addrs,
			"password":       redactSecret(c.Redis.Password),
			"db":             c.Redis.DB,
			"pool_size":      c.Redis.PoolSize,
			"min_idle_conns": c.Redis.MinIdleConns,
			"conn_timeout":   c.Redis.ConnTimeout,
			"read_timeout":   c.Redis.ReadTimeout,
			"write_timeout":  c.Redis.WriteTimeout,
			"max_retries":    c.Redis.MaxRetries,
			"cluster_mode":   c.Redis.ClusterMode,
		},
		"log": c.Log,
	}
}

func redactSecret(value string) string {
	if value == "" {
		return ""
	}
	return "***"
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("HM") // House Manager
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

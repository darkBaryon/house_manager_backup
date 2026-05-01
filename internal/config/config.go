package config

import (
	"fmt"

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
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

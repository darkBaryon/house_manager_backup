package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadEnvOverrideSensitiveFields(t *testing.T) {
	withWorkingDir(t, t.TempDir())
	configPath := writeConfigFile(t, `
server:
  port: 8080
  mode: debug
mongodb:
  addrs:
    - "127.0.0.1:27018"
  database: "rent-house"
  auth_source: "admin"
  username: "yaml-user"
  password: "yaml-pass"
  pool_size: 100
  min_pool_size: 10
  connect_timeout: "10s"
  socket_timeout: "30s"
  server_selection_timeout: "10s"
  retry_reads: true
  retry_writes: true
  replica_set: ""
redis:
  addrs:
    - "127.0.0.1:6380"
  password: "yaml-redis-pass"
  db: 0
  pool_size: 100
  min_idle_conns: 10
  conn_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"
  max_retries: 3
  cluster_mode: false
log:
  level: "info"
  format: "json"
  add_source: false
  service: "house-manager"
  env: "test"
`)

	t.Setenv("MONGODB_USERNAME", "env-user")
	t.Setenv("MONGODB_PASSWORD", "env-pass")
	t.Setenv("REDIS_PASSWORD", "env-redis-pass")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.MongoDB.Username != "env-user" {
		t.Fatalf("expected mongodb username from env, got %q", cfg.MongoDB.Username)
	}
	if cfg.MongoDB.Password != "env-pass" {
		t.Fatalf("expected mongodb password from env, got %q", cfg.MongoDB.Password)
	}
	if cfg.Redis.Password != "env-redis-pass" {
		t.Fatalf("expected redis password from env, got %q", cfg.Redis.Password)
	}
}

func TestLoadPreservesYAMLWhenEnvEmpty(t *testing.T) {
	withWorkingDir(t, t.TempDir())
	configPath := writeConfigFile(t, `
server:
  port: 8080
  mode: debug
mongodb:
  addrs:
    - "127.0.0.1:27018"
  database: "rent-house"
  auth_source: "admin"
  username: "yaml-user"
  password: "yaml-pass"
  pool_size: 100
  min_pool_size: 10
  connect_timeout: "10s"
  socket_timeout: "30s"
  server_selection_timeout: "10s"
  retry_reads: true
  retry_writes: true
  replica_set: ""
redis:
  addrs:
    - "127.0.0.1:6380"
  password: "yaml-redis-pass"
  db: 0
  pool_size: 100
  min_idle_conns: 10
  conn_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"
  max_retries: 3
  cluster_mode: false
log:
  level: "info"
  format: "json"
  add_source: false
  service: "house-manager"
  env: "test"
`)

	unsetEnv(t, "MONGODB_USERNAME")
	unsetEnv(t, "MONGODB_PASSWORD")
	unsetEnv(t, "REDIS_PASSWORD")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.MongoDB.Username != "yaml-user" {
		t.Fatalf("expected yaml mongodb username, got %q", cfg.MongoDB.Username)
	}
	if cfg.MongoDB.Password != "yaml-pass" {
		t.Fatalf("expected yaml mongodb password, got %q", cfg.MongoDB.Password)
	}
	if cfg.Redis.Password != "yaml-redis-pass" {
		t.Fatalf("expected yaml redis password, got %q", cfg.Redis.Password)
	}
}

func TestLoadIgnoresUnboundEnvironmentVariables(t *testing.T) {
	withWorkingDir(t, t.TempDir())
	configPath := writeConfigFile(t, `
server:
  port: 8080
  mode: debug
mongodb:
  addrs:
    - "127.0.0.1:27018"
  database: "rent-house"
  auth_source: "admin"
  username: "yaml-user"
  password: "yaml-pass"
  pool_size: 100
  min_pool_size: 10
  connect_timeout: "10s"
  socket_timeout: "30s"
  server_selection_timeout: "10s"
  retry_reads: true
  retry_writes: true
  replica_set: ""
redis:
  addrs:
    - "127.0.0.1:6380"
  password: "yaml-redis-pass"
  db: 0
  pool_size: 100
  min_idle_conns: 10
  conn_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"
  max_retries: 3
  cluster_mode: false
log:
  level: "info"
  format: "text"
  add_source: false
  service: ""
  env: ""
`)

	t.Setenv("LOG_FORMAT", "json")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Log.Format != "text" {
		t.Fatalf("expected yaml log format, got %q", cfg.Log.Format)
	}
}

func TestLoadDotEnvForLocalDevelopment(t *testing.T) {
	dir := t.TempDir()
	withWorkingDir(t, dir)
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(`
server:
  port: 8080
  mode: debug
mongodb:
  addrs:
    - "127.0.0.1:27018"
  database: "rent-house"
  auth_source: "admin"
  username: ""
  password: ""
  pool_size: 100
  min_pool_size: 10
  connect_timeout: "10s"
  socket_timeout: "30s"
  server_selection_timeout: "10s"
  retry_reads: true
  retry_writes: true
  replica_set: ""
redis:
  addrs:
    - "127.0.0.1:6380"
  password: ""
  db: 0
  pool_size: 100
  min_idle_conns: 10
  conn_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"
  max_retries: 3
  cluster_mode: false
log:
  level: "info"
  format: "json"
  add_source: false
  service: "house-manager"
  env: "test"
`), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	unsetEnv(t, "MONGODB_USERNAME")
	unsetEnv(t, "MONGODB_PASSWORD")
	unsetEnv(t, "REDIS_PASSWORD")
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("MONGODB_USERNAME=dotenv-user\nMONGODB_PASSWORD=dotenv-pass\nREDIS_PASSWORD=dotenv-redis\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.MongoDB.Username != "dotenv-user" || cfg.MongoDB.Password != "dotenv-pass" || cfg.Redis.Password != "dotenv-redis" {
		t.Fatalf("expected .env values, got mongodb=(%q,%q) redis=%q", cfg.MongoDB.Username, cfg.MongoDB.Password, cfg.Redis.Password)
	}
}

func TestDurationParsing(t *testing.T) {
	cfg := Config{
		MongoDB: MongoConfig{
			ConnectTimeout:         "10s",
			SocketTimeout:          "30s",
			ServerSelectionTimeout: "15s",
		},
		Redis: RedisConfig{
			ConnTimeout:  "5s",
			ReadTimeout:  "3s",
			WriteTimeout: "4s",
		},
	}

	got, err := cfg.MongoDB.ConnectTimeoutDuration()
	assertDuration(t, got, err, 10*time.Second)
	got, err = cfg.MongoDB.SocketTimeoutDuration()
	assertDuration(t, got, err, 30*time.Second)
	got, err = cfg.MongoDB.ServerSelectionTimeoutDuration()
	assertDuration(t, got, err, 15*time.Second)
	got, err = cfg.Redis.ConnTimeoutDuration()
	assertDuration(t, got, err, 5*time.Second)
	got, err = cfg.Redis.ReadTimeoutDuration()
	assertDuration(t, got, err, 3*time.Second)
	got, err = cfg.Redis.WriteTimeoutDuration()
	assertDuration(t, got, err, 4*time.Second)
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	return path
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
}

func assertDuration(t *testing.T, got time.Duration, err error, want time.Duration) {
	t.Helper()
	if err != nil {
		t.Fatalf("duration parse error: %v", err)
	}
	if got != want {
		t.Fatalf("expected duration %v, got %v", want, got)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	oldValue, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset env %s: %v", key, err)
	}
	t.Cleanup(func() {
		if !existed {
			_ = os.Unsetenv(key)
			return
		}
		_ = os.Setenv(key, oldValue)
	})
}

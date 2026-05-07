package redis

import (
	"fmt"
	"strings"
	"time"
)

type Config struct {
	Addrs        []string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	ConnTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	MaxRetries   int
	ClusterMode  bool
}

func (c Config) Validate() error {
	addrs := c.normalizedAddrs()
	if len(addrs) == 0 {
		if c.ClusterMode {
			return fmt.Errorf("redis cluster addrs are required")
		}
		return fmt.Errorf("redis addr is required")
	}
	return nil
}

func (c Config) normalizedAddrs() []string {
	addrs := make([]string, 0, len(c.Addrs))
	for _, addr := range c.Addrs {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		addrs = append(addrs, addr)
	}
	return addrs
}

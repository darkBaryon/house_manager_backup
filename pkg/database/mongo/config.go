package mongo

import (
	"fmt"
	"strings"
	"time"
)

type Config struct {
	URI                    string
	Addrs                  []string
	Database               string
	AuthSource             string
	Username               string
	Password               string
	PoolSize               uint64
	MinPoolSize            uint64
	ConnectTimeout         time.Duration
	SocketTimeout          time.Duration
	ServerSelectionTimeout time.Duration
	EnableRetryReads       bool
	EnableRetryWrites      bool
	ReplicaSet             string
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Database) == "" {
		return fmt.Errorf("mongo database is required")
	}
	if strings.TrimSpace(c.URI) != "" {
		return nil
	}

	addrs := c.normalizedAddrs()
	if len(addrs) == 0 {
		return fmt.Errorf("mongo addrs are required when uri is empty")
	}
	return nil
}

func (c Config) MongoURI() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	if uri := strings.TrimSpace(c.URI); uri != "" {
		return uri, nil
	}
	return "mongodb://" + strings.Join(c.normalizedAddrs(), ","), nil
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

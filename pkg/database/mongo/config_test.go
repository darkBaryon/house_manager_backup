package mongo

import (
	"strings"
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	t.Run("empty addrs", func(t *testing.T) {
		cfg := Config{Database: "house_manager"}
		err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), "mongo addrs are required") {
			t.Fatalf("expected addrs validation error, got %v", err)
		}
	})

	t.Run("empty database", func(t *testing.T) {
		cfg := Config{Addrs: []string{"127.0.0.1:27017"}}
		err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), "mongo database is required") {
			t.Fatalf("expected database validation error, got %v", err)
		}
	})

	t.Run("uri allows empty addrs", func(t *testing.T) {
		cfg := Config{
			URI:      "mongodb://127.0.0.1:27017",
			Database: "house_manager",
		}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("expected uri-based config to pass validation, got %v", err)
		}
	})
}

func TestConfigMongoURI(t *testing.T) {
	t.Run("uri preferred", func(t *testing.T) {
		cfg := Config{
			URI:      " mongodb://127.0.0.1:27017 ",
			Addrs:    []string{"127.0.0.1:27018"},
			Database: "house_manager",
		}
		uri, err := cfg.MongoURI()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if uri != "mongodb://127.0.0.1:27017" {
			t.Fatalf("expected preferred uri, got %q", uri)
		}
	})

	t.Run("build from addrs", func(t *testing.T) {
		cfg := Config{
			Addrs:    []string{" 127.0.0.1:27017 ", "", "127.0.0.1:27018"},
			Database: "house_manager",
		}
		uri, err := cfg.MongoURI()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if uri != "mongodb://127.0.0.1:27017,127.0.0.1:27018" {
			t.Fatalf("unexpected uri %q", uri)
		}
	})
}

func TestConfigDurationFields(t *testing.T) {
	cfg := Config{
		Addrs:                  []string{"127.0.0.1:27017"},
		Database:               "house_manager",
		ConnectTimeout:         5 * time.Second,
		ServerSelectionTimeout: 7 * time.Second,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

package pythonchat

import "time"

type Config struct {
	BaseURL       string
	InternalToken string
	Timeout       time.Duration
}

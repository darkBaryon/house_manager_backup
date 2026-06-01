package chat

import "time"

type ServiceConfig struct {
	SessionIdleTimeout time.Duration
	RecentMessageLimit int
}

type SendInput struct {
	RequestID  string
	UserID     string
	Message    string
	NewSession bool
}

type SendResult struct {
	SessionID        string
	AssistantMessage string
	Sentences        []string
	HouseList        []HouseItem
}

type HouseItem struct {
	ListingID       string
	Title           string
	Price           int
	PriceText       string
	District        string
	BizArea         string
	LayoutText      string
	SubwayDistanceM int
	Images          []TaggedImage
}

type TaggedImage struct {
	URL string
	Tag string
}

type RecentMessage struct {
	Role    string
	Content string
}

type AIRespondInput struct {
	RequestID      string
	SessionID      string
	Message        string
	RecentMessages []RecentMessage
	RuntimeContext map[string]any
}

type AIRespondOutput struct {
	AssistantMessage string
	Sentences        []string
	HouseList        []HouseItem
	UpdatedContext   map[string]any
	SafetyResult     map[string]any
	IntentResult     map[string]any
}

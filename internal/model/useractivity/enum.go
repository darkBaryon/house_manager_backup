package useractivity

const (
	HistorySourceDiscover = "discover"
	HistorySourceAI       = "ai"
	HistorySourceDetail   = "detail"
)

var validHistorySources = map[string]struct{}{
	HistorySourceDiscover: {},
	HistorySourceAI:       {},
	HistorySourceDetail:   {},
}

func ValidHistorySource(source string) bool {
	_, ok := validHistorySources[source]
	return ok
}

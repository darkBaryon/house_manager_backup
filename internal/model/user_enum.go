package model

type SourceChannel string

const (
	SourceChannelMini SourceChannel = "miniapp"
)

var validSourceChannels = map[SourceChannel]struct{}{
	SourceChannelMini: {},
}

func (v SourceChannel) ValidOptional() bool {
	if v == "" {
		return true
	}
	_, ok := validSourceChannels[v]
	return ok
}

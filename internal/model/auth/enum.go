package auth

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

type AuthProvider string

const (
	AuthProviderWechat AuthProvider = "wechat"
)

var validAuthProviders = map[AuthProvider]struct{}{
	AuthProviderWechat: {},
}

func (v AuthProvider) Valid() bool {
	_, ok := validAuthProviders[v]
	return ok
}

package model

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

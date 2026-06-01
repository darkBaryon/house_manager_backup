package pythonchat

type envelope struct {
	Code  int          `json:"code"`
	Error string       `json:"error"`
	Data  *respondData `json:"data"`
}

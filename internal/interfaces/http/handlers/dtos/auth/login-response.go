package authhttpdtos

type LoginResponse struct {
	Token string `json:"session_token"`
}

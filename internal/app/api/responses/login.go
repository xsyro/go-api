package responses

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	Exp          int    `json:"exp"`
}

func NewLoginResponse(token, refreshToken string, exp int) *LoginResponse {
	return &LoginResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		Exp:          exp,
	}
}

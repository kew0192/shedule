package users

type User struct {
	ID           int    `json:"id"`
	First_name   string `json:"first_name"`
	Last_name    string `json:"last_name"`
	Role         string `json:"role"`
	Code         string `json:"code"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

package services

type User struct {
	PublicID string `json:"public_id"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Email    string `json:"email"`
}

type Token struct {
	//PublicID    string `json:"public_id"`
	AccessToken string `json:"access_token"`
}

type Credentials struct {
	Email string `json:"email"`
}

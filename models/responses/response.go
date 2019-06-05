package responses

import (
	"fmt"
	"time"
)

type ResponseError struct {
	Status  string `json:"status"`
	Success bool   `json:"success"`
	Error   string `json:"errorMsg"`
}

type Login struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type Refresh struct {
	AccessToken string `json:"access_token"`
}

type UserInfo struct {
	ID                 uint       `json:"id"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at"`
	Login              string     `json:"login"`
	Name               string     `json:"name"`
	Email              string     `json:"email"`
	Enabled2fa         bool       `json:"enabled2fa"`
	SMSPhoneNumber     string     `json:"phone_number"`
	SMSSeedPhoneNumber string     `json:"seed_phone_number"`
}

func ErrorResponse(text string) ResponseError {
	return ResponseError {
		Status:  "error",
		Success: false,
		Error:   fmt.Sprintf(text),
	}
}
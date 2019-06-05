package responses

import "fmt"

type ResponseError struct {
	Status  string `json:"status"`
	Success bool   `json:"success"`
	Error   string `json:"errorMsg"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func ErrorResponse(text string) ResponseError {
	return ResponseError {
		Status:  "error",
		Success: false,
		Error:   fmt.Sprintf(text),
	}
}
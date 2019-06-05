package responses

import "fmt"

type ResponseError struct {
	Status  string `json:"status"`
	Success bool   `json:"success"`
	Error   string `json:"errorMsg"`
}

func ErrorResponse(text string) ResponseError {
	return ResponseError {
		Status:  "error",
		Success: false,
		Error:   fmt.Sprintf(text),
	}
}
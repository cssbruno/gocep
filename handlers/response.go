package handlers

import (
	"encoding/json"
	"net/http"
)

type apiErrorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var marshalAPIError = json.Marshal

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := apiErrorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	}

	payload, err := marshalAPIError(resp)
	if err != nil {
		return
	}
	_, _ = w.Write(payload)
}

package api

import (
	"encoding/json"
	"net/http"
)

func SendSuccessResponse(w http.ResponseWriter, data interface{}) {
	response := SuccessResponse{
		Status: "seccess",
		Data:   data,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func SendErrorResponse(w http.ResponseWriter, message string, code int) {
	response := ErrorResponse{
		Status:  "error",
		Message: message,
		Code:    code,
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

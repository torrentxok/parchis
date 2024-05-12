package auth

import (
	"encoding/json"
	"log"
	"net/http"
)

// Хендлер регистрации
func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	var regData RegistrationData
	log.Print("[INFO] Start executing signup.")
	err := json.NewDecoder(r.Body).Decode(&regData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := ErrorResponse{
			Error: struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				Code:    http.StatusBadRequest,
				Message: "Ошибка при декодировании JSON: " + err.Error(),
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	err = ValidateRegData(regData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := ErrorResponse{
			Error: struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	newUser, err := InsertUser(regData)
	if err != nil {
		log.Print("[ERROR] Ошибка создания пользователя: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := ErrorResponse{
			Error: struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				Code:    http.StatusBadRequest,
				Message: "Ошибка создания пользовтаеля",
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	err = SendConfirmationEmail(newUser)
	if err != nil {
		log.Print("[ERROR] Ошибка отправки письма: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := ErrorResponse{
			Error: struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				Code:    http.StatusBadRequest,
				Message: "Ошибка создания пользовтаеля",
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	errorResponse := ErrorResponse{
		Error: struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{
			Code:    http.StatusOK,
			Message: "",
		},
	}
	json.NewEncoder(w).Encode(errorResponse)
}

func ConfirmEmailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	log.Print("[INFO] Start confirm email")

	token := r.URL.Query().Get("token")
	if token == "" {
		log.Print("[ERROR] Token not detected")
		http.Error(w, "Token not detected", http.StatusBadRequest)
	}

	err := ConfirmEmail(token)
	if err != nil {
		log.Print("[ERROR] Ошибка подтверждения: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := ErrorResponse{
			Error: struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				Code:    http.StatusBadRequest,
				Message: "Ошибка подтвержения",
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	errorResponse := ErrorResponse{
		Error: struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{
			Code:    http.StatusOK,
			Message: "",
		},
	}
	json.NewEncoder(w).Encode(errorResponse)
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
}

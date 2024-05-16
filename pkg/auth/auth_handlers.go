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
		log.Print("Ошибка при декодировании JSON: " + err.Error())
		SendJSONResponse(w, "Ошибка при декодировании JSON", http.StatusBadRequest)
		return
	}

	err = ValidateRegData(regData)
	if err != nil {
		log.Print("[ERROR] Ошибка валидации пользователя: " + err.Error())
		SendJSONResponse(w, "Ошибка данных", http.StatusBadRequest)
		return
	}

	newUser, err := InsertUser(regData)
	if err != nil {
		log.Print("[ERROR] Ошибка создания пользователя: " + err.Error())
		SendJSONResponse(w, "Ошибка создания пользовтаеля", http.StatusBadRequest)
		return
	}

	err = SendConfirmationEmail(newUser)
	if err != nil {
		log.Print("[ERROR] Ошибка отправки письма: " + err.Error())
		SendJSONResponse(w, "Ошибка создания пользовтаеля", http.StatusBadRequest)
		return
	}
	SendJSONResponse(w, "", http.StatusOK)
}

func ConfirmEmailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	log.Print("[INFO] Start confirm email")

	token := r.URL.Query().Get("token")
	if token == "" {
		log.Print("[ERROR] Token not detected")
		http.Error(w, "Token not detected", http.StatusBadRequest)
		return
	}

	err := ConfirmEmail(token)
	if err != nil {
		log.Print("[ERROR] Ошибка подтверждения: " + err.Error())
		SendJSONResponse(w, "Ошибка подтверждения", http.StatusBadRequest)
		return
	}

	SendJSONResponse(w, "", http.StatusOK)
}

func SendJSONResponse(w http.ResponseWriter, msg string, code int) {
	w.WriteHeader(code)
	JSONResponse := JSONResponse{
		Error: struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{
			Code:    code,
			Message: msg,
		},
	}
	json.NewEncoder(w).Encode(JSONResponse)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	log.Print("[INFO] Start authentification")
	var login LoginData
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		log.Print("[ERROR] Ошибка запроса" + err.Error())
		SendJSONResponse(w, "Ошибка запроса", http.StatusBadRequest)
		return
	}

	err = ValidateLoginData(login)
	if err != nil {
		log.Print("[ERROR] Ошибка валидации" + err.Error())
		SendJSONResponse(w, "Неверные данные", http.StatusBadRequest)
		return
	}

	loginResponse, err := LoginUser(&login)
	if err != nil {
		log.Print("[ERROR] Ошибка авторизации" + err.Error())
		SendJSONResponse(w, "Ошибка авторизации", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loginResponse)

}

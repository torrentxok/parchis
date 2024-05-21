package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/torrentxok/parchis/pkg/api"
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
		api.SendErrorResponse(w, "Ошибка при декодировании JSON", http.StatusBadRequest)
		return
	}

	err = ValidateRegData(regData)
	if err != nil {
		log.Print("[ERROR] Ошибка валидации пользователя: " + err.Error())
		api.SendErrorResponse(w, "Ошибка данных", http.StatusBadRequest)
		return
	}

	newUser, err := InsertUser(regData)
	if err != nil {
		log.Print("[ERROR] Ошибка создания пользователя: " + err.Error())
		api.SendErrorResponse(w, "Ошибка создания пользовтаеля", http.StatusBadRequest)
		return
	}

	err = SendConfirmationEmail(newUser)
	if err != nil {
		log.Print("[ERROR] Ошибка отправки письма: " + err.Error())
		api.SendErrorResponse(w, "Ошибка создания пользовтаеля", http.StatusBadRequest)
		return
	}
	api.SendSuccessResponse(w, nil)
}

func ConfirmEmailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	log.Print("[INFO] Start confirm email")

	token := r.URL.Query().Get("token")
	if token == "" {
		log.Print("[ERROR] Token not detected")
		api.SendErrorResponse(w, "Token not detected", http.StatusBadRequest)
		return
	}

	err := ConfirmEmail(token)
	if err != nil {
		log.Print("[ERROR] Ошибка подтверждения: " + err.Error())
		api.SendErrorResponse(w, "Ошибка подтверждения", http.StatusBadRequest)
		return
	}

	api.SendSuccessResponse(w, nil)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	log.Print("[INFO] Start authentification")
	var login LoginData
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		log.Print("[ERROR] Ошибка запроса" + err.Error())
		api.SendErrorResponse(w, "Ошибка запроса", http.StatusBadRequest)
		return
	}

	err = ValidateLoginData(login)
	if err != nil {
		log.Print("[ERROR] Ошибка валидации" + err.Error())
		api.SendErrorResponse(w, "Неверные данные", http.StatusBadRequest)
		return
	}

	loginResponse, err := LoginUser(&login)
	if err != nil {
		log.Print("[ERROR] Ошибка авторизации" + err.Error())
		api.SendErrorResponse(w, "Ошибка авторизации", http.StatusBadRequest)
		return
	}

	api.SendSuccessResponse(w, loginResponse)

}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Print("[ERROR] Authorization header is required")
			api.SendErrorResponse(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" || !validateToken(token) {
			log.Print("[ERROR] Токен не прошел валидацию")
			api.SendErrorResponse(w, "Invalid or missing token", http.StatusUnauthorized)
			return
		}

		claims, err := ExtractClaims(token)
		if err != nil {
			log.Print("[ERROR] Error extracting claims: " + err.Error())
			api.SendErrorResponse(w, "Invalid or missing token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func AccessTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	log.Print("[INFO] Start authentification")
	var accessToken AccessTokenRequest
	err := json.NewDecoder(r.Body).Decode(&accessToken)
	if err != nil {
		log.Print("[ERROR] Ошибка запроса" + err.Error())
		api.SendErrorResponse(w, "Ошибка запроса", http.StatusBadRequest)
		return
	}
	if accessToken.Token == "" || !validateToken(accessToken.Token) {
		log.Print("[ERROR] Токен не прошел валидацию")
		api.SendErrorResponse(w, "Invalid or missing token", http.StatusUnauthorized)
		return
	}
}

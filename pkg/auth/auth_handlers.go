package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
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
	var refreshToken AccessTokenRequest
	err := json.NewDecoder(r.Body).Decode(&refreshToken)
	if err != nil {
		log.Print("[ERROR] Ошибка запроса" + err.Error())
		api.SendErrorResponse(w, "Ошибка запроса", http.StatusBadRequest)
		return
	}
	claims, ok := r.Context().Value(ClaimsKey).(jwt.MapClaims)
	if !ok {
		log.Print("[ERROR] No claims found")
		api.SendErrorResponse(w, "No claims found", http.StatusInternalServerError)
		return
	}
	userIdFloat64, ok := claims["UserId"].(float64)
	if !ok {
		log.Print("[ERROR] Invalid user ID in claims")
		api.SendErrorResponse(w, "Invalid user ID in claims", http.StatusInternalServerError)
		return
	}
	userId := int(userIdFloat64)
	accessTokenResponse, err := AccessTokenUpdate(userId, refreshToken.Token)
	if err != nil {
		log.Print("[ERROR] Error update access token")
		api.SendErrorResponse(w, "Error update access token", http.StatusInternalServerError)
		return
	}
	api.SendSuccessResponse(w, accessTokenResponse)
}

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	var refreshToken RefreshTokenRequest
	err := json.NewDecoder(r.Body).Decode(&refreshToken)
	if err != nil {
		log.Print("[ERROR] Ошибка запроса" + err.Error())
		api.SendErrorResponse(w, "Ошибка запроса", http.StatusBadRequest)
		return
	}
	claims, ok := r.Context().Value(ClaimsKey).(jwt.MapClaims)
	if !ok {
		log.Print("[ERROR] No claims found")
		api.SendErrorResponse(w, "No claims found", http.StatusInternalServerError)
		return
	}
	userIdFloat64, ok := claims["UserId"].(float64)
	if !ok {
		log.Print("[ERROR] Invalid user ID in claims")
		api.SendErrorResponse(w, "Invalid user ID in claims", http.StatusInternalServerError)
		return
	}
	userId := int(userIdFloat64)
	refreshTokenResponse, err := RefreshTokenUpdate(userId, refreshToken.Token)
	if err != nil {
		log.Print("[ERROR] Error update refresh token")
		api.SendErrorResponse(w, "Error update refresh token", http.StatusInternalServerError)
		return
	}
	api.SendSuccessResponse(w, refreshTokenResponse)
}

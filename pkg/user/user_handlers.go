package user

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/torrentxok/parchis/pkg/api"
	"github.com/torrentxok/parchis/pkg/auth"
)

func GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	profileId, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Print("[ERROR] Invalid profile ID: " + err.Error())
		api.SendErrorResponse(w, "Invalid profile ID", http.StatusBadRequest)
		return
	}
	claims, ok := r.Context().Value(auth.ClaimsKey).(jwt.MapClaims)
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

	userProfile, err := GetUserProfile(profileId, userId)
	if err != nil {
		log.Print("[ERROR] Error get profile")
		api.SendErrorResponse(w, "Error get profile", http.StatusInternalServerError)
		return
	}

	api.SendSuccessResponse(w, userProfile)
}

// Заявка в друзья
func FriendshipRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	var friend Friend
	err := json.NewDecoder(r.Body).Decode(&friend)
	if err != nil {
		log.Print("Ошибка при декодировании JSON: " + err.Error())
		api.SendErrorResponse(w, "Ошибка при декодировании JSON", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(auth.ClaimsKey).(jwt.MapClaims)
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

	if userId == friend.TargetUserId {
		log.Print("[ERROR] User ID and Friend ID are equal")
		api.SendErrorResponse(w, "User ID and Friend ID are equal", http.StatusBadRequest)
		return
	}

	err = FriendshipRequest(friend.TargetUserId, userId)
	if err != nil {
		log.Print("[ERROR] Error send friendship request: " + err.Error())
		api.SendErrorResponse(w, "Error send friendship request", http.StatusInternalServerError)
		return
	}
	api.SendSuccessResponse(w, nil)
}

// Принять в друзья
func FriendshipAcceptHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	var friend Friend
	err := json.NewDecoder(r.Body).Decode(&friend)
	if err != nil {
		log.Print("Ошибка при декодировании JSON: " + err.Error())
		api.SendErrorResponse(w, "Ошибка при декодировании JSON", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(auth.ClaimsKey).(jwt.MapClaims)
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
	if userId == friend.TargetUserId {
		log.Print("[ERROR] User ID and Friend ID are equal")
		api.SendErrorResponse(w, "User ID and Friend ID are equal", http.StatusBadRequest)
		return
	}

	err = FriendshipAccept(friend.TargetUserId, userId)
	if err != nil {
		log.Print("[ERROR] Error friendship accept: " + err.Error())
		api.SendErrorResponse(w, "Error send friendship request", http.StatusInternalServerError)
		return
	}
	api.SendSuccessResponse(w, nil)
}

// Удалить из друзей / отклонить заявку
func FriendshipRemoveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	var friend Friend
	err := json.NewDecoder(r.Body).Decode(&friend)
	if err != nil {
		log.Print("Ошибка при декодировании JSON: " + err.Error())
		api.SendErrorResponse(w, "Ошибка при декодировании JSON", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(auth.ClaimsKey).(jwt.MapClaims)
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
	if userId == friend.TargetUserId {
		log.Print("[ERROR] User ID and Friend ID are equal")
		api.SendErrorResponse(w, "User ID and Friend ID are equal", http.StatusBadRequest)
		return
	}

	err = FriendshipRemove(friend.TargetUserId, userId)
	if err != nil {
		log.Print("[ERROR] Error send friendship request: " + err.Error())
		api.SendErrorResponse(w, "Error send friendship request", http.StatusInternalServerError)
		return
	}
	api.SendSuccessResponse(w, nil)
}

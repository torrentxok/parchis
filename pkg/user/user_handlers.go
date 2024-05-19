package user

import (
	"log"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/torrentxok/parchis/pkg/api"
	"github.com/torrentxok/parchis/pkg/auth"
)

func GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {
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

// Заявка в друзья или подтверждение заявки
func AddFriendHandler(w http.ResponseWriter, r *http.Request) {

}

package user

import (
	"context"
	"log"

	database "github.com/torrentxok/parchis/pkg/db"
)

func GetUserProfile(targetUserId int, currentUserId int) (UserProfileResponse, error) {
	var response UserProfileResponse
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return response, err
	}
	defer db.Close(context.Background())
	response, err = GetUserProfileInfo(db, targetUserId, currentUserId)
	if err != nil {
		log.Print("[ERROR] Error getting user profile" + err.Error())
		return response, err
	}
	friends, err := GetUserFriends(db, targetUserId)
	if err != nil {
		log.Print("[ERROR] Error getting user friends: " + err.Error())
		return response, err
	}
	response.Friends = friends
	return response, nil
}

func FriendshipRequest(targetUserId int, currentUserId int) error {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return err
	}
	defer db.Close(context.Background())
	err = FriendshipsManaging(db, targetUserId, currentUserId, "request")
	if err != nil {
		log.Print("[ERROR] Ошибка добавления пользователя: " + err.Error())
		return err
	}
	return nil
}

func FriendshipAccept(targetUserId int, currentUserId int) error {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return err
	}
	defer db.Close(context.Background())
	err = FriendshipsManaging(db, targetUserId, currentUserId, "accept")
	if err != nil {
		log.Print("[ERROR] Ошибка добавления пользователя: " + err.Error())
		return err
	}
	return nil
}

func FriendshipRemove(targetUserId int, currentUserId int) error {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return err
	}
	defer db.Close(context.Background())
	err = FriendshipsManaging(db, targetUserId, currentUserId, "remove")
	if err != nil {
		log.Print("[ERROR] Ошибка добавления пользователя: " + err.Error())
		return err
	}
	return nil
}

package user

import "github.com/jackc/pgx/v5/pgtype"

type UserProfileResponse struct {
	Profile struct {
		Id       int    `json:"id"`
		Username string `json:"username"`
	} `json:"profile"`
	IsOwnProfile     bool        `json:"is_own_profile"`
	FriendshipStatus pgtype.Text `json:"friendship_status"`
	Friends          []struct {
		Id       int    `json:"id"`
		Username string `json:"username"`
	} `json:"friends"`
}

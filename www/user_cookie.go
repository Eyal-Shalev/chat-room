package www

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type UserCookie struct {
	Username string `json:"username"`
}

const UserCookieKey = "chat-room-user"

func GetUserCookie(r *http.Request) (*UserCookie, error) {
	cookie, err := r.Cookie(UserCookieKey)
	if err != nil {
		return nil, fmt.Errorf("cookie not found: %w", err)
	}

	value, err := base64.RawStdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("decode cookie: %w", err)
	}

	var userCookie UserCookie
	err = json.Unmarshal(value, &userCookie)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal user cookie: %w", err)
	}

	return &userCookie, nil
}

func SetUserCookie(w http.ResponseWriter, username string) {
	value, err := json.Marshal(UserCookie{Username: username})
	if err != nil {
		panic(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:  UserCookieKey,
		Value: base64.RawStdEncoding.EncodeToString(value),
	})
}

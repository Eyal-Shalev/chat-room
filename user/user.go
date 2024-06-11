package user

import (
	"fmt"

	"chat-room/data"
	"github.com/google/uuid"
)

type User struct {
	Name     string
	UUID     uuid.UUID
	Incoming chan []data.UserMessage
}

func (u User) IsAnonymous() bool {
	return u.Name == ""
}

func (u User) String() string {
	if u.IsAnonymous() {
		return u.UUID.String()
	}
	return fmt.Sprintf("%s (%s)", u.Name, u.UUID)
}

func New(name string) *User {
	return &User{
		Name:     name,
		UUID:     uuid.New(),
		Incoming: make(chan []data.UserMessage),
	}
}

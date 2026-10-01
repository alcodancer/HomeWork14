package domain

import (
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidName  = errors.New("invalid user name")
	ErrInvalidEmail = errors.New("invalid user email")
	ErrUserExists   = errors.New("user already exists")
)

type User struct {
	ID    bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name  string        `bson:"name"          json:"name"`
	Email string        `bson:"email"         json:"email"`
}
package users

import (
	"tapesonic/model"
	"tapesonic/util"

	"github.com/google/uuid"
)

type User struct {
	Id uuid.UUID

	Name         string
	PasswordHash string
	Role         model.Role

	ApiKey string

	CreatedAt util.TimestampWrapper
}

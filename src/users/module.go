package users

import "gorm.io/gorm"

type UsersModule struct {
	UserService *UserService
}

func NewUsersModule(db *gorm.DB) *UsersModule {
	storage := newUserStorage(db)
	service := newUserService(storage)

	return &UsersModule{
		UserService: service,
	}
}

package users

import "gorm.io/gorm"

type UsersModule struct {
	UserService *UserService
}

func NewUsersModule(db *gorm.DB) (*UsersModule, error) {
	storage, err := newUserStorage(db)
	if err != nil {
		return nil, err
	}

	service := newUserService(storage)

	return &UsersModule{
		UserService: service,
	}, nil
}

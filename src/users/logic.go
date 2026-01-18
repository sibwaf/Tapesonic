package users

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"tapesonic/model"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	users *UserStorage
}

func newUserService(users *UserStorage) *UserService {
	return &UserService{users: users}
}

func (svc *UserService) GetListForApi() ([]User, error) {
	return svc.users.GetAll()
}

func (svc *UserService) CreateFirstAdmin(username string, password string) (User, error) {
	// todo: full table exclusive lock

	adminExists, err := svc.CheckAdminAccountExists()
	if err != nil {
		return User{}, err
	}

	if adminExists {
		return User{}, fmt.Errorf("admin account already exists")
	}

	return svc.users.Create(
		User{
			Id:           uuid.New(),
			Name:         username,
			PasswordHash: getPasswordHash(password),
			ApiKey:       generateApiKey(),
			Role:         model.ROLE_ADMIN,
			CreatedAt:    util.NewTimestampWrapper(time.Now()),
		},
	)
}

func (svc *UserService) CreateUser(name string, password string, role model.Role) (User, error) {
	user := User{
		Id:           uuid.New(),
		Name:         name,
		PasswordHash: getPasswordHash(password),
		ApiKey:       generateApiKey(),
		Role:         role,
		CreatedAt:    util.NewTimestampWrapper(time.Now()),
	}

	return svc.users.Create(user)
}

func (svc *UserService) UpdateUser(id uuid.UUID, name string, password string, role model.Role) (User, error) {
	user, err := svc.users.FindById(id)
	if err != nil {
		return User{}, err
	}
	if user == nil {
		return User{}, model.ErrNotFound
	}

	if name != "" {
		user.Name = name
	}
	if password != "" {
		user.PasswordHash = getPasswordHash(password)
	}
	if role != "" {
		user.Role = role
	}

	return svc.users.Update(*user)
}

func (svc *UserService) UpdateApiKey(id uuid.UUID) (User, error) {
	user, err := svc.users.FindById(id)
	if err != nil {
		return User{}, err
	}
	if user == nil {
		return User{}, model.ErrNotFound
	}

	user.ApiKey = generateApiKey()

	return svc.users.Update(*user)
}

func (svc *UserService) CheckAdminAccountExists() (bool, error) {
	users, err := svc.users.GetListByRole(model.ROLE_ADMIN)
	if err != nil {
		return false, err
	}

	return len(users) > 0, nil
}

func (svc *UserService) TryAuthenticateWithPassword(username string, password string) (*User, error) {
	passwordHash := getPasswordHash(password)
	return svc.users.FindByCredentials(username, passwordHash)
}

func (svc *UserService) TryAuthenticateWithApiKey(apiKey string) (*User, error) {
	return svc.users.FindByApiKey(apiKey)
}

func (svc *UserService) FindByName(username string) (*User, error) {
	return svc.users.FindByName(username)
}

func getPasswordHash(password string) string {
	hashBytes := sha512.Sum512([]byte(password))
	return hex.EncodeToString(hashBytes[:])
}

func generateApiKey() string {
	return util.GenerateRandomString(24)
}

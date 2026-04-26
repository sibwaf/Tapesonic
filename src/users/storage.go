package users

import (
	"errors"
	"fmt"
	"tapesonic/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStorage struct {
	db *gorm.DB
}

func newUserStorage(db *gorm.DB) *UserStorage {
	return &UserStorage{db: db}
}

func (storage *UserStorage) Create(user User) (User, error) {
	sql := `
		INSERT INTO users (id, name, password_hash, role, api_key, created_at)
		VALUES (@id, @name, @passwordHash, @role, @apiKey, @createdAt)
		RETURNING *
	`
	params := map[string]any{
		"id":           user.Id,
		"name":         user.Name,
		"passwordHash": user.PasswordHash,
		"role":         user.Role,
		"apiKey":       user.ApiKey,
		"createdAt":    user.CreatedAt,
	}

	return user, storage.db.Raw(sql, params).First(&user).Error
}

func (storage *UserStorage) Update(user User) (User, error) {
	sql := `
		UPDATE users
		SET name = @name, password_hash = @passwordHash, role = @role, api_key = @apiKey
		WHERE id = @id
		RETURNING *
	`
	params := map[string]any{
		"id":           user.Id,
		"name":         user.Name,
		"passwordHash": user.PasswordHash,
		"role":         user.Role,
		"apiKey":       user.ApiKey,
	}

	err := storage.db.Raw(sql, params).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user, model.ErrNotFound
	} else {
		return user, err
	}
}

func (storage *UserStorage) GetAll() ([]User, error) {
	sql := `
		SELECT id, name, password_hash, role, api_key
		FROM users
		ORDER BY created_at
	`
	result := []User{}
	return result, storage.db.Raw(sql).Find(&result).Error
}

func (storage *UserStorage) GetListByRole(role model.Role) ([]User, error) {
	return storage.findAllMatching(
		`"role" = @role`,
		map[string]any{"role": role},
	)
}

func (storage *UserStorage) FindById(id uuid.UUID) (*User, error) {
	return storage.findSingleMatching(`id = @id`, map[string]any{"id": id})
}

func (storage *UserStorage) FindByCredentials(username string, passwordHash string) (*User, error) {
	return storage.findSingleMatching(
		`"name" = @username AND "password_hash" = @passwordHash`,
		map[string]any{"username": username, "passwordHash": passwordHash},
	)
}

func (storage *UserStorage) FindByApiKey(apiKey string) (*User, error) {
	return storage.findSingleMatching(
		`"api_key" = @apiKey`,
		map[string]any{"apiKey": apiKey},
	)
}

func (storage *UserStorage) FindByName(username string) (*User, error) {
	return storage.findSingleMatching(
		`"name" = @username`,
		map[string]any{"username": username},
	)
}

func (storage *UserStorage) findSingleMatching(filter string, params map[string]any) (*User, error) {
	users, err := storage.findAllMatching(filter, params)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	} else if len(users) == 1 {
		return &users[0], nil
	} else {
		return nil, fmt.Errorf("found multiple matching users")
	}
}

func (storage *UserStorage) findAllMatching(filter string, params map[string]any) ([]User, error) {
	query := fmt.Sprintf(
		`
			SELECT "id", "name", "password_hash", "role", "api_key"
			FROM users
			WHERE %s
		`,
		filter,
	)

	result := []User{}
	return result, storage.db.Raw(query, params).Find(&result).Error
}

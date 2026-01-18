package admin

import (
	"encoding/json"
	"net/http"
	"tapesonic/model"
	"tapesonic/users"
	"tapesonic/util"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserRs struct {
	Id     string
	Name   string
	Role   model.Role
	ApiKey string
}

func toUserRs(user users.User) UserRs {
	return UserRs{
		Id:   user.Id.String(),
		Name: user.Name,
		Role: user.Role,
	}
}

func GetUsers(auth *authenticator, users *users.UserService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authorize(r, model.ROLE_ADMIN)
		if err != nil {
			return nil, err
		}

		users, err := users.GetListForApi()
		if err != nil {
			return nil, err
		}

		return util.Map(users, toUserRs), nil
	}
}

type UserRq struct {
	Name     string
	Password string
	Role     model.Role
}

func PostUsers(auth *authenticator, users *users.UserService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authorize(r, model.ROLE_ADMIN)
		if err != nil {
			return nil, err
		}

		var userRq UserRq
		if err := json.NewDecoder(r.Body).Decode(&userRq); err != nil {
			return nil, err
		}

		user, err := users.CreateUser(userRq.Name, userRq.Password, userRq.Role)
		if err != nil {
			return nil, err
		}

		return toUserRs(user), nil
	}
}

func GetUserMe(auth *authenticator, users *users.UserService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		userRs := toUserRs(user)
		userRs.ApiKey = user.ApiKey
		return userRs, nil
	}
}

func PutUserRoot(users *users.UserService) WebappHandler {
	return func(r *http.Request) (any, error) {
		var userRq UserRq
		if err := json.NewDecoder(r.Body).Decode(&userRq); err != nil {
			return nil, err
		}

		user, err := users.CreateFirstAdmin(userRq.Name, userRq.Password)
		if err != nil {
			return nil, err
		}

		return toUserRs(user), nil
	}
}

func PatchUser(auth *authenticator, users *users.UserService) WebappHandler {
	return func(r *http.Request) (any, error) {
		currentUser, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		targetUserId, err := uuid.Parse(mux.Vars(r)["userId"])
		if err != nil {
			if currentUser.Role == model.ROLE_ADMIN {
				return nil, model.ErrNotFound
			} else {
				return nil, model.ErrInsufficientPermissions
			}
		}

		if targetUserId != currentUser.Id && currentUser.Role != model.ROLE_ADMIN {
			// only admin can patch other users
			return nil, model.ErrInsufficientPermissions
		}

		userRq := UserRq{}
		if err := json.NewDecoder(r.Body).Decode(&userRq); err != nil {
			return nil, err
		}

		if targetUserId == currentUser.Id && userRq.Role != "" {
			// changing your own role is not allowed
			return nil, model.ErrInsufficientPermissions
		}

		targetUser, err := users.UpdateUser(targetUserId, userRq.Name, userRq.Password, userRq.Role)
		if err != nil {
			return nil, err
		}

		userRs := toUserRs(targetUser)
		if currentUser.Id == targetUser.Id {
			userRs.ApiKey = targetUser.ApiKey
		}
		return userRs, nil
	}
}

func PostUserApiKeys(auth *authenticator, users *users.UserService) WebappHandler {
	return func(r *http.Request) (any, error) {
		currentUser, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		targetUserId, err := uuid.Parse(mux.Vars(r)["userId"])
		if err != nil {
			return nil, model.ErrInsufficientPermissions
		}

		if currentUser.Id != targetUserId {
			return nil, model.ErrInsufficientPermissions
		}

		currentUser, err = users.UpdateApiKey(currentUser.Id)
		if err != nil {
			return nil, err
		}

		userRs := toUserRs(currentUser)
		userRs.ApiKey = currentUser.ApiKey
		return userRs, nil
	}
}

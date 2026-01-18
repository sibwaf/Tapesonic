package subsonic

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"tapesonic/util"
)

type AuthMethod interface {
	ApplyTo(r *http.Request)
}

type EmptyAuth struct {
}

func (auth *EmptyAuth) ApplyTo(r *http.Request) {
	// do nothing
}

type PlainAuth struct {
	username string
	password string
}

func (auth *PlainAuth) ApplyTo(r *http.Request) {
	query := r.URL.Query()
	query.Add("u", auth.username)
	query.Add("p", auth.password)

	r.URL.RawQuery = query.Encode()
}

func NewPlainAuth(username string, password string) PlainAuth {
	return PlainAuth{username: username, password: password}
}

type SaltedAuth struct {
	username string
	password string
}

func (auth *SaltedAuth) ApplyTo(r *http.Request) {
	salt := util.GenerateRandomString(8)
	token := GenerateAuthToken(auth.password, salt)

	query := r.URL.Query()
	query.Add("u", auth.username)
	query.Add("t", token)
	query.Add("s", salt)

	r.URL.RawQuery = query.Encode()
}

func NewSaltedAuth(username string, password string) SaltedAuth {
	return SaltedAuth{username: username, password: password}
}

func GenerateAuthToken(password string, salt string) string {
	hash := md5.Sum([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

func EncodePassword(password string) string {
	return "enc:" + hex.EncodeToString([]byte(password))
}

func DecodePassword(encoded string) (string, error) {
	encoded, ok := strings.CutPrefix(encoded, "enc:")
	if !ok {
		return "", fmt.Errorf("not an encoded password")
	}

	chars, err := hex.DecodeString(encoded)
	return string(chars), err
}

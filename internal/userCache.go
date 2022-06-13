package internal

import (
	"errors"
	"strings"
	"sync"
)

var ErrUserNotFound = errors.New("no user")

type UserCache struct {
	userByNick      map[string]*User
	userByNickMutex sync.RWMutex

	nickByEmail      map[string]string
	nickByEmailMutex sync.RWMutex
}

func NewUserCache() *UserCache {
	return &UserCache{
		userByNick:  make(map[string]*User),
		nickByEmail: make(map[string]string),
	}
}

func (uc *UserCache) Add(data *User) {
	uc.userByNickMutex.Lock()
	uc.userByNick[strings.ToLower(data.Nickname)] = data
	uc.userByNickMutex.Unlock()

	uc.nickByEmailMutex.Lock()
	uc.nickByEmail[data.Email] = data.Nickname
	uc.nickByEmailMutex.Unlock()
}

func (uc *UserCache) GetUserByNickname(nick string) (*User, error) {
	uc.userByNickMutex.RLock()
	user, ok := uc.userByNick[strings.ToLower(nick)]
	uc.userByNickMutex.RUnlock()
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (uc *UserCache) GetUserByEmail(email string) (*User, error) {
	uc.nickByEmailMutex.RLock()
	nick, ok := uc.nickByEmail[email]
	uc.nickByEmailMutex.RUnlock()
	if !ok {
		return nil, ErrUserNotFound
	}
	return uc.GetUserByNickname(nick)
}

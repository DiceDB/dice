package core

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dicedb/dice/config"
	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultUserName = "default"
	AuthCmd         = "AUTH"

	SessionStatusPending SessionStatusT = SessionStatusT(0)
	SessionStatusActive  SessionStatusT = SessionStatusT(1)
	SessionStatusExpired SessionStatusT = SessionStatusT(2)
)

var (
	UserStore *Users = NewUsersStore()
)

type (
	SessionStatusT uint8
	Session        struct {
		ID   uint64
		User *User

		CreatedAt      time.Time
		LastAccessedAt time.Time

		Status SessionStatusT
	}

	Users struct {
		store  map[string]*User
		stLock *sync.RWMutex
	}

	User struct {
		Username          string
		Passwords         []string
		IsPasswordEnabled bool
	}
)

func NewUsersStore() (users *Users) {
	users = &Users{
		store:  make(map[string]*User),
		stLock: &sync.RWMutex{},
	}
	return
}

func (users *Users) Get(username string) (user *User, err error) {
	users.stLock.RLock()
	defer users.stLock.RUnlock()
	isPresent := false
	if user, isPresent = users.store[username]; !isPresent {
		err = fmt.Errorf("user %s not found", username)
		return
	}
	return
}

func (users *Users) Add(username string) (user *User, err error) {
	user = &User{
		Username: username,
	}
	users.stLock.Lock()
	defer users.stLock.Unlock()
	users.store[username] = user
	return
}

func (user *User) SetPassword(password string) (err error) {
	var (
		hashedPassword []byte
	)
	if hashedPassword, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); err != nil {
		return
	}
	user.Passwords = append(user.Passwords, string(hashedPassword))
	return
}

func NewSession() (session *Session) {
	session = &Session{
		ID: uint64(time.Now().UTC().Unix()),

		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),

		Status: SessionStatusPending,
	}
	return
}

func (session *Session) IsActive() (isActive bool) {
	if config.RequirePass == "" && session.Status != SessionStatusActive {
		if err := session.Activate(session.User); err != nil {
			return
		}
	}
	isActive = session.Status == SessionStatusActive
	if isActive {
		session.LastAccessedAt = time.Now().UTC()
	}
	return
}

func (session *Session) Activate(user *User) (err error) {
	session.User = user
	session.Status = SessionStatusActive
	session.CreatedAt = time.Now().UTC()
	session.LastAccessedAt = time.Now().UTC()
	return
}

func (session *Session) Validate(username, password string) (err error) {
	var (
		user *User
	)
	if user, err = UserStore.Get(username); err != nil {
		return
	}
	if username == DefaultUserName && len(user.Passwords) == 0 {
		if err = session.Activate(user); err != nil {
			return
		}
		return
	}
	for _, userPassword := range user.Passwords {
		if err = bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(password)); err != nil {
			continue
		}
		if err = session.Activate(user); err != nil {
			log.Println("error activating session for user", username, err)
			continue
		}
		// User has been validated and the session has been activated
		return
	}
	err = fmt.Errorf("no password matched for user %s", username)
	return
}

func (session *Session) Expire() (err error) {
	session.Status = SessionStatusExpired
	return
}

package auth

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dicedb/dice/internal/server/utils"

	"github.com/dicedb/dice/config"
	"golang.org/x/crypto/bcrypt"
)

const (
	Cmd = "AUTH"

	SessionStatusPending = SessionStatusT(0)
	SessionStatusActive  = SessionStatusT(1)
	SessionStatusExpired = SessionStatusT(2)
)

var (
	UserStore = NewUsersStore()
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
		return nil, fmt.Errorf("ERR user not found")
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
	if password == utils.EmptyStr {
		slog.Warn("DiceDB is running without authentication. Consider setting a password.")
	}

	if hashedPassword, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); err != nil {
		return
	}
	user.Passwords = append(user.Passwords, string(hashedPassword))
	return
}

func NewSession() (session *Session) {
	session = &Session{
		ID: uint64(utils.GetCurrentTime().UTC().Unix()),

		CreatedAt:      utils.GetCurrentTime(),
		LastAccessedAt: utils.GetCurrentTime(),

		Status: SessionStatusPending,
	}
	return
}

func (session *Session) IsActive() (isActive bool) {
	if config.DiceConfig.Auth.Password == utils.EmptyStr && session.Status != SessionStatusActive {
		session.Activate(session.User)
	}
	isActive = session.Status == SessionStatusActive
	if isActive {
		session.LastAccessedAt = utils.GetCurrentTime().UTC()
	}
	return
}

func (session *Session) Activate(user *User) {
	session.User = user
	session.Status = SessionStatusActive
	session.CreatedAt = utils.GetCurrentTime().UTC()
	session.LastAccessedAt = utils.GetCurrentTime().UTC()
}

func (session *Session) Validate(username, password string) error {
	var (
		err  error
		user *User
	)
	if user, err = UserStore.Get(username); err != nil {
		return err
	}
	if username == config.DiceConfig.Auth.UserName && len(user.Passwords) == 0 {
		session.Activate(user)
		return nil
	}
	for _, userPassword := range user.Passwords {
		if err = bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(password)); err != nil {
			continue
		}
		session.Activate(user)
		return nil
	}
	return fmt.Errorf("WRONGPASS invalid username-password pair or user is disabled")
}

func (session *Session) Expire() {
	session.Status = SessionStatusExpired
}

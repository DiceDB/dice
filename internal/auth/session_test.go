// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package auth

import (
	"testing"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/server/utils"
)

func TestNewUsers(t *testing.T) {
	users := NewUsersStore()
	if users == nil {
		t.Error("NewUsers() returned nil")
	}
	if users != nil && users.store == nil {
		t.Error("NewUsers() created Users with nil store")
	}
	if users != nil && users.stLock == nil {
		t.Error("NewUsers() created Users with nil stLock")
	}
}

func TestUsersAddAndGet(t *testing.T) {
	users := NewUsersStore()
	if users == nil {
		t.Fatal("NewUsers() returned nil")
	}
	username := "testuser"

	// Test Add
	user, err := users.Add(username)
	if err != nil {
		t.Errorf("Users.Add() returned an error: %v", err)
	}
	if user == nil {
		t.Fatal("Users.Add() returned nil user")
	}
	if user.Username != username {
		t.Errorf("Users.Add() created user with incorrect username. Got %s, want %s", user.Username, username)
	}

	// Test Get
	if users == nil {
		t.Fatal("Users is nil")
	}
	retrievedUser, err := users.Get(username)
	if err != nil {
		t.Errorf("Users.Get() returned an error: %v", err)
	}
	if retrievedUser == nil {
		t.Fatal("Users.Get() returned nil user")
	}
	if retrievedUser != user {
		t.Error("Users.Get() returned a different user than the one added")
	}
}

func TestUserSetPassword(t *testing.T) {
	user := &User{Username: "testuser"}
	password := "testpassword"

	err := user.SetPassword(password)
	if err != nil {
		t.Errorf("User.SetPassword() returned an error: %v", err)
	}
	if len(user.Passwords) != 1 {
		t.Errorf("User.SetPassword() did not add password. Got %d passwords, want 1", len(user.Passwords))
	}
}

func TestNewSession(t *testing.T) {
	session := NewSession()
	if session == nil {
		t.Error("NewSession() returned nil")
	}
	if session != nil && session.Status != SessionStatusPending {
		t.Errorf("NewSession() created session with incorrect status. Got %v, want %v", session.Status, SessionStatusPending)
	}
}

func TestSessionIsActive(t *testing.T) {
	mockTime := &utils.MockClock{CurrTime: time.Now()}
	utils.CurrentTime = mockTime

	config.DiceConfig.Auth.Password = "testpassword"
	session := NewSession()
	if session.IsActive() {
		t.Error("New session should not be active")
	}

	session.Status = SessionStatusActive
	if !session.IsActive() {
		t.Error("Session with SessionStatusActive should be active")
	}

	oldLastAccessed := session.LastAccessedAt
	mockTime.SetTime(time.Now().Add(5 * time.Millisecond))

	session.IsActive()
	if !session.LastAccessedAt.After(oldLastAccessed) {
		t.Error("IsActive() should update LastAccessedAt")
	}
	config.DiceConfig.Auth.Password = utils.EmptyStr
}

func TestSessionActivate(t *testing.T) {
	session := NewSession()
	user := &User{Username: config.DiceConfig.Auth.UserName}

	session.Activate(user)

	if session.Status != SessionStatusActive {
		t.Errorf("Session.Activate() did not set status to Active. Got %v, want %v", session.Status, SessionStatusActive)
	}
	if session.User != user {
		t.Error("Session.Activate() did not set the User correctly")
	}
}

func TestSessionValidate(t *testing.T) {
	username := config.DiceConfig.Auth.UserName
	password := "testpassword"

	user, _ := UserStore.Add(username)
	if err := user.SetPassword(password); err != nil {
		t.Fatalf("User.SetPassword() returned an error: %v", err)
	}

	session := NewSession()
	err := session.Validate(username, password)
	if err != nil {
		t.Errorf("Session.Validate() returned an error: %v", err)
	}
	if session.Status != SessionStatusActive {
		t.Error("Session.Validate() did not activate the session")
	}

	err = session.Validate(username, "wrongpassword")
	if err == nil {
		t.Error("Session.Validate() did not return an error for wrong password")
	}
}

func TestSessionExpire(t *testing.T) {
	session := NewSession()
	session.Expire()
	if session.Status != SessionStatusExpired {
		t.Errorf("Session.Expire() did not set status to Expired. Got %v, want %v", session.Status, SessionStatusExpired)
	}
}

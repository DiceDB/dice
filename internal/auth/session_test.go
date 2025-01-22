// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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

	config.Config.Password = "testpassword"
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
	config.Config.Password = utils.EmptyStr
}

func TestSessionActivate(t *testing.T) {
	session := NewSession()
	user := &User{Username: config.Config.Username}

	session.Activate(user)

	if session.Status != SessionStatusActive {
		t.Errorf("Session.Activate() did not set status to Active. Got %v, want %v", session.Status, SessionStatusActive)
	}
	if session.User != user {
		t.Error("Session.Activate() did not set the User correctly")
	}
}

func TestSessionValidate(t *testing.T) {
	username := config.Config.Username
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

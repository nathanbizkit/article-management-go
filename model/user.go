package model

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User model
type User struct {
	ID               uint
	Username         string
	Email            string
	Password         string
	Name             string
	Bio              string
	Image            string
	CreatedAt        time.Time
	UpdatedAt        *time.Time
	Follows          []User
	FavoriteArticles []Article
}

// HashPassword makes password field crypted
func (u *User) HashPassword() error {
	if len(u.Password) == 0 {
		return errors.New("password should not be empty")
	}

	h, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(h)
	return nil
}

// CheckPassword checks if user password is matched
func (u *User) CheckPassword(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plain))
	return err == nil
}

package model

import (
	"errors"
	"regexp"
	"time"
	"unicode"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
)

const passwordMinLen = 7

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

// Validate validates fields of user model
func (u User) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(
			&u.Username,
			validation.Required,
			validation.Match(regexp.MustCompile("[a-zA-Z0-9]+")),
		),
		validation.Field(
			&u.Email,
			validation.Required,
			is.Email,
		),
		validation.Field(
			&u.Password,
			validation.Required,
			validation.Length(passwordMinLen, 0),
			validation.By(isStrongPassword),
		),
	)
}

func isStrongPassword(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.New("must be a string")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	for _, char := range s {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsSymbol(char) || unicode.IsPunct(char):
			hasSpecial = true
		}
	}

	if hasUpper && hasLower && hasNumber && hasSpecial {
		return nil
	}

	return errors.New("must have at least one uppercase, one lowercase, one number, and one symbol")
}

// Overwrite overwrites each field if it's not zero-value
func (u *User) Overwrite(username, email, password, name, bio, image string) {
	if username != "" {
		u.Username = username
	}

	if email != "" {
		u.Email = email
	}

	if password != "" {
		u.Password = password
	}

	if name != "" {
		u.Name = name
	}

	if bio != "" {
		u.Bio = bio
	}

	if image != "" {
		u.Image = image
	}
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

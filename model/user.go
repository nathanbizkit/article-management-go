package model

import (
	"errors"
	"regexp"
	"time"
	"unicode"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/nathanbizkit/article-management/message"
	"golang.org/x/crypto/bcrypt"
)

const (
	userShortMinLen = 5
	userShortMaxLen = 100
	userLongMaxLen  = 255
	passwordMinLen  = 7
	passwordMaxLen  = 50
)

// User model
type User struct {
	ID        uint
	Username  string
	Email     string
	Password  string
	Name      string
	Bio       string
	Image     string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// Validate validates fields of user model
func (u User) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(
			&u.Username,
			validation.Required,
			validation.Length(userShortMinLen, userShortMaxLen),
			validation.Match(
				regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9_]+[a-zA-Z0-9]$")),
		),
		validation.Field(
			&u.Email,
			validation.Required,
			validation.Length(userShortMinLen, userShortMaxLen),
			is.Email,
		),
		validation.Field(
			&u.Password,
			validation.Required,
			validation.Length(passwordMinLen, passwordMaxLen),
			validation.By(isStrongPassword),
		),
		validation.Field(
			&u.Name,
			validation.Required,
			validation.Length(userShortMinLen, userShortMaxLen),
		),
		validation.Field(
			&u.Bio,
			validation.Length(0, userLongMaxLen),
		),
		validation.Field(
			&u.Image,
			validation.Length(0, userLongMaxLen),
			is.URL,
		),
	)
}

func isStrongPassword(value interface{}) error {
	s, _ := value.(string)

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
func (u *User) Overwrite(username, email, password, name, bio, image string) (requirePasswordHashing bool) {
	if username != "" {
		u.Username = username
	}

	if email != "" {
		u.Email = email
	}

	if password != "" {
		u.Password = password
		requirePasswordHashing = true
	}

	if name != "" {
		u.Name = name
	}

	u.Bio = bio
	u.Image = image

	return
}

// HashPassword makes password field crypted
func (u *User) HashPassword() error {
	if u.Password == "" {
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

// ResponseProfile generates response message for user's profile
func (u *User) ResponseProfile(following bool) message.ProfileResponse {
	return message.ProfileResponse{
		Username:  u.Username,
		Name:      u.Name,
		Bio:       u.Bio,
		Image:     u.Image,
		Following: following,
	}
}

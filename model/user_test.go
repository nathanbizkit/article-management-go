package model

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nathanbizkit/article-management/message"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUserModel_Validate(t *testing.T) {
	longString := strings.Repeat("a", 101)

	tests := []struct {
		title    string
		in       *User
		hasError bool
	}{
		{
			"validate user: success",
			&User{
				Username: "foo_user",
				Email:    "foo@example.com",
				Name:     "FooUser",
				Password: "pA55w0Rd!",
			},
			false,
		},
		{
			"validate user: no username",
			&User{
				Username: "",
				Email:    "foo@example.com",
				Name:     "FooUser",
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: username is too short",
			&User{
				Username: "foo",
				Email:    "foo@example.com",
				Name:     "FooUser",
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: username is too long",
			&User{
				Username: longString,
				Email:    "foo@example.com",
				Name:     "FooUser",
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: no underscore at the beginning of username",
			&User{
				Username: "_fooname",
				Email:    "foo@example.com",
				Name:     "FooUser",
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: no underscore at the end of username",
			&User{
				Username: "fooname_",
				Email:    "foo@example.com",
				Name:     "FooUser",
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: username is other than a-z, A-Z, 0-9, _ in the middle",
			&User{
				Username: "foo_name-@-",
				Email:    "foo@example.com",
				Name:     "FooUser",
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: no email",
			&User{
				Username: "foo_user",
				Email:    "",
				Name:     "FooUser",
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: email is in invalid format",
			&User{
				Username: "foo_user",
				Email:    "email.com",
				Name:     "FooUser",
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: email is too short",
			&User{
				Username: "foo_user",
				Email:    "foo",
				Name:     "FooUser",
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: email is too long",
			&User{
				Username: "foo_user",
				Email:    longString,
				Name:     "FooUser",
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: no name",
			&User{
				Username: "foo_user",
				Email:    "foo@example.com",
				Name:     "",
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: name is too short",
			&User{
				Username: "foo_user",
				Email:    "foo@example.com",
				Name:     "Foo",
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: name is too long",
			&User{
				Username: "foo_user",
				Email:    "foo@example.com",
				Name:     longString,
				Password: "pA55w0Rd!",
			},
			true,
		},
		{
			"validate user: no password",
			&User{
				Username: "foo_user",
				Email:    "foo@example.com",
				Name:     "FooUser",
				Password: "",
			},
			true,
		},
		{
			"validate user: password is too short",
			&User{
				Username: "foo_user",
				Email:    "foo@example.com",
				Name:     "FooUser",
				Password: "pass",
			},
			true,
		},
		{
			"validate user: password is too long",
			&User{
				Username: "foo_user",
				Email:    "foo@example.com",
				Name:     "FooUser",
				Password: strings.Repeat("a", 51),
			},
			true,
		},
		{
			"validate user: weak password",
			&User{
				Username: "foo_user",
				Email:    "foo@example.com",
				Name:     "FooUser",
				Password: "password",
			},
			true,
		},
	}

	for _, tt := range tests {
		err := tt.in.Validate()
		if tt.hasError {
			assert.Error(t, err, fmt.Sprintf("%s: expect an error", tt.title))
		} else {
			assert.NoError(t, err, fmt.Sprintf("%s: expect no error", tt.title))
		}
	}
}

func TestUserModel_Overwrite(t *testing.T) {
	tests := []struct {
		title          string
		u              *User
		in             *User
		expectedUser   *User
		expectedReturn bool
	}{
		{
			"overwrite user: success",
			&User{
				ID:        1,
				Username:  "foo_user",
				Email:     "foo@example.com",
				Password:  "encrypted_password",
				Name:      "FooUser",
				Bio:       "This is my bio.",
				Image:     "https://imgur.com/image.jpeg",
				CreatedAt: time.Now(),
				UpdatedAt: nil,
			},
			&User{
				Username: "new_foo_user",
				Email:    "foo_bar@example.com",
				Password: "new_password",
				Name:     "NewFooUser",
				Bio:      "This is my new bio.",
				Image:    "https://imgur.com/new_image.jpeg",
			},
			&User{
				ID:        1,
				Username:  "new_foo_user",
				Email:     "foo_bar@example.com",
				Password:  "new_password",
				Name:      "NewFooUser",
				Bio:       "This is my new bio.",
				Image:     "https://imgur.com/new_image.jpeg",
				CreatedAt: time.Now(),
				UpdatedAt: nil,
			},
			true,
		},
		{
			"overwrite user: empty bio and image, no changes for other fields",
			&User{
				ID:        1,
				Username:  "foo_user",
				Email:     "foo@example.com",
				Password:  "encrypted_password",
				Name:      "FooUser",
				Bio:       "This is my bio.",
				Image:     "https://imgur.com/image.jpeg",
				CreatedAt: time.Now(),
				UpdatedAt: nil,
			},
			&User{
				Bio:   "",
				Image: "",
			},
			&User{
				ID:        1,
				Username:  "foo_user",
				Email:     "foo@example.com",
				Password:  "encrypted_password",
				Name:      "FooUser",
				Bio:       "",
				Image:     "",
				CreatedAt: time.Now(),
				UpdatedAt: nil,
			},
			false,
		},
	}

	for _, tt := range tests {
		isPlainPassword := tt.u.Overwrite(
			tt.in.Username, tt.in.Email, tt.in.Password,
			tt.in.Name, tt.in.Bio, tt.in.Image)

		assert.Equal(t, tt.expectedReturn, isPlainPassword, tt.title,
			fmt.Sprintf("%s: expect return isPlainPassword (%v)=%v",
				tt.title, tt.expectedReturn, isPlainPassword))

		assert.Equal(t, tt.u.Username, tt.expectedUser.Username,
			fmt.Sprintf("%s: expect username (%s)=%s",
				tt.title, tt.u.Username, tt.expectedUser.Username))
		assert.Equal(t, tt.u.Email, tt.expectedUser.Email,
			fmt.Sprintf("%s: expect email (%s)=%s",
				tt.title, tt.u.Email, tt.expectedUser.Email))
		assert.Equal(t, tt.u.Password, tt.expectedUser.Password,
			fmt.Sprintf("%s: expect password (%s)=%s",
				tt.title, tt.u.Password, tt.expectedUser.Password))
		assert.Equal(t, tt.u.Name, tt.expectedUser.Name,
			fmt.Sprintf("%s: expect name (%s)=%s",
				tt.title, tt.u.Name, tt.expectedUser.Name))
		assert.Equal(t, tt.u.Bio, tt.expectedUser.Bio,
			fmt.Sprintf("%s: expect bio (%s)=%s",
				tt.title, tt.u.Bio, tt.expectedUser.Bio))
		assert.Equal(t, tt.u.Image, tt.expectedUser.Image,
			fmt.Sprintf("%s: expect image (%s)=%s",
				tt.title, tt.u.Image, tt.expectedUser.Image))
	}
}

func TestUserModel_HashPassword(t *testing.T) {
	tests := []struct {
		title    string
		u        *User
		hasError bool
	}{
		{
			"hash password user: success",
			&User{Password: "pA55w0Rd!"},
			false,
		},
		{
			"hash password user: empty password",
			&User{Password: ""},
			true,
		},
		{
			"hash password user: password is too long",
			&User{Password: strings.Repeat("a", 73)},
			true,
		},
	}

	for _, tt := range tests {
		tempPassword := tt.u.Password
		err := tt.u.HashPassword()

		if tt.hasError {
			assert.Error(t, err, fmt.Sprintf("%s: expect an error", tt.title))
			assert.Equal(t, tempPassword, tt.u.Password,
				fmt.Sprintf("%s: expect no hashing", tt.title))
		} else {
			assert.NoError(t, err, fmt.Sprintf("%s: expect no error", tt.title))

			err = bcrypt.CompareHashAndPassword([]byte(tt.u.Password), []byte(tempPassword))
			assert.NoError(t, err, fmt.Sprintf("%s: expect hashing", tt.title))
		}
	}
}

func TestUserModel_CheckPassword(t *testing.T) {
	plainPassword := "pA55w0Rd!"
	h, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	tests := []struct {
		title          string
		u              *User
		in             string
		expectedReturn bool
	}{
		{
			"check password user: success",
			&User{Password: string(h)},
			plainPassword,
			true,
		},
		{
			"check password user: wrong password",
			&User{Password: string(h)},
			"password",
			false,
		},
	}

	for _, tt := range tests {
		pass := tt.u.CheckPassword(tt.in)
		assert.Equal(t, tt.expectedReturn, pass,
			fmt.Sprintf("%s: expect return (%v)=%v",
				tt.title, pass, tt.expectedReturn))
	}
}

func TestUserModel_ResponseProfile(t *testing.T) {
	u := User{
		ID:        1,
		Username:  "foo_user",
		Email:     "foo@example.com",
		Password:  "encrypted_password",
		Name:      "FooUser",
		Bio:       "This is my bio.",
		Image:     "https://imgur.com/image.jpeg",
		CreatedAt: time.Now(),
		UpdatedAt: nil,
	}
	expected := message.ProfileResponse{
		Username:  "foo_user",
		Name:      "FooUser",
		Bio:       "This is my bio.",
		Image:     "https://imgur.com/image.jpeg",
		Following: false,
	}

	profile := u.ResponseProfile(false)
	assert.Equal(t, expected, profile)
}

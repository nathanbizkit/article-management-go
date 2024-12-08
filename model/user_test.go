package model

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nathanbizkit/article-management-go/message"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUnit_UserModel(t *testing.T) {
	if !testing.Short() {
		t.Skip("skipping unit tests.")
	}

	t.Run("Validate", func(t *testing.T) {
		shortMaxLenString := strings.Repeat("a", 101)
		longMaxLenString := strings.Repeat("a", 256)
		passwordMaxLenString := strings.Repeat("a", 51)

		tests := []struct {
			title           string
			user            *User
			isPlainPassword bool
			hasError        bool
		}{
			{
				"validate user: success",
				&User{
					Username: "foo_user",
					Email:    "foo@example.com",
					Password: "pA55w0Rd!",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				false,
			},
			{
				"validate user: no username",
				&User{
					Username: "",
					Email:    "foo@example.com",
					Password: "pA55w0Rd!",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: username is too short",
				&User{
					Username: "foo",
					Email:    "foo@example.com",
					Password: "pA55w0Rd!",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: username is too long",
				&User{
					Username: shortMaxLenString,
					Email:    "foo@example.com",
					Password: "pA55w0Rd!",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: no underscore at the beginning of username",
				&User{
					Username: "_fooname",
					Email:    "foo@example.com",
					Password: "pA55w0Rd!",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: no underscore at the end of username",
				&User{
					Username: "fooname_",
					Email:    "foo@example.com",
					Password: "pA55w0Rd!",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: username is other than a-z, A-Z, 0-9, _, . in the middle",
				&User{
					Username: "foo_name.-@-",
					Email:    "foo@example.com",
					Password: "pA55w0Rd!",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: no email",
				&User{
					Username: "foo_user",
					Email:    "",
					Password: "pA55w0Rd!",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: email is in invalid format",
				&User{
					Username: "foo_user",
					Email:    "email.com",
					Password: "pA55w0Rd!",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: email is too short",
				&User{
					Username: "foo_user",
					Email:    "foo",
					Password: "pA55w0Rd!",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: email is too long",
				&User{
					Username: "foo_user",
					Email:    shortMaxLenString,
					Password: "pA55w0Rd!",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: no name",
				&User{
					Username: "foo_user",
					Email:    "foo@example.com",
					Password: "pA55w0Rd!",
					Name:     "",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: name is too short",
				&User{
					Username: "foo_user",
					Email:    "foo@example.com",
					Password: "pA55w0Rd!",
					Name:     "Foo",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: name is too long",
				&User{
					Username: "foo_user",
					Email:    "foo@example.com",
					Password: "pA55w0Rd!",
					Name:     shortMaxLenString,
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: no password",
				&User{
					Username: "foo_user",
					Email:    "foo@example.com",
					Password: "",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: skip password validation (already hashed)",
				&User{
					Username: "foo_user",
					Email:    "foo@example.com",
					Password: "already_hashed_password",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				false,
				false,
			},
			{
				"validate user: password is too short",
				&User{
					Username: "foo_user",
					Email:    "foo@example.com",
					Password: "pass",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: password is too long",
				&User{
					Username: "foo_user",
					Email:    "foo@example.com",
					Password: passwordMaxLenString,
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: weak password",
				&User{
					Username: "foo_user",
					Email:    "foo@example.com",
					Password: "password",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: bio is too long",
				&User{
					Username: "foo_user",
					Email:    "foo@example.com",
					Password: "password",
					Name:     "FooUser",
					Bio:      longMaxLenString,
					Image:    "https://imgur.com/image.jpeg",
				},
				true,
				true,
			},
			{
				"validate user: image is too long",
				&User{
					Username: "foo_user",
					Email:    "foo@example.com",
					Password: "password",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    longMaxLenString,
				},
				true,
				true,
			},
			{
				"validate user: invalid type image (url)",
				&User{
					Username: "foo_user",
					Email:    "foo@example.com",
					Password: "password",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "invalid_image_url",
				},
				true,
				true,
			},
		}

		for _, tt := range tests {
			err := tt.user.Validate(tt.isPlainPassword)

			if tt.hasError {
				assert.Error(t, err, tt.title)
			} else {
				assert.NoError(t, err, tt.title)
			}
		}
	})

	t.Run("Overwrite", func(t *testing.T) {
		now := time.Now()

		tests := []struct {
			title                   string
			user                    *User
			input                   *User
			expectedUser            *User
			expectedIsPlainPassword bool
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
					CreatedAt: now,
					UpdatedAt: now,
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
					CreatedAt: now,
					UpdatedAt: now,
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
					CreatedAt: now,
					UpdatedAt: now,
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
					CreatedAt: now,
					UpdatedAt: now,
				},
				false,
			},
		}

		for _, tt := range tests {
			actualIsPlainPassword := tt.user.Overwrite(
				tt.input.Username, tt.input.Email, tt.input.Password,
				tt.input.Name, tt.input.Bio, tt.input.Image)

			assert.Equal(t, tt.expectedIsPlainPassword, actualIsPlainPassword, tt.title)
			assert.Equal(t, tt.expectedUser, tt.user, tt.title)
		}
	})

	t.Run("HashPassword", func(t *testing.T) {
		tests := []struct {
			title    string
			user     *User
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
			tempPassword := tt.user.Password
			err := tt.user.HashPassword()

			if tt.hasError {
				assert.Error(t, err, tt.title)
				assert.Equal(t, tempPassword, tt.user.Password, tt.title)
			} else {
				assert.NoError(t, err, tt.title)
				assert.NotEqual(t, tempPassword, tt.user.Password, tt.title)

				err = bcrypt.CompareHashAndPassword([]byte(tt.user.Password), []byte(tempPassword))
				assert.NoError(t, err, fmt.Sprintf("%s: expect password to be hashed", tt.title))
			}
		}
	})

	t.Run("CheckPassword", func(t *testing.T) {
		plain := "pA55w0Rd!"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)

		tests := []struct {
			title           string
			user            *User
			input           string
			expectedMatched bool
		}{
			{
				"check password user: success",
				&User{Password: string(hashedPassword)},
				plain,
				true,
			},
			{
				"check password user: wrong password",
				&User{Password: string(hashedPassword)},
				"password",
				false,
			},
		}

		for _, tt := range tests {
			actualMatched := tt.user.CheckPassword(tt.input)
			assert.Equal(t, tt.expectedMatched, actualMatched, tt.title)
		}
	})

	t.Run("ResponseProfile", func(t *testing.T) {
		now := time.Now()
		user := User{
			ID:        1,
			Username:  "foo_user",
			Email:     "foo@example.com",
			Password:  "encrypted_password",
			Name:      "FooUser",
			Bio:       "This is my bio.",
			Image:     "https://imgur.com/image.jpeg",
			CreatedAt: now,
			UpdatedAt: now,
		}

		expected := message.ProfileResponse{
			Username:  "foo_user",
			Name:      "FooUser",
			Bio:       "This is my bio.",
			Image:     "https://imgur.com/image.jpeg",
			Following: false,
		}

		actual := user.ResponseProfile(false)
		assert.Equal(t, expected, actual)
	})
}

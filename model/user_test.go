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

func TestUnit_UserModel(t *testing.T) {
	if !testing.Short() {
		t.Skip("skipping unit tests.")
	}

	t.Run("Validate", func(t *testing.T) {
		shortMaxLenString := strings.Repeat("a", 101)
		longMaxLenString := strings.Repeat("a", 256)
		passwordMaxLenString := strings.Repeat("a", 51)

		tests := []struct {
			title    string
			u        *User
			hasError bool
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
			},
			{
				"validate user: username is other than a-z, A-Z, 0-9, _ in the middle",
				&User{
					Username: "foo_name-@-",
					Email:    "foo@example.com",
					Password: "pA55w0Rd!",
					Name:     "FooUser",
					Bio:      "This is my bio.",
					Image:    "https://imgur.com/image.jpeg",
				},
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
			},
		}

		for _, tt := range tests {
			err := tt.u.Validate()

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
			title        string
			u            *User
			in           *User
			expectedUser *User
			expected     bool
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
					CreatedAt: now,
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
					CreatedAt: now,
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
					CreatedAt: now,
					UpdatedAt: nil,
				},
				false,
			},
		}

		for _, tt := range tests {
			actual := tt.u.Overwrite(
				tt.in.Username, tt.in.Email, tt.in.Password,
				tt.in.Name, tt.in.Bio, tt.in.Image)

			assert.Equal(t, tt.expected, actual, tt.title)
			assert.Equal(t, tt.expectedUser, tt.u, tt.title)
		}
	})

	t.Run("HashPassword", func(t *testing.T) {
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
				assert.Error(t, err, tt.title)
				assert.Equal(t, tempPassword, tt.u.Password, tt.title)
			} else {
				assert.NoError(t, err, tt.title)
				assert.NotEqual(t, tempPassword, tt.u.Password, tt.title)

				err = bcrypt.CompareHashAndPassword([]byte(tt.u.Password), []byte(tempPassword))
				assert.NoError(t, err, fmt.Sprintf("%s: expect password to be hashed", tt.title))
			}
		}
	})

	t.Run("CheckPassword", func(t *testing.T) {
		plain := "pA55w0Rd!"
		h, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)

		tests := []struct {
			title    string
			u        *User
			in       string
			expected bool
		}{
			{
				"check password user: success",
				&User{Password: string(h)},
				plain,
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
			actual := tt.u.CheckPassword(tt.in)
			assert.Equal(t, tt.expected, actual, tt.title)
		}
	})

	t.Run("ResponseProfile", func(t *testing.T) {
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

		following := false
		expected := message.ProfileResponse{
			Username:  "foo_user",
			Name:      "FooUser",
			Bio:       "This is my bio.",
			Image:     "https://imgur.com/image.jpeg",
			Following: following,
		}

		actual := u.ResponseProfile(following)
		assert.Equal(t, expected, actual)
	})
}

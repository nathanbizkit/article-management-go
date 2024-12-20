package message

/* Request message */

// LoginUserRequest definition
type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateUserRequest definition
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// UpdateUserRequest definition
type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

/* Response message */

// ProfileResponse definition
type ProfileResponse struct {
	Username  string `json:"username"`
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

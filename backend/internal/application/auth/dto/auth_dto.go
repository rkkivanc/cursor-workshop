package dto

// RegisterRequest is the input for user registration.
type RegisterRequest struct {
	Email       string `validate:"required,email"`
	Password    string `validate:"required,min=8"`
	DisplayName string `validate:"required,min=2,max=100"`
}

// LoginRequest is the input for user login.
type LoginRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

// RefreshRequest contains the refresh token to rotate.
type RefreshRequest struct {
	UserID       string `validate:"required,uuid4"`
	RefreshToken string `validate:"required"`
}

// LogoutRequest contains tokens to revoke.
type LogoutRequest struct {
	UserID       string `validate:"required,uuid4"`
	AccessToken  string `validate:"required"`
	RefreshToken string `validate:"required"`
}

// AuthResponse is returned after a successful login or register.
type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in"`
	User         UserData `json:"user"`
}

// UserData is a safe subset of user fields for client responses.
type UserData struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Role        string `json:"role"`
}

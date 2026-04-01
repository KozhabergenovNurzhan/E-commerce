// Package pb contains request/response types that mirror the proto definitions
// in proto/*.proto. These are used by the gRPC handlers and are encoded/decoded
// via the JSON codec registered in internal/grpc/codec.
package pb

// ── Auth ────────────────────────────────────────────────────────────────────

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	User   *UserMessage   `json:"user"`
	Tokens *TokensResponse `json:"tokens"`
}

type TokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserMessage struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	CreatedAt int64  `json:"created_at"`
}

type LogoutResponse struct{}

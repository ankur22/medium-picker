package rest

type SignupRequest struct {
	Email string `json:"email"`
}

type SignupResponse struct {
	UserID string `json:"userId"`
}

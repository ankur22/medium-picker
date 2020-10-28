package rest

type SignupRequest struct {
	Email string `json:"email"`
}

type SignupResponse struct {
	UserID string `json:"userId"`
}

type SignInRequest struct {
	Email string `json:"email"`
}

type SignInResponse struct {
	UserID string `json:"userId"`
}

type NewMediumSourceRequest struct {
	Source string `json:"source"`
}

type Source struct {
	URL string `json:"url"`
	ID  string `json:"id"`
}

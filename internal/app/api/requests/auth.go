package requests

type BasicAuth struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type Login struct {
	BasicAuth
}

type Signup struct {
	BasicAuth
	Name string `json:"name" validate:"required,nameWithSpace"`
}

type Refresh struct {
	Token string `json:"token" validate:"required"`
}

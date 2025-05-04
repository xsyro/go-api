package requests

type PostTodo struct {
	UserID string `json:"user_id" validate:"required,uuid" example:"uuid"`
	Task   string `json:"task" validate:"required" example:"Todo this"`
}

type PatchTodo struct {
	Task string `json:"task" validate:"required" example:"Todo that"`
	Done bool   `json:"done" validate:"required" example:"false"`
}
